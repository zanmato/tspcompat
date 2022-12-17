package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/zanmato/tspcompat/internal/sign"
)

func main() {
	// Read required parameters
	oldURL := flag.String("old-url", "", "URL of the old API")
	newURL := flag.String("new-url", "", "URL of the new API")
	listen := flag.String("listen", ":8000", "Listen address")

	flag.Parse()

	if newURL == nil || oldURL == nil {
		flag.Usage()
		os.Exit(1)
	}

	// Read the cached indexes if they exist
	ti, err := os.ReadFile("tag_index.json")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error reading tag index: %v", err)
	}

	si, err := os.ReadFile("sign_index.json")
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error reading sign index: %v", err)
	}

	var (
		tagIndex  map[string]int
		signIndex map[int]int
	)
	if ti != nil && si != nil {
		if err := json.Unmarshal(ti, &tagIndex); err != nil {
			log.Fatalf("error unmarshaling tag index: %v", err)
		}

		if err := json.Unmarshal(si, &signIndex); err != nil {
			log.Fatalf("error unmarshaling sign index: %v", err)
		}
	} else {
		res, err := http.Get(*oldURL)
		if err != nil {
			log.Fatalf("error creating request to %s: %v", *oldURL, err)
		}

		tagIndex, signIndex, err = sign.BuildIndexes(res.Body)
		res.Body.Close()

		if err != nil {
			log.Fatalf("error building tag index: %v", err)
		}

		// Write the indexes to disk
		ti, err := json.Marshal(tagIndex)
		if err != nil {
			log.Fatalf("error marshaling tag index: %v", err)
		}

		if err := os.WriteFile("tag_index.json", ti, 0644); err != nil {
			log.Fatalf("error writing tag index: %v", err)
		}

		si, err := json.Marshal(signIndex)
		if err != nil {
			log.Fatalf("error marshaling sign index: %v", err)
		}

		if err := os.WriteFile("sign_index.json", si, 0644); err != nil {
			log.Fatalf("error writing sign index: %v", err)
		}
	}

	srv := http.Server{
		Addr: *listen,
	}

	// Add a handler proxying the new API format to the old API format
	http.HandleFunc("/signs.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		u := *newURL
		if r.URL.RawQuery != "" {
			u += "?" + r.URL.RawQuery
		}

		res, err := http.Get(u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error creating request to %s: %v", *newURL+r.URL.Query().Encode(), err)
			return
		}
		defer res.Body.Close()

		w.Header().Add("Content-Type", "application/json")
		if err := sign.TransformJSONStream(w, res.Body, tagIndex, signIndex); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed transforming the request to %s: %v", *newURL+r.URL.Query().Encode(), err)
			return
		}
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
