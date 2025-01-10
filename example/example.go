package main

import (
	"context"
	"log"
	"strings"

	"requests-go"
)

func main() {
	ctx := context.Background()
	cfg := requests.Config{}
	c := requests.NewClient(ctx, cfg)

	p, err := c.GetHTML("https://en.wikipedia.org/wiki/Main_Page")
	if err != nil {
		log.Fatal(err)
	}

	links, err := p.GetLinks()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(strings.Join(links, "\n"))

	nodes, err := p.GetElementsByClassName("vector-header")
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		log.Println(node)
	}
}
