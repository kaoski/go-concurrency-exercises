//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"sync"
	"time"
)

var tweetChan chan *Tweet

func producer(stream Stream, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		tweet, err := stream.Next()
		if err == nil {
			go func() { tweetChan <- tweet }()
		}
		if err == ErrEOF {
			close(tweetChan)
			return
		}
	}
}

func consumer(wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range tweetChan {
		go func(t *Tweet) {
			if t.IsTalkingAboutGo() {
				fmt.Println(t.Username, "\ttweets about golang")
			} else {
				fmt.Println(t.Username, "\tdoes not tweet about golang")
			}
		}(t)
	}
}

func main() {
	var wg sync.WaitGroup
	start := time.Now()
	stream := GetMockStream()
	tweetChan = make(chan *Tweet, 100)

	// Producer
	wg.Add(1)
	go producer(stream, &wg)

	// Consumer
	wg.Add(1)
	go consumer(&wg)

	wg.Wait()
	fmt.Printf("Process took %s\n", time.Since(start))
}
