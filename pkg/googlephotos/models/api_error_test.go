package models

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseApiError(t *testing.T) {
	json := `{
  "error": {
    "code": 400,
    "message": "Range 0 (year: 2021\nmonth: 12\nday: 22\n - ): Both start and end dates must follow the same format.",
    "status": "INVALID_ARGUMENT"
  }
}`

	testUrl, _ := url.Parse("http://google.com")
	response := http.Response{
		Request:    &http.Request{URL: testUrl},
		Status:     "Bad Request",
		StatusCode: 400,
		Header:     map[string][]string{"Content-Type": {"application/json"}},
	}

	err := ParseErrorReponse(&response, []byte(json))

	assert.Nil(t, err.Err)
	assert.Equal(t, json, err.Raw)
	assert.Equal(t, 400, err.Response.Code)
	assert.Equal(t, "Range 0 (year: 2021\nmonth: 12\nday: 22\n - ): Both start and end dates must follow the same format.", err.Response.Message)
	assert.Equal(t, "INVALID_ARGUMENT", err.Response.Status)

	assert.Equal(t, "http://google.com", err.Url)
	assert.Equal(t, response.Header, err.Headers)

	expectedMessage := fmt.Sprintf("Error (url: 'http://google.com', status code: 400) %s", json)
	assert.Equal(t, expectedMessage, err.Error())
}
