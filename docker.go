package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// listContainers lists all Docker containers.
func listContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got /containers request from %s\n", ctx.Value(keyServerAddress))

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		http.Error(w, "Failed to create Docker client", http.StatusInternalServerError)
		return
	}
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		http.Error(w, "Failed to list containers", http.StatusInternalServerError)
		return
	}

	for _, container := range containers {
		fmt.Fprintf(w, "ID: %s, Name: %s, Image: %s, Status: %s\n", container.ID[:10], container.Names, container.Image, container.Status)
	}
}

// getContainer returns a handler function that handles requests for a specific container.
// It inspects the container and returns its details in JSON format.
func getContainer(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		containerID := strings.TrimPrefix(r.URL.Path, basePath)

		fmt.Printf("got request for container %s from %s\n", containerID, ctx.Value(keyServerAddress))

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			http.Error(w, "Failed to create Docker client", http.StatusInternalServerError)
			return
		}

		containerJSON, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			http.Error(w, "Failed to inspect container", http.StatusInternalServerError)
			return
		}

		query := r.URL.Query()
		response := make(map[string]interface{})

		if _, ok := query["status"]; ok {
			response["Status"] = containerJSON.State.Status
		}
		if _, ok := query["running"]; ok {
			response["Running"] = containerJSON.State.Running
		}
		if _, ok := query["paused"]; ok {
			response["Paused"] = containerJSON.State.Paused
		}
		if _, ok := query["restarting"]; ok {
			response["Restarting"] = containerJSON.State.Restarting
		}
		if _, ok := query["dead"]; ok {
			response["Dead"] = containerJSON.State.Dead
		}
		if _, ok := query["error"]; ok {
			response["Error"] = containerJSON.State.Error
		}
		if _, ok := query["exitcode"]; ok {
			response["ExitCode"] = containerJSON.State.ExitCode
		}
		if _, ok := query["state"]; ok {
			response["State"] = containerJSON.State
		}

		if len(response) == 0 {
			response = map[string]interface{}{
				"ID":     containerJSON.ID,
				"Name":   containerJSON.Name,
				"Image":  containerJSON.Image,
				"Status": containerJSON.State.Status,
				"State":  containerJSON.State,
			}
		}

		jsonOutput, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			http.Error(w, "Failed to marshal container JSON", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s", jsonOutput)
	}
}


// getContainerHealth returns a handler function that handles health check requests for a specific container.
//
// Endpoint: /container/health/{containerID}
// Method: GET
// Description: Checks the health of a specific Docker container.
// Parameters:
//   - containerID: The ID of the container to check health for.
// Response: A message indicating the container's status.
//           If the container is not running, an appropriate HTTP status code is returned.
func getContainerHealth(basePath string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        containerID := strings.TrimPrefix(r.URL.Path, basePath)

        fmt.Printf("got /health request for container %s from %s\n", containerID, ctx.Value(keyServerAddress))

        cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
        if err != nil {
            http.Error(w, "Failed to create Docker client", http.StatusInternalServerError)
            return
        }

        containerJSON, err := cli.ContainerInspect(ctx, containerID)
        if err != nil {
            http.Error(w, "Failed to inspect container", http.StatusInternalServerError)
            return
        }

        // if containerJSON.State.Status != "running" {
        //     http.Error(w, fmt.Sprintf("Container %s is not running. Current status: %s", containerID, containerJSON.State.Status), http.StatusInternalServerError)
        //     return
        // }

		status := containerJSON.State.Status
        var httpStatus int
        var message string

		switch status {
        case "running":
            httpStatus = http.StatusOK
            message = fmt.Sprintf("Container %s is healthy and running.", containerID)
        case "created":
            httpStatus = http.StatusAccepted
            message = fmt.Sprintf("Container %s is created but not running.", containerID)
        case "restarting":
            httpStatus = http.StatusServiceUnavailable
            message = fmt.Sprintf("Container %s is restarting.", containerID)
        case "removing":
            httpStatus = http.StatusServiceUnavailable
            message = fmt.Sprintf("Container %s is being removed.", containerID)
        case "paused":
            httpStatus = http.StatusLocked
            message = fmt.Sprintf("Container %s is paused.", containerID)
        case "exited":
            httpStatus = http.StatusGone
            message = fmt.Sprintf("Container %s has exited.", containerID)
        case "dead":
            httpStatus = http.StatusInternalServerError
            message = fmt.Sprintf("Container %s is dead.", containerID)
        default:
            httpStatus = http.StatusInternalServerError
            message = fmt.Sprintf("Container %s is in an unknown state: %s.", containerID, status)
        }

		http.Error(w, message, httpStatus)
    }
}