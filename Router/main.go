package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var listenAddress = flag.String("l", "0.0.0.0:23333", "Listen address.")
var serverRawURL = flag.String("s", "http://127.0.0.1:23333/", "Server URL.")
var ServerURL *url.URL
var Rander = rand.New(rand.NewSource(time.Now().UnixMilli()))

func main() {
	flag.Parse()
	initServerURL()
	InitRouteServer()
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("main: Listen failed: %s.\n", err)
	}
}

func initServerURL() {
	var err error
	ServerURL, err = url.Parse(*serverRawURL)
	if err != nil {
		log.Fatalf("main: Parse url failed: %s.\n", err)
	}
	if ServerURL.Scheme != "http" && ServerURL.Scheme != "https" {
		log.Fatalf("main: Wrong protocol!\n")
	}
	if len(ServerURL.Fragment) > 0 {
		log.Fatalf("main: Has fragment!\n")
	}
	if len(ServerURL.Path) == 0 || ServerURL.Path[len(ServerURL.Path)-1] != '/' {
		ServerURL.Path += "/"
	}
}
