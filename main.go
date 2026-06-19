package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func healthZ_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metrics_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) reset_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	type returnVals struct {
		Error string `json:"error"`
	}

	log.Printf("%s: %s", msg, err)
	respBody := returnVals{
		Error: msg,
	}

	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Error marshalling JSON", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func validation_handler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		BodyVal string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Error decoding parameters", err)
		return
	}
	if len(params.BodyVal) > 140 {
		err := errors.New("Length of body exceeds 140")
		respondWithError(w, 400, "Chirp is too long", err)
		return
	}

	cleaned_body := badWordReplacer(params.BodyVal)
	type returnVals struct {
		Body string `json:"cleaned_body"`
	}
	respBody := returnVals{
		Body: cleaned_body,
	}
	respondWithJSON(w, 200, respBody)

}

func badWordReplacer(text string) string {
	badWordList := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for idx, word := range words {
		if slices.Contains(badWordList, strings.ToLower(word)) {
			words[idx] = "****"
		}
	}
	return strings.Join(words, " ")
}

func main() {
	apiCfg := apiConfig{}
	smu := http.NewServeMux()
	smu.HandleFunc("GET /api/healthz", healthZ_handler)
	smu.HandleFunc("GET /admin/metrics", apiCfg.metrics_handler)
	smu.HandleFunc("POST /admin/reset", apiCfg.reset_handler)
	smu.HandleFunc("POST /api/validate_chirp", validation_handler)
	smu.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	server := http.Server{
		Addr:    ":8080",
		Handler: smu,
	}
	server.ListenAndServe()
}
