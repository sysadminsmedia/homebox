package v1

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func entityCtrlTracer() trace.Tracer {
	return otel.Tracer("controller")
}

func recordCtrlSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func startEntityCtrlSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return entityCtrlTracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

// HandleEntitiesGetAll godoc
//
//	@Summary	Query All Entities
//	@Tags		Entities
//	@Produce	json
//	@Param		q			query		string		false	"search string"
//	@Param		page		query		int			false	"page number"
//	@Param		pageSize	query		int			false	"items per page"
//	@Param		tags		query		[]string	false	"tags Ids"		collectionFormat(multi)
//	@Param		parentIds	query		[]string	false	"parent Ids"	collectionFormat(multi)
//	@Success	200			{object}	repo.EntityListResult
//	@Router		/v1/entities [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesGetAll() errchain.HandlerFunc {
	extractQuery := func(r *http.Request) repo.EntityQuery {
		params := r.URL.Query()

		filterFieldItems := func(raw []string) []repo.FieldQuery {
			return lo.FilterMap(raw, func(v string, _ int) (repo.FieldQuery, bool) {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) != 2 {
					return repo.FieldQuery{}, false
				}
				return repo.FieldQuery{
					Name:  parts[0],
					Value: parts[1],
				}, true
			})
		}

		v := repo.EntityQuery{
			Page:             queryIntOrNegativeOne(params.Get("page")),
			PageSize:         queryIntOrNegativeOne(params.Get("pageSize")),
			Search:           params.Get("q"),
			ParentIDs:        queryUUIDList(params, "parentIds"),
			TagIDs:           queryUUIDList(params, "tags"),
			NegateTags:       queryBool(params.Get("negateTags")),
			OnlyWithoutPhoto: queryBool(params.Get("onlyWithoutPhoto")),
			OnlyWithPhoto:    queryBool(params.Get("onlyWithPhoto")),
			IncludeArchived:  queryBool(params.Get("includeArchived")),
			Fields:           filterFieldItems(params["fields"]),
			OrderBy:          params.Get("orderBy"),
		}

		// Parse isLocation filter: "true" = locations only, "false" = items only, absent = default (items only)
		if isLocStr := params.Get("isLocation"); isLocStr != "" {
			isLoc := queryBool(isLocStr)
			v.IsLocation = &isLoc
		}

		v.FilterChildren = queryBool(params.Get("filterChildren"))

		if strings.HasPrefix(v.Search, "#") {
			aidStr := strings.TrimPrefix(v.Search, "#")

			aid, ok := repo.ParseAssetID(aidStr)
			if ok {
				v.Search = ""
				v.AssetID = aid
			}
		}

		return v
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		query := extractQuery(r)
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntitiesGetAll",
			attribute.String("query.search", query.Search),
			attribute.Int("query.page", query.Page),
			attribute.Int("query.page_size", query.PageSize),
			attribute.Int("query.tag_ids.count", len(query.TagIDs)),
			attribute.Int("query.parent_ids.count", len(query.ParentIDs)),
			attribute.Int("query.fields.count", len(query.Fields)),
			attribute.Bool("query.include_archived", query.IncludeArchived),
			attribute.Bool("query.filter_children", query.FilterChildren),
			attribute.Bool("query.only_with_photo", query.OnlyWithPhoto),
			attribute.Bool("query.only_without_photo", query.OnlyWithoutPhoto),
			attribute.String("query.order_by", query.OrderBy),
			attribute.Bool("query.is_location.set", query.IsLocation != nil),
			attribute.Bool("query.is_location.value", query.IsLocation != nil && *query.IsLocation),
			attribute.Bool("query.asset_id.set", !query.AssetID.Nil()),
		)
		defer span.End()

		ctx := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", ctx.GID.String()))

		items, err := ctrl.repo.Entities.QueryByGroup(ctx, ctx.GID, query)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				span.SetAttributes(attribute.Int("response.items.count", 0))
				return server.JSON(w, http.StatusOK, repo.PaginationResult[repo.EntitySummary]{
					Items: []repo.EntitySummary{},
				})
			}
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to get entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		_, totalSpan := startEntityCtrlSpan(spanCtx, "controller.V1.HandleEntitiesGetAll.totalPrice",
			attribute.Int("items.count", len(items.Items)))
		totalPrice := new(big.Int)
		for _, item := range items.Items {
			if !item.SoldTime.IsZero() {
				continue
			}
			totalPrice.Add(totalPrice, big.NewInt(int64(math.Round(item.PurchasePrice*100))))
		}

		totalPriceFloat, _ := new(big.Float).Quo(new(big.Float).SetInt(totalPrice), big.NewFloat(100)).Float64()
		totalSpan.SetAttributes(attribute.Float64("total_price", totalPriceFloat))
		totalSpan.End()

		span.SetAttributes(
			attribute.Int("response.items.count", len(items.Items)),
			attribute.Int("response.total", items.Total),
			attribute.Float64("response.total_price", totalPriceFloat),
		)

		return server.JSON(w, http.StatusOK, repo.EntityListResult{
			PaginationResult: items,
			TotalPrice:       totalPriceFloat,
		})
	}
}

// HandleEntityFullPath godoc
//
//	@Summary	Get the full path of an entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path		string	true	"Entity ID"
//	@Success	200	{object}	[]repo.EntityPath
//	@Router		/v1/entities/{id}/path [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityFullPath() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) ([]repo.EntityPath, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityFullPath",
			attribute.String("entity.id", ID.String()))
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		out, err := ctrl.repo.Entities.PathForEntity(auth, auth.GID, ID)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.Int("path.depth", len(out)))
		return out, nil
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntitiesCreate godoc
//
//	@Summary	Create Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		payload	body		repo.EntityCreate	true	"Entity Data"
//	@Success	201		{object}	repo.EntityOut
//	@Router		/v1/entities [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.EntityCreate) (repo.EntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntitiesCreate",
			attribute.String("entity.name", body.Name),
			attribute.Float64("entity.quantity", body.Quantity),
			attribute.Bool("entity.parent_id.set", body.ParentID != uuid.Nil),
			attribute.Bool("entity.entity_type_id.set", body.EntityTypeID != uuid.Nil),
			attribute.Int("entity.tags.count", len(body.TagIDs)),
		)
		defer span.End()

		ctx := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", ctx.GID.String()))
		out, err := ctrl.svc.Entities.Create(ctx, body)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.String("entity.id", out.ID.String()))
		return out, nil
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleEntityGet godoc
//
//	@Summary	Get Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path		string	true	"Entity ID"
//	@Success	200	{object}	repo.EntityOut
//	@Router		/v1/entities/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.EntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityGet",
			attribute.String("entity.id", ID.String()))
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		out, err := ctrl.repo.Entities.GetOneByGroup(auth, auth.GID, ID)
		if err != nil {
			recordCtrlSpanError(span, err)
		}
		return out, err
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityDelete godoc
//
//	@Summary	Delete Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id	path	string	true	"Entity ID"
//	@Success	204
//	@Router		/v1/entities/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityDelete",
			attribute.String("entity.id", ID.String()))
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		err := ctrl.repo.Entities.DeleteByGroup(auth, auth.GID, ID)
		if err != nil {
			recordCtrlSpanError(span, err)
		}
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleEntityUpdate godoc
//
//	@Summary	Update Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id		path		string				true	"Entity ID"
//	@Param		payload	body		repo.EntityUpdate	true	"Entity Data"
//	@Success	200		{object}	repo.EntityOut
//	@Router		/v1/entities/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.EntityUpdate) (repo.EntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityUpdate",
			attribute.String("entity.id", ID.String()),
			attribute.String("entity.name", body.Name),
			attribute.Float64("entity.quantity", body.Quantity),
			attribute.Bool("entity.archived", body.Archived),
			attribute.Bool("entity.parent_id.set", body.ParentID != uuid.Nil),
			attribute.Int("entity.tags.count", len(body.TagIDs)),
			attribute.Int("entity.fields.count", len(body.Fields)),
			attribute.Bool("entity.sync_child_locations", body.SyncChildEntityLocations),
		)
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))

		body.ID = ID
		out, err := ctrl.repo.Entities.UpdateByGroup(auth, auth.GID, body)
		if err != nil {
			recordCtrlSpanError(span, err)
		}
		return out, err
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityPatch godoc
//
//	@Summary	Patch Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id		path		string				true	"Entity ID"
//	@Param		payload	body		repo.EntityPatch	true	"Entity Data"
//	@Success	200		{object}	repo.EntityOut
//	@Router		/v1/entities/{id} [Patch]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityPatch() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.EntityPatch) (repo.EntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityPatch",
			attribute.String("entity.id", ID.String()),
			attribute.Bool("patch.import_ref.set", body.ImportRef != nil),
			attribute.Bool("patch.quantity.set", body.Quantity != nil),
			attribute.Bool("patch.parent_id.set", body.ParentID != uuid.Nil),
			attribute.Bool("patch.entity_type_id.set", body.EntityTypeID != uuid.Nil),
			attribute.Bool("patch.tag_ids.set", body.TagIDs != nil),
		)
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))

		body.ID = ID
		err := ctrl.repo.Entities.Patch(auth, auth.GID, ID, body)
		if err != nil {
			recordCtrlSpanError(span, err)
			return repo.EntityOut{}, err
		}

		out, err := ctrl.repo.Entities.GetOneByGroup(auth, auth.GID, ID)
		if err != nil {
			recordCtrlSpanError(span, err)
		}
		return out, err
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityDuplicate godoc
//
//	@Summary	Duplicate Entity
//	@Tags		Entities
//	@Produce	json
//	@Param		id		path		string					true	"Entity ID"
//	@Param		payload	body		repo.DuplicateOptions	true	"Duplicate Options"
//	@Success	201		{object}	repo.EntityOut
//	@Router		/v1/entities/{id}/duplicate [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityDuplicate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, options repo.DuplicateOptions) (repo.EntityOut, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityDuplicate",
			attribute.String("entity.source_id", ID.String()),
			attribute.Bool("options.copy_maintenance", options.CopyMaintenance),
			attribute.Bool("options.copy_attachments", options.CopyAttachments),
			attribute.Bool("options.copy_custom_fields", options.CopyCustomFields),
		)
		defer span.End()

		ctx := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", ctx.GID.String()))
		out, err := ctrl.svc.Entities.Duplicate(ctx, ctx.GID, ID, options)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.String("entity.id", out.ID.String()))
		return out, nil
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}

// HandleGetAllCustomFieldNames godoc
//
//	@Summary	Get All Custom Field Names
//	@Tags		Entities
//	@Produce	json
//	@Success	200	{array}		string
//	@Router		/v1/entities/fields [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetAllCustomFieldNames() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]string, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleGetAllCustomFieldNames")
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		out, err := ctrl.repo.Entities.GetAllCustomFieldNames(auth, auth.GID)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.Int("names.count", len(out)))
		return out, nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGetAllCustomFieldValues godoc
//
//	@Summary	Get All Custom Field Values
//	@Tags		Entities
//	@Produce	json
//	@Param		field	query		string	true	"Field name"
//	@Success	200		{array}		string
//	@Router		/v1/entities/fields/values [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetAllCustomFieldValues() errchain.HandlerFunc {
	type query struct {
		Field string `schema:"field" validate:"required"`
	}

	fn := func(r *http.Request, q query) ([]string, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleGetAllCustomFieldValues",
			attribute.String("field.name", q.Field))
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		out, err := ctrl.repo.Entities.GetAllCustomFieldValues(auth, auth.GID, q.Field)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.Int("values.count", len(out)))
		return out, nil
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleEntitiesImport godoc
//
//	@Summary	Import Entities
//	@Tags		Entities
//	@Accept		multipart/form-data
//	@Produce	json
//	@Success	204
//	@Param		csv	formData	file	true	"CSV file to upload"
//	@Router		/v1/entities/import [Post]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesImport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntitiesImport")
		defer span.End()

		_, parseSpan := startEntityCtrlSpan(spanCtx, "controller.V1.HandleEntitiesImport.parseForm")
		err := r.ParseMultipartForm(ctrl.maxUploadSize << 20)
		if err != nil {
			recordCtrlSpanError(parseSpan, err)
			parseSpan.End()
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to parse multipart form")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		file, _, err := r.FormFile("csv")
		if err != nil {
			recordCtrlSpanError(parseSpan, err)
			parseSpan.End()
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to get file from form")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		parseSpan.End()
		defer func() { _ = file.Close() }()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))

		count, err := ctrl.svc.Entities.CsvImport(spanCtx, auth.GID, file)
		if err != nil {
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to import entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Int("rows.imported.count", count))

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// HandleLocationTreeQuery godoc
//
//	@Summary	Get Locations Tree
//	@Tags		Entities
//	@Produce	json
//	@Param		withItems	query		bool	false	"include items in response tree"
//	@Success	200			{object}	[]repo.TreeItem
//	@Router		/v1/entities/tree [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLocationTreeQuery() errchain.HandlerFunc {
	fn := func(r *http.Request, query repo.TreeQuery) ([]repo.TreeItem, error) {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleLocationTreeQuery",
			attribute.Bool("query.with_items", query.WithItems))
		defer span.End()

		auth := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", auth.GID.String()))
		out, err := ctrl.repo.Entities.Tree(auth, auth.GID, query)
		if err != nil {
			recordCtrlSpanError(span, err)
			return out, err
		}
		span.SetAttributes(attribute.Int("tree.roots.count", len(out)))
		return out, nil
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleEntitiesExport godoc
//
//	@Summary	Export Entities
//	@Tags		Entities
//	@Success	200	{string}	string	"text/csv"
//	@Router		/v1/entities/export [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntitiesExport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntitiesExport")
		defer span.End()

		ctx := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("group.id", ctx.GID.String()))

		csvData, err := ctrl.svc.Entities.ExportCSV(spanCtx, ctx.GID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
		if err != nil {
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to export entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Int("csv.rows.count", len(csvData)))

		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("homebox-entities_%s.csv", timestamp)

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filename))

		_, writeSpan := startEntityCtrlSpan(spanCtx, "controller.V1.HandleEntitiesExport.write",
			attribute.Int("csv.rows.count", len(csvData)))
		defer writeSpan.End()
		writer := csv.NewWriter(w)
		if err := writer.WriteAll(csvData); err != nil {
			recordCtrlSpanError(writeSpan, err)
			log.Err(err).Msg("failed to write CSV export response")
		}
		return nil
	}
}
