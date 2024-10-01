package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type contextKey string
const appVersion string = "0.1.1"
const keyServerAddress contextKey = "server_address"

const (
	pathRoot               = "/"
	pathHealth             = "/health"
	pathContainers         = "/containers"
	getContainerPath       = "/container/"
	getContainerHealthPath = getContainerPath + "health/"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got / request\n", ctx.Value(keyServerAddress))
	io.WriteString(w, fmt.Sprintf("Conscript version %s\n", appVersion))
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got /health request from %s\n", ctx.Value(keyServerAddress))
	io.WriteString(w, "OK\n")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(pathRoot, getRoot)
	mux.HandleFunc(pathHealth, getHealth)

	mux.HandleFunc(pathContainers, listContainers)
	mux.HandleFunc(getContainerPath, getContainer(getContainerPath))
	mux.HandleFunc(getContainerHealthPath, getContainerHealth(getContainerHealthPath))

	ctx, cancelCtx := context.WithCancel(context.Background())

	healthServer := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(ctx, keyServerAddress, listener.Addr().String())
		},
	}

	go func() {
		err := healthServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("health server closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server two: %s\n", err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
}
