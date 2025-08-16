package courses

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"backendgo/internal/types"
)

type idLookupRow struct {
	ID    int `json:"ID"`
	INDEX int `json:"index"`
}

type CourseClient struct {
	coursesByCode map[string]types.CourseWithId
	coursesByID   map[int]types.CourseWithId
	idToIndex     map[int]int
	indexToID     map[int]int
}

func NewCourseClient(assetsDir string) (*CourseClient, error) {
	cc := &CourseClient{
		coursesByCode: map[string]types.CourseWithId{},
		coursesByID:   map[int]types.CourseWithId{},
		idToIndex:     map[int]int{},
		indexToID:     map[int]int{},
	}

	coursesJSON := filepath.Join(assetsDir, "courses", "courses.parquet.json")
	idLookupJSON := filepath.Join(assetsDir, "courses", "id_lookup.parquet.json")
	if err := cc.loadCoursesJSON(coursesJSON); err != nil {
		return nil, fmt.Errorf("load courses json: %w", err)
	}
	if err := cc.loadIdLookupJSON(idLookupJSON); err != nil {
		return nil, fmt.Errorf("load id_lookup json: %w", err)
	}
	return cc, nil
}

func (c *CourseClient) loadCoursesJSON(path string) error {
	b, err := os.ReadFile(path)
	if err != nil { return err }
	var rows []types.CourseWithId
	if err := json.Unmarshal(b, &rows); err != nil { return err }
	for _, rec := range rows {
		if rec.CODE == "" { continue }
		if rec.ID != nil { c.coursesByID[*rec.ID] = rec }
		c.coursesByCode[rec.CODE] = rec
	}
	return nil
}

func (c *CourseClient) loadIdLookupJSON(path string) error {
	b, err := os.ReadFile(path)
	if err != nil { return err }
	var rows []idLookupRow
	if err := json.Unmarshal(b, &rows); err != nil { return err }
	for _, r := range rows {
		c.idToIndex[r.ID] = r.INDEX
		c.indexToID[r.INDEX] = r.ID
	}
	return nil
}

func (c *CourseClient) GetCourseByCode(code string) (types.CourseWithId, bool) {
	v, ok := c.coursesByCode[code]
	return v, ok
}

func (c *CourseClient) GetCourseByID(id int) (types.CourseWithId, bool) {
	v, ok := c.coursesByID[id]
	return v, ok
}

func (c *CourseClient) GetCourseByIndex(index int) (types.CourseWithId, bool) {
	if id, ok := c.indexToID[index]; ok {
		return c.GetCourseByID(id)
	}
	return types.CourseWithId{}, false
}

func (c *CourseClient) GetCourseIDsByCodes(codes []string) []int {
	res := make([]int, 0, len(codes))
	for _, code := range codes {
		if v, ok := c.coursesByCode[code]; ok && v.ID != nil {
			res = append(res, *v.ID)
		}
	}
	return res
}

func (c *CourseClient) IndexForID(id int) (int, bool) {
	idx, ok := c.idToIndex[id]
	return idx, ok
}

func (c *CourseClient) AllCoursesSortedByID() []types.CourseWithId {
	res := make([]types.CourseWithId, 0, len(c.coursesByCode))
	for _, v := range c.coursesByCode { res = append(res, v) }
	// sort by ID
	slices.SortFunc(res, func(a, b types.CourseWithId) int {
		ai, bi := 0, 0
		if a.ID != nil { ai = *a.ID }
		if b.ID != nil { bi = *b.ID }
		return ai - bi
	})
	return res
}

func (c *CourseClient) FilterCourse(course types.CourseWithId) bool {
	name := strings.ToLower(course.NAME)
	nameEn := strings.ToLower(course.NAME_EN)
	contains := func(s string, parts []string) bool {
		for _, p := range parts { if strings.Contains(s, p) { return true } }
		return false
	}
	return (name != "" && contains(name, []string{"thesis", "diplomov", "bakalářsk", "státnic"})) ||
		(nameEn != "" && contains(nameEn, []string{"thesis", "state exam"}))
}