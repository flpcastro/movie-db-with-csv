package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flpcastro/movie-db-with-csv/database"
	"github.com/jackc/pgx/v4/pgxpool"
)

var writer database.Writer

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbConnString := os.Getenv("DB_CONN")
	if dbConnString == "" {
		dbConnString = "postgresql://root:root@localhost:5432/movies"
	}

	conn, err := pgxpool.Connect(ctx, dbConnString)
	if err != nil {
		fmt.Printf("could not connect to database: " + err.Error())
		os.Exit(1)
	}

	writer = database.Writer{Pool: conn}
	if err := clearDatabase(); err != nil {
		fmt.Printf("could not clear database: " + err.Error())
		os.Exit(1)
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}
