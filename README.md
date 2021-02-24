# simple-go-reverse-proxy

A simple HTTP reverse proxy written in Go which has basic support for websockets.

# Installation

1. Install golang [here](https://golang.org/doc/install).
2. Run the following command to install dependencies:
    ```
    go get github.com/gorilla/websocket
    ```
3. Run the following commands to clone the repo and build the proxy:
    ```
    git clone https://github.com/dli357/simple-go-reverse-proxy
    cd simple-go-reverse-proxy
    go build
    ```

# Running the Proxy
Run the following command to run the proxy after building the proxy:
```
./simple-go-reverse-proxy
  --port=<port> 
  --backend=<address:port>
  [--insert-header=<header> --insert-header-val=<header-value>]
  [--websocket-scheme=<scheme>]
```
 - `port`: The port to serve from.
 - `backend`: The backend URL to proxy requests to. Should use the `http` scheme.
 - `insert-header`: An optional header to inject into all requests.
 - `insert-header-val`: An optional header value for the `insert-header` header. Defaults to `test-value` if `insert-header` is set but `insert-header-val` is not.
 - `websocket-scheme`: Used to specify the scheme for the proxy to use when proxying a websocket connection to the backend. Default is `ws`.

 # Example usage:
 The following command proxies all HTTP and websocket requests to localhost:8002 to localhost:8000. Additionally, it injects an `Authorization` header into every request as well.
 ```
 ./simple-go-reverse-proxy --backend=http://localhost:8000 --port=8002 --insert-header=Authorization --insert-header-val="Bearer test-value"
 ```