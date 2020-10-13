package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkPolyfill(b *testing.B) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createTestPolyfill(b, tmp)

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		polyfill := NewPolyfill(filepath.Join(tmp, "Golden"), "Golden")
		err = polyfill.loadConfig()
		if err != nil {
			b.Fatal(err)
		}

		err = polyfill.checkLicense()
		if err != nil {
			b.Fatal(err)
		}

		err = polyfill.loadSources()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestPolyfillLoadConfig(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createTestPolyfill(t, tmp)

	polyfill := NewPolyfill(filepath.Join(tmp, "Golden"), "Golden")
	err = polyfill.loadConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPolyfillCheckLicense(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createTestPolyfill(t, tmp)

	polyfill := NewPolyfill(filepath.Join(tmp, "Golden"), "Golden")
	err = polyfill.loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	err = polyfill.checkLicense()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPolyfillWriteOutput(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createTestPolyfill(t, tmp)

	polyfill := NewPolyfill(filepath.Join(tmp, "Golden"), "Golden")
	err = polyfill.loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	err = polyfill.checkLicense()
	if err != nil {
		t.Fatal(err)
	}

	err = polyfill.writeOutput(filepath.Join(tmp, "dist"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPolyfillProcess(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createTestPolyfill(t, tmp)

	polyfill := NewPolyfill(filepath.Join(tmp, "Golden"), "Golden")
	err = polyfill.process(filepath.Join(tmp, "dist"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailingPolyfillCheckLicense(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	createBadTestPolyfill(t, tmp)

	polyfill := NewPolyfill(filepath.Join(tmp, "Bad"), "Bad")
	err = polyfill.loadConfig()
	if err != nil {
		t.Fatal(err)
	}

	err = polyfill.checkLicense()
	if err == nil {
		t.Fatal("Expected failure")
	}

	if err.Error() != "The license Abstyles (Bad) is not OSI approved." {
		t.Fatal("Expected : \"The license Abstyles (Bad) is not OSI approved.\"")
	}
}

func createTestPolyfill(t testing.TB, dir string) {
	err := os.Mkdir(filepath.Join(dir, "Golden"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	polyfillFile, err := os.Create(filepath.Join(dir, "Golden/polyfill.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer polyfillFile.Close()

	configFile, err := os.Create(filepath.Join(dir, "Golden/config.toml"))
	if err != nil {
		t.Fatal(err)
	}

	defer configFile.Close()

	detectFile, err := os.Create(filepath.Join(dir, "Golden/detect.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer detectFile.Close()

	testFile, err := os.Create(filepath.Join(dir, "Golden/tests.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer testFile.Close()

	_, err = polyfillFile.WriteString(`
function Golden() {
	console.log('hello world!');
}
`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = configFile.WriteString(`
aliases = []
dependencies = [
  "window"
]
license = "MIT"
docs = ""
spec = ""

[browsers]
ie = "*"
ie_mob = "10 - 11"
safari = "<9"
chrome = "<44"
firefox = "<40"
android = "*"
ios_saf = "<9"
opera = "<25"
firefox_mob = "<36"
bb = "10 - *"

`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = detectFile.WriteString(`
(function() {
	try {
		if ('Golden' in self) {
			return true	
		}
	} catch (err) {
		return false;
	}

	return false;
}())
`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testFile.WriteString(`
it('is a function', function () {
	proclaim.isFunction(Golden);
});
`)
	if err != nil {
		t.Fatal(err)
	}
}

func createBadTestPolyfill(t testing.TB, dir string) {
	err := os.Mkdir(filepath.Join(dir, "Bad"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	polyfillFile, err := os.Create(filepath.Join(dir, "Bad/polyfill.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer polyfillFile.Close()

	configFile, err := os.Create(filepath.Join(dir, "Bad/config.toml"))
	if err != nil {
		t.Fatal(err)
	}

	defer configFile.Close()

	detectFile, err := os.Create(filepath.Join(dir, "Bad/detect.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer detectFile.Close()

	testFile, err := os.Create(filepath.Join(dir, "Bad/tests.js"))
	if err != nil {
		t.Fatal(err)
	}

	defer testFile.Close()

	_, err = polyfillFile.WriteString(`
function Bad() {
	console.log('hello world!');
}
`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = configFile.WriteString(`
aliases = []
dependencies = [
  "window"
]
license = "Abstyles"
docs = ""
spec = ""

[browsers]
ie = "*"
ie_mob = "10 - 11"
safari = "<9"
chrome = "<44"
firefox = "<40"
android = "*"
ios_saf = "<9"
opera = "<25"
firefox_mob = "<36"
bb = "10 - *"

`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = detectFile.WriteString(`
(function() {
	try {
		if ('Bad' in self) {
			return true	
		}
	} catch (err) {
		return false;
	}

	return false;
}())
`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testFile.WriteString(`
it('is a function', function () {
	proclaim.isFunction(Bad);
});
`)
	if err != nil {
		t.Fatal(err)
	}
}
