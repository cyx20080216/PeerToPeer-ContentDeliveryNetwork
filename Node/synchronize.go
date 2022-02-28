package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func Synchronize() {
	fileList, err := getFileList()
	if err != nil {
		log.Printf("Synchronize: Failed to get file list: %s.\n", err)
		return
	}
	for _, file := range fileList {
		go checkAndSynchronize(file)
	}
}

func checkAndSynchronize(path string) {
	sha1HashValue, err := getSHA1HashValue(path)
	if err != nil {
		log.Printf("Synchronize: Failed to get SHA1 hash value of %s: %s.\n", path, err)
		return
	}
	var file *os.File
	file, err = os.OpenFile("file/" + path, os.O_RDONLY, 0)
	if err == nil {
		var localValue string
		localValue, err = calcSHA1HashValue(file)
		if err != nil {
			log.Printf("Synchronize: Failed to get SHA1 hash value of local file %s: %s.\n", path, err)
			return
		}
		if localValue == sha1HashValue {
			return
		}
	}
	file.Close()
	err = synchronizeFile(path, sha1HashValue)
	if err != nil {
		log.Printf("Synchronize: Failed to update file %s: %s.\n", path, err)
		return
	} else {
		log.Printf("Synchronize: Update file %s finished.\n", path)
		return
	}
}

func synchronizeFile(path string, sha1HashValue string) error {
	nodeList, err := getNodeList()
	if err != nil {
		return err
	}
	nodeList = append(nodeList, ServerURL.String())
	for len(nodeList) > 0 {
		index := Rander.Intn(len(nodeList))
		if checkNode(nodeList[index], path, sha1HashValue) {
			file, err := os.OpenFile("file/" + path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)
			if err != nil {
				return err
			}
			fileURL, _ := url.Parse(nodeList[index])
			fileURL.Path += "file/" + path
			err = downloadFile(file, fileURL.String())
			if err != nil {
				nodeList = append(nodeList[:index], nodeList[index+1:]...)
				continue
			}
			file.Close()
			file, err = os.OpenFile("file/" + path, os.O_RDONLY, 0)
			if err != nil {
				return err
			}
			var fileSHA1HashValue string
			fileSHA1HashValue, err = calcSHA1HashValue(file)
			if err != nil {
				return err
			}
			file.Close()
			if fileSHA1HashValue == sha1HashValue {
				return nil
			}
			nodeList = append(nodeList[:index], nodeList[index+1:]...)
		} else {
			nodeList = append(nodeList[:index], nodeList[index+1:]...)
		}
	}
	return fmt.Errorf("can not find a right node")
}

func getFileList() ([]string, error) {
	nodeListURL := *ServerURL
	nodeListURL.Path += "filelist"
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
	return nodeList, nil
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

func calcSHA1HashValue(src io.Reader) (string, error) {
	sha1Hasher := sha1.New()
	_, err := io.Copy(sha1Hasher, src)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1Hasher.Sum(nil)), nil
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

func downloadFile(dst io.Writer, url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("the status code is %d", res.StatusCode)
	}
	_, err = io.Copy(dst, res.Body)
	return err
}
