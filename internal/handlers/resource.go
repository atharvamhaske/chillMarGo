package handlers

import (
	"net/http"

	"github.com/chillMarGO/internal/types"
	"github.com/chillMarGO/internal/utils"
)

func Resource(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.WriteJSON(w, http.StatusOK, types.Response{
			Success: true,
			Data:    "You will get resource",
		})
	default:
		utils.WriteJSON(w, http.StatusMethodNotAllowed, types.Response{
			Success: false,
			Error:   "Only `GET` method is supported",
		})
	}
}
