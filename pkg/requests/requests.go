package requests

import (
	"net/http"
	"net/textproto"
	"net/url"
)

type HTTPTest struct {
	Name           string
	URL            *url.URL
	Method         string
	Headers        http.Header
	Body           []byte
	MultipartParts []MultipartPart
}

type MultipartPart struct {
	Headers textproto.MIMEHeader
	Content []byte
}
