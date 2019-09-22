package main

import (
  "fmt"
  "net/http"
  "net/http/httputil"
  "net/url"
  "encoding/json"
  "os"
  "log"
)

type Configuration struct {
  Servers []string
  Algorithm string
}

func main() {
  configFile, err := os.Open("config.json")
  if err != nil {
      log.Fatal(err)
  }
  defer configFile.Close()

  decoder := json.NewDecoder(configFile)
  configuration := Configuration{}
  err = decoder.Decode(&configuration)
  if err != nil {
    log.Fatal(err)
  }

fmt.Println(configuration.Servers)

serverNumber := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    serverNumber = (serverNumber+1)%len(configuration.Servers)
    targetUrl, _ := url.Parse(configuration.Servers[serverNumber])
    httputil.NewSingleHostReverseProxy(targetUrl).ServeHTTP(w, r)
	})

	http.ListenAndServe(":4000", nil)
}
