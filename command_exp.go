package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LocationDetail struct {
	PokemonEncounters []struct {
		Pokemon nameUrl `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func commandExp(cfg *config, args ...string) error {

	if len(args) == 0 {
		fmt.Println("A location name must be provided")
		return nil
	}

	url := "https://pokeapi.co/api/v2/location-area/" + args[0]

	fmt.Printf("Exploring %s\n", args[0])

	if cacheExp(cfg, url) == nil {
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

	var locationDetail LocationDetail
	err = json.Unmarshal(storage, &locationDetail)
	if err != nil {
		return fmt.Errorf("Error decoding response: %w\n", err)
	}

	lenPoke := len(locationDetail.PokemonEncounters)

	for i := range lenPoke {
		fmt.Println(" - ", locationDetail.PokemonEncounters[i].Pokemon.Name)
	}

	return nil
}

func cacheExp(cfg *config, url string) error {
	got, ok := cfg.cache.Get(url)

	if !ok {
		return fmt.Errorf("Not found in cache\n")
	}
	
	var locationDetail LocationDetail
	err := json.Unmarshal(got.Val, &locationDetail)
	if err != nil {
		return fmt.Errorf("Error decoding cache data: %w\n", err)
	}

	lenPoke := len(locationDetail.PokemonEncounters)

	for i := range lenPoke {
		fmt.Println(" - ", locationDetail.PokemonEncounters[i].Pokemon.Name)
	}

	return nil
}