package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"github.com/emrebaba-ui/pokedexcli/internal"
)

type config struct {
	commands map[string]cliCommand
	client *http.Client
	cache *internal.Cache
	next string
	previous string
	pokedex map[string]Pokemon
}

type cliCommand struct {
	name		string
	description	string
	callback func(*config, ...string) error	
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        	"help",
			description: 	"Displays a help message",
			callback:    	commandHelp,
		},
		"exit": {
			name:			"exit",
			description: 	"Exit the Pokedex",
			callback:    	commandExit,
		},
		"map": {
			name:			"map",
			description:	"Displays the location names",
			callback:		commandMap,
		},
		"mapb": {
			name:			"mapb",
			description:	"Displays the previous location names",
			callback:		commandMapb,
		},
		"explore": {
			name:			"explore",
			description:	"Displays the names of all the Pokémon",
			callback:		commandExp,
		},
		"catch": {
			name:			"catch",
			description:	"Adds Pokemon into the User's Pokedex",
			callback:		commandCatch,
		},
		"inspect": {
			name:			"inspect",
			description:	"Checks if a Pokemon is in Pokedex",
			callback:		commandIns,
		},
		"pokedex": {
			name:			"pokedex",
			description:	"Lists all the Pokémon in the Pokedex",
			callback:		commandPok,
		},
	}
}

func cleanInput(text string) []string {
	lowerText := strings.ToLower(text)
	words := strings.Fields(lowerText)
	return words
}

func startRepl(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner == nil {
		return
	}

	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			userInput := scanner.Text()
			words := cleanInput(userInput)
			if len(words) == 0 {
				continue
			}
			cmdName := words[0]

			if command, ok := cfg.commands[cmdName]; ok {
				err := command.callback(cfg, words[1:]...)
				if err != nil {
					fmt.Println("Error callback:", err)
					return
				}
			} else {
				fmt.Println("Unknown command")
				continue
			}
		} else {
			return
		}
	}
}

