package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Configuration struct {
	Servers   []string
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

	var serverNumber int

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if configuration.Algorithm == "RoundRobin" {
			serverNumber = (serverNumber + 1) % len(configuration.Servers)
		} else if configuration.Algorithm == "Random" {
			serverNumber = rand.Intn(len(configuration.Servers))
		}

		targetUrl, _ := url.Parse(configuration.Servers[serverNumber])
		httputil.NewSingleHostReverseProxy(targetUrl).ServeHTTP(w, r)
	})

	http.ListenAndServe(":4000", nil)
}
