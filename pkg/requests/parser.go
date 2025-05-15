package requests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
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

type MultipartPart struct {
	Headers textproto.MIMEHeader
	Content []byte
}

// ParseTests reads a file and parses it into a slice of HTTPTest structs.
func ParseTests(f io.Reader) ([]HTTPTest, error) {
	variables := map[string]string{}

	var tests []HTTPTest
	test := newTest()

	currentStep := ParseStepMethodURL

	rdr := bufio.NewScanner(f)
	lineNum := 0
	for rdr.Scan() {
		lineNum++
		line := rdr.Text()
		switch {
		case len(line) == 0:
			if currentStep == ParseStepBody {
				// Finalize multipart if needed
				if isMultipart(test.Headers) && len(test.Body) > 0 {
					parts, err := parseMultipartBody(test.Headers, test.Body)
					if err != nil {
						return nil, fmt.Errorf("line %d: multipart parse error: %w", lineNum, err)
					}
					test.MultipartParts = parts
				}
				tests = append(tests, test)
				test = newTest()
				currentStep = ParseStepMethodURL
				continue
			}
			currentStep++
			continue
		case strings.HasPrefix(line, "//"):
			continue
		case line[0] == '@':
			key, value, ok := strings.Cut(line[1:], "=")
			if !ok {
				return nil, fmt.Errorf("line %d: invalid variable declaration: %q", lineNum, line)
			}
			variables[key] = value
			continue
		}

		if strings.HasPrefix(line, "###") {
			if test.Method != "" {
				if isMultipart(test.Headers) && len(test.Body) > 0 {
					parts, err := parseMultipartBody(test.Headers, test.Body)
					if err != nil {
						return nil, fmt.Errorf("line %d: multipart parse error: %w", lineNum, err)
					}
					test.MultipartParts = parts
				}
				tests = append(tests, test)
				currentStep = ParseStepMethodURL
				test = newTest()
			}
			test.Name = strings.TrimSpace(strings.TrimPrefix(line, "###"))
			currentStep = ParseStepMethodURL
			continue
		}

		if currentStep == ParseStepMethodURL {
			method, urlStr, ok := strings.Cut(line, " ")
			if !ok {
				return nil, fmt.Errorf("line %d: invalid test method/url: %q", lineNum, line)
			}
			test.Method = method
			u, err := url.Parse(injectVariables(variables, urlStr))
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid url: %q: %w", lineNum, urlStr, err)
			}
			test.URL = *u
			continue
		}

		if currentStep == ParseStepHeader {
			key, value, ok := strings.Cut(line, ": ")
			if !ok {
				return nil, fmt.Errorf("line %d: invalid header: %q", lineNum, line)
			}
			test.Headers.Add(key, injectVariables(variables, value))
			continue
		}

		if currentStep == ParseStepBody {
			test.Body = append(test.Body, injectVariables(variables, line)+"\n"...)
			continue
		}
	}
	if err := rdr.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	if test.Method != "" {
		if isMultipart(test.Headers) && len(test.Body) > 0 {
			parts, err := parseMultipartBody(test.Headers, test.Body)
			if err != nil {
				return nil, fmt.Errorf("final multipart parse error: %w", err)
			}
			test.MultipartParts = parts
		}
		tests = append(tests, test)
	}

	return tests, nil
}

// Helper to check if Content-Type is multipart/form-data
func isMultipart(headers http.Header) bool {
	ct := headers.Get("Content-Type")
	return strings.HasPrefix(ct, "multipart/form-data")
}

// Parse multipart body into parts using mime/multipart
func parseMultipartBody(headers http.Header, body []byte) ([]MultipartPart, error) {
	ct := headers.Get("Content-Type")
	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("parse content-type: %w", err)
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("missing boundary in multipart/form-data")
	}

	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	var parts []MultipartPart
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading part: %w", err)
		}
		content, err := io.ReadAll(p)
		if err != nil {
			return nil, fmt.Errorf("reading part content: %w", err)
		}
		parts = append(parts, MultipartPart{
			Headers: p.Header,
			Content: content,
		})
	}

	return parts, nil
}

const (
	placeholderStart = "{{"
	placeholderEnd   = "}}"
)

var variablePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

func injectVariables(variables map[string]string, input string) string {
	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		key := variablePattern.FindStringSubmatch(match)[1]
		if val := os.Getenv(key); val != "" {
			return val
		}
		if val, ok := variables[key]; ok {
			return val
		}
		return match // leave as-is if not found
	})
}
