package utils

import (
	"io"
	"net/url"
)

func CheckClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}

func BuildUrl(baseUrl string, queryString map[string][]string) (*url.URL, error) {
	fullUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	q := fullUrl.Query()
	for key, value := range queryString {
		for _, v := range value {
			if v != "" {
				q.Add(key, v)
			}
		}
	}
	fullUrl.RawQuery = q.Encode()
	return fullUrl, nil
}
