package types

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
	CODE                       string   `json:"CODE"`
	FACULTY                    string   `json:"FACULTY"`
	NAME                       string   `json:"NAME"`
	NAME_EN                    string   `json:"NAME_EN"`
	LANGUAGE                   string   `json:"LANGUAGE"`
	SEMESTER                   string   `json:"SEMESTER"`
	CREDITS                    string   `json:"CREDITS"`
	DEPARTMENT                 string   `json:"DEPARTMENT"`
	TEACHERS                   string   `json:"TEACHERS"`
	COMPLETION                 string   `json:"COMPLETION"`
	PREREQUISITES              string   `json:"PREREQUISITES"`
	FIELDS_OF_STUDY            *string  `json:"FIELDS_OF_STUDY"`
	TYPE_OF_STUDY              *string  `json:"TYPE_OF_STUDY"`
	LECTURES_SEMINARS_HOMEWORK string   `json:"LECTURES_SEMINARS_HOMEWORK"`
	SYLLABUS                   string   `json:"SYLLABUS"`
	OBJECTIVES                 string   `json:"OBJECTIVES"`
	TEXT_PREREQUISITS          *string  `json:"TEXT_PREREQUISITS"`
	ASSESMENT_METHODS          string   `json:"ASSESMENT_METHODS"`
	TEACHING_METHODS           string   `json:"TEACHING_METHODS"`
	TEACHER_INFO               *string  `json:"TEACHER_INFO"`
	LEARNING_OUTCOMES          string   `json:"LEARNING_OUTCOMES"`
	LITERATURE                 string   `json:"LITERATURE"`
	STUDENTS_ENROLLED          string   `json:"STUDENTS_ENROLLED"`
	STUDENTS_PASSED            string   `json:"STUDENTS_PASSED"`
	AVERAGE_GRADE              string   `json:"AVERAGE_GRADE"`
	FOLLOWUP_COURSES           *string  `json:"FOLLOWUP_COURSES"`
	KEYWORDS                   []string `json:"KEYWORDS"`
	DESCRIPTION                string   `json:"DESCRIPTION"`
	RATINGS                    Ratings  `json:"RATINGS"`
}

type CourseWithId struct {
	Course
	ID               *int     `json:"ID"`
	SIMILARITY       float64  `json:"SIMILARITY"`
	RECOMMENDED_FROM []string `json:"RECOMMENDED_FROM"`
}

type RecommendationFeedbackLog struct {
	Liked           []string `json:"liked"`
	Disliked        []string `json:"disliked"`
	Skipped         []string `json:"skipped"`
	Course          string   `json:"course"`
	Action          string   `json:"action"`
	UserID          string   `json:"user_id"`
	Model           string   `json:"model"`
	Phrases         []string `json:"phrases"`
	RecommendedFrom []string `json:"recommended_from"`
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