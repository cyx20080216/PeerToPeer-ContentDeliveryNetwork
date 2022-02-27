package main

import (
	"net/http"
)

func initFileServer() {
	http.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("file/"))))
}
