package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	loadondb "github.com/flpcastro/movie-db-with-csv"
	"github.com/flpcastro/movie-db-with-csv/database"
	"github.com/flpcastro/movie-db-with-csv/errors"
	"github.com/flpcastro/movie-db-with-csv/parser"
)

type Concurrently struct {
	Writer    database.Writer
	ChunkSize int
	Workers   int
	Timeout   time.Duration
}

func (c Concurrently) worker(chMovies chan []loadondb.Movie, done chan bool, chErr chan error) {
	defer func() {
		if r := recover(); r != nil {
			chErr <- fmt.Errorf("%s", r)
		}
	}()

	for {
		select {
		case movies, more := <-chMovies:
			if !more {
				done <- true
				return
			}

			cctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
			err := c.Writer.InsertMovies(cctx, movies)
			cancel()
			if err != nil {
				chErr <- err
			}
		}
	}
}

func (c Concurrently) ImportMovies(ctx context.Context, reader *csv.Reader) error {
	if c.Timeout.Nanoseconds() <= 0 {
		return errors.NewTimeoutRequired()
	}

	log.Printf("Inserting movies concurrently with chunkSize=%d and consumers=%d", c.ChunkSize, c.Workers)
	chMovies := make(chan []loadondb.Movie)
	chErr := make(chan error)
	done := make(chan bool)

	for i := 0; i < c.Workers; i++ {
		go c.worker(chMovies, done, chErr)
	}

	i := 0
	movies := []loadondb.Movie{}
	errors := errors.List{}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if i == 0 {
			i++
			continue
		}

		movie, err := parser.ParseLine(line)
		if err != nil {
			errors.AddError(err)
			continue
		}

		movies = append(movies, movie)
		i++

		if len(movies)%(c.ChunkSize) == 0 {
			chMovies <- movies
		}
	}

	if len(movies) > 0 {
		chMovies <- movies
	}

	close(chMovies)
	i = 0

RUN:
	for {
		select {
		case <-done:
			i++
			if i == c.Workers {
				break RUN
			}
		case err := <-chErr:
			errors.AddError(err)
		case <-ctx.Done():
			break RUN
		}
	}

	if errors.Len() == 0 {
		return nil
	}

	return &errors
}
