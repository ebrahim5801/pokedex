package main

import "strings"

func cleanInput(text string) []string {
	words := strings.Fields(text)

	var output []string
	for _, w := range words {
		output = append(output, strings.ToLower(w))
	}
	return output
}
