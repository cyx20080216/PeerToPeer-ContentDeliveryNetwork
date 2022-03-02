package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var fileSet = make(map[string]byte)
var fileSetLock sync.RWMutex

func initFileListServer() {
	http.HandleFunc("/filelist", responseFileList)
	fileList, dirList := GetFileListAndDirList("file/")
	for _, each := range fileList {
		fileSet[each] = byte(0)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("fileListServer: Create the watcher failed: %s.\n", err)
		return
	}
	watcher.Add("file/")
	for _, each := range dirList {
		watcher.Add("file/" + each)
	}
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

func processFileListAndFileSystemNotify(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			{
				// log.Println(event.Op, event.Name, len(event.Name))
				unixPath := strings.Replace(event.Name, "\\", "/", -1)
				if len(unixPath) >= 5 && unixPath[:5] == "file/" {
					if (event.Op & fsnotify.Create) != 0 {
						if IsFileOrDir(unixPath) == 0 {
							fileSetLock.Lock()
							fileSet[unixPath[5:]] = byte(0)
							fileSetLock.Unlock()
							log.Printf("fileListServer: Add %s in the set. Because it created.\n", unixPath)
						} else if IsFileOrDir(unixPath) == 1 {
							if unixPath[len(unixPath)-1] != '/' {
								unixPath += "/"
							}
							watcher.Add(unixPath)
							fileList, dirList := GetFileListAndDirList(unixPath)
							fileSetLock.Lock()
							for _, each := range fileList {
								fileSet[each[5:]] = byte(0)
							}
							fileSetLock.Unlock()
							for _, each := range dirList {
								watcher.Add(unixPath + each)
							}
						}
					}
					if (event.Op&fsnotify.Remove) != 0 || (event.Op&fsnotify.Rename) != 0 {
						fileSetLock.Lock()
						delete(fileSet, unixPath[5:])
						fileSetLock.Unlock()
						if unixPath[len(unixPath)-1] != '/' {
							unixPath += "/"
						}
						willRemove := make([]string, 0)
						fileSetLock.RLock()
						for each, _ := range fileSet {
							if len(each) >= len(unixPath[5:]) && each[:len(unixPath[5:])] == unixPath[5:] {
								willRemove = append(willRemove, each)
							}
						}
						fileSetLock.RUnlock()
						fileSetLock.Lock()
						for _, each := range willRemove {
							delete(fileSet, each)
						}
						fileSetLock.Unlock()
						watcher.Remove(unixPath)
						log.Printf("fileListServer: Remove %s in the set. Because it removed.\n", unixPath)
					}
				}
			}
		case err := <-watcher.Errors:
			log.Printf("fileListServer: Watcher has error: %s.\n", err)
		}
	}
}
