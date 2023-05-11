package sign

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncClient struct {
	db         *pgxpool.Pool
	httpClient *http.Client
}

func NewSyncClient(db *pgxpool.Pool) (*SyncClient, error) {
	return &SyncClient{
		db: db,
		httpClient: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}, nil
}

func (c *SyncClient) Sync(ctx context.Context, dataURL string) error {
	u, err := url.Parse(dataURL)
	if err != nil {
		return fmt.Errorf("unable to parse data URL: %w", err)
	}

	// Remove all the existing data
	if _, err := c.db.Exec(
		ctx,
		`TRUNCATE TABLE signs RESTART IDENTITY CASCADE`,
	); err != nil {
		return fmt.Errorf("error truncating signs table: %w", err)
	}

	if _, err := c.db.Exec(
		ctx,
		`INSERT INTO categories (name, slug)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO NOTHING`,
		"Odefinierade",
		"odefinierade",
	); err != nil {
		return fmt.Errorf("error inserting default category: %w", err)
	}

	for i := 1; i < 1000; i++ {
		q := u.Query()
		q.Set("page", strconv.Itoa(i))
		u.RawQuery = q.Encode()

		ins, err := c.loadFromURL(ctx, u)
		if err != nil {
			return fmt.Errorf("error loading data from %s: %w", u.String(), err)
		}

		if ins == 0 {
			break
		}
	}

	return nil
}

func (c *SyncClient) loadFromURL(ctx context.Context, dataURL *url.URL) (int, error) {
	j := 0

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dataURL.String(), http.NoBody)
	if err != nil {
		return j, fmt.Errorf("error creating request to %s: %v", dataURL, err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return j, fmt.Errorf("error creating request to %s: %v", dataURL, err)
	}
	defer res.Body.Close()

	// Read the opening bracket
	dec := json.NewDecoder(res.Body)
	if _, err := dec.Token(); err != nil {
		return j, err
	}

	// Read the array, record by record
	batch := &pgx.Batch{}
	baseURL := dataURL.Scheme + "://" + dataURL.Host + "/"
	var s NewSign
	for dec.More() {
		if err := dec.Decode(&s); err != nil {
			return j, err
		}

		id, err := strconv.Atoi(strings.TrimPrefix(s.ID, "0"))
		if err != nil {
			return j, fmt.Errorf("error parsing sign id: %w", err)
		}

		updatedAt, err := time.ParseInLocation("2006-01-02", s.LastUpdate, time.Local)
		if err != nil {
			return j, fmt.Errorf("error parsing sign last update: %w", err)
		}

		// Insert the sign
		batch.Queue(
			`INSERT INTO signs (id, updated_at, video_url, description, unusual, frequency, transcription, vocable, hidden_words)
			VALUES ($1, $2, $3, $4, NULLIF($5, '') IS NOT NULL, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ARRAY['']))`,
			id,
			updatedAt,
			baseURL+s.Movie,
			s.Description,
			s.Frequency,
			s.Transcription,
			s.Glosa,
			strings.Split(s.HiddenAlsoMeans, ", "),
		)

		// Insert words
		words := strings.Split(s.AlsoMeans, ", ")
		for _, w := range words {
			w = strings.TrimSpace(w)
			if w == "" {
				continue
			}

			batch.Queue(
				`INSERT INTO words (sign_id, word)
				VALUES ($1, $2)`,
				s.ID,
				w,
			)
		}

		// Insert the categories
		for _, t := range s.Categories {
			batch.Queue(
				`INSERT INTO categories (name, slug)
				VALUES ($1, $2)
				ON CONFLICT (slug) DO NOTHING`,
				t.Name, t.Slug,
			)

			batch.Queue(
				`INSERT INTO signs_categories (sign_id, category_id)
				SELECT
				$1, id
				FROM categories
				WHERE categories.slug = $2
				ON CONFLICT (sign_id, category_id) DO NOTHING`,
				s.ID, t.Slug,
			)
		}

		if len(s.Categories) == 0 {
			batch.Queue(
				`INSERT INTO signs_categories (sign_id, category_id)
				SELECT
				$1, id
				FROM categories
				WHERE categories.slug = 'odefinierade'`,
				s.ID,
			)
		}

		// Insert the phrases
		for _, e := range s.Phrases {
			batch.Queue(
				`INSERT INTO phrases (sign_id, video_url, phrase)
				VALUES ($1, $2, $3)`,
				s.ID,
				baseURL+e.Movie,
				e.Phrase,
			)
		}

		j++
	}

	// Send the batch
	if l := batch.Len(); l > 0 {
		res := c.db.SendBatch(ctx, batch)
		if err := res.Close(); err != nil {
			return j, fmt.Errorf("error executing batch: %w", err)
		}
	}

	return j, nil
}
