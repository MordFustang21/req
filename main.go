package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/MordFustang21/req/pkg/requests"
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

	tests := requests.ParseTests(f)
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

	req, err := http.NewRequest(test.Method, test.URL.String(), strings.NewReader(string(test.Body)))
	if err != nil {
		panic(err)
	}

	for key, values := range test.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
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

	io.Copy(os.Stdout, resp.Body)
	resp.Body.Close()
}
