package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zanmato/tspcompat/internal/sign"
)

func main() {
	// Read required parameters
	loadURL := flag.String("load-from", "", "URL to the API which to load data from")
	listen := flag.String("listen", ":8000", "Listen address")

	flag.Parse()

	dsn, exists := os.LookupEnv("DB_DSN")
	if !exists {
		log.Fatal("DB_DSN environment variable not set")
	}

	// Open DB
	connConf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("unable to parse database DSN: %v", err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), connConf)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Load data?
	if loadURL != nil && *loadURL != "" {
		log.Printf("loading data from %s", *loadURL)
		if err := loadData(dbpool, *loadURL); err != nil {
			log.Fatalf("failed loading data from %s: %v", *loadURL, err)
		}
	}

	srv := http.Server{
		Addr: *listen,
	}

	// Add a handler returning signs one by one
	http.HandleFunc("/signs.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		params := []interface{}{}
		where := []string{"TRUE"}

		// Parse query params
		if r.URL.Query().Has("changed_at") {
			changedAt, err := time.Parse("2006-01-02 15:04", r.URL.Query().Get("changed_at"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			params = append(params, changedAt)
			where = append(where, "updated_at > $1")
		}

		start := time.Now()

		rows, err := dbpool.Query(
			r.Context(),
			fmt.Sprintf(
				`SELECT 
				json_build_object(
					'id', signs.id,
					'deleted', signs.deleted,
					'unusual', signs.unusual,
					'ref_id', signs.ref_id,
					'video_url', signs.video_url,
					'updated_at', signs.updated_at,
					'description', signs.description,
					'frequency', signs.frequency,
					'tags', (
						SELECT json_agg(
							json_build_object(
								'id', tags.id,
								'tag', tags.name
							)
						)
						FROM tags
						INNER JOIN signs_tags ON signs_tags.sign_id = signs.id AND signs_tags.tag_id = tags.id
					),
					'words', (
						SELECT json_agg(
							json_build_object(
								'id', words.id,
								'word', words.word
							)
						)
						FROM words
						WHERE words.sign_id = signs.id
					),
					'examples', (
						SELECT json_agg(
							json_build_object(
								'id', examples.id,
								'video_url', examples.video_url,
								'description', examples.description
							)
						)
						FROM examples
						WHERE examples.sign_id = signs.id
					)
				)
				FROM signs
				WHERE %s
				ORDER BY signs.id`,
				strings.Join(where, " AND "),
			),
			params...,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error querying database: %v", err)
			return
		}
		defer rows.Close()

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("["))

		var (
			b []byte
			i int
		)
		for rows.Next() {
			if err := rows.Scan(&b); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("error scanning database row: %v", err)
				return
			}

			if i > 0 {
				w.Write([]byte(","))
			}

			w.Write(b)
			i++
		}

		w.Write([]byte("]"))

		log.Printf("served %d signs in %s", i, time.Since(start))
	})

	// Gracefully shutdown the server
	connClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
		close(connClosed)
	}()

	log.Printf("listening on %s", *listen)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server quit unexpectedly: %v", err)
	}

	<-connClosed
}

func loadData(db *pgxpool.Pool, dataURL string) error {
	u, err := url.Parse(dataURL)
	if err != nil {
		return fmt.Errorf("unable to parse data URL: %w", err)
	}

	res, err := http.Get(u.String())
	if err != nil {
		log.Fatalf("error creating request to %s: %v", u, err)
	}
	defer res.Body.Close()

	ctx := context.Background()

	// Remove all the existing data
	if _, err := db.Exec(
		ctx,
		`TRUNCATE TABLE signs CASCADE`,
	); err != nil {
		return fmt.Errorf("error truncating signs table: %w", err)
	}

	dec := json.NewDecoder(res.Body)

	// Read the opening bracket
	if _, err := dec.Token(); err != nil {
		return err
	}

	batch := &pgx.Batch{}

	// Read the array, record by record
	var s sign.NewSign
	for dec.More() {
		if err := dec.Decode(&s); err != nil {
			return err
		}

		if l := batch.Len(); l >= 1000 {
			res := db.SendBatch(ctx, batch)
			if err := res.Close(); err != nil {
				return fmt.Errorf("error executing batch: %w", err)
			}
			batch = &pgx.Batch{}

			log.Printf("inserted %d signs", l)
		}

		// Insert the sign
		batch.Queue(
			`INSERT INTO signs (id, ref_id, video_url, description, unusual, frequency, deleted)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			s.ID, s.RefID, s.VideoURL, s.Description, s.Unusual, s.Frequency, s.Deleted,
		)

		// Insert the tags
		for _, t := range s.Tags {
			batch.Queue(
				`INSERT INTO tags (id, name)
				VALUES ($1, $2)
				ON CONFLICT (id) DO NOTHING`,
				t.ID, t.Tag,
			)

			batch.Queue(
				`INSERT INTO signs_tags (sign_id, tag_id)
				VALUES ($1, $2)
				ON CONFLICT (sign_id, tag_id) DO NOTHING`,
				s.ID, t.ID,
			)
		}

		// Insert the words
		for _, w := range s.Words {
			batch.Queue(
				`INSERT INTO words (id, sign_id, word)
				VALUES ($1, $2, $3)
				ON CONFLICT (id) DO NOTHING`,
				w.ID, s.ID, w.Word,
			)
		}

		// Insert the examples
		for _, e := range s.Examples {
			batch.Queue(
				`INSERT INTO examples (id, sign_id, video_url, description)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id) DO NOTHING`,
				e.ID, s.ID, e.VideoURL, e.Description,
			)
		}
	}

	// Send the last batch
	if l := batch.Len(); l > 0 {
		res := db.SendBatch(ctx, batch)
		if err := res.Close(); err != nil {
			return fmt.Errorf("error executing batch: %w", err)
		}

		log.Printf("inserted %d signs", l)
	}

	return nil
}
