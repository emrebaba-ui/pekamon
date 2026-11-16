package main

import (
	"time"
	"net/http"
	"github.com/emrebaba-ui/pokedexcli/internal"
)

func main() {
	c := internal.NewCache(5 * time.Minute)
	p := make(map[string]Pokemon)
	cl := &http.Client{Timeout: 3 * time.Second}
	
	cfg := &config{
		commands: getCommands(),
		client: cl,
		cache: c,
		next: "https://pokeapi.co/api/v2/location-area",
		previous: "",
		pokedex: p,
	}
	startRepl(cfg)
}

