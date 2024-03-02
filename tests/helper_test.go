package tests

import (
	"context"
	"time"

	loadondb "github.com/flpcastro/movie-db-with-csv"
)

func teardown() {
	if err := clearDatabase(); err != nil {
		panic(err)
	}
}

func clearDatabase() error {
	conn := writer.Pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := conn.Exec(ctx, "TRUNCATE movies")
	if err != nil {
		return err
	}

	return nil
}

func findMovieByID(id int) (loadondb.Movie, error) {
	conn := writer.Pool

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var movie loadondb.Movie
	row := conn.QueryRow(ctx, "SELECT id, title, year, genres FROM movies WHERE id = $1", id)
	row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Genres,
	)

	return movie, nil
}

func countMovies() int {
	conn := writer.Pool

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var counter int
	conn.QueryRow(ctx, "SELECT count(*) FROM movies").Scan(&counter)

	return counter
}
