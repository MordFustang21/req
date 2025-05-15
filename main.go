package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/MordFustang21/req/pkg/requests"
	"github.com/TylerBrock/colorjson"
	"github.com/manifoldco/promptui"
)

var (
	preview = flag.Bool("preview", false, "Preview the request")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		panic("must provide an .http file")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tests, err := requests.ParseTests(f)
	if err != nil {
		panic(err)
	}

	runPromptUI(tests)
}

func runPromptUI(tests []requests.HTTPTest) {
	customFuncMap := promptui.FuncMap
	customFuncMap["url"] = func(t requests.HTTPTest) string {
		return t.URL.String()
	}

	subtestPrompt := promptui.Select{
		Label: "Select a request",
		Items: tests,
		Templates: &promptui.SelectTemplates{
			Active:   "> {{ .Name }} - {{ .Method}} {{ url . }}",
			Inactive: "  {{ .Name }} - {{ .Method }} {{ url . }}",
			Selected: "{{ .Name }} - {{ .Method }} {{ url . }}",
			FuncMap:  customFuncMap,
		},
		Searcher: func(input string, index int) bool {
			test := tests[index]
			input = strings.ToLower(input)
			switch {
			case strings.Contains(strings.ToLower(test.Name), input):
				return true
			case strings.Contains(strings.ToLower(test.URL.String()), input):
				return true
			case strings.Contains(strings.ToLower(test.Method), input):
				return true
			default:
				return false
			}
		},
	}

	index, _, err := subtestPrompt.Run()
	switch {
	case err == nil:
		executeTest(tests[index])
	case err == promptui.ErrInterrupt:
		fmt.Println("No Test Selected")
		os.Exit(0)
	default:
		panic(err)
	}

}

func executeTest(test requests.HTTPTest) {
	if *preview {
		fmt.Println(test.Method + " " + test.URL.String())
		for key, values := range test.Headers {
			for _, value := range values {
				fmt.Println(key + ": " + value)
			}
		}
		fmt.Println()
		fmt.Println(string(test.Body))
		return
	}

	var req *http.Request
	var err error

	if len(test.MultipartParts) > 0 {
		// Build multipart body
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		for _, part := range test.MultipartParts {
			hdr := make(textproto.MIMEHeader)
			for k, vals := range part.Headers {
				for _, v := range vals {
					hdr.Add(k, v)
				}
			}
			w, err := writer.CreatePart(hdr)
			if err != nil {
				panic(err)
			}
			_, err = w.Write(part.Content)
			if err != nil {
				panic(err)
			}
		}
		writer.Close()

		req, err = http.NewRequest(test.Method, test.URL.String(), &buf)
		if err != nil {
			panic(err)
		}

		// Set Content-Type to multipart with boundary
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Add other headers except Content-Type (already set)
		for key, values := range test.Headers {
			if strings.ToLower(key) == "content-type" {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	} else {
		req, err = http.NewRequest(test.Method, test.URL.String(), strings.NewReader(string(test.Body)))
		if err != nil {
			panic(err)
		}
		for key, values := range test.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

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

func printFormattedBody(headers http.Header, resp *http.Response) {
	// Detect response encoding
	encoding := headers.Get("Content-Type")
	switch encoding {
	case "application/json":
		// Decode into a usable JSON object
		var obj interface{}
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
