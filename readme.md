# Gin Swagger Proxy
This project is a Gin-based HTTP proxy that serves Swagger documentation from an upstream server and provides a Swagger UI for API exploration.

## Features
- Fetches Swagger JSON from an upstream server.
- Serves the Swagger JSON on `/swagger`.
- Integrates Swagger UI available at `/docs/index.html`.
- Forwards incoming requests to an upstream server and returns the response.

## Prerequisites

- Ensure that Kubernetes is running.
- Ensure that `kubectl` is installed and configured to interact with your Kubernetes cluster.


## Installation
To install the required dependencies, run:

```shell
go get github.com/gin-gonic/gin
go get github.com/swaggo/files
go get github.com/swaggo/gin-swagger
```

## Usage
1. Clone the repository:
    ```shell
    git clone https://github.com/jobayer12/kubeSwag
    cd kubeSwag
    ```
2. Start the kubectl proxy:
    ```shell
    kubectl proxy --port=8000
    ````
3. Run the application:
    ```shell
    go run main.go
    ```
4. Access the Swagger UI:

    Open your web browser and navigate to http://localhost:8080/docs/index.html

# Endpoints
- `GET /swagger`: Returns the Swagger JSON.

- `GET /docs/*any`: Displays the Swagger UI.
- `/*any`: Forwards all other requests to the upstream server.


Feel free to contribute to this project by submitting issues or pull requests.