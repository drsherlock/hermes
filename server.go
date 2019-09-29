package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/drsherlock/hermes/configuration"
)

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
	cnf := configuration.Load()

	servers := cnf.Servers

	fmt.Println(servers)

	serverNumber := 0
	serverRequestNumber := 0

	var mutex = &sync.Mutex{}

	var cc = make([]int, len(servers))
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
		switch cnf.Algorithm {
		case "RoundRobin":
			mutex.Lock()
			serverNumber = (serverNumber + 1) % len(servers)
			mutex.Unlock()
		case "Random":
			mutex.Lock()
			serverNumber = rand.Intn(len(servers))
			mutex.Unlock()
		case "WeightedRoundRobin":
			mutex.Lock()
			if serverRequestNumber < cnf.ServerWeights[serverNumber] {
				serverRequestNumber++
			} else {
				serverNumber = (serverNumber + 1) % len(servers)
				serverRequestNumber = 1
			}
			mutex.Unlock()
		case "LeastConnections":
			leastConnectionsCount := math.MaxInt32

			for i := 0; i < len(servers); i++ {
				connectionsCount := scc.Value(i)
				if connectionsCount <= leastConnectionsCount {
					leastConnectionsCount = connectionsCount
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
		targetUrl, err := url.Parse(servers[serverNumber])
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
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(w, r)

		if cnf.Algorithm == "LeastConnections" {
			mutex.Lock()
			scc.Dec(serverNumber) // Decrement connection count for server
			mutex.Unlock()
		}

		mutex.Lock()
		fmt.Println(a, b, c, total)
		fmt.Println(scc.cc)
		mutex.Unlock()
	})

	http.ListenAndServe(":"+cnf.Port, nil)
}
