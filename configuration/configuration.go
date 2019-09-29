package configuration

import (
	"encoding/json"
	"log"
	"os"
)


type Configuration struct {
	Port          string
	Servers       []string
	Algorithm     string
	ServerWeights []int
}

func Load() Configuration {
	cF, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer cF.Close()

  cFDecoder := json.NewDecoder(cF)
	cnf := Configuration{}
	err = cFDecoder.Decode(&cnf)
	if err != nil {
		log.Fatal(err)
	}

  return cnf
}
