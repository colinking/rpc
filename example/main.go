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

	routes := Routes{}
	generated.Register(mux, routes)

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

type Routes struct{}

var _ generated.Routes = Routes{}

// GET /v0/runs/get
func (r Routes) V0RunsGet(ctx context.Context, request generated.V0RunsGetRequest) (generated.V0RunsGetResponse, error) {
	return generated.V0RunsGetResponse{
		RunID: request.ID,
	}, nil
}

// POST /v0/tasks/execute
func (r Routes) V0TasksExecute(ctx context.Context, request generated.V0TasksExecuteRequest) (generated.V0TasksExecuteResponse, error) {
	return generated.V0TasksExecuteResponse{}, errors.New("not implemented")
}
