package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/ItencY/internal/runner"
)

const maxRequestSize = 1024 * 1024 // 1 MB payload protection

type AppHandler struct {
	runner runner.Runner
}

func NewAppHandler(r runner.Runner) *AppHandler {
	return &AppHandler{
		runner: r,
	}
}

func (ah *AppHandler) handlerRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	limitedBody := io.LimitReader(r.Body, maxRequestSize)
	var req runner.CommandRequest
	decoder := json.NewDecoder(limitedBody)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON structure", http.StatusBadRequest)
		return
	}
	if len(req.Args) == 0 {
		http.Error(w, "Arguments list cannot be empty", http.StatusBadRequest)
		return
	}
	resp, err := ah.runner.Run(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, runner.ErrUnauthorizedCommand):
			http.Error(w, "Unauthorized command", http.StatusForbidden)
			return
		case errors.Is(err, runner.ErrCommandTimeout):
			http.Error(w, "command timed out", http.StatusGatewayTimeout)
			return
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response JSON: %v", err)
	}
}
