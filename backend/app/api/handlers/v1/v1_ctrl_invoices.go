package v1

import (
	"io"
	"net/http"

	"github.com/hay-kot/httpkit/errchain"
)

// HandleProcessInvoice actúa como proxy para enviar el archivo PDF al contenedor de Python
func (c *V1Controller) HandleProcessInvoice() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1. Redirigimos la petición al contenedor de Python a través de la red local
		proxyReq, err := http.NewRequest(r.Method, "http://procesador_pdf:8000/procesar", r.Body)
		if err != nil {
			return err // errchain se encargará de formatear el error
		}

		// Copiamos los headers para preservar el multipart/form-data
		proxyReq.Header = r.Header

		// 2. Ejecutamos la petición hacia Python
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// 3. Devolvemos la respuesta exacta (JSON) al Frontend
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		
		return err
	}
}