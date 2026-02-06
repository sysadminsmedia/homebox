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

// HandleItemsGetAll godoc
//
//	@Summary	Query All Items
//	@Tags		Items
//	@Produce	json
//	@Param		q			query		string		false	"search string"
//	@Param		page		query		int			false	"page number"
//	@Param		pageSize	query		int			false	"items per page"
//	@Param		tags		query		[]string	false	"tags Ids"		collectionFormat(multi)
//	@Param		locations	query		[]string	false	"location Ids"	collectionFormat(multi)
//	@Param		parentIds	query		[]string	false	"parent Ids"	collectionFormat(multi)
//	@Success	200			{object}	repo.PaginationResult[repo.ItemSummary]{}
//	@Router		/v1/items [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsGetAll() errchain.HandlerFunc {
	extractQuery := func(r *http.Request) repo.ItemQuery {
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

		v := repo.ItemQuery{
			Page:             queryIntOrNegativeOne(params.Get("page")),
			PageSize:         queryIntOrNegativeOne(params.Get("pageSize")),
			Search:           params.Get("q"),
			LocationIDs:      queryUUIDList(params, "locations"),
			TagIDs:           queryUUIDList(params, "tags"),
			NegateTags:       queryBool(params.Get("negateTags")),
			OnlyWithoutPhoto: queryBool(params.Get("onlyWithoutPhoto")),
			OnlyWithPhoto:    queryBool(params.Get("onlyWithPhoto")),
			ParentItemIDs:    queryUUIDList(params, "parentIds"),
			IncludeArchived:  queryBool(params.Get("includeArchived")),
			Fields:           filterFieldItems(params["fields"]),
			OrderBy:          params.Get("orderBy"),
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

		items, err := ctrl.repo.Items.QueryByGroup(ctx, ctx.GID, extractQuery(r))
		totalPrice := new(big.Int)
		for _, item := range items.Items {
			if !item.SoldTime.IsZero() { // Skip items with a non-null SoldDate
				continue
			}
			totalPrice.Add(totalPrice, big.NewInt(int64(item.PurchasePrice*100)))
		}

		totalPriceFloat := new(big.Float).SetInt(totalPrice)
		totalPriceFloat.Quo(totalPriceFloat, big.NewFloat(100))

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return server.JSON(w, http.StatusOK, repo.PaginationResult[repo.ItemSummary]{
					Items: []repo.ItemSummary{},
				})
			}
			log.Err(err).Msg("failed to get items")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusOK, items)
	}
}

// HandleItemFullPath godoc
//
//	@Summary	Get the full path of an item
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	[]repo.ItemPath
//	@Router		/v1/items/{id}/path [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemFullPath() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) ([]repo.ItemPath, error) {
		auth := services.NewContext(r.Context())
		item, err := ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
		if err != nil {
			return nil, err
		}

		paths, err := ctrl.repo.Locations.PathForLoc(auth, auth.GID, item.Location.ID)
		if err != nil {
			return nil, err
		}

		if item.Parent != nil {
			paths = append(paths, repo.ItemPath{
				Type: repo.ItemTypeItem,
				ID:   item.Parent.ID,
				Name: item.Parent.Name,
			})
		}

		paths = append(paths, repo.ItemPath{
			Type: repo.ItemTypeItem,
			ID:   item.ID,
			Name: item.Name,
		})

		return paths, nil
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleItemsCreate godoc
//
//	@Summary	Create Item
//	@Tags		Items
//	@Produce	json
//	@Param		payload	body		repo.ItemCreate	true	"Item Data"
//	@Success	201		{object}	repo.ItemSummary
//	@Router		/v1/items [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.ItemCreate) (repo.ItemOut, error) {
		return ctrl.svc.Items.Create(services.NewContext(r.Context()), body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleItemGet godocs
//
//	@Summary	Get Item
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	repo.ItemOut
//	@Router		/v1/items/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.ItemOut, error) {
		auth := services.NewContext(r.Context())

		return ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleItemDelete godocs
//
//	@Summary	Delete Item
//	@Tags		Items
//	@Produce	json
//	@Param		id	path	string	true	"Item ID"
//	@Success	204
//	@Router		/v1/items/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Items.DeleteByGroup(auth, auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleItemUpdate godocs
//
//	@Summary	Update Item
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string			true	"Item ID"
//	@Param		payload	body		repo.ItemUpdate	true	"Item Data"
//	@Success	200		{object}	repo.ItemOut
//	@Router		/v1/items/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.ItemUpdate) (repo.ItemOut, error) {
		auth := services.NewContext(r.Context())

		body.ID = ID
		return ctrl.repo.Items.UpdateByGroup(auth, auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleItemPatch godocs
//
//	@Summary	Update Item
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string			true	"Item ID"
//	@Param		payload	body		repo.ItemPatch	true	"Item Data"
//	@Success	200		{object}	repo.ItemOut
//	@Router		/v1/items/{id} [Patch]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemPatch() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.ItemPatch) (repo.ItemOut, error) {
		auth := services.NewContext(r.Context())

		body.ID = ID
		err := ctrl.repo.Items.Patch(auth, auth.GID, ID, body)
		if err != nil {
			return repo.ItemOut{}, err
		}

		return ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleItemDuplicate godocs
//
//	@Summary	Duplicate Item
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string					true	"Item ID"
//	@Param		payload	body		repo.DuplicateOptions	true	"Duplicate Options"
//	@Success	201		{object}	repo.ItemOut
//	@Router		/v1/items/{id}/duplicate [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemDuplicate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, options repo.DuplicateOptions) (repo.ItemOut, error) {
		ctx := services.NewContext(r.Context())
		return ctrl.svc.Items.Duplicate(ctx, ctx.GID, ID, options)
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}

// HandleGetAllCustomFieldNames godocs
//
//	@Summary	Get All Custom Field Names
//	@Tags		Items
//	@Produce	json
//	@Success	200
//	@Router		/v1/items/fields [GET]
//	@Success	200	{object}	[]string
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetAllCustomFieldNames() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]string, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Items.GetAllCustomFieldNames(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGetAllCustomFieldValues godocs
//
//	@Summary	Get All Custom Field Values
//	@Tags		Items
//	@Produce	json
//	@Success	200
//	@Router		/v1/items/fields/values [GET]
//	@Success	200	{object}	[]string
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetAllCustomFieldValues() errchain.HandlerFunc {
	type query struct {
		Field string `schema:"field" validate:"required"`
	}

	fn := func(r *http.Request, q query) ([]string, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Items.GetAllCustomFieldValues(auth, auth.GID, q.Field)
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleItemsImport godocs
//
//	@Summary	Import Items
//	@Tags		Items
//	@Accept		multipart/form-data
//	@Produce	json
//	@Success	204
//	@Param		csv	formData	file	true	"Image to upload"
//	@Router		/v1/items/import [Post]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsImport() errchain.HandlerFunc {
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

		tenant := services.UseTenantCtx(r.Context())

		_, err = ctrl.svc.Items.CsvImport(r.Context(), tenant, file)
		if err != nil {
			log.Err(err).Msg("failed to import items")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusNoContent, nil)
	}
}

// HandleItemsExport godocs
//
//	@Summary	Export Items
//	@Tags		Items
//	@Success	200	{string}	string	"text/csv"
//	@Router		/v1/items/export [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsExport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		csvData, err := ctrl.svc.Items.ExportCSV(r.Context(), ctx.GID, GetHBURL(r, &ctrl.config.Options, ctrl.url))
		if err != nil {
			log.Err(err).Msg("failed to export items")
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
