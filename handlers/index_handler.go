package handlers

import (
	"encoding/json"
	"net/http"

	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
)

type Index struct {
	Version string `json:"version"`
}

func IndexHandler(app *App) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, c Context) {
		index := Index{"Carts API 1.0.0"}

		json.NewEncoder(w).Encode(index)
	}
}
