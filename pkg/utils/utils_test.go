package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildUrlNoQueryString(t *testing.T) {
	newUrl, err := BuildUrl("http://www.google.com", map[string][]string{})
	assert.NoError(t, err)
	assert.Equal(t, "http://www.google.com", newUrl.String())
}

func TestBuildUrlWithQueryString(t *testing.T) {
	queryString := map[string][]string{}
	queryString["pageSize"] = []string{"23"}
	queryString["pageToken"] = []string{"abc"}

	newUrl, err := BuildUrl("http://www.google.com", queryString)
	assert.NoError(t, err)
	assert.Equal(t, "http://www.google.com?pageSize=23&pageToken=abc", newUrl.String())
}

func TestBuildUrlWithArrayQueryString(t *testing.T) {
	queryString := map[string][]string{}
	queryString["ids"] = []string{"23", "abc"}

	newUrl, err := BuildUrl("http://www.google.com", queryString)
	assert.NoError(t, err)
	assert.Equal(t, "http://www.google.com?ids=23&ids=abc", newUrl.String())
}
