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

// Simple backend for testing the simple-go-proxy which echos all data sent to it. Usage:
//    ./test-backend --port=<port>
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	port = flag.Int("port", 0, "Port on which to listen")
)

func main() {
	flag.Parse()

	if *port == 0 {
		log.Fatal("You must specify a local port number on which to listen")
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request Receieved: %+v", r)
		if websocket.IsWebSocketUpgrade(r) {
			upgrader := websocket.Upgrader{}
			clientConn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("Error opening websocket connection for request %v: %v", r.Host, err)
				w.WriteHeader(500)
				w.Write([]byte("error upgrading websocket connection"))
				return
			}
			go echoWebSocketMessages(clientConn)
			return
		}
		if r.Body != nil {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body for request %v: %v", r.Host, err)
				w.WriteHeader(500)
				w.Write([]byte("error reading body"))
				return
			}
			w.Write(data)
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *port), nil))
}

func echoWebSocketMessages(conn *websocket.Conn) {
	for {
		messageType, reader, err := conn.NextReader()
		if err != nil {
			log.Printf("err while getting next reader for %q: %v", conn.RemoteAddr().String(), err)
			return
		}
		writer, err := conn.NextWriter(messageType)
		if err != nil {
			log.Printf("err while getting next writer for %q: %v", conn.RemoteAddr().String(), err)
			return
		}
		msg, err := ioutil.ReadAll(reader)
		log.Printf("Echo websocket message: %q", string(msg))
		if err != nil {
			log.Printf("err while reading from %q: %v", conn.RemoteAddr().String(), err)
			return
		}
		_, err = writer.Write(msg)
		if err != nil {
			log.Printf("err while writing to %q: %v", conn.RemoteAddr().String(), err)
			return
		}
	}
}
