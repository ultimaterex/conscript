package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerInfo struct {
	ID     string   `json:"id"`
	Names  []string `json:"names"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
}


func createDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	return cli, nil
}

// handleError sends an error response with the given message and status code.
func handleError(w http.ResponseWriter, message string, err error, statusCode int) {
    http.Error(w, fmt.Sprintf("%s: %v", message, err), statusCode)
}

// formatContainerInfo formats the container information for JSON output.
func formatContainerInfo(containers []types.Container) []ContainerInfo {
    var containerInfos []ContainerInfo
    for _, container := range containers {
        containerInfos = append(containerInfos, ContainerInfo{
            ID:     container.ID[:10],
            Names:  container.Names,
            Image:  container.Image,
            Status: container.Status,
        })
    }
    return containerInfos
}


// listContainers returns a handler function that handles requests for listing all Docker containers.
//
// Endpoint: /containers
// Method: GET
// Description: Lists all Docker containers.
//
// Response: A list of Docker containers.
//
//	The response object may contain the following fields:
//   - ID: The ID of the container.
//   - Names: The names of the container.
//   - Image: The image used to create the container.
//   - Status: The current status of the container.
//
//	The response object may also contain additional fields based on query parameters:
//   - json: Whether to return the response as JSON.
//   - all: Whether to list all containers, including stopped containers.
// 
// 	Example response (text):
// 	ID: f7b4d9f2f2, Name: [/my-container], Image: my-image:latest, Status: running
// 
// 	Example response (json):
// 	[
// 	  {
// 	    "id": "f7b4d9f2f2",
// 	    "names": [
// 	      "/my-container"
// 	    ],
// 	    "image": "my-image:latest",
// 	    "status": "running"
// 	  }
// 	]
func listContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("got /containers request from %s\n", ctx.Value(keyServerAddress))

    cli, err := createDockerClient()
    if err != nil {
        handleError(w, "Failed to create Docker client", err, http.StatusInternalServerError)
        return
    }

    // Check for the "all" query parameter
    query := r.URL.Query()
    listAll := query.Get("all") == "true"

    containers, err := cli.ContainerList(ctx, container.ListOptions{All: listAll})
    if err != nil {
        handleError(w, "Failed to list containers", err, http.StatusInternalServerError)
        return
    }

    // Check for the "json" query parameter
    if query.Get("json") == "true" {
        w.Header().Set("Content-Type", "application/json")
        containerInfos := formatContainerInfo(containers)
        if err := json.NewEncoder(w).Encode(containerInfos); err != nil {
            handleError(w, "Failed to encode JSON response", err, http.StatusInternalServerError)
        }
        return
    }

    // Default output format
    for _, container := range containers {
        fmt.Fprintf(w, "ID: %s, Name: %s, Image: %s, Status: %s\n", container.ID[:10], container.Names, container.Image, container.Status)
    }
}


// getContainer returns a handler function that handles requests for a specific container.
//
// Endpoint: /container/{containerID}
// Method: GET
// Description: Retrieves information about a specific Docker container.
// Parameters:
//   - containerID: The ID of the container to retrieve information for.
//
// Response: A JSON object containing information about the container.
//
//	The response object may contain the following fields:
//   - ID: The ID of the container.
//   - Name: The name of the container.
//   - Image: The image used to create the container.
//   - Status: The current status of the container.
//   - State: The state of the container.
//
//	The response object may also contain additional fields based on query parameters:
//   - status: The status of the container.
//   - running: Whether the container is running.
//   - paused: Whether the container is paused.
//   - restarting: Whether the container is restarting.
//   - dead: Whether the container is dead.
//   - error: Any error message associated with the container.
//   - exitcode: The exit code of the container.
//   - state: The full state object of the container.
func getContainer(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		containerID := strings.TrimPrefix(r.URL.Path, basePath)

		fmt.Printf("got request for container %s from %s\n", containerID, ctx.Value(keyServerAddress))

		// Check if containerID is empty
		if containerID == "" {
			http.Error(w, "Container name is required", http.StatusBadRequest)
			return
		}

        cli, err := createDockerClient()
        if err != nil {
            handleError(w, "Failed to create Docker client", err, http.StatusInternalServerError)
            return
        }

        containerJSON, err := cli.ContainerInspect(ctx, containerID)
        if err != nil {
            handleError(w, fmt.Sprintf("Failed to inspect container '%s'", containerID), err, http.StatusInternalServerError)
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
            handleError(w, "Failed to marshal container JSON", err, http.StatusInternalServerError)
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
//
// Response: A message indicating the container's status.
//
//	If the container is not running, an appropriate HTTP status code is returned.
func getContainerHealth(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		containerID := strings.TrimPrefix(r.URL.Path, basePath)

		fmt.Printf("got /health request for container %s from %s\n", containerID, ctx.Value(keyServerAddress))

		// Check if containerID is empty
		if containerID == "" {
			http.Error(w, "Container name is required", http.StatusBadRequest)
			return
		}

        cli, err := createDockerClient()
        if err != nil {
            handleError(w, "Failed to create Docker client", err, http.StatusInternalServerError)
            return
        }

        containerJSON, err := cli.ContainerInspect(ctx, containerID)
        if err != nil {
            handleError(w, fmt.Sprintf("Failed to inspect container '%s'", containerID), err, http.StatusInternalServerError)
            return
        }

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
