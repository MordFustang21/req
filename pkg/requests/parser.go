package requests

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	ParseStepMethodURL = iota
	ParseStepHeader
	ParseStepBody
)

func newTest() HTTPTest {
	return HTTPTest{
		Name:    "Unnamed",
		Headers: http.Header{},
	}
}

func ParseTests(f *os.File) []HTTPTest {
	variables := map[string]string{}

	var tests []HTTPTest
	test := newTest()

	currentStep := ParseStepMethodURL

	// Read line by line each test is separated by ###
	rdr := bufio.NewScanner(f)
	for rdr.Scan() {
		line := rdr.Text()
		// if line starts with a comment skip it
		switch {
		// On empty line increment the step
		case len(line) == 0:
			if currentStep == ParseStepBody {
				// We're done with the test, add it to the tests
				tests = append(tests, test)
				test = newTest()
				currentStep = ParseStepMethodURL
				continue
			}

			currentStep++
			continue
		case strings.HasPrefix(line, "//"):
			continue

		// This is a variable we need to parse and inject.
		case line[0] == '@':
			// split line into key and value
			key, value, ok := strings.Cut(line[1:], "=")
			if !ok {
				panic("invalid variable " + line)
			}

			variables[key] = value
			continue
		}

		if strings.HasPrefix(line, "###") {
			// Check for existing test and add it to the tests.
			if test.Method != "" {
				tests = append(tests, test)
				currentStep = ParseStepMethodURL
				test = newTest()
			}

			// Set test name
			test.Name = strings.TrimSpace(strings.TrimPrefix(line, "###"))
			// Reset step
			currentStep = ParseStepMethodURL
			continue
		}

		if currentStep == ParseStepMethodURL {
			// split line into method and url
			method, urlStr, ok := strings.Cut(line, " ")
			if !ok {
				panic("invalid test url " + line)
			}

			test.Method = method

			// parse the url
			u, err := url.Parse(injectVariables(variables, urlStr))
			if err != nil {
				panic(err)
			}

			test.URL = *u

			continue
		}

		if currentStep == ParseStepHeader {
			// Parse header line into key and value
			key, value, ok := strings.Cut(line, ": ")
			if !ok {
				panic("invalid header: " + line)
			}

			// Add header to test
			test.Headers.Add(key, injectVariables(variables, value))
			continue
		}

		if currentStep == ParseStepBody {
			// Add body to test
			test.Body = append(test.Body, injectVariables(variables, line)+"\n"...)
			continue
		}
	}

	// if test is not empty add it to the tests
	// this is for the last test
	if test.Method != "" {
		tests = append(tests, test)
	}

	return tests
}

func injectVariables(variables map[string]string, input string) string {
	// Replace all variables in the input with their values
	for key, value := range variables {
		input = strings.ReplaceAll(input, "{{"+key+"}}", value)
	}

	return input
}
