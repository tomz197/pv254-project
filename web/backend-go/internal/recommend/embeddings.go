package recommend

import (
	"sort"

	"backendgo/internal/courses"
	"backendgo/internal/embeddings"
	"backendgo/internal/types"
	"gonum.org/v1/gonum/floats"
)

func to64(v []float32) []float64 {
	out := make([]float64, len(v))
	for i, x := range v { out[i] = float64(x) }
	return out
}

func cosine(a32, b32 []float32) float32 {
	a := to64(a32)
	b := to64(b32)
	na := floats.Norm(a, 2)
	nb := floats.Norm(b, 2)
	if na == 0 || nb == 0 { return 0 }
	return float32(floats.Dot(a, b) / (na * nb))
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

	// Precompute normalized rows in float64-compatible form
	normRows := make([][]float64, emb.Rows)
	for i := 0; i < emb.Rows; i++ {
		row := to64(emb.Row(i))
		n := floats.Norm(row, 2)
		if n != 0 { floats.Scale(1.0/n, row) }
		normRows[i] = row
	}

	scores := make([]float32, emb.Rows)
	bestPair := make([][2]int, emb.Rows)
	for i := 0; i < emb.Rows; i++ { scores[i] = -1e9 }
	for tIdx, t := range targets {
		tt := to64(t)
		n := floats.Norm(tt, 2)
		if n == 0 { continue }
		floats.Scale(1.0/n, tt)
		for i := 0; i < emb.Rows; i++ {
			sc := float32(floats.Dot(normRows[i], tt))
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