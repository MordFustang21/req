package requests

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parse(t *testing.T) {
	f, err := os.Open("testdata/main.http")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tests := ParseTests(f)

	expected := []HTTPTest{
		{
			Name:   "Named Request",
			Method: http.MethodGet,
			URL: url.URL{
				Scheme: "http",
				Host:   "go.dev",
			},
			Headers: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
			},
			Body: []byte(`{"name": "John Doe"}`),
		},
		{
			Name:   "Test 2",
			Method: http.MethodGet,
			URL: url.URL{
				Scheme: "http",
				Host:   "zed.dev",
			},
			Headers: http.Header{},
		},
		{
			Name:   "Request with variables",
			Method: http.MethodGet,
			URL: url.URL{
				Scheme: "http",
				Host:   "zed.dev:8080",
			},
			Headers: http.Header{
				"Authorization": []string{"Bearer my-token-variable"},
			},
		},
	}

	assert.Equal(t, expected, tests, "got invalid tests")
}
