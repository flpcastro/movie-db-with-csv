package main

import (
	"context"
	"encoding/csv"
	"flag"
	"os"
	"time"

	"github.com/flpcastro/movie-db-with-csv/database"
	"github.com/flpcastro/movie-db-with-csv/importer"
	"github.com/jackc/pgx/v4/pgxpool"
)

var chunkSize *int
var sequentialExecution *bool
var concurrency *int
var filePath *string
var durationFlag *string

const defaultChunkSize = 100
const defaultConcurrency = 2
const defaultDuration = "1m"

type Importer interface {
	ImportMovies(ctx context.Context, reader *csv.Reader) error
}

func main() {
	chunkSize = flag.Int("n", defaultChunkSize, "number of movies to be persisted on database at once")
	concurrency = flag.Int("c", defaultConcurrency, "number of goroutines that persists movies")
	sequentialExecution = flag.Bool("s", false, "force the job to execute without concurrency")
	filePath = flag.String("f", "movie.csv", "csv file containing the list of movies")
	durationFlag = flag.String("t", defaultDuration, "timeout for database")
	flag.Parse()

	duration, err := time.ParseDuration(defaultDuration)
	if err != nil {
		panic(err)
	}
	if durationFlag != nil {
		var err error
		duration, err = time.ParseDuration(*durationFlag)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Open(*filePath)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	dbConnString := os.Getenv("DB_CONN")
	pool, err := pgxpool.Connect(ctx, dbConnString)
	if err != nil {
		panic(err)
	}
	writer := database.Writer{Pool: pool}

	reader := csv.NewReader(file)
	var im Importer
	im = importer.Concurrently{
		Writer:    writer,
		ChunkSize: *chunkSize,
		Workers:   *concurrency,
		Timeout:   duration,
	}
	if sequentialExecution != nil && *sequentialExecution == true {
		im = importer.Sequentially{
			Writer:    writer,
			ChunkSize: *chunkSize,
		}
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err = im.ImportMovies(ctx, reader)
	if err != nil {
		panic(err)
	}
}
