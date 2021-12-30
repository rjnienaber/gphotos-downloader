package models

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchOptionsSerialization(t *testing.T) {
	searchOptions := SearchOptions{
		Filters: SearchFilters{
			DateFilter: SearchDateFilter{
				Ranges: []SearchDateRange{
					{
						StartDate: SearchDate{
							Year:  2021,
							Month: 12,
							Day:   22,
						},
						EndDate: SearchDate{
							Year:  2021,
							Month: 12,
							Day:   31,
						},
					},
				},
			},
		},
		Size: 2,
	}

	jsonBytes, err := searchOptions.Serialize()
	assert.NoError(t, err)
	json, err := io.ReadAll(jsonBytes)
	assert.NoError(t, err)
	expected := `{
  "pageSize": 2,
  "filters": {
    "dateFilter": {
      "ranges": [
        {
          "startDate": {
            "year": 2021,
            "month": 12,
            "day": 22
          },
          "endDate": {
            "year": 2021,
            "month": 12,
            "day": 31
          }
        }
      ]
    }
  }
}`
	assert.Equal(t, expected, string(json))
}
