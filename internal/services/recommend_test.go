package services

import (
	"math"
	"testing"
)

func TestCalculateScore(t *testing.T) {
	s := calculateScore(0.8, 0.6, 0.4, 0.5)
	if s <= 0 {
		t.Fatalf("expected positive score")
	}
	expected := 0.67
	if math.Abs(s-expected) > 1e-9 {
		t.Fatalf("unexpected score: got=%f expected=%f", s, expected)
	}
}
