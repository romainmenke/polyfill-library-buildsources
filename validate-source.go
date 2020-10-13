package main

import (
	"fmt"
	"log"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
)

func validateSource(code []byte, label string) error {
	result := api.Transform(string(code), api.TransformOptions{
		Loader: api.LoaderJS,
		Target: api.ES5,
	})

	hasRealWarningsOrErrors := false
	if len(result.Errors) > 0 || len(result.Warnings) > 0 {
		for _, err := range result.Errors {
			log.Println("err", err)
			hasRealWarningsOrErrors = true
		}

		for _, warning := range result.Warnings {
			if warning.Text == "Comparison with -0 using the \"===\" operator will also match 0" {
				continue
			}

			log.Println("warning", warning)
			hasRealWarningsOrErrors = true
		}
	}

	if hasRealWarningsOrErrors {
		return errors.New(fmt.Sprintf("Error parsing source code for %s", label))
	}

	return nil
}
