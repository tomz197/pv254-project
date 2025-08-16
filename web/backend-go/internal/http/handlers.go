package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"backendgo/internal/courses"
	"backendgo/internal/db"
	"backendgo/internal/embeddings"
	"backendgo/internal/recommend"
	"backendgo/internal/sparse"
	"backendgo/internal/types"
)

type Server struct {
	CC   *courses.CourseClient
	Emb  *embeddings.Matrix
	KWTf *sparse.CSR
	DB   *db.MongoLogger
}

type recReq struct {
	Liked    []string `json:"liked"`
	Disliked []string `json:"disliked"`
	Skipped  []string `json:"skipped"`
	N        int      `json:"n"`
	Model    string   `json:"model"`
	Relevance float64 `json:"relevance"`
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleModels(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []string{"max_with_combinations", "keywords_tfidf"})
}

func (s *Server) HandleCourse(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/course/"):]
	if v, ok := s.CC.GetCourseByCode(code); ok {
		respondJSON(w, http.StatusOK, v)
		return
	}
	http.Error(w, "Course not found", http.StatusNotFound)
}

func (s *Server) HandleRecommendations(w http.ResponseWriter, r *http.Request) {
	var req recReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.N <= 0 { req.N = 10 }
	model := req.Model
	if model == "" { model = "average" }

	var out []types.CourseWithId
	switch model {
	case "keywords_tfidf":
		out = recommend.RecommendKeywords(req.Liked, req.Disliked, req.Skipped, s.CC, req.N, s.KWTf)
	case "max_with_combinations":
		out = recommend.RecommendMaxWithCombinations(req.Liked, req.Disliked, req.Skipped, s.CC, req.N, s.Emb)
	default:
		out = recommend.RecommendMaxWithCombinations(req.Liked, req.Disliked, req.Skipped, s.CC, req.N, s.Emb)
	}
	respondJSON(w, http.StatusOK, types.RecommendationResponse{RecommendedCourses: out})
}

func (s *Server) HandleLogRecommendationFeedback(w http.ResponseWriter, r *http.Request) {
	var logReq types.RecommendationFeedbackLog
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
	_ = s.DB.LogRecommendationFeedback(r.Context(), logReq)
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleLogUserFeedback(w http.ResponseWriter, r *http.Request) {
	var logReq types.UserFeedbackLog
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
	_ = s.DB.LogUserFeedback(r.Context(), logReq)
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func NewServer(assetsDir string) (*Server, error) {
	cc, err := courses.NewCourseClient(assetsDir)
	if err != nil { return nil, err }
	emb, err := embeddings.LoadNPY(filepath.Join(assetsDir, "embeddings_tomas_03.npy"))
	if err != nil { return nil, err }
	kw, err := sparse.LoadCSRFromNPZ(filepath.Join(assetsDir, "intersects_tfidf.npz"))
	if err != nil { return nil, err }
	dbLogger, err := db.NewMongoLogger(context.Background())
	if err != nil { return nil, err }
	return &Server{CC: cc, Emb: emb, KWTf: kw, DB: dbLogger}, nil
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := map[string]bool{}
		if v := os.Getenv("FRONTEND_URL"); v != "" { allowed[v] = true }
		allowed["https://recommend.muni.courses"] = true
		allowed["https://muni.courses"] = true
		origin := r.Header.Get("Origin")
		if allowed[origin] { w.Header().Set("Access-Control-Allow-Origin", origin) }
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
		next.ServeHTTP(w, r)
	})
}