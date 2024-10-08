package v1

import (
	"bytes"
	"image/png"
	"io"
	"net/http"
	"net/url"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	_ "embed"
)

//go:embed assets/QRIcon.png
var qrcodeLogo []byte

// HandleGenerateQRCode godoc
//
//	@Summary	Create QR Code
//	@Tags		Items
//	@Produce	json
//	@Param		data	query		string	false	"data to be encoded into qrcode"
//	@Success	200		{string}	string	"image/jpeg"
//	@Router		/v1/qrcode [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleGenerateQRCode() errchain.HandlerFunc {
	type query struct {
		// 4,296 characters is the maximum length of a QR code
		Data string `schema:"data" validate:"required,max=4296"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		q, err := adapters.DecodeQuery[query](r)
		if err != nil {
			return err
		}

		image, err := png.Decode(bytes.NewReader(qrcodeLogo))
		if err != nil {
			panic(err)
		}

		decodedStr, err := url.QueryUnescape(q.Data)
		if err != nil {
			return err
		}

		qrc, err := qrcode.New(decodedStr)
		if err != nil {
			return err
		}

		toWriteCloser := struct {
			io.Writer
			io.Closer
		}{
			Writer: w,
			Closer: io.NopCloser(nil),
		}

		qrwriter := standard.NewWithWriter(toWriteCloser, standard.WithLogoImage(image))

		// Return the QR code as a jpeg image
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=qrcode.jpg")
		return qrc.Save(qrwriter)
	}
}
