package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"github.com/sysadminsmedia/homebox/backend/pkgs/labelmaker"
)

func generateOrPrint(ctrl *V1Controller, w http.ResponseWriter, r *http.Request, title string, description string, url string) error {
	params := labelmaker.NewGenerateParams(int(ctrl.config.LabelMaker.Width), int(ctrl.config.LabelMaker.Height), int(ctrl.config.LabelMaker.Margin), int(ctrl.config.LabelMaker.Padding), ctrl.config.LabelMaker.FontSize, title, description, url, ctrl.config.LabelMaker.DynamicLength, ctrl.config.LabelMaker.AdditionalInformation)

	print := queryBool(r.URL.Query().Get("print"))

	if print {
		err := labelmaker.PrintLabel(ctrl.config, &params)
		if err != nil {
			return err
		}

		_, err = w.Write([]byte("Printed!"))
		return err
	} else {
		return labelmaker.GenerateLabel(w, &params, ctrl.config)
	}
}

// HandleGetLocationLabel godoc
//
//	@Summary	Get Location label
//	@Tags		Locations
//	@Produce	json
//	@Param		id		path		string	true	"Location ID"
//	@Param		print	query		bool	false	"Print this label, defaults to false"
//	@Success	200		{string}	string	"image/png"
//	@Router		/v1/labelmaker/location/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetLocationLabel() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ID, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		auth := services.NewContext(r.Context())
		location, err := ctrl.repo.Locations.GetOneByGroup(auth, auth.GID, ID)
		if err != nil {
			return err
		}

		hbURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)
		return generateOrPrint(ctrl, w, r, location.Name, "Homebox Location", fmt.Sprintf("%s/location/%s", hbURL, location.ID))
	}
}

// HandleGetItemLabel godoc
//
//	@Summary	Get Item label
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string	true	"Item ID"
//	@Param		print	query		bool	false	"Print this label, defaults to false"
//	@Success	200		{string}	string	"image/png"
//	@Router		/v1/labelmaker/item/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetItemLabel() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ID, err := adapters.RouteUUID(r, "id")
		if err != nil {
			return err
		}

		auth := services.NewContext(r.Context())
		item, err := ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
		if err != nil {
			return err
		}

		description := ""

		if item.Location != nil {
			description += fmt.Sprintf("\nLocation: %s", item.Location.Name)
		}

		hbURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)
		return generateOrPrint(ctrl, w, r, item.Name, description, fmt.Sprintf("%s/item/%s", hbURL, item.ID))
	}
}

// HandleGetAssetLabel godoc
//
//	@Summary	Get Asset label
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string	true	"Asset ID"
//	@Param		print	query		bool	false	"Print this label, defaults to false"
//	@Success	200		{string}	string	"image/png"
//	@Router		/v1/labelmaker/assets/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGetAssetLabel() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		assetIDParam := chi.URLParam(r, "id")
		assetIDParam = strings.ReplaceAll(assetIDParam, "-", "")
		assetID, err := strconv.ParseInt(assetIDParam, 10, 64)
		if err != nil {
			return err
		}

		auth := services.NewContext(r.Context())
		item, err := ctrl.repo.Items.QueryByAssetID(auth, auth.GID, repo.AssetID(assetID), 0, 1)
		if err != nil {
			return err
		}

		if len(item.Items) == 0 {
			return validate.NewRequestError(fmt.Errorf("failed to find asset id"), http.StatusNotFound)
		}

		description := item.Items[0].Name

		if item.Items[0].Location != nil {
			description += fmt.Sprintf("\nLocation: %s", item.Items[0].Location.Name)
		}

		hbURL := GetHBURL(r, &ctrl.config.Options, ctrl.url)
		return generateOrPrint(ctrl, w, r, item.Items[0].AssetID.String(), description, fmt.Sprintf("%s/a/%s", hbURL, item.Items[0].AssetID.String()))
	}
}
