package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	requests "github.com/mreck/requests-go"
)

func main() {
	ctx := context.Background()
	cfg := requests.Config{CreateDirs: true}
	clt := requests.NewClient(ctx, cfg)

	p, err := clt.GetHTML("https://en.wikipedia.org/wiki/Main_Page")
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
		log.Println(node.ClassList())
	}

	queue := clt.CreateDownloadQueue(requests.DownloadQueueConfig{
		WorkerCount: 2,
	})

	imgs, err := p.GetElementsByTagName("img")
	if err != nil {
		log.Fatal(err)
	}

	var imgLinks []string

	for _, img := range imgs {
		if src, ok := img.Src(); ok {
			imgLinks = append(imgLinks, "https://en.wikipedia.org/"+src)
		}
	}

	for i, link := range imgLinks {
		ext := filepath.Ext(link)
		err := queue.Enqueue(link, filepath.Join("tmp", fmt.Sprintf("%d.%s", i, ext)))
		if err != nil {
			log.Fatal(err)
		}
	}

	queue.WaitUntilDone()

	for _, err := range queue.Errors() {
		log.Println(err)
	}
}
