# Go OpenAI Proxy Server

## Overview
This project serves as a primitive proxy server for OpenAI-compatible API requests, including HTTP Streaming. The server listens on a specified interface and port, and handles incoming HTTP requests and redirects them to an OpenAI compatible API.

The server also includes a request interceptor mechanism, which allows you to modify the request data before it is sent to the OpenAI API. Currently, it includes a Google Search interceptor as an example.

## Features
- HTTP/HTTPS server using Go's standard `net/http` package
- Request interceptors for modifying request data
- Logging using Uber's Zap logging library
- Configurable listening interface and port via command-line flags

## Requirements
- Go 1.x
- Uber's Zap logging library

## Installation
1. Clone the repository:

    ```bash
    git clone https://github.com/your-repo/go-openai-proxy.git
    ```

2. Navigate to the project directory:

    ```bash
    cd go-openai-proxy
    ```

3. Build the project:

    ```bash
    go build
    ```

## Usage
Run the server with the default settings:

```bash
./go-openai-proxy
```

Or specify the interface and port:
```
./go-openai-proxy --iface 192.168.1.1
```

#### API Endpoints
POST /: Main endpoint for handling OpenAI API requests

#### Interceptors
You can add your own request interceptors by modifying the interceptors slice in the internal package.

#### Logging
Logs are generated using Uber's Zap logging library and are written to stdout.

## License
This project is licensed under the MIT License - see the LICENSE.md file for details.

