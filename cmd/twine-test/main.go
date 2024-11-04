package main

import (
	"fmt"
	"os"

	"github.com/m110/secrets/assets/twine"
)

func main() {
	file, err := os.ReadFile("cmd/twine-test/test.twee")
	if err != nil {
		panic(err)
	}

	story, err := twine.ParseStory(string(file))
	if err != nil {
		panic(err)
	}

	for _, p := range story.Passages {
		fmt.Println(p.Title)
		fmt.Println("\t", p.Tags)
		fmt.Println("\t", p.Content)
		fmt.Println("\t", p.Links)

		fmt.Println()
	}

}
