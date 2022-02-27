package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func InitRouteServer() {
	http.HandleFunc("/", responseFile)
}

func responseFile(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("hashServer: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		log.Println("hashServer: Wrong method! Response statu set as 404.")
		http.NotFound(responseWriter, request)
		return
	}
	path := request.URL.Path
	if len(path) < 2 {
		log.Println("hashServer: Wrong path! Response statu set as 404.")
		http.NotFound(responseWriter, request)
		return
	}
	path = path[1:]
	node, err := chooseNode(request.URL.Path)
	if err != nil {
		log.Printf("hashServer: Choose node failed: %s. Response statu set as 404.\n", err)
		http.NotFound(responseWriter, request)
		return
	}
	fileURL, err := url.Parse(node)
	if err != nil {
		log.Printf("hashServer: Parse node url failed: %s. Response statu set as 404.\n", err)
		http.NotFound(responseWriter, request)
		return
	}
	fileURL.Path += "file/" + path
	http.Redirect(responseWriter, request, fileURL.String(), http.StatusFound)
	log.Printf("hashServer: Redirect to %s finished.", fileURL.String())
}

func chooseNode(path string) (string, error) {
	sha1HashValue, err := getSHA1HashValue(path)
	if err != nil {
		return "", err
	}
	var nodeList []string
	nodeList, err = getNodeList()
	if err != nil {
		return "", err
	}
	nodeList = append(nodeList, ServerURL.String())
	for len(nodeList) > 0 {
		index := Rander.Intn(len(nodeList))
		if checkNode(nodeList[index], path, sha1HashValue) {
			return nodeList[index], nil
		} else {
			nodeList = append(nodeList[:index], nodeList[index+1:]...)
		}
	}
	return "", fmt.Errorf("can not find a right node")
}

func getSHA1HashValue(path string) (string, error) {
	hashURL := *ServerURL
	hashURL.Path += "hash/" + path
	res, err := http.Get(hashURL.String())
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("the status code is %d", res.StatusCode)
	}
	var content []byte
	content, err = io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getNodeList() ([]string, error) {
	nodeListURL := *ServerURL
	nodeListURL.Path += "nodelist"
	res, err := http.Get(nodeListURL.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("the status code is %d", res.StatusCode)
	}
	var content []byte
	content, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var nodeList []string
	err = json.Unmarshal(content, &nodeList)
	if err != nil {
		return nil, err
	}
	// var nodeSet = make(map[string]byte)
	// for _, nodeURL := range nodeList {
	// 	nodeSet[nodeURL] = byte(0)
	// }
	return nodeList, nil
}

func checkNode(nodeURL string, path string, sha1HashValue string) bool {
	hashURL, _ := url.Parse(nodeURL)
	hashURL.Path += "hash/" + path
	res, err := http.Get(hashURL.String())
	if err != nil {
		return false
	}
	if res.StatusCode != 200 {
		return false
	}
	var content []byte
	content, err = io.ReadAll(res.Body)
	if err != nil {
		return false
	}
	if string(content) != sha1HashValue {
		return false
	}
	return true
}
