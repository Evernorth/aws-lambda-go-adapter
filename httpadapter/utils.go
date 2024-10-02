package httpadapter

import (
	"net/http"
	"strings"
)

func formatHeaders(httpRequest *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range httpRequest.Header {
		headers[key] = strings.Join(values, ",")
	}
	return headers
}
