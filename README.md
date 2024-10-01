# Conscript

Conscript is a Go-based HTTP server that provides various read only endpoints to interact with Docker containers. It includes endpoints for health checks, listing containers, and retrieving container details based on query parameters.

## Features

- **Health Check**: Endpoint to check the health of the server.
- **List Containers**: Endpoint to list all Docker containers.
- **Get Container Details**: Endpoint to get details of a specific Docker container based on query parameters.

## Endpoints

### Root Endpoint

- **URL**: `/`
- **Method**: `GET`
- **Description**: Returns a welcome message.
- **Response**:
  ```json
  Conscript alpha
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
- **Description**: Lists all Docker containers.
- **Response**: JSON array of container details.

### Get Container Details Endpoint

- **URL**: `/containerById/{containerID}`
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
- **Response**: JSON object with the requested container details.

## Running the Application

### Prerequisites

- Go 1.23.1 or later
- Docker

### Building and Running Locally

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

4. **Access the application**:
   Open your browser and navigate to `http://localhost:3333`.

### Running with Docker

1. **Build the Docker image**:
   ```sh
   docker build -t conscript:latest .
   ```

2. **Run the Docker container**:
   ```sh
   docker run -p 3333:3333 -v /var/run/docker.sock:/var/run/docker.sock conscript:latest
   ```

3. **Access the application**:
   Open your browser and navigate to `http://localhost:3333`.


## License

This project is licensed under the GNU General Public License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Contact

For any questions or inquiries, please contact [selby@serubii.com](mailto:selby@serubii.com).