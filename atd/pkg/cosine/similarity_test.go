package cosine

import (
	"math"
	"testing"
)

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{"identical", []float64{1, 0, 1}, []float64{1, 0, 1}, 1.0},
		{"orthogonal", []float64{1, 0}, []float64{0, 1}, 0.0},
		{"different", []float64{1, 2, 3}, []float64{4, 5, 6}, 0.9746318461970762},
		{"empty a", []float64{}, []float64{1, 2}, 0.0},
		{"mismatched length", []float64{1, 2}, []float64{1, 2, 3}, 0.0},
		{"zero vector", []float64{0, 0}, []float64{1, 1}, 0.0},
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
