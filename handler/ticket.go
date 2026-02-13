package handler

import (
	"encoding/json"
	"net/http"
	"ticket-system/service"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Stock   int    `json:"stock"`
}

type TicketHandler struct {
	Service *service.TicketService
}

func NewTicketHandler(s *service.TicketService) *TicketHandler {
	return &TicketHandler{
		Service: s,
	}
}

func (h *TicketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	success, remaining := h.Service.BuyTicket(userID)

	if success {
		// 200 OK: 예매 성공
		json.NewEncoder(w).Encode(Response{
			Success: true,
			Message: "예매 성공!",
			Stock:   remaining,
		})
	} else if remaining <= 0 {
		// 410 Gone: 진짜 품절
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "매진되었습니다.",
			Stock:   0,
		})
	} else {
		// 409 Conflict: 락 경쟁 실패 (다시 시도 가능)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "접속자가 많아 실패했습니다. 다시 시도해 주세요.",
			Stock:   remaining,
		})
	}
}
