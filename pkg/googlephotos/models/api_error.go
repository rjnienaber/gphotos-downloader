package models

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ApiError struct {
	Url        string
	StatusCode int
	Headers    map[string][]string
	Raw        string
	Err        error
	Response   apiErrorResponse
}

func (a ApiError) Error() string {
	if a.Err != nil {
		return a.Err.Error()
	}

	return fmt.Sprintf("Error (url: '%s', status code: %d) %s", a.Url, a.StatusCode, a.Raw)
}

type apiError struct {
	Error apiErrorResponse `json:"error"`
}

type apiErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func ParseErrorReponse(resp *http.Response, responseBody []byte) (err ApiError) {
	err = ApiError{
		Url:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Raw:        string(responseBody),
	}

	var errorResponse apiError
	jsonErr := json2.Unmarshal(responseBody, &errorResponse)
	if jsonErr != nil {
		msg := fmt.Sprintf("failed deserialising error response: '%s'", string(responseBody))
		err.Err = errors.New(msg)
		return
	}
	err.Response = errorResponse.Error
	return
}
