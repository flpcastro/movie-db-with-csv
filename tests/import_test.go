package tests

import (
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	loadondb "github.com/flpcastro/movie-db-with-csv"
	"github.com/flpcastro/movie-db-with-csv/importer"
	"github.com/stretchr/testify/assert"
)

type Importer interface {
	ImportMovies(ctx context.Context, reader *csv.Reader) error
}

func TestInsertOneMovie(t *testing.T) {
	importers := map[string]Importer{
		"sequentially": importer.Sequentially{
			Writer:    writer,
			ChunkSize: 100,
		},
		"concurrently": importer.Concurrently{
			Writer:    writer,
			ChunkSize: 100,
			Workers:   2,
			Timeout:   10 * time.Second,
		},
	}

	for title, imp := range importers {
		t.Run(title, func(t *testing.T) {
			teardown()
			file := "\"movieId\",\"title\",\"genres\"\n10,\"GoldenEye (1995)\",\"Action|Adventure|Thriller\""

			reader := csv.NewReader(strings.NewReader(file))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := imp.ImportMovies(ctx, reader)
			assert.Nil(t, err)

			expected := loadondb.Movie{
				ID:     10,
				Title:  "GoldenEye",
				Year:   1995,
				Genres: []string{"Action", "Adventure", "Thriller"},
			}

			movie, err := findMovieByID(10)
			assert.Nil(t, err)
			assert.Equal(t, expected, movie)

			count := countMovies()
			assert.Equal(t, 1, count)
		})
	}
}

func TestFailToInsertOneMovieSaveTheOthers(t *testing.T) {
	importers := map[string]Importer{
		"sequentially": importer.Sequentially{
			Writer:    writer,
			ChunkSize: 2,
		},
		"concurrently": importer.Concurrently{
			Writer:    writer,
			ChunkSize: 2,
			Workers:   2,
			Timeout:   10 * time.Second,
		},
	}

	for title, imp := range importers {
		t.Run(title, func(t *testing.T) {
			teardown()

			file := `"movieId","title","genres"
2,"Jumanji (1995)","Adventure|Children|Fantasy"
xpto,"Toy Story (1995)","Adventure|Children|Fantasy"
4,"GoldenEye (1995)","Action|Adventure|Thriller"`

			reader := csv.NewReader(strings.NewReader(file))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := imp.ImportMovies(ctx, reader)
			assert.Error(t, err)

			count := countMovies()
			assert.Equal(t, 2, count)
		})
	}
}

func TestFailToInsertOneMovieDoesNotAffectTheOthers(t *testing.T) {
	importers := map[string]Importer{
		"sequentially": importer.Sequentially{
			Writer:    writer,
			ChunkSize: 2,
		},
		"concurrently": importer.Concurrently{
			Writer:    writer,
			ChunkSize: 2,
			Workers:   1,
			Timeout:   30 * time.Second,
		},
	}

	for title, imp := range importers {
		t.Run(title, func(t *testing.T) {
			teardown()

			file := `"movieId","title","genres"
2,"Jumanji (1995)","Adventure|Children|Fantasy"
4,"GoldenEye (1995)","Action|Adventure|Thriller"
xpto,"Toy Story (1995)","Adventure|Children|Fantasy"`

			reader := csv.NewReader(strings.NewReader(file))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			err := imp.ImportMovies(ctx, reader)
			assert.Error(t, err)

			count := countMovies()
			assert.Equal(t, 2, count)

			expected := loadondb.Movie{
				ID:     4,
				Title:  "GoldenEye",
				Year:   1995,
				Genres: []string{"Action", "Adventure", "Thriller"},
			}
			movie, err := findMovieByID(4)
			assert.NoError(t, err)
			assert.Equal(t, expected, movie)
		})
	}
}

func TestIdempotencyBasedOnID(t *testing.T) {
	importers := map[string]Importer{
		"sequentially": importer.Sequentially{
			Writer:    writer,
			ChunkSize: 1,
		},
		"concurrently": importer.Concurrently{
			Writer:    writer,
			ChunkSize: 3,
			Workers:   2,
			Timeout:   30 * time.Second,
		},
	}

	for title, imp := range importers {
		t.Run(title, func(t *testing.T) {
			teardown()
			file := `"movieId","title","genres"
2,"Jumanji (1995)","Adventure|Children|Fantasy"`

			reader := csv.NewReader(strings.NewReader(file))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := imp.ImportMovies(ctx, reader)
			assert.Nil(t, err)

			expected1 := loadondb.Movie{
				ID:     2,
				Title:  "Jumanji",
				Year:   1995,
				Genres: []string{"Adventure", "Children", "Fantasy"},
			}

			movie, err := findMovieByID(2)
			assert.NoError(t, err)
			assert.Equal(t, expected1, movie)

			file = `"movieId","title","genres"
2,"Jumanji (1995)","Adventure|Children|Fantasy"
11,"GoldenEye (1995)","Action|Adventure|Thriller"`

			reader = csv.NewReader(strings.NewReader(file))
			err = imp.ImportMovies(ctx, reader)
			assert.Nil(t, err)

			expected2 := loadondb.Movie{
				ID:     11,
				Title:  "GoldenEye",
				Year:   1995,
				Genres: []string{"Action", "Adventure", "Thriller"},
			}

			count := countMovies()
			assert.Equal(t, 2, count)

			movie, err = findMovieByID(2)
			assert.NoError(t, err)
			assert.Equal(t, expected1, movie)

			movie, err = findMovieByID(11)
			assert.NoError(t, err)
			assert.Equal(t, expected2, movie)
		})
	}
}
