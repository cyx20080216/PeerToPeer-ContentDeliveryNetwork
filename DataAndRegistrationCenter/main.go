package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var listenAddress = flag.String("l", "0.0.0.0:23333", "Listen address.")

func main() {
	flag.Parse()
	os.Mkdir("file/", os.ModePerm)
	GetFileListAndDirList("file/")
	initFileServer()
	initHashServer()
	initFileListServer()
	initRegistrationServer()
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatalf("main: Listen failed: %s.\n", err)
	}
}

func GetFileListAndDirList(dir string) (fileList []string, dirList []string) {
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			unixPath := strings.Replace(path, "\\", "/", -1)
			dirList = append(dirList, unixPath[len(dir):])
		} else {
			unixPath := strings.Replace(path, "\\", "/", -1)
			fileList = append(fileList, unixPath[len(dir):])
		}
		return nil
	})
	return
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
