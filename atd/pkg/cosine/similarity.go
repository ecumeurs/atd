package cosine

import "math"

// Similarity computes the cosine similarity between two float32 vectors.
// Returns 0.0 if either vector is empty or lengths don't match.
func Similarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0.0
	}
	var dot, magA, magB float32
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA*magB == 0 {
		return 0.0
	}
	return float64(dot) / (math.Sqrt(float64(magA)) * math.Sqrt(float64(magB)))
}
