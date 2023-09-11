# Go OpenAI Proxy Server

## Overview
The server listens on a specified interface and port, and handles incoming HTTP requests and redirects them to an OpenAI compatible API or Azure API based on the configuration. 

The server also includes a request interceptor mechanism, which allows you to modify the request data before it is sent to the OpenAI API. Currently, it includes a Google Search interceptor as an example.

## Features
- HTTP/HTTPS server using Go's standard `net/http` package
- Configurable listening interface, port, and upstreams via command-line flags or a YAML configuration file
- Conveniently log your requests to an OpenAI-compatible API using Uber's Zap logging library
- Request interceptors for modifying request data
- Support for multiple upstream types (Azure, OpenAI)

## Requirements
- Go 1.x

## Installation
```bash
git clone https://github.com/go-openai-proxy/go-openai-proxy.git
cd go-openai-proxy
go get -u
go build
```

## Usage
Run the server with the default settings:

```
bash
./go-openai-proxy
```

3. Build the project:

```
bash
go build
```

## Usage
Run the server with the default settings:

```
./go-openai-proxy
```

Or specify the configuration options:
```
./go-openai-proxy --config config.yaml --listeners 192.168.1.1:6001 --logLevel debug --certFile /path/to/cert.crt --keyFile /path/to/key.key --useTLS false
```

## Configuration File
You can also use a YAML configuration file to set up the server. Here's an example:
```
certFile: "/path/to/cert/file.crt"
keyFile: "/path/to/key/file.key"
logLevel: "debug"
listeners:
  - interface: "0.0.0.0"
    port: "6001"
upstreams:
  Primary:
    type: "azure"
    model: "default"
    url: "http://10.10.0.127:5001"
    priority: 2
  Secondary:
    type: "openai"
    model: "default"
    priority: 1

```

#### Example Output
Here's some example output you can get out of the logger:
```
$ go run cmd/main.go
Debug: Added config to context in main.go: &{map[Primary:{azure http://10.10.0.127:5001 2} Secondary:{openai  1}] [{0.0.0.0 6001}]}
{"level":"info","ts":1694269942.962039,"caller":"cmd/main.go:90","msg":"Hostname","hostname":"Administrators-MacBook-Pro.local"}
{"level":"info","ts":1694269942.962371,"caller":"cmd/main.go:100","msg":"Starting listener","address":"0.0.0.0:6001"}
Debug: Retrieved config from context in handlers.go: &{map[Primary:{azure http://10.10.0.127:5001 2} Secondary:{openai  1}] [{0.0.0.0 6001}]}, Type: *internal.Config
Debug: Retrieved config from context in handlers.go: &{map[Primary:{azure http://10.10.0.127:5001 2} Secondary:{openai  1}] [{0.0.0.0 6001}]}, Type: *internal.Config
{"level":"info","ts":1694269949.982666,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":""}
{"level":"info","ts":1694269949.983812,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":"Hello"}
{"level":"info","ts":1694269950.029989,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":"!"}
{"level":"info","ts":1694269950.205655,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" How"}
{"level":"info","ts":1694269950.206295,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" can"}
{"level":"info","ts":1694269950.2064052,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" I"}
{"level":"info","ts":1694269950.2065,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" assist"}
{"level":"info","ts":1694269950.209876,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" you"}
{"level":"info","ts":1694269950.250054,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":" today"}
{"level":"info","ts":1694269950.284432,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":"?"}
{"level":"info","ts":1694269950.3289962,"caller":"internal/handlers.go:90","msg":"JSON Response Segment","content":""}
{"level":"info","ts":1694269950.329086,"caller":"internal/handlers.go:107","msg":"JSON Completed Response","response":"{\"completedResponse\":\"Hello! How can I assist you today?\",\"requestMessages\":[{\"role\":\"user\",\"content\":\"Hello world\"}]}"}
```

## License
This project is licensed under the MIT License - see the LICENSE.md file for details.

