package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"

	"github.com/TylerBrock/colorjson"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

// ExecuteTest executes an HTTP test and prints the response and timing information.
func ExecuteTest(test HTTPTest, withTiming bool) {
	var start, dnsStart, connectStart, tlsStart, responseStart time.Time
	var dnsDuration, connectDuration, tlsDuration, requestDuration, responseDuration time.Duration

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsDuration = time.Since(dnsStart)
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connectDuration = time.Since(connectStart)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsDuration = time.Since(tlsStart)
		},
		GotFirstResponseByte: func() {
			responseDuration = time.Since(responseStart)
		},
	}

	var bodyReader io.Reader

	if len(test.MultipartParts) > 0 {
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		for _, part := range test.MultipartParts {
			fw, err := writer.CreatePart(part.Headers)
			if err != nil {
				panic(err)
			}

			_, err = fw.Write(part.Content)
			if err != nil {
				panic(err)
			}
		}

		if err := writer.Close(); err != nil {
			panic(err)
		}

		bodyReader = &b

		// Set Content-Type header for multipart
		if test.Headers == nil {
			test.Headers = http.Header{}
		}
		test.Headers = test.Headers.Clone()
		test.Headers.Set("Content-Type", writer.FormDataContentType())
	} else {
		bodyReader = strings.NewReader(string(test.Body))
	}

	req, err := http.NewRequest(test.Method, test.URL.String(), bodyReader)
	if err != nil {
		panic(err)
	}

	for key, values := range test.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Attach the trace to the request's context
	ctx := httptrace.WithClientTrace(req.Context(), trace)
	req = req.WithContext(ctx)

	start = time.Now()
	responseStart = time.Now()

	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	requestDuration = time.Since(start)

	if withTiming {
		// Print timing information
		fmt.Println("\n--- Timing Information ---")
		fmt.Printf("DNS Lookup:         %v\n", dnsDuration)
		fmt.Printf("TCP Connection:     %v\n", connectDuration)
		fmt.Printf("TLS Handshake:      %v\n", tlsDuration)
		fmt.Printf("Time to First Byte: %v\n", responseDuration)
		fmt.Printf("Total Request Time: %v\n", requestDuration)
	}

	// Print all response info
	fmt.Println(resp.Proto + " " + resp.Status)

	// Print out headers
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Println(key + ": " + value)
		}
	}

	fmt.Println()

	printFormattedBody(resp.Header, resp)
}

const (
	EncodingJSON = "application/json"
)

func printFormattedBody(headers http.Header, resp *http.Response) {
	// Detect response encoding
	encoding := headers.Get("Content-Type")
	switch {
	case strings.HasPrefix(encoding, EncodingJSON):
		// Decode into a usable JSON object
		var obj any
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&obj)
		if err != nil {
			panic(err)
		}

		// Pretty print JSON
		f := colorjson.NewFormatter()
		f.Indent = 2
		out, _ := f.Marshal(obj)
		os.Stdout.Write(out)

	default:
		io.Copy(os.Stdout, resp.Body)
	}
}
