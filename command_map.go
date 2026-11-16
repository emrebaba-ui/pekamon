package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io"
)

type nameUrl struct {
	Name string
	Url string
}

type location struct {
	Count int
	Next *string
	Previous *string
	Results []nameUrl
}

func commandMap(cfg *config, _ ...string) error {

	if cacheMap(cfg) == nil {
		return nil
	}

	req, err := http.NewRequest("GET", cfg.next, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}

	res, err := cfg.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error getting response: %w", err)
	}
	defer res.Body.Close()

	storage, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error storing cache:", err)
	} else {
		cfg.cache.Add(cfg.next, storage)
	}

	var curLocation location
	err = json.Unmarshal(storage, &curLocation)
	if err != nil {
		return fmt.Errorf("Error decoding response: %w", err)
	}

	if curLocation.Next != nil {
		cfg.next = *curLocation.Next}
	if curLocation.Previous != nil {
		cfg.previous = *curLocation.Previous}

	for i := range 20 {
		fmt.Println(curLocation.Results[i].Name)
	}

	return nil
}

func cacheMap(cfg *config) error {
	got, ok := cfg.cache.Get(cfg.next)
	if !ok {
		return fmt.Errorf("Not found in cache")
	} else {
		var curLocation location
		err := json.Unmarshal(got.Val, &curLocation)
		if err != nil {
			return fmt.Errorf("Error decoding cache data: %w", err)}
		if curLocation.Next != nil {
			cfg.next = *curLocation.Next}
		if curLocation.Previous != nil {
			cfg.previous = *curLocation.Previous}
		for i := range 20 {
			fmt.Println(curLocation.Results[i].Name)
		}
		return nil
	}
}