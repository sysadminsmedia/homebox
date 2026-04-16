package v1

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
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
)

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
//	@Success	200			{object}	repo.PaginationResult[repo.EntitySummary]{}
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
		ctx := services.NewContext(r.Context())
		query := extractQuery(r)

		// When querying for locations specifically, use the container query
		// which includes item count aggregation, then normalize to the same
		// PaginationResult shape so the endpoint always returns a consistent type.
		if query.IsLocation != nil && *query.IsLocation {
			filterChildren := queryBool(r.URL.Query().Get("filterChildren"))
			containers, err := ctrl.repo.Entities.GetAllContainers(ctx, ctx.GID, repo.ContainerQuery{
				FilterChildren: filterChildren,
			})
			if err != nil {
				log.Err(err).Msg("failed to get containers")
				return validate.NewRequestError(err, http.StatusInternalServerError)
			}

			summaries := make([]repo.EntitySummary, len(containers))
			for i, c := range containers {
				s := c.EntitySummary
				s.ItemCount = c.ItemCount
				summaries[i] = s
			}

			return server.JSON(w, http.StatusOK, struct {
				repo.PaginationResult[repo.EntitySummary]
				TotalPrice float64 `json:"totalPrice"`
			}{
				PaginationResult: repo.PaginationResult[repo.EntitySummary]{
					Page:     1,
					PageSize: len(summaries),
					Total:    len(summaries),
					Items:    summaries,
				},
				TotalPrice: 0,
			})
		}

		items, err := ctrl.repo.Entities.QueryByGroup(ctx, ctx.GID, query)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return server.JSON(w, http.StatusOK, repo.PaginationResult[repo.EntitySummary]{
					Items: []repo.EntitySummary{},
				})
			}
			log.Err(err).Msg("failed to get entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		totalPrice := new(big.Int)
		for _, item := range items.Items {
			if !item.SoldTime.IsZero() {
				continue
			}
			totalPrice.Add(totalPrice, big.NewInt(int64(item.PurchasePrice*100)))
		}

		totalPriceFloat, _ := new(big.Float).SetInt(totalPrice).Quo(new(big.Float).SetInt(totalPrice), big.NewFloat(100)).Float64()

		return server.JSON(w, http.StatusOK, struct {
			repo.PaginationResult[repo.EntitySummary]
			TotalPrice float64 `json:"totalPrice"`
		}{
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
		auth := services.NewContext(r.Context())
		return ctrl.repo.Entities.PathForEntity(auth, auth.GID, ID)
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
		return ctrl.svc.Entities.Create(services.NewContext(r.Context()), body)
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
		auth := services.NewContext(r.Context())

		return ctrl.repo.Entities.GetOneByGroup(auth, auth.GID, ID)
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
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Entities.DeleteByGroup(auth, auth.GID, ID)
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
		auth := services.NewContext(r.Context())

		body.ID = ID
		return ctrl.repo.Entities.UpdateByGroup(auth, auth.GID, body)
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
		auth := services.NewContext(r.Context())

		body.ID = ID
		err := ctrl.repo.Entities.Patch(auth, auth.GID, ID, body)
		if err != nil {
			return repo.EntityOut{}, err
		}

		return ctrl.repo.Entities.GetOneByGroup(auth, auth.GID, ID)
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
		ctx := services.NewContext(r.Context())
		return ctrl.svc.Entities.Duplicate(ctx, ctx.GID, ID, options)
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
		auth := services.NewContext(r.Context())
		return ctrl.repo.Entities.GetAllCustomFieldNames(auth, auth.GID)
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
		auth := services.NewContext(r.Context())
		return ctrl.repo.Entities.GetAllCustomFieldValues(auth, auth.GID, q.Field)
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
		err := r.ParseMultipartForm(ctrl.maxUploadSize << 20)
		if err != nil {
			log.Err(err).Msg("failed to parse multipart form")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		file, _, err := r.FormFile("csv")
		if err != nil {
			log.Err(err).Msg("failed to get file from form")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		defer func() { _ = file.Close() }()

		tenant := services.UseTenantCtx(r.Context())

		_, err = ctrl.svc.Entities.CsvImport(r.Context(), tenant, file)
		if err != nil {
			log.Err(err).Msg("failed to import entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusNoContent, nil)
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
		auth := services.NewContext(r.Context())
		return ctrl.repo.Entities.Tree(auth, auth.GID, query)
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
		ctx := services.NewContext(r.Context())

		csvData, err := ctrl.svc.Entities.ExportCSV(r.Context(), ctx.GID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
		if err != nil {
			log.Err(err).Msg("failed to export entities")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		timestamp := time.Now().Format("2006-01-02_15-04-05")      // YYYY-MM-DD_HH-MM-SS format
		filename := fmt.Sprintf("homebox-items_%s.csv", timestamp) // add timestamp to filename

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filename))

		writer := csv.NewWriter(w)
		writer.Comma = ','
		return writer.WriteAll(csvData)
	}
}
