package requests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
