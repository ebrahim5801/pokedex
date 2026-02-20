package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/ebrahim5801/pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(c *config, location string) error
}

type config struct {
	next     string
	previous string
	cache    *pokecache.Cache
	pokedex  map[string]Pokemon
}

type results struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []location `json:"results"`
}

type location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Pokemon struct {
	Name       string        `json:"name"`
	Experience int           `json:"base_experience"`
	Height     int           `json:"height"`
	Weight     int           `json:"weight"`
	Stats      []PokemonStat `json:"stats"`
	Types      []PokemonType `json:"types"`
}

type PokemonStat struct {
	BaseStat int      `json:"base_stat"`
	Stat     StatName `json:"stat"`
}

type StatName struct {
	Name string `json:"name"`
}

type PokemonType struct {
	Type TypeName `json:"type"`
}

type TypeName struct {
	Name string `json:"name"`
}

type PokemonEncounters struct {
	Pokemon Pokemon `json:"pokemon"`
}

type locationAreaResult struct {
	PokemonEncounters []PokemonEncounters `json:"pokemon_encounters"`
}

func main() {
	cfg := &config{
		next:    "https://pokeapi.co/api/v2/location-area",
		cache:   pokecache.NewCache(5 * time.Second),
		pokedex: map[string]Pokemon{},
	}

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
		"map": {
			name:        "map",
			description: "Display the next page of locations",
			callback:    commandMap,
		},
		"explore": {
			name:        "explore",
			description: "Display a list of Pokemon in a location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon",
			callback:    commandInspect,
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
		cmd, ok := supportedCommands[cleanedText[0]]
		if ok {
			location := ""
			if len(cleanedText) > 1 {
				location = cleanedText[1]
			}
			err := cmd.callback(cfg, location)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(c *config, location string) error {
	fmt.Printf("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *config, location string) error {
	fmt.Println(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex`)
	return nil
}

func commandMap(c *config, location string) error {
	if c.next == "" {
		return errors.New("you're on the last page")
	}

	if data, ok := c.cache.Get(c.next); ok {
		var res results
		if err := json.Unmarshal(data, &res); err != nil {
			return err
		}
		for _, loc := range res.Results {
			fmt.Println(loc.Name)
		}
		c.next = res.Next
		c.previous = res.Previous
		return nil
	}

	resp, err := http.Get(c.next)
	if err != nil {
		return errors.New("error fetching the data")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("error reading the data")
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	c.cache.Add(c.next, body)

	var res results
	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}
	for _, loc := range res.Results {
		fmt.Println(loc.Name)
	}
	c.next = res.Next
	c.previous = res.Previous
	return nil
}

func commandExplore(c *config, location string) error {
	if location == "" {
		return fmt.Errorf("usage: explore <location-area>")
	}
	url := "https://pokeapi.co/api/v2/location-area/" + location

	if data, ok := c.cache.Get(url); ok {
		var res locationAreaResult
		if err := json.Unmarshal(data, &res); err != nil {
			return err
		}
		for _, e := range res.PokemonEncounters {
			fmt.Println(" -", e.Pokemon.Name)
		}
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	c.cache.Add(url, body)

	var res locationAreaResult
	if err = json.Unmarshal(body, &res); err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\n", location)
	fmt.Println("Found Pokemon:")
	for _, e := range res.PokemonEncounters {
		fmt.Println(" -", e.Pokemon.Name)
	}
	return nil
}

func commandCatch(c *config, name string) error {
	if name == "" {
		return fmt.Errorf("usage: catch <pokemon-name>")
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	url := "https://pokeapi.co/api/v2/pokemon/" + name

	var pokemon Pokemon

	if data, ok := c.cache.Get(url); ok {
		if err := json.Unmarshal(data, &pokemon); err != nil {
			return err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode > 299 {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		c.cache.Add(url, body)
		if err = json.Unmarshal(body, &pokemon); err != nil {
			return err
		}
	}

	if rand.Intn(100) > pokemon.Experience {
		fmt.Printf("%s was caught!\n", name)
		c.pokedex[name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", name)
	}

	return nil
}

func commandInspect(c *config, name string) error {
	if name == "" {
		return fmt.Errorf("usage: inspect <pokemon-name>")
	}
	pokemon, ok := c.pokedex[name]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", s.Stat.Name, s.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}
