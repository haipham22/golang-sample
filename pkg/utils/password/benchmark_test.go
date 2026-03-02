package password

import (
	"bytes"
	"fmt"
	"testing"
)

// BenchmarkHashPassword benchmarks password hashing performance
// Following Uber: "Benchmark before optimizing"
// Following Uber: "Use meaningful benchmark names"
func BenchmarkHashPassword(b *testing.B) {
	password := "TestPassword123!"
	b.ResetTimer()
	b.ReportAllocs() // Following Uber: "Report allocations to detect leaks"

	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

// BenchmarkHashPassword_Parallel benchmarks concurrent hashing
// Following Uber: "Test parallel execution patterns"
func BenchmarkHashPassword_Parallel(b *testing.B) {
	password := "TestPassword123!"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := HashPassword(password)
			if err != nil {
				b.Fatalf("HashPassword failed: %v", err)
			}
		}
	})
}

// BenchmarkCheckPasswordHash benchmarks verification performance
// Following Uber: "Benchmark both sides of operations (hash + verify)"
func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ok := CheckPasswordHash(password, hash)
		if !ok {
			b.Fatalf("CheckPasswordHash returned false for valid password")
		}
	}
}

// BenchmarkCheckPasswordHash_Parallel benchmarks concurrent verification
func BenchmarkCheckPasswordHash_Parallel(b *testing.B) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ok := CheckPasswordHash(password, hash)
			if !ok {
				b.Fatalf("CheckPasswordHash returned false for valid password")
			}
		}
	})
}

// BenchmarkHashPassword_VaryingLength benchmarks different password lengths
// Following Uber: "Benchmark across different input sizes"
func BenchmarkHashPassword_VaryingLength(b *testing.B) {
	lengths := []int{8, 16, 32, 64, 72} // bcrypt max is 72 bytes

	for _, length := range lengths {
		// Create a password of exactly 'length' bytes filled with 'a'
		password := string(bytes.Repeat([]byte("a"), length))

		b.Run(fmt.Sprintf("length-%d", length), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := HashPassword(password)
				if err != nil {
					b.Fatalf("HashPassword failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkOperation measures full hash + verify cycle
// Following Uber: "Benchmark realistic workflows"
func BenchmarkOperation(b *testing.B) {
	password := "TestPassword123!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Full hash + verify cycle each iteration
		hash, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
		ok := CheckPasswordHash(password, hash)
		if !ok {
			b.Fatalf("CheckPasswordHash returned false for valid password")
		}
	}
}
