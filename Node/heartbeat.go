package main

import (
	"log"
	"net/http"
)

func initHeartbeat() {
	http.HandleFunc("/heartbeat", func(responseWriter http.ResponseWriter, request *http.Request) {
		log.Printf("heartbeat: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			log.Println("heartbeat: Wrong method! Response statu set as 404.")
			http.NotFound(responseWriter, request)
			return
		}
	})
}
