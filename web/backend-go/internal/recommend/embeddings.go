package recommend

import (
	"math"
	"sort"

	"backendgo/internal/courses"
	"backendgo/internal/embeddings"
	"backendgo/internal/types"
)

func norm(v []float32) float32 {
	var s float64
	for _, x := range v { s += float64(x) * float64(x) }
	return float32(math.Sqrt(s))
}

func dot(a, b []float32) float32 {
	var s float64
	for i := range a { s += float64(a[i]) * float64(b[i]) }
	return float32(s)
}

func cosine(a, b []float32) float32 {
	na := norm(a); nb := norm(b)
	if na == 0 || nb == 0 { return 0 }
	return dot(a, b) / (na * nb)
}

func RecommendMaxWithCombinations(liked, disliked, skipped []string, cc *courses.CourseClient, n int, emb *embeddings.Matrix) []types.CourseWithId {
	likedIDs := cc.GetCourseIDsByCodes(liked)
	dislikedIDs := cc.GetCourseIDsByCodes(disliked)

	// map liked IDs to indices in embeddings
	likedIdx := make([]int, 0, len(likedIDs))
	for _, id := range likedIDs { if idx, ok := cc.IndexForID(id); ok { likedIdx = append(likedIdx, idx) } }
	dislikedIdx := make([]int, 0, len(dislikedIDs))
	for _, id := range dislikedIDs { if idx, ok := cc.IndexForID(id); ok { dislikedIdx = append(dislikedIdx, idx) } }

	if len(likedIdx) == 0 { return nil }

	// Build target vectors as averages of pairs (including i==j)
	targets := make([][]float32, 0)
	pairs := make([][2]int, 0)
	for i := 0; i < len(likedIdx); i++ {
		for j := i; j < len(likedIdx); j++ {
			a := emb.Row(likedIdx[i])
			b := emb.Row(likedIdx[j])
			avg := make([]float32, emb.Cols)
			for k := 0; k < emb.Cols; k++ { avg[k] = 0.5 * (a[k] + b[k]) }
			targets = append(targets, avg)
			pairs = append(pairs, [2]int{likedIdx[i], likedIdx[j]})
		}
	}

	// Precompute normalized rows
	normRows := make([][]float32, emb.Rows)
	for i := 0; i < emb.Rows; i++ {
		row := emb.Row(i)
		n := norm(row)
		normed := make([]float32, emb.Cols)
		if n == 0 { copy(normed, row) } else {
			for k := 0; k < emb.Cols; k++ { normed[k] = row[k] / n }
		}
		normRows[i] = normed
	}

	scores := make([]float32, emb.Rows)
	bestPair := make([][2]int, emb.Rows)
	for i := 0; i < emb.Rows; i++ { scores[i] = -1e9 }
	for tIdx, t := range targets {
		nt := norm(t); if nt == 0 { continue }
		for i := 0; i < emb.Rows; i++ {
			// cosine = dot(normRows[i], t/|t|)
			var s float64
			for k := 0; k < emb.Cols; k++ { s += float64(normRows[i][k]) * float64(t[k]/nt) }
			sc := float32(s)
			if sc > scores[i] {
				scores[i] = sc
				bestPair[i] = pairs[tIdx]
			}
		}
	}

	// Filter out too similar to disliked
	if len(dislikedIdx) > 0 {
		for i := 0; i < emb.Rows; i++ {
			maxDis := float32(-1)
			for _, d := range dislikedIdx { v := cosine(emb.Row(i), emb.Row(d)); if v > maxDis { maxDis = v } }
			if maxDis > 0.8 { scores[i] = float32(-1e9) }
		}
	}

	// Rank by score
	type pair struct{ I int; S float32 }
	arr := make([]pair, emb.Rows)
	for i := range arr { arr[i] = pair{I: i, S: scores[i]} }
	sort.Slice(arr, func(i, j int) bool { return arr[i].S > arr[j].S })

	res := make([]types.CourseWithId, 0, n)
	for _, p := range arr {
		if len(res) >= n { break }
		if course, ok := cc.GetCourseByIndex(p.I); ok {
			bp := bestPair[p.I]
			c1, _ := cc.GetCourseByIndex(bp[0])
			c2, _ := cc.GetCourseByIndex(bp[1])
			if c1.CODE == c2.CODE {
				course.RECOMMENDED_FROM = []string{c1.CODE}
			} else {
				course.RECOMMENDED_FROM = []string{c1.CODE, c2.CODE}
			}
			res = append(res, course)
		}
	}
	return res
}