package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// Allowlist of permitted utilities (RCE protection)
var allowedBinaries = map[string]string{
	"echo": "/usr/bin/echo",
}

const (
	maxRequestSize = 1024 * 1024     // 1 MB for memory protection
	commandTimeout = 5 * time.Second // Command execution timeout
)

type RunRequest struct {
	Command string   `json:"command"` // Passing a fixed cmd name
	Args    []string `json:"args"`
}

type RunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func handlerRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	limitedBody := io.LimitReader(r.Body, maxRequestSize)
	var req RunRequest
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
	// Command validation via whitelist
	binaryPath, exists := allowedBinaries[req.Command]
	if !exists {
		http.Error(w, "Unauthorized command", http.StatusForbidden)
		return
	}
	// We create a context with a timeout so that the server doesn't hang.
	ctx, cancel := context.WithTimeout(r.Context(), commandTimeout)
	defer cancel()
	// #nosec G204
	cmd := exec.CommandContext(ctx, binaryPath, req.Args...)
	stdoutBuf, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	stderrBuf, err := cmd.StderrPipe()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err = cmd.Start(); err != nil {
		http.Error(w, "Failed to start command", http.StatusInternalServerError)
		return
	}
	// Reading the output (limiting the volume to avoid overflowing server memory)
	stdoutBytes, _ := io.ReadAll(io.LimitReader(stdoutBuf, maxRequestSize))
	stderrBytes, _ := io.ReadAll(io.LimitReader(stderrBuf, maxRequestSize))
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		// Check whether the process has terminated due to a context timeout.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			http.Error(w, "Command timed out", http.StatusGatewayTimeout)
			return
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			http.Error(w, "Execution failed", http.StatusInternalServerError)
			return
		}
	}
	resp := RunResponse{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response JSON: %v", err)
	}
}
