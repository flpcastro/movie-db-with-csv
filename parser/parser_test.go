package parser

import (
	"testing"

	moviedbwithcsv "github.com/flpcastro/movie-db-with-csv"
	"github.com/flpcastro/movie-db-with-csv/errors"
	"github.com/stretchr/testify/assert"
)

func Test_ParseValidLines(t *testing.T) {
	scenarios := map[string]struct {
		line     []string
		expected moviedbwithcsv.Movie
	}{
		"movie with year": {
			line: []string{"1", "Toy Story (1995)", "Adventure|Animation|Children|Comedy|Fantasy"},
			expected: moviedbwithcsv.Movie{
				ID:     1,
				Title:  "Toy Story",
				Year:   1995,
				Genres: []string{"Adventure", "Animation", "Children", "Comedy", "Fantasy"},
			},
		},
		"movie with portuguese name": {
			line: []string{"1", "Toy Story (história de brinquedos) (1995)", "Adventure|Animation|Children|Comedy|Fantasy"},
			expected: moviedbwithcsv.Movie{
				ID:     1,
				Title:  "Toy Story (história de brinquedos)",
				Year:   1995,
				Genres: []string{"Adventure", "Animation", "Children", "Comedy", "Fantasy"},
			},
		},
		"movie without year": {
			line: []string{"2", "Toy Story", "Adventure|Animation|Children|Comedy|Fantasy"},
			expected: moviedbwithcsv.Movie{
				ID:     2,
				Title:  "Toy Story",
				Year:   0,
				Genres: []string{"Adventure", "Animation", "Children", "Comedy", "Fantasy"},
			},
		},
		"movie without genres": {
			line: []string{"3", "Toy Story (2000)", ""},
			expected: moviedbwithcsv.Movie{
				ID:     3,
				Title:  "Toy Story",
				Year:   2000,
				Genres: []string{},
			},
		},
		"movie with bad formatted year": {
			line: []string{"1", "Toy Story (2000a)", "Adventure|Animation|Children|Comedy|Fantasy"},
			expected: moviedbwithcsv.Movie{
				ID:     1,
				Title:  "Toy Story",
				Year:   0,
				Genres: []string{"Adventure", "Animation", "Children", "Comedy", "Fantasy"},
			},
		},
	}

	for title, value := range scenarios {
		t.Run(title, func(t *testing.T) {
			found, err := ParseLine(value.line)
			assert.NoError(t, err)
			assert.Equal(t, value.expected, found)
		})
	}
}

func Test_ParseInvalidLines(t *testing.T) {
	scenarios := map[string]struct {
		line              []string
		expectedErrorCode int
	}{
		"line missing column": {
			line:              []string{"1", "Toy Story (1995)"},
			expectedErrorCode: errors.NonFormattedLine,
		},
		"line with non valid id": {
			line:              []string{"", "Toy Story (1995)", "Adventure|Animation|Children|Comedy|Fantasy"},
			expectedErrorCode: errors.NonValidID,
		},
		"line with non numeric id": {
			line:              []string{"a", "Toy Story (1995)", "Adventure|Animation|Children|Comedy|Fantasy"},
			expectedErrorCode: errors.NonValidID,
		},
	}

	for title, value := range scenarios {
		t.Run(title, func(t *testing.T) {
			_, err := ParseLine(value.line)
			assert.Error(t, err)
			e, ok := err.(errors.Error)
			if !ok {
				assert.Fail(t, "error should be type custom")
			}
			assert.Equal(t, value.expectedErrorCode, e.Code)
		})
	}
}
