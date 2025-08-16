package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Data models mirroring Python dataclasses

type Ratings struct {
	TheoreticalVsPractical string `json:"theoretical_vs_practical"`
	Usefulness             string `json:"usefulness"`
	Interest               string `json:"interest"`
	StemVsHumanities       string `json:"stem_vs_humanities"`
	AbstractVsSpecific     string `json:"abstract_vs_specific"`
	Difficulty             string `json:"difficulty"`
	Multidisciplinary      string `json:"multidisciplinary"`
	ProjectBased           string `json:"project_based"`
	Creative               string `json:"creative"`
}

type Course struct {
	CODE                     string   `json:"CODE"`
	FACULTY                  string   `json:"FACULTY"`
	NAME                     string   `json:"NAME"`
	NAME_EN                  string   `json:"NAME_EN"`
	LANGUAGE                 string   `json:"LANGUAGE"`
	SEMESTER                 string   `json:"SEMESTER"`
	CREDITS                  string   `json:"CREDITS"`
	DEPARTMENT               string   `json:"DEPARTMENT"`
	TEACHERS                 string   `json:"TEACHERS"`
	COMPLETION               string   `json:"COMPLETION"`
	PREREQUISITES            string   `json:"PREREQUISITES"`
	FIELDS_OF_STUDY          *string  `json:"FIELDS_OF_STUDY"`
	TYPE_OF_STUDY            *string  `json:"TYPE_OF_STUDY"`
	LECTURES_SEMINARS_HOMEWORK string `json:"LECTURES_SEMINARS_HOMEWORK"`
	SYLLABUS                 string   `json:"SYLLABUS"`
	OBJECTIVES               string   `json:"OBJECTIVES"`
	TEXT_PREREQUISITS        *string  `json:"TEXT_PREREQUISITS"`
	ASSESMENT_METHODS        string   `json:"ASSESMENT_METHODS"`
	TEACHING_METHODS         string   `json:"TEACHING_METHODS"`
	TEACHER_INFO             *string  `json:"TEACHER_INFO"`
	LEARNING_OUTCOMES        string   `json:"LEARNING_OUTCOMES"`
	LITERATURE               string   `json:"LITERATURE"`
	STUDENTS_ENROLLED        string   `json:"STUDENTS_ENROLLED"`
	STUDENTS_PASSED          string   `json:"STUDENTS_PASSED"`
	AVERAGE_GRADE            string   `json:"AVERAGE_GRADE"`
	FOLLOWUP_COURSES         *string  `json:"FOLLOWUP_COURSES"`
	KEYWORDS                 []string `json:"KEYWORDS"`
	DESCRIPTION              string   `json:"DESCRIPTION"`
	RATINGS                  Ratings  `json:"RATINGS"`
}

type CourseWithId struct {
	Course
	ID               *int     `json:"ID"`
	SIMILARITY       float64  `json:"SIMILARITY"`
	RECOMMENDED_FROM []string `json:"RECOMMENDED_FROM"`
}

type RecommendationFeedbackLog struct {
	Liked            []string `json:"liked"`
	Disliked         []string `json:"disliked"`
	Skipped          []string `json:"skipped"`
	Course           string   `json:"course"`
	Action           string   `json:"action"`
	UserID           string   `json:"user_id"`
	Model            string   `json:"model"`
	Phrases          []string `json:"phrases"`
	RecommendedFrom  []string `json:"recommended_from"`
}

type UserFeedbackLog struct {
	Text      *string  `json:"text"`
	Rating    *int     `json:"rating"`
	Faculty   *string  `json:"faculty"`
	StudyType *string  `json:"study_type"`
	Semester  *string  `json:"semester"`
	Phrases   []string `json:"phrases"`
	Model     *string  `json:"model"`
	UserID    string   `json:"user_id"`
}

type RecommendationResponse struct {
	RecommendedCourses []CourseWithId `json:"recommended_courses"`
}

// In-memory placeholders. Real implementation should load Parquet/NPZ equivalents.
var allCoursesByCode = map[string]CourseWithId{}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := map[string]bool{}
		if v := os.Getenv("FRONTEND_URL"); v != "" {
			allowed[v] = true
		}
		allowed["https://recommend.muni.courses"] = true
		allowed["https://muni.courses"] = true

		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, []string{"max_with_combinations", "keywords_tfidf"})
}

func handleCourse(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/course/"):]
	if code == "" {
		http.Error(w, "missing course code", http.StatusBadRequest)
		return
	}
	c, ok := allCoursesByCode[code]
	if !ok {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, c)
}

type RecommendationsRequest struct {
	Liked    []string `json:"liked"`
	Disliked []string `json:"disliked"`
	Skipped  []string `json:"skipped"`
	N        int      `json:"n"`
	Model    string   `json:"model"`
	Relevance float64 `json:"relevance"`
}

func handleRecommendations(w http.ResponseWriter, r *http.Request) {
	var req RecommendationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.N <= 0 {
		req.N = 10
	}
	if req.Model == "" {
		req.Model = "average"
	}
	// Placeholder: return first N courses ignoring model
	res := RecommendationResponse{RecommendedCourses: []CourseWithId{}}
	count := 0
	for _, c := range allCoursesByCode {
		res.RecommendedCourses = append(res.RecommendedCourses, c)
		count++
		if count >= req.N {
			break
		}
	}
	respondJSON(w, http.StatusOK, res)
}

func handleLogRecommendationFeedback(w http.ResponseWriter, r *http.Request) {
	var logReq RecommendationFeedbackLog
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// TODO: persist to Mongo similar to Python backend
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleLogUserFeedback(w http.ResponseWriter, r *http.Request) {
	var logReq UserFeedbackLog
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// TODO: persist to Mongo similar to Python backend
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/models", handleModels)
	mux.HandleFunc("/course/", handleCourse)
	mux.HandleFunc("/recommendations", handleRecommendations)
	mux.HandleFunc("/log_recommendation_feedback", handleLogRecommendationFeedback)
	mux.HandleFunc("/log_user_feedback", handleLogUserFeedback)

	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	log.Printf("Starting backend-go on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}