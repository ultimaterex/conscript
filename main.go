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

const keyServerAddress contextKey = "server_address"

const (
	pathRoot         = "/"
	pathHealth       = "/health"
	pathContainers   = "/containers"
	getContainerPath = "/container/"
	getContainerHealthPath = getContainerPath + "health/"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got / request\n", ctx.Value(keyServerAddress))
	io.WriteString(w, "Conscript alpha\n")
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
	// configServer := &http.Server{
	// 	Addr:    ":3332",
	// 	Handler: mux,
	// 	BaseContext: func(listener net.Listener) context.Context {
	// 		return context.WithValue(ctx, keyServerAddress, listener.Addr().String())
	// 	},
	// }

	healthServer := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(listener net.Listener) context.Context {
			return context.WithValue(ctx, keyServerAddress, listener.Addr().String())
		},
	}

	// go func() {
	// 	err := configServer.ListenAndServe()
	// 	if errors.Is(err, http.ErrServerClosed) {
	// 		fmt.Printf("config server closed\n")
	// 	} else if err != nil {
	// 		fmt.Printf("error listening for server one: %s\n", err)
	// 	}
	// 	cancelCtx()
	// }()

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
