package parser

import (
	"regexp"
	"strconv"
	"strings"

	moviedbwithcsv "github.com/flpcastro/movie-db-with-csv"
	"github.com/flpcastro/movie-db-with-csv/errors"
)

func ParseLine(line []string) (moviedbwithcsv.Movie, error) {
	if len(line) < 3 {
		return moviedbwithcsv.Movie{}, errors.NewNonFormattedLine(line)
	}

	id, err := strconv.Atoi(line[0])
	if err != nil {
		return moviedbwithcsv.Movie{}, errors.NewNonValidID(line[0])
	}

	title := line[1]
	year := 0
	re, err := regexp.Compile("(.*)\\s*\\((.*)\\)")
	if err == nil {
		titleMatches := re.FindStringSubmatch(line[1])
		if len(titleMatches) == 3 {
			title = strings.Trim(titleMatches[1], " ")
			year, err = strconv.Atoi(titleMatches[2])
			if err != nil {
				// NOTHING
			}
		}
	}

	genres := strings.Split(strings.Trim(line[2], "\""), "|")
	if len(genres) == 1 && genres[0] == "" {
		genres = []string{}
	}

	return moviedbwithcsv.Movie{
		ID:     id,
		Title:  title,
		Year:   year,
		Genres: genres,
	}, nil
}
