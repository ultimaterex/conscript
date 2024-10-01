package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

type contextKey string

const appVersion string = "0.1.2"
const keyServerAddress contextKey = "server_address"

var startTime = time.Now()

const (
	pathRoot               = "/"
	pathHealth             = "/health"
	pathInfo               = "/info"
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

type Info struct {
	ApplicationVersion string `json:"application_version"`
	Hostname           string `json:"hostname"`
	Uptime             string `json:"uptime"`
	CurrentTime        string `json:"current_time"`
	GoVersion          string `json:"go_version"`
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got /info request from %s\n", ctx.Value(keyServerAddress))
	hostname, err := os.Hostname()
	if err != nil {
		http.Error(w, "Failed to get hostname", http.StatusInternalServerError)
		return
	}

	uptime := time.Since(startTime).String()
	currentTime := time.Now().Format(time.RFC1123)
	goVersion := runtime.Version()

	info := Info{
		ApplicationVersion: appVersion,
		Hostname:           hostname,
		Uptime:             uptime,
		CurrentTime:        currentTime,
		GoVersion:          goVersion,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(pathRoot, getRoot)
	mux.HandleFunc(pathHealth, getHealth)
	mux.HandleFunc(pathInfo, getInfo)

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
