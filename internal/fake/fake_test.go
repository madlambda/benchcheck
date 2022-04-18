package fake_test

import (
	"testing"

	"github.com/madlambda/benchcheck/internal/fake"
)

func TestFake(t *testing.T) {
	// playing around with coverage
	fake.Do()
}

func BenchmarkFake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fake.Do()
	}
}
