package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

type ActionAmountResult struct {
	Completed int `json:"completed"`
}

func actionHandlerFactory(ref string, fn func(context.Context, uuid.UUID) (int, error)) errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())

		totalCompleted, err := fn(ctx, ctx.GID)
		if err != nil {
			log.Err(err).Str("action_ref", ref).Msg("failed to run action")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, ActionAmountResult{Completed: totalCompleted})
	}
}

// HandleEnsureAssetID godoc
//
//	@Summary		Ensure Asset IDs
//	@Description	Ensures all items in the database have an asset ID
//	@Tags			Actions
//	@Produce		json
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/ensure-asset-ids [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleEnsureAssetID() errchain.HandlerFunc {
	return actionHandlerFactory("ensure asset IDs", ctrl.svc.Items.EnsureAssetID)
}

// HandleEnsureImportRefs godoc
//
//	@Summary		Ensures Import Refs
//	@Description	Ensures all items in the database have an import ref
//	@Tags			Actions
//	@Produce		json
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/ensure-import-refs [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleEnsureImportRefs() errchain.HandlerFunc {
	return actionHandlerFactory("ensure import refs", ctrl.svc.Items.EnsureImportRef)
}

// HandleItemDateZeroOut godoc
//
//	@Summary		Zero Out Time Fields
//	@Description	Resets all item date fields to the beginning of the day
//	@Tags			Actions
//	@Produce		json
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/zero-item-time-fields [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleItemDateZeroOut() errchain.HandlerFunc {
	return actionHandlerFactory("zero out date time", ctrl.repo.Items.ZeroOutTimeFields)
}

// HandleSetPrimaryPhotos godoc
//
//	@Summary		Set Primary Photos
//	@Description	Sets the first photo of each item as the primary photo
//	@Tags			Actions
//	@Produce		json
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/set-primary-photos [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleSetPrimaryPhotos() errchain.HandlerFunc {
	return actionHandlerFactory("ensure asset IDs", ctrl.repo.Items.SetPrimaryPhotos)
}

// HandleCreateMissingThumbnails godoc
//
//	@Summary		Create Missing Thumbnails
//	@Description	Creates thumbnails for items that are missing them
//	@Tags			Actions
//	@Produce		json
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/create-missing-thumbnails [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleCreateMissingThumbnails() errchain.HandlerFunc {
	return actionHandlerFactory("create missing thumbnails", ctrl.repo.Attachments.CreateMissingThumbnails)
}

// WipeInventoryOptions represents the options for wiping inventory
type WipeInventoryOptions struct {
	WipeTags        bool `json:"wipeTags"`
	WipeLocations   bool `json:"wipeLocations"`
	WipeMaintenance bool `json:"wipeMaintenance"`
}

// HandleWipeInventory godoc
//
//	@Summary		Wipe Inventory
//	@Description	Deletes all items in the inventory
//	@Tags			Actions
//	@Produce		json
//	@Param			options	body	WipeInventoryOptions	false	"Wipe options"
//	@Success		200	{object}	ActionAmountResult
//	@Router			/v1/actions/wipe-inventory [Post]
//	@Security		Bearer
func (ctrl *V1Controller) HandleWipeInventory() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if ctrl.isDemo {
			return validate.NewRequestError(errors.New("wipe inventory is not allowed in demo mode"), http.StatusForbidden)
		}

		ctx := services.NewContext(r.Context())

		// Check if user is owner
		if !ctx.User.IsOwner {
			return validate.NewRequestError(errors.New("only group owners can wipe inventory"), http.StatusForbidden)
		}

		// Parse options from request body
		var options WipeInventoryOptions
		if err := server.Decode(r, &options); err != nil {
			// If no body provided, use default (false for all)
			options = WipeInventoryOptions{
				WipeTags:        false,
				WipeLocations:   false,
				WipeMaintenance: false,
			}
		}

		totalCompleted, err := ctrl.repo.Items.WipeInventory(ctx, ctx.GID, options.WipeTags, options.WipeLocations, options.WipeMaintenance)
		if err != nil {
			log.Err(err).Str("action_ref", "wipe inventory").Msg("failed to run action")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		// Publish mutation events for wiped resources
		if ctrl.bus != nil {
			if options.WipeTags {
				ctrl.bus.Publish(eventbus.EventTagMutation, eventbus.GroupMutationEvent{GID: ctx.GID})
			}
			if options.WipeLocations {
				ctrl.bus.Publish(eventbus.EventLocationMutation, eventbus.GroupMutationEvent{GID: ctx.GID})
			}
		}

		return server.JSON(w, http.StatusOK, ActionAmountResult{Completed: totalCompleted})
	}
}
