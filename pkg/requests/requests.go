package requests

import (
	"net/http"
	"net/url"
)

type HTTPTest struct {
	Name           string
	URL            url.URL
	Method         string
	Headers        http.Header
	Body           []byte
	MultipartParts []MultipartPart
}
