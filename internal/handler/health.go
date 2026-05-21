package handler

import "net/http"

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
