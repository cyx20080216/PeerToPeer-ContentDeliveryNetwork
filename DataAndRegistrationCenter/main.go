package main

import (
	"flag"
	"log"
	"net/http"
)

var listenAddress = flag.String("l", "0.0.0.0:23333", "Listen address.")

func main() {
	flag.Parse()
	initFileServer()
	initHashServer()
	initFileListServer()
	initRegistrationServer()
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("main: Listen failed: %s.\n", err)
	}
}
