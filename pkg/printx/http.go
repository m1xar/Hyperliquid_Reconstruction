package printx

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tidwall/pretty"
)

// PrintHTTPResponseJSON prints the content of an HTTP response in a beautifully formatted JSON.
// response: a pointer to the HTTP response to be formatted.
// header: a string containing the header or description of the response content.
func PrintHTTPResponseJSON(response *http.Response, header string) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	result := pretty.Pretty(bodyBytes)

	fmt.Printf("\n%v :\n%s\n", header, string(result))
}
