package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

var listenAddress = flag.String("l", "0.0.0.0:23333", "Listen address.")
var nodeRawURL = flag.String("n", "http://127.0.0.1:23333/", "Node URL.")
var serverRawURL = flag.String("s", "http://127.0.0.1:23333/", "Server URL.")
var synchronizeTime = flag.Int("t", 300, "Synchronize time.")
var ServerURL *url.URL
var Rander = rand.New(rand.NewSource(time.Now().UnixMilli()))

func main() {
	flag.Parse()
	os.Mkdir("file/", os.ModePerm)
	initServerURL()
	initFileServer()
	initHashServer()
	initHeartbeat()
	go func() {
		time.Sleep(1000000000)
		register()
	}()
	go func() {
		time.Sleep(1000000000)
		for {
			Synchronize()
			time.Sleep(time.Duration((*synchronizeTime) * 1000000000))
		}
	}()
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

func register() {
	serverRegisterURL := *ServerURL
	serverRegisterURL.Path += "register"
	serverRegisterRawURL := serverRegisterURL.String()
	var form = make(url.Values)
	form["url"] = append(form["url"], *nodeRawURL)
	res, err := http.PostForm(serverRegisterRawURL, form)
	if err != nil {
		log.Fatalf("main: Register failed: %s.\n", err)
	}
	if res.StatusCode != 200 {
		content, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("main: Read the content failed: %s. The status code %d is wrong.", err, res.StatusCode)
		}
		log.Printf("main: The status code %d is wrong. The content is %s.", res.StatusCode, string(content))
	}
	log.Println("main: Finish register")
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
