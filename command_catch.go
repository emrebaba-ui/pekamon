package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

type Pokemon struct {
	Name string `json:"name"`
	BaseExperience int `json:"base_experience"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	// the rest was not there
}

func commandCatch(cfg *config, args ...string) error {

	if len(args) == 0 {
		fmt.Println("A Pokemon name must be provided")
		return nil
	}

	url := "https://pokeapi.co/api/v2/pokemon/" + args[0]

	fmt.Printf("Throwing a Pokeball at %s...\n", args[0])

	if cacheCatch(cfg, url) == nil {
		return nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %w\n", err)
	}

	res, err := cfg.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error getting response: %w\n", err)
	}
	defer res.Body.Close()

	storage, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Error storing response: %w\n", err)
	}
	cfg.cache.Add(url, storage)

	var pokemon Pokemon
	err = json.Unmarshal(storage, &pokemon)
	if err != nil {
		return fmt.Errorf("Error decoding response: %w\n", err)
	}

	nicerDicer(cfg, pokemon)

	return nil
}

func cacheCatch(cfg *config, url string) error {
	got, ok := cfg.cache.Get(url)
	if !ok {
		return fmt.Errorf("Not found in the cache\n")
	}

	var pokemon Pokemon
	err := json.Unmarshal(got.Val, &pokemon)
	if err != nil {
		return fmt.Errorf("Error decoding cache data: %w\n", err)
	}

	nicerDicer(cfg, pokemon)

	fmt.Println("*cache was used*")
	return nil
}

func nicerDicer (cfg *config, pokemon Pokemon) bool {
	success := false
	baseNumber := rand.Float64()
	threshold := 1 - 10 / float64(pokemon.BaseExperience)
	if baseNumber > threshold {
		success = true
	}

	if success {
		cfg.pokedex[pokemon.Name] = pokemon
		fmt.Printf("%s was caught!\n", pokemon.Name)
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}

	return success
}