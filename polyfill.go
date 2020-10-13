package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/evanw/esbuild/pkg/api"
	spdx "github.com/mitchellh/go-spdx"
	"github.com/pkg/errors"
)

var spdxList *spdx.LicenseList

func init() {
	var err error
	spdxList, err = spdx.List()
	if err != nil {
		panic(err)
	}

}

type pathInfo struct {
	absolute string
	relative string
}

type polyfillConfig struct {
	Size         uint64 `json:"size"`
	DetectSource string `json:"detectSource"`
	BaseDir      string `json:"baseDir"`
	HasTests     bool   `json:"hasTests"`
	IsTestable   bool   `json:"isTestable"`
	IsPublic     bool   `json:"isPublic"`

	Aliases      []string          `toml:"aliases" json:"aliases"`
	Dependencies []string          `toml:"dependencies" json:"dependencies"`
	Spec         string            `toml:"spec" json:"spec"`
	Docs         string            `toml:"docs" json:"docs"`
	License      string            `toml:"license" json:"license"`
	Browsers     map[string]string `toml:"browsers" json:"browsers"`
	Test         map[string]*bool  `toml:"test" json:"test"`
	Build        map[string]*bool  `toml:"build" json:"build"`
}

type source struct {
	min []byte
	raw []byte
}

func NewPolyfill(absolutePath string, relativePath string) *Polyfill {
	return &Polyfill{
		name: strings.ReplaceAll(strings.ReplaceAll(relativePath, "/", "."), "\\", "."),
		path: pathInfo{
			relative: relativePath,
			absolute: absolutePath,
		},
	}
}

type Polyfill struct {
	path    pathInfo
	name    string
	config  polyfillConfig
	sources source
}

func (x *Polyfill) process(outputDir string) error {
	if !x.hasConfigFile() {
		return nil
	}

	err := x.loadConfig()
	if err != nil {
		return err
	}

	err = x.checkLicense()
	if err != nil {
		return err
	}

	err = x.loadSources()
	if err != nil {
		return err
	}

	x.updateConfig()

	err = x.writeOutput(outputDir)
	if err != nil {
		return err
	}

	return nil
}

func (x Polyfill) aliases() []string {
	return append([]string{"all"}, x.config.Aliases...)
}

func (x Polyfill) dependencies() []string {
	return x.config.Dependencies
}

func (x Polyfill) configPath() string {
	return filepath.Join(x.path.absolute, "config.toml")
}

func (x Polyfill) detectPath() string {
	return filepath.Join(x.path.absolute, "detect.js")
}

func (x Polyfill) sourcePath() string {
	return filepath.Join(x.path.absolute, "polyfill.js")
}

func (x Polyfill) hasConfigFile() bool {
	if _, err := os.Stat(x.configPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (x *Polyfill) updateConfig() {
	x.config.Size = uint64(len(x.sources.min))
}

func (x *Polyfill) loadConfig() error {
	{
		b, err := ioutil.ReadFile(x.configPath())
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to read config from %s", x.configPath()))
		}

		if _, err := toml.Decode(string(b), &x.config); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to read config from %s", x.configPath()))
		}
	}

	if strings.HasPrefix(x.path.relative, "_") {
		for ua := range uaBaselines {
			if version, ok := x.config.Browsers[ua]; !ok || version != "*" {
				browserSupport := map[string]string{}
				for ua := range uaBaselines {
					browserSupport[ua] = "*"
				}

				buf := new(bytes.Buffer)
				if err := toml.NewEncoder(buf).Encode(browserSupport); err != nil {
					panic(err) // should never happen, errors indicate a bug in this codebase
				}

				return errors.New(fmt.Sprintf("Internal polyfill called %s is not targeting all supported browsers correctly. It should be: \n%s", x.name, buf.String()))
			}
		}
	}

	x.config.BaseDir = x.path.relative

	if _, err := os.Stat(filepath.Join(x.path.absolute, "tests.js")); !os.IsNotExist(err) {
		x.config.HasTests = true
	}

	x.config.IsTestable = true
	if ci, ok := x.config.Test["ci"]; ok && ci != nil && *ci == false {
		x.config.IsTestable = false
	}

	x.config.IsPublic = !strings.HasPrefix(x.name, "_")

	if _, err := os.Stat(filepath.Join(x.path.absolute, "tests.js")); os.IsExist(err) {
		x.config.HasTests = true
	}

	if _, err := os.Stat(x.detectPath()); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(x.detectPath())
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to detect source from %s", x.detectPath()))
		}

		minified, err := x.minifyDetect(b)
		if err != nil {
			return err
		}

		err = validateSource([]byte(fmt.Sprintf("if (%s) true;", string(minified.min))), fmt.Sprintf("%s feature detect from %s", x.name, x.detectPath()))
		if err != nil {
			return err
		}

		x.config.DetectSource = string(minified.min)
	}

	return nil
}

func (x *Polyfill) loadSources() error {
	b, err := ioutil.ReadFile(x.sourcePath())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to polyfill source from %s", x.sourcePath()))
	}

	minified, err := x.minifyPolyfill(b)
	if err != nil {
		return err
	}

	x.sources = minified
	x.config.Size = uint64(len(x.sources.min))
	return nil
}

func (x *Polyfill) checkLicense() error {
	if x.config.License == "" {
		return nil
	}

	license := spdxList.License(x.config.License)
	if license == nil {
		return errors.New(fmt.Sprintf("The license %s is not on the SPDX list of licenses ( https://spdx.org/licenses/ ).", x.config.License))
	}

	if x.config.License != "CC0-1.0" && x.config.License != "WTFPL" && !license.OSIApproved {
		return errors.New(fmt.Sprintf("The license %s (%s) is not OSI approved.", x.config.License, x.name))
	}

	return nil
}

func (x *Polyfill) minifyDetect(detectSource []byte) (source, error) {
	raw := bytes.NewBuffer(nil)
	_, err := raw.WriteString(fmt.Sprintf("\n// %s\n", x.name))
	if err != nil {
		return source{}, err
	}

	_, err = raw.Write(detectSource)
	if err != nil {
		return source{}, err
	}

	if minify, ok := x.config.Build["minify"]; ok && minify != nil && *minify == false {
		return source{
			raw: raw.Bytes(),
			min: append(detectSource, []byte("\n")...),
		}, nil
	}

	err = validateSource(detectSource, fmt.Sprintf(`%s from %s`, x.name, x.sourcePath()))
	if err != nil {
		return source{}, err
	}

	result := api.Transform(string(detectSource), api.TransformOptions{
		Loader:            api.LoaderJS,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})

	if len(result.Errors) > 0 || len(result.Warnings) > 0 {
		return source{}, errors.New(fmt.Sprintf("Error minifying detect source for %s", x.name))
	}

	return source{
		raw: raw.Bytes(),
		min: []byte(strings.TrimSuffix(string(result.JS), ";\n")),
	}, nil
}

func (x *Polyfill) minifyPolyfill(polyfillSource []byte) (source, error) {
	raw := bytes.NewBuffer(nil)
	_, err := raw.WriteString(fmt.Sprintf("\n// %s\n", x.name))
	if err != nil {
		return source{}, err
	}

	_, err = raw.Write(polyfillSource)
	if err != nil {
		return source{}, err
	}

	if minify, ok := x.config.Build["minify"]; ok && minify != nil && *minify == false {
		return source{
			raw: raw.Bytes(),
			min: append(polyfillSource, []byte("\n")...),
		}, nil
	}

	err = validateSource(polyfillSource, fmt.Sprintf(`%s from %s`, x.name, x.sourcePath()))
	if err != nil {
		return source{}, err
	}

	result := api.Transform(string(polyfillSource), api.TransformOptions{
		Loader:            api.LoaderJS,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})

	// if len(result.Errors) > 0 || len(result.Warnings) > 0 {
	// 	return source{}, errors.New(fmt.Sprintf("Error minifying polyfill source for %s", x.name))
	// }

	return source{
		raw: removeSourceMaps(raw.Bytes()),
		min: removeSourceMaps(result.JS),
	}, nil
}

func (x *Polyfill) writeOutput(root string) error {
	encodedConfig := bytes.NewBuffer(nil)
	configEncoder := json.NewEncoder(encodedConfig)
	configEncoder.SetEscapeHTML(false)
	configEncoder.SetIndent("", "\t")

	err := configEncoder.Encode(x.config)
	if err != nil {
		return err
	}

	destination := filepath.Join(root, x.name)
	files := map[string][]byte{
		filepath.Join(destination, "meta.json"): encodedConfig.Bytes(),
		filepath.Join(destination, "raw.js"):    x.sources.raw,
		filepath.Join(destination, "min.js"):    x.sources.min,
	}

	err = os.MkdirAll(destination, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	for path, contents := range files {
		f, err := os.Create(path)
		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write(contents)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func removeSourceMaps(in []byte) []byte {
	r := regexp.MustCompile(`^\/\/#\ssourceMappingURL(.+)$`)
	return r.ReplaceAll(in, []byte{})
}
