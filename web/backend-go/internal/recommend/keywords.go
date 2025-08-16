package recommend

import (
	"sort"

	"backendgo/internal/courses"
	"backendgo/internal/sparse"
	"backendgo/internal/types"
)

type Scored struct { Idx int; Score float32 }

func RecommendKeywords(liked, disliked, skipped []string, cc *courses.CourseClient, n int, m *sparse.CSR) []types.CourseWithId {
	likedIdx := toIndices(cc, liked)
	dislikedIdx := toIndices(cc, disliked)
	skippedIdx := toIndices(cc, skipped)

	if len(likedIdx) == 0 { return nil }

	base := m.SumRows(likedIdx)
	if len(dislikedIdx) > 0 {
		pen := m.SumRows(dislikedIdx)
		w := float32(0.5 / float64(len(dislikedIdx)))
		for i := range base { base[i] -= w * pen[i] }
	}

	scores := make([]Scored, m.Rows)
	for i := 0; i < m.Rows; i++ {
		scores[i] = Scored{Idx: i, Score: base[i]}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].Score > scores[j].Score })

	excluded := toSet(append(append(likedIdx, dislikedIdx...), skippedIdx...))
	res := make([]types.CourseWithId, 0, n)
	for _, s := range scores {
		if len(res) >= n { break }
		if _, ok := excluded[s.Idx]; ok { continue }
		if c, ok := cc.GetCourseByID(s.Idx); ok {
			if cc.FilterCourse(c) { continue }
			// compute recommended_from: top 2 liked by direct score m[liked,row]
			var rf []string
			for _, li := range likedIdx {
				idx, val := m.Row(li)
				for j, col := range idx {
					if col == s.Idx {
						if v, ok := cc.GetCourseByID(li); ok { rf = append(rf, v.CODE) }
						_ = val[j]
						break
					}
				}
			}
			if len(rf) > 2 { rf = rf[:2] }
			c.RECOMMENDED_FROM = rf
			res = append(res, c)
		}
	}
	return res
}

func toIndices(cc *courses.CourseClient, codes []string) []int {
	ids := cc.GetCourseIDsByCodes(codes)
	res := make([]int, 0, len(ids))
	for _, id := range ids {
		if idx, ok := cc.IndexForID(id); ok { res = append(res, idx) }
	}
	return res
}

func toSet(v []int) map[int]struct{} {
	m := make(map[int]struct{}, len(v))
	for _, x := range v { m[x] = struct{}{} }
	return m
}