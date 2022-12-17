package sign

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"
)

// NewSign represents the new sign format that needs to be transformed.
type NewSign struct {
	Deleted  bool   `json:"deleted"`
	Unusual  bool   `json:"unusual"`
	ID       int    `json:"id"`
	RefID    int    `json:"ref_id"`
	VideoURL string `json:"video_url"`
	// UpdatedAt   bool    `json:"updated_at"`
	Description string  `json:"description"`
	Frequency   *string `json:"frequency"`
	Tags        []struct {
		ID  int    `json:"id"`
		Tag string `json:"tag"`
	} `json:"tags"`
	Examples []struct {
		ID          int    `json:"id"`
		VideoURL    string `json:"video_url"`
		Description string `json:"description"`
	} `json:"examples"`
	Words []struct {
		ID   int    `json:"id"`
		Word string `json:"word"`
	} `json:"words"`
}

type Tag struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

type Example struct {
	ID          string `json:"id"`
	VideoURL    string `json:"video_url"`
	Description string `json:"description"`
}

type Word struct {
	ID   string `json:"id"`
	Word string `json:"word"`
}

// OldSign represents the old sign format.
type OldSign struct {
	Deleted     bool      `json:"deleted"`
	Unusual     bool      `json:"unusual"`
	ID          string    `json:"id"`
	RefID       string    `json:"ref_id"`
	VideoURL    string    `json:"video_url"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
	Frequency   *string   `json:"frequency"`
	Tags        []Tag     `json:"tags"`
	Examples    []Example `json:"examples"`
	Words       []Word    `json:"words"`
}

func (s NewSign) MarshalJSON() ([]byte, error) {
	r := OldSign{
		Deleted:     s.Deleted,
		Unusual:     s.Unusual,
		ID:          strconv.Itoa(s.ID),
		RefID:       strconv.Itoa(s.RefID),
		VideoURL:    s.VideoURL,
		UpdatedAt:   time.Now(),
		Description: s.Description,
		Frequency:   s.Frequency,
	}

	r.Tags = make([]Tag, len(s.Tags))
	for i, t := range s.Tags {
		r.Tags[i] = Tag{
			ID:  strconv.Itoa(t.ID),
			Tag: t.Tag,
		}
	}

	r.Examples = make([]Example, len(s.Examples))
	for i, e := range s.Examples {
		r.Examples[i] = Example{
			ID:          strconv.Itoa(e.ID),
			VideoURL:    e.VideoURL,
			Description: e.Description,
		}
	}

	r.Words = make([]Word, len(s.Words))
	for i, w := range s.Words {
		r.Words[i] = Word{
			ID:   strconv.Itoa(w.ID),
			Word: w.Word,
		}
	}

	return json.Marshal(r)
}

// TransformJSONStream decodes a JSON stream consiting of a sequence of objects and encodes it to the given writer.
func TransformJSONStream(w io.Writer, r io.Reader, tagIndex map[string]int, signIndex map[int]int) error {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)

	// Read the opening bracket
	if _, err := dec.Token(); err != nil {
		return err
	}

	// Write the opening bracket
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}

	var (
		s NewSign
		i int
	)

	// Read the whole array
	for dec.More() {
		if err := dec.Decode(&s); err != nil {
			return err
		}

		// Map sign index
		if n, exists := signIndex[s.ID]; exists {
			s.ID = n
		}

		// Map tags
		for j, t := range s.Tags {
			if n, exists := tagIndex[strings.ToLower(t.Tag)]; exists {
				s.Tags[j].ID = n
			}
		}

		// Write a comma if this is not the first element
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}

		if err := enc.Encode(s); err != nil {
			return err
		}

		i++
	}

	// Write the closing bracket
	if _, err := w.Write([]byte{']'}); err != nil {
		return err
	}

	return nil
}

// BuildIndexes builds an index over all tags and signs in the given JSON stream.
func BuildIndexes(r io.Reader) (map[string]int, map[int]int, error) {
	tagIdx := map[string]int{}
	signIdx := map[int]int{}
	dec := json.NewDecoder(r)

	// Read the opening bracket
	if _, err := dec.Token(); err != nil {
		return nil, nil, err
	}

	var (
		s     OldSign
		refID int
	)

	// Read the whole array
	for dec.More() {
		if err := dec.Decode(&s); err != nil {
			return nil, nil, err
		}

		refID, _ = strconv.Atoi(s.RefID)
		if _, exists := signIdx[refID]; !exists {
			signIdx[refID], _ = strconv.Atoi(s.ID)
		}

		for _, t := range s.Tags {
			if _, exists := tagIdx[strings.ToLower(t.Tag)]; exists {
				continue
			}

			tagIdx[strings.ToLower(t.Tag)], _ = strconv.Atoi(t.ID)
		}
	}

	return tagIdx, signIdx, nil
}
