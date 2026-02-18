package main

import (
	"bufio"
	"fmt"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func main() {
	supportedCommands := map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display a help message",
			callback:    commandHelp,
		},
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		line := scanner.Text()
		cleanedText := cleanInput(line)
		if len(cleanedText) == 0 {
			continue
		}
		_, ok := supportedCommands[cleanedText[0]]
		if ok {
			err := supportedCommands[cleanedText[0]].callback()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}
}

func commandExit() error {
	fmt.Printf("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}
