package widget

import (
	"testing"
	"time"
)

func TestJitteredDuration(t *testing.T) {
	tests := []struct {
		name string
		base time.Duration
	}{
		{"5s base", 5 * time.Second},
		{"1m base", time.Minute},
		{"10m base", 10 * time.Minute},
		{"1s base", time.Second},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			expectedLow := time.Duration(float64(testCase.base) * (1.0 - jitterFraction))
			expectedHigh := time.Duration(float64(testCase.base) * (1.0 + jitterFraction))

			for range 1000 {
				got := jitteredDuration(testCase.base)
				if got < expectedLow {
					t.Errorf("jitteredDuration(%v) = %v, below floor %v", testCase.base, got, expectedLow)
				}
				if got > expectedHigh {
					t.Errorf("jitteredDuration(%v) = %v, above ceiling %v", testCase.base, got, expectedHigh)
				}
			}
		})
	}
}

func TestJitteredDurationFloor(t *testing.T) {
	// A very small base should be clamped to minJitteredInterval.
	tinyBase := 100 * time.Millisecond

	for range 100 {
		got := jitteredDuration(tinyBase)
		if got < minJitteredInterval {
			t.Errorf("jitteredDuration(%v) = %v, below minJitteredInterval %v", tinyBase, got, minJitteredInterval)
		}
	}
}

func TestJitteredDurationDistribution(t *testing.T) {
	// Verify the output is not degenerate — multiple distinct values over 100 samples.
	base := 5 * time.Second
	seen := make(map[time.Duration]bool)

	for range 100 {
		seen[jitteredDuration(base)] = true
	}

	if len(seen) < 10 {
		t.Errorf("expected at least 10 distinct values from 100 samples, got %d", len(seen))
	}
}
