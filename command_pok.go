package main

import (
	"fmt"
)

func commandPok(cfg *config, _ ...string) error {
	
	if len(cfg.pokedex) == 0 {
		fmt.Println("Dein Pokedex ist vollig leer")
	} else {
		for _, pokemon := range cfg.pokedex {
			fmt.Println(" - ", pokemon.Name)
		}
	}

	return nil
}