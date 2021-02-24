/*
Copyright 2021 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Simple go proxy which supports websockets and inserts a specified header and value. Usage:
//    ./simple-go-proxy --backend=<address:port> --port=<port> [--insert-header=<header> --insert-header-val=<header-value>] [--websocket-scheme=<scheme>]
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/websocket"
)

var (
	port            = flag.Int("port", 0, "Port on which to listen")
	backend         = flag.String("backend", "", "URL of the backend HTTP server to proxy")
	insertHeader    = flag.String("insert-header", "", "The header to inject into all requests")
	insertHeaderVal = flag.String("insert-header-val", "test-value", "The value to inject for insert-header")
	websocketScheme = flag.String("websocket-scheme", "ws", "The scheme to use for opening backend websocket connections. Default is `ws`.")
)

func main() {
	flag.Parse()

	if *backend == "" {
		log.Fatal("You must specify the address of the backend server to proxy")
	}
	if *port == 0 {
		log.Fatal("You must specify a local port number on which to listen")
	}
	backendURL, err := url.Parse(*backend)
	if err != nil {
		log.Fatalf("Failure parsing the address of the backend server: %v", err)
	}
	backendProxy := httputil.NewSingleHostReverseProxy(backendURL)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			newHeader := http.Header{}
			if *insertHeader != "" {
				newHeader.Add(*insertHeader, *insertHeaderVal)
			}
			backendWebsocketURL := *backendURL
			backendWebsocketURL.Scheme = *websocketScheme
			backendConn, _, err := websocket.DefaultDialer.Dial(backendWebsocketURL.String(), newHeader)
			if err != nil {
				log.Printf("Error opening websocket connection for request %v: %v", r.Host, err)
				w.WriteHeader(500)
				w.Write([]byte("error opening websocket connection"))
				return
			}
			upgrader := websocket.Upgrader{}
			clientConn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("Error upgrading websocket connection for request %v: %v", r.Host, err)
				w.WriteHeader(500)
				w.Write([]byte("error upgrading websocket connection"))
				return
			}
			log.Printf("Opened backend connection to %v", backendWebsocketURL.String())
			go proxyWebSocketMessagesOneWay(clientConn, backendConn)
			go proxyWebSocketMessagesOneWay(backendConn, clientConn)
			return
		}
		if *insertHeader != "" {
			r.Header.Add(*insertHeader, *insertHeaderVal)
		}
		backendProxy.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil))
}

func proxyWebSocketMessagesOneWay(client, server *websocket.Conn) {
	for {
		messageType, reader, err := client.NextReader()
		if err != nil {
			log.Printf("err while getting next reader for %q: %v", client.RemoteAddr().String(), err)
			return
		}
		writer, err := server.NextWriter(messageType)
		if err != nil {
			log.Printf("err while getting next writer for %q: %v", server.RemoteAddr().String(), err)
			return
		}
		msg, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Printf("err while reading from %q: %v", client.RemoteAddr().String(), err)
			return
		}
		_, err = writer.Write(msg)
		if err != nil {
			log.Printf("err while writing to %q: %v", client.RemoteAddr().String(), err)
			return
		}
	}
}
