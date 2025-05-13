package requests

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parse(t *testing.T) {
	// Define table-driven test cases
	testCases := []struct {
		filePath string
		expected []HTTPTest
	}{
		{
			filePath: "testdata/main.http",
			expected: []HTTPTest{
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
					Body: []byte("{\"name\": \"John Doe\"}\n"),
				},
				{
					Name:   "Multi-line Body",
					Method: http.MethodGet,
					URL: url.URL{
						Scheme: "http",
						Host:   "go.dev",
					},
					Headers: http.Header{
						"Authorization": []string{"Bearer token"},
						"Content-Type":  []string{"application/json"},
					},
					Body: []byte("{\n    \"name\": \"John Doe\"\n}\n"),
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
			},
		},
		{
			filePath: "testdata/readme.http",
			expected: []HTTPTest{
				{
					Name:   "Named requests",
					Method: http.MethodGet,
					URL: url.URL{
						Scheme: "https",
						Host:   "jsonplaceholder.typicode.com",
						Path:   "/todos/1",
					},
					Headers: http.Header{
						"Accept": []string{"application/json"},
					},
					Body: []byte("{\n  \"userId\": 1,\n  \"id\": 1,\n  \"title\": \"delectus aut autem\",\n  \"completed\": false\n}\n"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filePath, func(t *testing.T) {
			// Open the test file
			f, err := os.Open(tc.filePath)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			// Parse the tests
			tests := ParseTests(f)

			// Assert the parsed tests match the expected tests
			assert.Equal(t, tc.expected, tests, "got invalid tests for file: %s", tc.filePath)
		})
	}
}
