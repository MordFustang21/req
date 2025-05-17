package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MordFustang21/req/pkg/requests"
	"github.com/manifoldco/promptui"
)

var (
	preview = flag.Bool("preview", false, "Preview the request")
	timing  = flag.Bool("timing", false, "Show timing information")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		panic("must provide an .http file")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal("error opening .http file", err)
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
		requests.ExecuteTest(tests[index], *timing)
	case err == promptui.ErrInterrupt:
		fmt.Println("No Test Selected")
		os.Exit(0)
	default:
		fmt.Printf("Prompt failed %v\n", err)
	}

}
