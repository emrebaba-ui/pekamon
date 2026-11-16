package main

import (
	"fmt"
)

func commandIns(cfg *config, args ...string) error {
	
	if _, ok := cfg.pokedex[args[0]]; ok {
		fmt.Printf("%s was caught!\n", args[0])
	} else {
		fmt.Println("you have not caught that pokemon")
	}

	return nil
}