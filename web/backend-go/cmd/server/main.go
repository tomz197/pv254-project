package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"backendgo/internal/http"
)

func main() {
	assets := os.Getenv("ASSETS_DIR")
	if assets == "" {
		assets = filepath.Clean("../backend/assets")
	}

	srv, err := httpapi.NewServer(assets)
	if err != nil { log.Fatalf("startup error: %v", err) }

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HandleHealth)
	mux.HandleFunc("/models", srv.HandleModels)
	mux.HandleFunc("/course/", srv.HandleCourse)
	mux.HandleFunc("/recommendations", srv.HandleRecommendations)
	mux.HandleFunc("/log_recommendation_feedback", srv.HandleLogRecommendationFeedback)
	mux.HandleFunc("/log_user_feedback", srv.HandleLogUserFeedback)

	handler := httpapi.Cors(mux)

	port := os.Getenv("PORT")
	if port == "" { port = "8000" }
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("Starting backend-go on :%s with assets at %s", port, assets)
	log.Fatal(server.ListenAndServe())
}