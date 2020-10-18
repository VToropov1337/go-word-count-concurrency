package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type result struct {
	URL   string
	Qty   int
	Error error
}

func main() {
	start := time.Now()
	resultChan := make(chan result)
	done := make(chan bool)
	maxWorkers := 5
	total := 0

	go worker(maxWorkers, resultChan, done)

	go func() {
		<-done
		close(resultChan)
	}()

	for item := range resultChan {
		if item.Error != nil {
			continue
		}
		fmt.Printf("Count for %s: %d\n", item.URL, item.Qty)

		total += item.Qty
	}
	fmt.Printf("Total: %v\n", total)
	elapsed := time.Since(start)
	fmt.Printf("The data obtained for %s\n", elapsed)
}

func wordCounter(url string) result {
	resp, err := http.Get(url)
	if err != nil {
		return result{url, 0, err}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result{url, 0, err}
	}

	words := bytes.Count(body, []byte("Go"))
	return result{url, words, nil}

}

func worker(threadLimit int, resultChan chan result, done chan bool) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, threadLimit)

	defer func() {
		close(semaphore)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(s string, semaphore chan struct{}, wg *sync.WaitGroup) {
			defer wg.Done()
			resultChan <- wordCounter(s)
			<-semaphore
		}(scanner.Text(), semaphore, &wg)
	}
	wg.Wait()
	done <- true
}
