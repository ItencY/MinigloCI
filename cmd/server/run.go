package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
)

type Request struct {
	Args []string `json:"args"`
}

type Response struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func handlerRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON structure", http.StatusBadRequest)
		return
	}
	if len(req.Args) == 0 {
		http.Error(w, "Arguments list cannot be empty", http.StatusBadRequest)
		return
	}
	binary := req.Args[0]
	args := req.Args[1:]
	// #nosec G204
	cmd := exec.Command(binary, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	resp := Response{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response JSON: %v", err)
	}
}
