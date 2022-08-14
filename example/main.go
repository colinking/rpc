package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/colinking/rpc/example/generated"
)

func main() {
	mux := http.NewServeMux()

	handler := &Handler{}
	generated.Register(mux, handler)

	server := &http.Server{
		Addr:    "0.0.0.0:4000",
		Handler: mux,
	}
	fmt.Printf("Listening on %s...\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Closed successfully")
			return
		}
		fmt.Printf("Server errored: %+v\n", err)
	}
}

type Handler struct{}

var _ generated.Handler = &Handler{}

// GET /v0/runs/get
func (h *Handler) V0RunsGet(ctx context.Context, req generated.V0RunsGetRequest) (generated.V0RunsGetResponse, error) {
	return generated.V0RunsGetResponse{
		RunID: req.ID,
	}, nil
}

// POST /v0/tasks/execute
func (h *Handler) V0TasksExecute(ctx context.Context, req generated.V0TasksExecuteRequest) (generated.V0TasksExecuteResponse, error) {
	return generated.V0TasksExecuteResponse{}, errors.New("not implemented")
}
