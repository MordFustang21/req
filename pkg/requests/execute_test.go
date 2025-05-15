package requests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteTestWithHttpTrace(t *testing.T) {
	// Mock server to simulate HTTP responses
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Simulate processing delay
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer mockServer.Close()

	// Create a mock HTTPTest
	test := HTTPTest{
		Name:   "Mock Test",
		Method: http.MethodGet,
		URL:    mustParseURL(mockServer.URL),
		Headers: http.Header{
			"Authorization": []string{"Bearer mock-token"},
		},
		Body: []byte{},
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the test
	ExecuteTest(test, true)

	// Restore stdout and read output
	w.Close()
	var output bytes.Buffer
	io.Copy(&output, r)
	os.Stdout = oldStdout

	// Convert output to string
	outputStr := output.String()

	// Validate output contains timing information
	assert.Contains(t, outputStr, "--- Timing Information ---")
	assert.Contains(t, outputStr, "DNS Lookup:")
	assert.Contains(t, outputStr, "TCP Connection:")
	assert.Contains(t, outputStr, "TLS Handshake:")
	assert.Contains(t, outputStr, "Time to First Byte:")
	assert.Contains(t, outputStr, "Total Request Time:")

	// Validate response details
	assert.Contains(t, outputStr, "HTTP/1.1 200 OK")
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

// End-to-end test: load multipart form from .http file, parse, and execute
func TestExecuteTestWithMultipartFormFromHTTPFile(t *testing.T) {
	// Mock server to verify multipart form data
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		err := r.ParseMultipartForm(10 << 20)
		assert.NoError(t, err)

		formValue := r.FormValue("field1")
		assert.Equal(t, "value1", formValue)

		file, header, err := r.FormFile("file1")
		require.NoError(t, err)

		defer file.Close()
		assert.Equal(t, "test.txt", header.Filename)
		content, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, "file content", string(content))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "multipart received"}`))
	}))
	defer mockServer.Close()

	// Load and parse the .http file
	f, err := os.Open("testdata/multipart.http")
	assert.NoError(t, err)
	defer f.Close()

	tests, err := ParseTests(f)
	assert.NoError(t, err)

	// Find the multipart test
	var multipartTest *HTTPTest
	for _, test := range tests {
		if strings.Contains(strings.ToLower(test.Name), "multipart") {
			multipartTest = &test
			break
		}
	}
	if multipartTest == nil {
		t.Fatal("Multipart test not found in .http file")
	}

	// Update the URL to point to the mock server
	multipartTest.URL = mustParseURL(mockServer.URL)

	// Capture output
	oldStdout := os.Stdout
	rPipe, wPipe, _ := os.Pipe()
	os.Stdout = wPipe

	ExecuteTest(*multipartTest, false)

	wPipe.Close()
	var output bytes.Buffer
	io.Copy(&output, rPipe)
	os.Stdout = oldStdout

	outputStr := output.String()
	assert.Contains(t, outputStr, "HTTP/1.1 200 OK")
	assert.JSONEq(t, `{"message": "multipart received"}`, strings.SplitN(outputStr, "\n\n", 2)[1])
}
