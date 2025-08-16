package http_handlers

import (
	"net/http"

	"github.com/sunr3d/order-stream-processor/internal/httpx"
)

func (h *httpHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "order-stream-processor",
	})
}
