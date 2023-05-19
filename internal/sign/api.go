package sign

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

var wordsRe = regexp.MustCompile(`(?m)((?:\b|\p{L})[^\s]+(?:\b|\p{L}))`)

type APIController struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAPIController(db *pgxpool.Pool, logger *zap.Logger) *APIController {
	return &APIController{
		db:     db,
		logger: logger,
	}
}

func (c *APIController) WordIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	where := []string{"TRUE"}
	orderBy := `ORDER BY LOWER(word->>2) COLLATE "se-SE-x-icu"`
	params := []interface{}{}
	pc := 1

	q := r.URL.Query()
	if q.Get("category_id") != "" {
		categoryID, err := strconv.Atoi(q.Get("category_id"))
		if err != nil {
			c.logger.Error("unable to parse category_id", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		where = append(where, fmt.Sprintf("words_view.categories @> $%d", pc))
		params = append(params, []int{categoryID})
		pc++
	}

	// Build search query
	if searchQuery := q.Get("q"); searchQuery != "" {
		var (
			w           string
			words       []string
			wordsPrefix []string
		)
		for _, match := range wordsRe.FindAllString(searchQuery, -1) {
			w = "'" + strings.ReplaceAll(match, "'", "\\'") + "'"
			if len(match) == 1 {
				words = append(words, w)
				wordsPrefix = append(wordsPrefix, w)
				continue
			}

			words = append(words, w)
			wordsPrefix = append(wordsPrefix, w+":*")
		}

		query := "(" + strings.Join(words, " <-> ") + ") | (" + strings.Join(wordsPrefix, " & ") + ")"

		params = append(params, query)
		where = append(where, fmt.Sprintf("words_view.tsv @@ to_tsquery('swedish', $%d)", pc))
		orderBy = fmt.Sprintf("ORDER BY ts_rank_cd(words_view.tsv, to_tsquery('swedish', $%d), 8|1) DESC", pc)
	}

	// Query
	rows, err := c.db.Query(
		r.Context(),
		fmt.Sprintf(
			`SELECT
			words_view.word
			FROM words_view
			WHERE %s
			%s`,
			strings.Join(where, " AND "),
			orderBy,
		),
		params...,
	)
	if err != nil {
		c.logger.Error("unable to query words", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Write response
	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Cache-Control", "public, max-age=86400")
	w.Write([]byte("["))

	var (
		b []byte
		i int
	)
	for rows.Next() {
		if err := rows.Scan(&b); err != nil {
			c.logger.Error("unable to scan word", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if i > 0 {
			w.Write([]byte(","))
		}

		w.Write(b)
		i++
	}
	w.Write([]byte("]"))
}

func (c *APIController) CategoryIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Query
	rows, err := c.db.Query(
		r.Context(),
		`SELECT
		categories_view.category
		FROM categories_view
		ORDER BY LOWER(category->>1) COLLATE "se-SE-x-icu"`,
	)
	if err != nil {
		c.logger.Error("unable to query categories", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Write response
	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte("["))

	var (
		b []byte
		i int
	)
	for rows.Next() {
		if err := rows.Scan(&b); err != nil {
			c.logger.Error("unable to scan category", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if i > 0 {
			w.Write([]byte(","))
		}

		w.Write(b)
		i++
	}
	w.Write([]byte("]"))
}

func (c *APIController) SignShow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		res []byte
	)

	if err := c.db.QueryRow(
		r.Context(),
		`SELECT
		sign
		FROM signs_view
		WHERE id = $1`,
		ps.ByName("id"),
	).Scan(&res); err != nil {
		c.logger.Error("unable to query sign", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Cache-Control", "public, max-age=3600")
	w.Write(res)
}
