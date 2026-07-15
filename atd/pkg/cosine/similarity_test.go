package cosine

import (
	"math"
	"testing"
)

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 0, 1}, []float32{1, 0, 1}, 1.0},
		{"orthogonal", []float32{1, 0}, []float32{0, 1}, 0.0},
		{"different", []float32{1, 2, 3}, []float32{4, 5, 6}, 0.9746318461970762},
		{"empty a", []float32{}, []float32{1, 2}, 0.0},
		{"mismatched length", []float32{1, 2}, []float32{1, 2, 3}, 0.0},
		{"zero vector", []float32{0, 0}, []float32{1, 1}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Similarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-10 {
				t.Errorf("Similarity() = %v, want %v", got, tt.want)
			}
		})
	}
}
