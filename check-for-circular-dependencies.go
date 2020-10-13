package main

import (
	"fmt"

	toposort "github.com/philopon/go-toposort"
	"github.com/pkg/errors"
)

func checkForCircularDependencies(polyfills []*Polyfill) error {
	graph := toposort.NewGraph(len(polyfills))

	for _, polyfill := range polyfills {
		ok := graph.AddNode(polyfill.name)
		if !ok {
			return errors.New(fmt.Sprintf("Unable to construct dependency graph for %s", polyfill.name))
		}
	}

	for _, polyfill := range polyfills {
		for _, dependency := range polyfill.dependencies() {
			ok := graph.AddEdge(dependency, polyfill.name)
			if !ok {
				return errors.New(fmt.Sprintf("Unable to construct dependency graph for %s which depends on %s", polyfill.name, dependency))
			}
		}
	}

	_, ok := graph.Toposort()
	if !ok {
		return errors.New("Unable to construct dependency graph")
	}

	return nil
}
