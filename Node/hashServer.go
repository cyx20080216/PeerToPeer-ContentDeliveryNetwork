package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var sha1HashValueCache = make(map[string]string)
var sha1HashValueCacheLock sync.RWMutex

func initHashServer() {
	http.HandleFunc("/hash/", responseHash)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("hashServer: Create the watcher failed: %s.\n", err)
		return
	}
	watcher.Add("file")
	_, dirList := GetFileListAndDirList("file/")
	for _, each := range dirList {
		watcher.Add("file/" + each)
	}
	go processSHA1HashValueCacheAndFileSystemNotify(watcher)
}

func responseHash(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("hashServer: Request from %s. Method is %s. Path is %s.\n", request.RemoteAddr, request.Method, request.URL.Path)
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		log.Println("hashServer: Wrong method! Response statu set as 404.")
		http.NotFound(responseWriter, request)
		return
	}
	sha1HashValueCacheLock.RLock()
	sha1HashValue, isPresent := sha1HashValueCache["file"+request.URL.Path[5:]]
	sha1HashValueCacheLock.RUnlock()
	if isPresent {
		fmt.Fprint(responseWriter, sha1HashValue)
		log.Printf("hashServer: Cache hit the file %s. Send the SHA-1 hash value %s.\n", request.URL.Path[6:], sha1HashValue)
		return
	}
	sha1Hasher := sha1.New()
	fileHandle, err := os.OpenFile("file"+request.URL.Path[5:], os.O_RDONLY, 0)
	if err != nil {
		log.Printf("hashServer: Open the file %s failed: %s. Response statu set as 404.\n", request.URL.Path[6:], err)
		http.NotFound(responseWriter, request)
		return
	}
	_, err = io.Copy(sha1Hasher, fileHandle)
	if err != nil {
		log.Printf("hashServer: Read the file %s failed: %s. Response statu set as 404.\n", request.URL.Path[6:], err)
		fileHandle.Close()
		http.NotFound(responseWriter, request)
		return
	}
	fileHandle.Close()
	sha1HashValue = fmt.Sprintf("%x", sha1Hasher.Sum(nil))
	sha1HashValueCacheLock.Lock()
	sha1HashValueCache[request.URL.Path[6:]] = sha1HashValue
	sha1HashValueCacheLock.Unlock()
	responseWriter.Write([]byte(sha1HashValue))
	log.Printf("hashServer: Finish reading the file %s and sending the SHA-1 hash value %s.\n", request.URL.Path[6:], sha1HashValue)
}

func processSHA1HashValueCacheAndFileSystemNotify(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			{
				unixPath := strings.Replace(event.Name, "\\", "/", -1)
				if (event.Op&fsnotify.Write) != 0 || (event.Op&fsnotify.Remove) != 0 || (event.Op&fsnotify.Rename) != 0 {
					sha1HashValueCacheLock.Lock()
					delete(sha1HashValueCache, unixPath[5:])
					sha1HashValueCacheLock.Unlock()
					log.Printf("hashServer: Remove the cache of %s. Because it wrote or removed.\n", unixPath)
				} else if event.Op&fsnotify.Create != 0 {
					if IsFileOrDir(unixPath) == 1 {
						if unixPath[len(unixPath)-1] != '/' {
							unixPath += "/"
						}
						watcher.Add(unixPath)
						_, dirList := GetFileListAndDirList(unixPath)
						for _, each := range dirList {
							watcher.Add(unixPath + each)
						}
					}
				}
			}
		case err := <-watcher.Errors:
			log.Printf("hashServer: Watcher has error: %s.\n", err)
		}
	}
}
