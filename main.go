package main

import (
	"sync"
)

var wg sync.WaitGroup

func main() {
	wg.Add(4)
	go Starter()
	wg.Wait()
}
