package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var nodeSet = make(map[string]byte)
var nodeSetLock sync.RWMutex

func initRegistrationServer() {
	http.HandleFunc("/register", responseRegister)
	http.HandleFunc("/nodelist", responseNodeList)
}

func responseRegister(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("registrationServer: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
	if request.Method != http.MethodPost {
		log.Printf("registrationServer: Wrong method! Response statu set as 404.\n")
		http.NotFound(responseWriter, request)
		return
	}
	err := request.ParseForm()
	if err != nil {
		log.Printf("registrationServer: Parse form failed: %s. Response statu set as 404.\n", err)
		http.NotFound(responseWriter, request)
		return
	}
	if !request.Form.Has("url") {
		log.Println("registrationServer: Key \"url\" not found in the form! Response statu set as 404 and send error.")
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte(`{"statu":"error","describe":"Key \"url\" not found in the form."}`))
	}
	var nodeURL *url.URL
	nodeURL, err = url.Parse(request.Form.Get("url"))
	if err != nil {
		log.Printf("registrationServer: Parse url failed: %s. Response statu set as 404 and send error.\n", err)
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte(`{"statu":"error","describe":"The value of the key \"url\" should be a right url."}`))
		return
	}
	if nodeURL.Scheme != "http" && nodeURL.Scheme != "https" {
		log.Println("registrationServer: Wrong protocol! Response statu set as 404 and send error.")
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte(`{"statu":"error","describe":"The protocol should be a \"http\" or \"https\"."}`))
		return
	}
	if len(nodeURL.Fragment) > 0 {
		log.Println("registrationServer: Has fragment! Response statu set as 404 and send error.")
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte(`{"statu":"error","describe":"Can not has fragment."}`))
		return
	}
	if len(nodeURL.Path) == 0 || nodeURL.Path[len(nodeURL.Path)-1] != '/' {
		nodeURL.Path += "/"
	}
	rawURL := nodeURL.String()
	nodeSetLock.RLock()
	_, isPresent := nodeSet[rawURL]
	nodeSetLock.RUnlock()
	if isPresent {
		log.Println("registrationServer: Is present! Response statu set as 404 and send error.")
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte(`{"statu":"error","describe":"Is present."}`))
		return
	}
	nodeSetLock.Lock()
	nodeSet[rawURL] = byte(0)
	nodeSetLock.Unlock()
	go checkHeartbeat(rawURL)
	responseWriter.Write([]byte(`{"statu":"ok"}`))
	log.Printf("registrationServer: Finish adding node %s", rawURL)
}
func checkHeartbeat(rawURL string) {
	heartbeatURL, _ := url.Parse(rawURL)
	heartbeatURL.Path += "heartbeat"
	heartbeatRawURL := heartbeatURL.String()
	for {
		res, err := http.Get(heartbeatRawURL)
		if err != nil {
			log.Printf("registrationServer: Check heartbeat of node %s failed: %s. Remove it.\n", rawURL, err)
			nodeSetLock.Lock()
			delete(nodeSet, rawURL)
			nodeSetLock.Unlock()
			return
		}
		if res.StatusCode != 200 {
			log.Printf("registrationServer: Get wrong status code %d when check heartbeat of node %s. Remove it.\n", res.StatusCode, rawURL)
			nodeSetLock.Lock()
			delete(nodeSet, rawURL)
			nodeSetLock.Unlock()
			return
		}
		time.Sleep(5000000000)
	}
}
func responseNodeList(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("registrationServer: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		log.Printf("registrationServer: Wrong method! Response statu set as 404.\n")
		http.NotFound(responseWriter, request)
		return
	}
	var nodeList = make([]string, 0)
	nodeSetLock.RLock()
	for file, _ := range nodeSet {
		nodeList = append(nodeList, file)
	}
	nodeSetLock.RUnlock()
	js, err := json.Marshal(nodeList)
	if err != nil {
		log.Printf("registrationServer: Marshal node list failed: %s. Response statu set as 404.\n", err)
		http.NotFound(responseWriter, request)
		return
	}
	responseWriter.Write(js)
	log.Println("registrationServer: Finish sending the node list.")
}
