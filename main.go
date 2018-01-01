package main

import (
  "net/http"
  "time"
  "fmt"
  "io/ioutil"
  "os"
)

func isDirectory(path string) (bool, error) {
  fileInfo, err := os.Stat(path) // get info from path, including type
  if err != nil {
    return false, err
  }
  return fileInfo.IsDir(), nil // true if path is a directory
}

func loadPage(path string) ([]byte, error) {
  isDir, err := isDirectory(path)
  if err != nil {
    return nil, err
  }
  if(isDir){
    path += "/index.html" // default page to render, i apologize for the lack of error handling
  }
  body, err := ioutil.ReadFile(path) // read byte stream from file
  if err != nil {
    return nil, err
  }
  return body, nil
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("." + r.URL.Path)
  page, err := loadPage("." + r.URL.Path)
  if err != nil {
    fmt.Fprintf(w, "An error occurred")
  } else {
    fmt.Fprintf(w, string(page)) // convert byte stream to string and return it
  }
}

func makeTimestamp() uint {
  return uint(time.Now().UnixNano() / int64(time.Millisecond))
}

func main() {
  hub := newHub()
  http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("websocket request")
    serveWs(hub, w, r)
  })
  http.HandleFunc("/", requestHandler) // handler for all requests
  http.ListenAndServe(":8080", nil) // default port
}
