package main

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"io"
)


func commandMapb(cfg *config, _ ...string) error {
	if cacheMapb(cfg) == nil {
		return nil
	}

	req, err := http.NewRequest("GET", cfg.previous, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Do(req)
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

func cacheMapb(cfg *config) error {
	got, ok := cfg.cache.Get(cfg.previous)
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