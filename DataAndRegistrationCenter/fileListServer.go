package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var fileSet = make(map[string]byte)
var fileSetLock sync.RWMutex

func initFileListServer() {
	http.HandleFunc("/filelist", responseFileList)
	getFileList()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("fileListServer: Create the watcher failed: %s.\n", err)
		return
	}
	watcher.Add("file/")
	go processFileListAndFileSystemNotify(watcher)
}

func responseFileList(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("fileListServer: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		log.Println("fileListServer: Wrong method! Response statu set as 404.")
		http.NotFound(responseWriter, request)
		return
	}
	var fileList = make([]string, 0)
	fileSetLock.RLock()
	for file, _ := range fileSet {
		fileList = append(fileList, file)
	}
	fileSetLock.RUnlock()
	js, err := json.Marshal(fileList)
	if err != nil {
		log.Printf("fileListServer: Marshal file list failed: %s. Response statu set as 404.\n", err)
		http.NotFound(responseWriter, request)
		return
	}
	responseWriter.Write(js)
	log.Println("fileListServer: Finish sending the file list.")
}

func getFileList() {
	fileSetLock.Lock()
	filepath.Walk("file", func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			unixPath := strings.Replace(path, "\\", "/", -1)
			fileSet[unixPath[5:]] = byte(0)
		}
		return nil
	})
	fileSetLock.Unlock()
}

func processFileListAndFileSystemNotify(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			{
				unixPath := strings.Replace(event.Name, "\\", "/", -1)
				if (event.Op & fsnotify.Create) != 0 {
					if IsFileOrDir(unixPath) == 0 {
						fileSetLock.Lock()
						fileSet[unixPath[5:]] = byte(0)
						fileSetLock.Unlock()
						log.Printf("fileListServer: Add %s in the set. Because it created.\n", unixPath)
					} else if IsFileOrDir(unixPath) == 1 {
						watcher.Add(unixPath)
					}
				} else if (event.Op&fsnotify.Remove) != 0 || (event.Op&fsnotify.Rename) != 0 {
					fileSetLock.Lock()
					delete(fileSet, unixPath[5:])
					fileSetLock.Unlock()
					log.Printf("fileListServer: Remove %s in the set. Because it removed.\n", unixPath)
				}
			}
		case err := <-watcher.Errors:
			log.Printf("fileListServer: Watcher has error: %s.\n", err)
		}
	}
}
