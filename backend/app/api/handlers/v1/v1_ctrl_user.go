package v1

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"go.opentelemetry.io/otel/attribute"
)

// HandleUserRegistration godoc
//
//		@Summary	Register New User
//		@Tags		User
//		@Produce	json
//		@Param		payload	body	services.UserRegistration	true	"User Data"
//		@Success	204
//	 @Failure    403 {string} string "Local login is not enabled"
//		@Router		/v1/users/register [Post]
func (ctrl *V1Controller) HandleUserRegistration() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserRegistration")
		defer span.End()

		if !ctrl.config.Options.AllowLocalLogin {
			span.SetAttributes(attribute.String("registration.outcome", "local_login_disabled"))
			return validate.NewRequestError(fmt.Errorf("local login is not enabled"), http.StatusForbidden)
		}

		regData := services.UserRegistration{}

		if err := server.Decode(r, &regData); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("registration.outcome", "decode_failed"))
			log.Err(err).Msg("failed to decode user registration data")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(
			attribute.Int("user.name.length", len(regData.Name)),
			attribute.Int("user.email.length", len(regData.Email)),
			attribute.Int("user.password.length", len(regData.Password)),
			attribute.Bool("registration.has_group_token", regData.GroupToken != ""),
		)

		if !ctrl.allowRegistration && regData.GroupToken == "" {
			span.SetAttributes(attribute.String("registration.outcome", "registration_disabled"))
			return validate.NewRequestError(fmt.Errorf("user registration disabled"), http.StatusForbidden)
		}

		usr, err := ctrl.svc.User.RegisterUser(spanCtx, regData)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("registration.outcome", "register_failed"))
			log.Err(err).Msg("failed to register user")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(
			attribute.String("registration.outcome", "success"),
			attribute.String("user.id", usr.ID.String()),
			attribute.String("group.id", usr.DefaultGroupID.String()),
		)

		return server.JSON(w, http.StatusNoContent, nil)
	}
}

// HandleUserSelf godoc
//
//	@Summary	Get User Self
//	@Tags		User
//	@Produce	json
//	@Success	200	{object}	Wrapped{item=repo.UserOut}
//	@Router		/v1/users/self [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelf() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelf")
		defer span.End()

		token := services.UseTokenCtx(spanCtx)
		span.SetAttributes(attribute.Bool("token.present", token != ""))
		usr, err := ctrl.svc.User.GetSelf(spanCtx, token)
		if usr.ID == uuid.Nil || err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("self.outcome", "lookup_failed"))
			log.Err(err).Msg("failed to get user")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.String("user.id", usr.ID.String()))

		return server.JSON(w, http.StatusOK, Wrap(usr))
	}
}

// HandleUserSelfUpdate godoc
//
//	@Summary	Update Account
//	@Tags		User
//	@Produce	json
//	@Param		payload	body		repo.UserUpdate	true	"User Data"
//	@Success	200		{object}	Wrapped{item=repo.UserUpdate}
//	@Router		/v1/users/self [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelfUpdate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelfUpdate")
		defer span.End()

		updateData := repo.UserUpdate{}
		if err := server.Decode(r, &updateData); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("update.outcome", "decode_failed"))
			log.Err(err).Msg("failed to decode user update data")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		actor := services.UseUserCtx(spanCtx)
		span.SetAttributes(attribute.String("user.id", actor.ID.String()))
		newData, err := ctrl.svc.User.UpdateSelf(spanCtx, actor.ID, updateData)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("update.outcome", "service_failed"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.String("update.outcome", "success"))

		return server.JSON(w, http.StatusOK, Wrap(newData))
	}
}

// HandleUserSelfDelete godoc
//
//	@Summary	Delete Account
//	@Tags		User
//	@Produce	json
//	@Success	204
//	@Router		/v1/users/self [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelfDelete() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelfDelete")
		defer span.End()

		if ctrl.isDemo {
			span.SetAttributes(attribute.String("delete.outcome", "demo_blocked"))
			return validate.NewRequestError(nil, http.StatusForbidden)
		}

		actor := services.UseUserCtx(spanCtx)
		span.SetAttributes(attribute.String("user.id", actor.ID.String()))
		if err := ctrl.svc.User.DeleteSelf(spanCtx, actor.ID); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("delete.outcome", "delete_failed"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.String("delete.outcome", "success"))

		return server.JSON(w, http.StatusNoContent, nil)
	}
}

// HandleUserSelfSettingsGet godoc
//
//	@Summary	Get user settings
//	@Tags		User
//	@Produce	json
//	@Success	200	{object}	Wrapped{item=map[string]interface{}}
//	@Router		/v1/users/self/settings [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelfSettingsGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelfSettingsGet")
		defer span.End()

		actor := services.UseUserCtx(spanCtx)
		span.SetAttributes(attribute.String("user.id", actor.ID.String()))
		settings, err := ctrl.svc.User.GetSettings(spanCtx, actor.ID)
		if err != nil {
			recordCtrlSpanError(span, err)
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Int("settings.keys.count", len(settings)))

		w.Header().Set("Cache-Control", "no-store")
		return server.JSON(w, http.StatusOK, Wrap(settings))
	}
}

// HandleUserSelfSettingsUpdate godoc
//
//	@Summary	Update user settings
//	@Tags		User
//	@Produce	json
//	@Success	200	{object}	Wrapped{item=map[string]interface{}}
//	@Router		/v1/users/self/settings [PUT]
//	@Param		payload	body	map[string]interface{}	true	"Settings Data"
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelfSettingsUpdate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelfSettingsUpdate")
		defer span.End()

		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
		var settings map[string]interface{}
		if err := server.Decode(r, &settings); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("settings.outcome", "decode_failed"))
			log.Err(err).Msg("failed to decode user settings data")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		span.SetAttributes(attribute.Int("settings.keys.count", len(settings)))

		actor := services.UseUserCtx(spanCtx)
		span.SetAttributes(attribute.String("user.id", actor.ID.String()))
		if err := ctrl.svc.User.SetSettings(spanCtx, actor.ID, settings); err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("settings.outcome", "set_failed"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		ctx := services.NewContext(spanCtx)
		ctrl.bus.Publish(eventbus.EventUserMutation, eventbus.GroupMutationEvent{GID: ctx.GID})

		newSettings, err := ctrl.svc.User.GetSettings(spanCtx, actor.ID)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("settings.outcome", "reload_failed"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.String("settings.outcome", "success"))

		w.Header().Set("Cache-Control", "no-store")
		return server.JSON(w, http.StatusOK, Wrap(newSettings))
	}
}

type (
	ChangePassword struct {
		Current string `json:"current,omitempty"`
		New     string `json:"new,omitempty"`
	}
)

// HandleUserSelfChangePassword godoc
//
//	@Summary	Change Password
//	@Tags		User
//	@Success	204
//	@Param		payload	body	ChangePassword	true	"Password Payload"
//	@Router		/v1/users/change-password [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleUserSelfChangePassword() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleUserSelfChangePassword")
		defer span.End()

		if ctrl.isDemo {
			span.SetAttributes(attribute.String("change_password.outcome", "demo_blocked"))
			return validate.NewRequestError(nil, http.StatusForbidden)
		}

		var cp ChangePassword
		err := server.Decode(r, &cp)
		if err != nil {
			recordCtrlSpanError(span, err)
			span.SetAttributes(attribute.String("change_password.outcome", "decode_failed"))
			log.Err(err).Msg("user failed to change password")
		}
		span.SetAttributes(
			attribute.Int("password.current.length", len(cp.Current)),
			attribute.Int("password.new.length", len(cp.New)),
		)

		ctx := services.NewContext(spanCtx)
		span.SetAttributes(attribute.String("user.id", ctx.UID.String()))

		ok := ctrl.svc.User.ChangePassword(ctx, cp.Current, cp.New)
		span.SetAttributes(attribute.Bool("change_password.ok", ok))
		if !ok {
			span.SetAttributes(attribute.String("change_password.outcome", "service_rejected"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.String("change_password.outcome", "success"))

		return server.JSON(w, http.StatusNoContent, nil)
	}
}
