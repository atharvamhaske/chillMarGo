package middleware

import (
	"net/http"

	"github.com/chillMarGO/internal/types"
	"github.com/chillMarGO/internal/utils"
)

func RateLimiterWrapper(next http.HandlerFunc, f func(string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(utils.UserIP(r.RemoteAddr)); 
		err != nil {
			utils.WriteJSON(w, http.StatusTooManyRequests, types.Response{
				Success: false,
				Error: err.Error(),
			})
			return 
		}
		next(w, r)
	}
}