package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func writeAliasFile(polyfills []*Polyfill, directory string) error {
	aliases := map[string][]string{}

	for _, polyfill := range polyfills {
		for _, alias := range polyfill.aliases() {
			if existing, ok := aliases[alias]; ok {
				aliases[alias] = append(existing, polyfill.name)
			} else {
				aliases[alias] = []string{polyfill.name}
			}
		}
	}

	f, err := os.Create(filepath.Join(directory, "aliases.json"))
	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")

	err = enc.Encode(aliases)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
