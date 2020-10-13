package main

import (
	"fmt"

	"github.com/pkg/errors"
)

func checkDependenciesExist(polyfills []*Polyfill) error {
	mapped := map[string]struct{}{}
	for _, polyfill := range polyfills {
		mapped[polyfill.name] = struct{}{}
	}

	for _, polyfill := range polyfills {
		for _, dependency := range polyfill.dependencies() {
			if _, ok := mapped[dependency]; !ok {
				return errors.New(fmt.Sprintf("Polyfill %s depends on %s, which does not exist within the polyfill-service. Recommended to either add the missing polyfill or remove the dependency.", polyfill.name, dependency))
			}
		}
	}

	return nil
}
