package models

import (
	"bytes"
	json2 "encoding/json"
	"io"
)

type SearchOptions struct {
	AlbumId string        `json:"albumId,omitempty"`
	Size    int           `json:"pageSize,omitempty"`
	Token   string        `json:"pageToken,omitempty"`
	Filters SearchFilters `json:"filters,omitempty"`
	OrderBy string        `json:"orderBy,omitempty"`
}

type SearchFilters struct {
	DateFilter SearchDateFilter `json:"dateFilter,omitempty"`
}

type SearchDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type SearchDateRange struct {
	StartDate SearchDate `json:"startDate,omitempty"`
	EndDate   SearchDate `json:"endDate,omitempty"`
}

type SearchDateFilter struct {
	Dates  []SearchDate      `json:"dates,omitempty"`
	Ranges []SearchDateRange `json:"ranges,omitempty"`
}

func (options *SearchOptions) Serialize() (reader io.Reader, err error) {
	json, err := json2.MarshalIndent(options, "", "  ")
	reader = bytes.NewReader(json)
	return
}
