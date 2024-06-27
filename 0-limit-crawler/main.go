//////////////////////////////////////////////////////////////////////
//
// Your task is to change the code to limit the crawler to at most one
// page per second, while maintaining concurrency (in other words,
// Crawl() must be called concurrently)
//
// @hint: you can achieve this by adding 3 lines
//

package main

import (
	"fmt"
	"sync"
	"time"
)

var dataStream chan struct{}

func run(url string) (string, []string, error) {
	<-dataStream
	body, urls, err := fetcher.Fetch(url)
	return body, urls, err
}

// Crawl uses `fetcher` from the `mockfetcher.go` file to imitate a
// real crawler. It crawls until the maximum depth has reached.
func Crawl(url string, depth int, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth <= 0 {
		return
	}
	body, urls, err := run(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("found: %s %q\n", url, body)

	wg.Add(len(urls))
	for _, u := range urls {
		// Do not remove the `go` keyword, as Crawl() must be
		// called concurrently
		go Crawl(u, depth-1, wg)
	}
	return
}

func main() {
	var wg sync.WaitGroup
	dataStream = make(chan struct{}, 1)
	defer func() { close(dataStream) }()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case dataStream <- struct{}{}:
			default:
			}
		}
	}()
	wg.Add(1)
	Crawl("http://golang.org/", 4, &wg)
	wg.Wait()
}
