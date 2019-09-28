package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
)

type Configuration struct {
	Port          string
	Servers       []string
	Algorithm     string
	ServerWeights []int
}

type Server struct {
	value    int
	priority int
	index    int
}

type ServerConnnectionsCount struct {
	cc []int
	l  sync.Mutex
}

// Inc increments the counter for the given index.
func (scc *ServerConnnectionsCount) Inc(idx int) {
	scc.l.Lock()
	defer scc.l.Unlock()

	scc.cc[idx]++
}

// Dec decrements the counter for the given index.
func (scc *ServerConnnectionsCount) Dec(idx int) {
	scc.l.Lock()
	defer scc.l.Unlock()

	scc.cc[idx]--
}

// Value returns the current value of the counter for the given key.
func (scc *ServerConnnectionsCount) Value(idx int) int {
	scc.l.Lock()
	defer scc.l.Unlock()

	return scc.cc[idx]
}

func main() {
	// Load configuration
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
	serverRequestNumber := 0

	var mutex = &sync.Mutex{}

	var cc = make([]int, len(configuration.Servers))
	var l = sync.Mutex{}
	scc := ServerConnnectionsCount{
		cc,
		l,
	}

	a := 0
	b := 0
	c := 0
	total := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if configuration.Algorithm == "RoundRobin" {
			mutex.Lock()
			serverNumber = (serverNumber + 1) % len(configuration.Servers)
			mutex.Unlock()
		} else if configuration.Algorithm == "Random" {
			mutex.Lock()
			serverNumber = rand.Intn(len(configuration.Servers))
			mutex.Unlock()
		} else if configuration.Algorithm == "WeightedRoundRobin" {
			mutex.Lock()
			if serverRequestNumber < configuration.ServerWeights[serverNumber] {
				serverRequestNumber++
			} else {
				serverNumber = (serverNumber + 1) % len(configuration.Servers)
				serverRequestNumber = 1
			}
			mutex.Unlock()
		} else if configuration.Algorithm == "LeastConnections" {
			leastConnectionsCount := math.MaxInt32

			for i := 0; i < len(configuration.Servers); i++ {
				count := scc.Value(i)
				if count <= leastConnectionsCount {
					leastConnectionsCount = count
					mutex.Lock()
					serverNumber = i
					mutex.Unlock()
				}
			}

			mutex.Lock()
			scc.Inc(serverNumber) // Increment connection count for server
			mutex.Unlock()

		}
		mutex.Lock()
		targetUrl, err := url.Parse(configuration.Servers[serverNumber])
		if serverNumber == 0 {
			a++
		} else if serverNumber == 1 {
			b++
		} else if serverNumber == 2 {
			c++
		}
		total++
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}

		// Send the request to the selected server
		httputil.NewSingleHostReverseProxy(targetUrl).ServeHTTP(w, r)

		if configuration.Algorithm == "LeastConnections" {
			mutex.Lock()
			scc.Dec(serverNumber) // Decrement connection count for server
			mutex.Unlock()
		}

		mutex.Lock()
		fmt.Println(a, b, c, total)
		fmt.Println(scc.cc)
		mutex.Unlock()
	})

	http.ListenAndServe(":"+configuration.Port, nil)
}
