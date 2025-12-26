package v1

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"github.com/sysadminsmedia/homebox/backend/pkgs/printer"
)

// HandlePrintersGetAll godoc
//
//	@Summary	Get All Printers
//	@Tags		Printers
//	@Produce	json
//	@Success	200	{object}	[]repo.PrinterSummary
//	@Router		/v1/printers [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.PrinterSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Printers.GetAll(r.Context(), auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandlePrintersGet godoc
//
//	@Summary	Get Printer
//	@Tags		Printers
//	@Produce	json
//	@Param		id	path		string	true	"Printer ID"
//	@Success	200	{object}	repo.PrinterOut
//	@Router		/v1/printers/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.PrinterOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Printers.GetOne(r.Context(), auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandlePrintersCreate godoc
//
//	@Summary	Create Printer
//	@Tags		Printers
//	@Produce	json
//	@Param		payload	body		repo.PrinterCreate	true	"Printer Data"
//	@Success	201		{object}	repo.PrinterOut
//	@Router		/v1/printers [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.PrinterCreate) (repo.PrinterOut, error) {
		// Validate printer address to prevent SSRF
		if err := printer.ValidatePrinterAddress(body.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return repo.PrinterOut{}, err
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.Printers.Create(r.Context(), auth.GID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandlePrintersUpdate godoc
//
//	@Summary	Update Printer
//	@Tags		Printers
//	@Produce	json
//	@Param		id		path		string				true	"Printer ID"
//	@Param		payload	body		repo.PrinterUpdate	true	"Printer Data"
//	@Success	200		{object}	repo.PrinterOut
//	@Router		/v1/printers/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.PrinterUpdate) (repo.PrinterOut, error) {
		// Validate printer address to prevent SSRF
		if err := printer.ValidatePrinterAddress(body.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return repo.PrinterOut{}, err
		}

		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.Printers.Update(r.Context(), auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandlePrintersDelete godoc
//
//	@Summary	Delete Printer
//	@Tags		Printers
//	@Produce	json
//	@Param		id	path	string	true	"Printer ID"
//	@Success	204
//	@Router		/v1/printers/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Printers.Delete(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandlePrintersSetDefault godoc
//
//	@Summary	Set Default Printer
//	@Tags		Printers
//	@Produce	json
//	@Param		id	path	string	true	"Printer ID"
//	@Success	204
//	@Router		/v1/printers/{id}/set-default [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersSetDefault() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Printers.SetDefault(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// PrinterStatusResponse represents the status of a printer
type PrinterStatusResponse struct {
	Status      string   `json:"status"`
	Message     string   `json:"message,omitempty"`
	MediaReady  []string `json:"mediaReady,omitempty"`
	SupportsIPP bool     `json:"supportsIpp"`
}

// HandlePrintersStatus godoc
//
//	@Summary	Get Printer Status
//	@Tags		Printers
//	@Produce	json
//	@Param		id	path		string	true	"Printer ID"
//	@Success	200	{object}	PrinterStatusResponse
//	@Router		/v1/printers/{id}/status [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersStatus() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (PrinterStatusResponse, error) {
		auth := services.NewContext(r.Context())

		p, err := ctrl.repo.Printers.GetOne(r.Context(), auth.GID, ID)
		if err != nil {
			return PrinterStatusResponse{}, err
		}

		// Re-validate printer address for defense-in-depth (prevents SSRF if DB is compromised)
		if err := printer.ValidatePrinterAddress(p.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return PrinterStatusResponse{
				Status:      "invalid",
				Message:     "Printer address failed validation: " + err.Error(),
				SupportsIPP: false,
			}, nil
		}

		// Create printer client
		client, err := printer.NewPrinterClient(printer.PrinterType(p.PrinterType), p.Address)
		if err != nil {
			return PrinterStatusResponse{
				Status:      "offline",
				Message:     err.Error(),
				SupportsIPP: false,
			}, nil
		}

		// Get printer info
		info, err := client.GetPrinterInfo(r.Context())
		if err != nil {
			// Update status in database
			_ = ctrl.repo.Printers.UpdateStatus(r.Context(), auth.GID, ID, "offline")
			return PrinterStatusResponse{
				Status:      "offline",
				Message:     err.Error(),
				SupportsIPP: true,
			}, nil
		}

		// Update status in database
		_ = ctrl.repo.Printers.UpdateStatus(r.Context(), auth.GID, ID, string(info.State))

		return PrinterStatusResponse{
			Status:      string(info.State),
			Message:     info.StateMessage,
			MediaReady:  info.MediaReady,
			SupportsIPP: true,
		}, nil
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// PrinterTestRequest represents a test print request
type PrinterTestRequest struct {
	Message string `json:"message,omitempty"`
}

// PrinterTestResponse represents the result of a test print
type PrinterTestResponse struct {
	Success bool   `json:"success"`
	JobID   int    `json:"jobId,omitempty"`
	Message string `json:"message"`
}

// HandlePrintersTest godoc
//
//	@Summary	Test Print
//	@Tags		Printers
//	@Produce	json
//	@Param		id		path		string				true	"Printer ID"
//	@Param		payload	body		PrinterTestRequest	false	"Test options"
//	@Success	200		{object}	PrinterTestResponse
//	@Router		/v1/printers/{id}/test [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersTest() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		auth := services.NewContext(r.Context())

		idParam := r.PathValue("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return err
		}

		p, err := ctrl.repo.Printers.GetOne(r.Context(), auth.GID, id)
		if err != nil {
			return err
		}

		// Re-validate printer address for defense-in-depth (prevents SSRF if DB is compromised)
		if err := printer.ValidatePrinterAddress(p.Address, ctrl.config.Printer.AllowPublicAddresses); err != nil {
			return server.JSON(w, http.StatusOK, PrinterTestResponse{
				Success: false,
				Message: "Printer address failed validation: " + err.Error(),
			})
		}

		// Create printer client
		client, err := printer.NewPrinterClient(printer.PrinterType(p.PrinterType), p.Address)
		if err != nil {
			return server.JSON(w, http.StatusOK, PrinterTestResponse{
				Success: false,
				Message: "Failed to connect to printer: " + err.Error(),
			})
		}

		// Create a simple test label (1x1 inch white PNG)
		testData := createTestLabelPNG()

		// Send test print
		result, err := client.Print(r.Context(), &printer.PrintJob{
			DocumentName: "HomeBox Test Print",
			ContentType:  "image/png",
			Data:         testData,
			Copies:       1,
		})

		if err != nil {
			return server.JSON(w, http.StatusOK, PrinterTestResponse{
				Success: false,
				Message: "Print failed: " + err.Error(),
			})
		}

		return server.JSON(w, http.StatusOK, PrinterTestResponse{
			Success: result.Success,
			JobID:   result.JobID,
			Message: result.Message,
		})
	}
}

// createTestLabelPNG creates a simple test PNG image with "TEST" text pattern
func createTestLabelPNG() []byte {
	// Create a 100x100 image with a simple pattern
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with white background
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, white)
		}
	}

	// Draw a simple border
	for i := 0; i < 100; i++ {
		img.Set(i, 0, black)
		img.Set(i, 99, black)
		img.Set(0, i, black)
		img.Set(99, i, black)
	}

	// Draw a diagonal line as a simple test pattern
	for i := 10; i < 90; i++ {
		img.Set(i, i, black)
		img.Set(i, 100-i, black)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Error().Err(err).Msg("failed to encode test pattern PNG")
	}
	return buf.Bytes()
}

// BrotherMediaInfo represents a Brother media type for the frontend
type BrotherMediaInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	WidthMM      int    `json:"widthMm"`
	LengthMM     int    `json:"lengthMm"` // 0 for continuous
	IsContinuous bool   `json:"isContinuous"`
	TwoColor     bool   `json:"twoColor"`
}

// HandlePrintersMediaTypes godoc
//
//	@Summary	Get Brother Media Types
//	@Tags		Printers
//	@Produce	json
//	@Success	200	{object}	[]BrotherMediaInfo
//	@Router		/v1/printers/media-types [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandlePrintersMediaTypes() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]BrotherMediaInfo, error) {
		mediaTypes := printer.BrotherMediaTypes
		result := make([]BrotherMediaInfo, 0, len(mediaTypes))

		for id, media := range mediaTypes {
			result = append(result, BrotherMediaInfo{
				ID:           id,
				Name:         media.Name,
				WidthMM:      media.WidthMM,
				LengthMM:     media.LengthMM,
				IsContinuous: media.LengthMM == 0,
				TwoColor:     media.TwoColor,
			})
		}

		return result, nil
	}

	return adapters.Command(fn, http.StatusOK)
}
