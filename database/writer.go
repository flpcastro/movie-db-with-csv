package database

import (
	"context"
	"fmt"

	loadondb "github.com/flpcastro/movie-db-with-csv"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Writer struct {
	Pool *pgxpool.Pool
}

func (repo *Writer) InsertMovies(ctx context.Context, movies []loadondb.Movie) error {
	query := "INSERT INTO movies (id, title, year, genres) VALUES"

	var values []interface{}
	mult := 1
	for i, m := range movies {
		query += fmt.Sprintf(" ($%d, $%d, $%d, $%d)", mult+i, mult+i+1, mult+i+2, mult+i+3)
		if i < len(movies)-1 {
			query += ", "
		}

		values = append(values, m.ID, m.Title, m.Year, m.Genres)
		mult += 3
	}

	query += " ON CONFLICT (id) DO UPDATE SET created_at = NOW();"

	_, err := repo.Pool.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("cannot save movies with query=%s: %w", query, err)
	}

	return nil
}
