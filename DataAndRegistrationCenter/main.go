package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

var listenAddress = flag.String("l", "0.0.0.0:23333", "Listen address.")

func main() {
	flag.Parse()
	os.Mkdir("file/", os.ModePerm)
	initFileServer()
	initHashServer()
	initFileListServer()
	initRegistrationServer()
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("main: Listen failed: %s.\n", err)
	}
}

func IsFileOrDir(path string) int {
	statu, err := os.Stat(path)
	if err != nil {
		return -1
	}
	if statu.IsDir() {
		return 1
	} else {
		return 0
	}
}
