package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Sync
	wg := sync.WaitGroup{}

	// Results
	polyfills := []*Polyfill{}
	results := make(chan *Polyfill)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case r := <-results:
				polyfills = append(polyfills, r)
				wg.Done()
			}
		}
	}()

	// Workers
	jobs := make(chan string)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case j := <-jobs:
					abspath, err := filepath.Abs(filepath.Join("./polyfills", j))
					if err != nil {
						log.Fatal(err)
					}

					polyfill := NewPolyfill(abspath, j)
					err = polyfill.process("./polyfills/__dist")
					if err != nil {
						log.Fatal(err)
					}

					results <- polyfill
				}
			}
		}()
	}

	err := os.RemoveAll("./polyfills/__dist")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	err = os.MkdirAll("./polyfills/__dist", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	polyfillPaths, err := flattenPolyfillDirectories("./polyfills")
	if err != nil {
		log.Fatal(err)
	}

	for _, polyfillPath := range polyfillPaths {
		wg.Add(1)
		jobs <- polyfillPath
	}

	wg.Wait()

	err = checkForCircularDependencies(polyfills)
	if err != nil {
		log.Fatal(err)
	}

	err = checkDependenciesExist(polyfills)
	if err != nil {
		log.Fatal(err)
	}

	err = writeAliasFile(polyfills, "./polyfills/__dist")
	if err != nil {
		log.Fatal(err)
	}
}
