package embedfs

import "testing"

type testDatum struct {
	basename string
	dirname  string
}

var (
	testData = map[string]testDatum{
		"/": testDatum{
			basename: "/",
			dirname:  "/",
		},
		"/app.js": testDatum{
			basename: "app.js",
			dirname:  "/",
		},
		"/folder1/folder2/index.html": testDatum{
			basename: "index.html",
			dirname:  "/folder1/folder2",
		},
	}
)

func TestBasename(t *testing.T) {
	for value, expected := range testData {
		actual := basename(value)
		if actual != expected.basename {
			t.Errorf("Basename of '%s' should return '%s', but got '%s'", value, expected.basename, actual)
		}
	}
}

func TestDirname(t *testing.T) {
	for value, expected := range testData {
		actual := dirname(value)
		if actual != expected.dirname {
			t.Errorf("Dirname of '%s' should return '%s', but got '%s'", value, expected.dirname, actual)
		}
	}
}
