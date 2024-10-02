# Conscript

Conscript is a Go-based HTTP server that provides various read only endpoints to interact with Docker containers. It includes endpoints for health checks, listing containers, and retrieving container details based on query parameters.

## Features

- **Health Check**: Endpoint to check the health of the server.
- **List Containers**: Endpoint to list Docker containers.
- **Get Container Details**: Endpoint to get details of a specific Docker container based on query parameters.

## Running the Application

**Access the application**: once deployed open your browser and navigate to `http://localhost:3333`.

### Running from GitHub Container Registry

**Run the Docker container**:

```sh
docker run -d \
  -p 3333:3333 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /etc/hostname:/etc/host_hostname \ # Optional, mount if you want machine_hostname to be available when running in docker environments
  ghcr.io/ultimaterex/conscript/conscript:latest
```

### Running with Docker Compose

1. **Create a `docker-compose.yml` file with the following content**:

   ```yaml
   services:
     conscript:
       command:
         - './conscript'
       container_name: 'conscript'
       hostname: 'conscript'
       image: 'ghcr.io/ultimaterex/conscript/conscript:latest'
       ports:
         - '3333:3333/tcp'
       restart: 'unless-stopped'
       volumes:
         - '/var/run/docker.sock:/var/run/docker.sock'
         - "/etc/hostname:/etc/host_hostname" # Optional, mount if you want machine_hostname to be available when running in docker environments
       working_dir: '/root'
   ```

2. **Run the Docker Compose setup**:

   ```sh
   docker-compose up -d
   ```

## Building and Running Locally

### Prerequisites

- Go 1.23.1 or later
- or Docker

1. **Clone the repository**:

   ```sh
   git clone https://github.com/yourusername/conscript.git
   cd conscript
   ```

2. **Build the application**:

   ```sh
   go build -o conscript .
   ```

3. **Run the application**:

   ```sh
   ./conscript
   ```

### Running with Docker

1. **Build the Docker image**:

   ```sh
   docker build -t conscript:latest .
   ```

2. **Run the Docker container**:

   ```sh
   docker run -p 3333:3333 -v /var/run/docker.sock:/var/run/docker.sock conscript:latest
   ```

## Endpoints

### Root Endpoint

- **URL**: `/`
- **Method**: `GET`
- **Description**: Returns a welcome message.
- **Response**:
  ```json
  Conscript version 0.1.0
  ```

### Health Check Endpoint

- **URL**: `/health`
- **Method**: `GET`
- **Description**: Returns the health status of the server.
- **Response**:
  ```json
  OK
  ```

### List Containers Endpoint

- **URL**: `/containers`
- **Method**: `GET`
- **Description**: Lists Docker containers. By default, it lists only running containers. If the "all" query parameter is set, it lists all containers.
- **Query Parameters**:
  - `all`: If set to `true`, lists all containers.
  - `json`: If set to `true`, returns the response in JSON format.
- **Response**: JSON array of container details if `json=true` is set, otherwise plain text.
  - **Default (Plain Text)**:
    ```
    ID: 1234567890, Name: /my-container, Image: my-image:latest, Status: running
    ID: abcdef1234, Name: /another-container, Image: another-image:latest, Status: exited
    ```
  - **JSON Format**:
    ```json
    [
      {
        "id": "1234567890",
        "names": ["/my-container"],
        "image": "my-image:latest",
        "status": "running"
      },
      {
        "id": "abcdef1234",
        "names": ["/another-container"],
        "image": "another-image:latest",
        "status": "exited"
      }
    ]
    ```

### Get Container Details Endpoint

- **URL**: `/container/{containerID}`
- **Method**: `GET`
- **Description**: Retrieves details of a specific Docker container based on query parameters.
- **Query Parameters**:

  - `status`: Returns the status of the container.
  - `running`: Returns whether the container is running.
  - `paused`: Returns whether the container is paused.
  - `restarting`: Returns whether the container is restarting.
  - `dead`: Returns whether the container is dead.
  - `error`: Returns any error associated with the container.
  - `exitcode`: Returns the exit code of the container.
  - `state`: Returns the entire state object of the container.
  - **Response**:

    - **Example Request**:

      ```sh
      curl http://localhost:3333/container/1234567890?status=true&running=true
      ```

    - **Example Response**:
      ```json
      {
        "Running": true,
        "Status": "running"
      }
      ```

### Container Health Check Endpoint

- **URL**: `/container/health/{containerID}`
- **Method**: `GET`
- **Description**: Returns the health status of a specific Docker container.
- **Response**: Plain text message indicating the health status of the container.

  - **Example Request**:

    ```sh
    curl http://localhost:3333/container/health/1234567890
    ```

  - **Example Response**:
    ```plaintext
    "Container 1234567890 is healthy and running."
    ```

## License

This project is licensed under the GNU General Public License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Contact

For any questions or inquiries, please contact [selby@serubii.com](mailto:selby@serubii.com).
