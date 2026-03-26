package base62

import (
	"testing"
)

func TestEncode_Zero(t *testing.T) {
	result := Encode(0)
	if result != "0" {
		t.Errorf("Encode(0) = %q, want %q", result, "0")
	}
}

func TestEncode_One(t *testing.T) {
	result := Encode(1)
	if result != "1" {
		t.Errorf("Encode(1) = %q, want %q", result, "1")
	}
}

func TestEncode_SmallNumbers(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{10, "A"},
		{36, "a"},
		{61, "z"},
		{62, "10"},
		{63, "11"},
		{124, "20"},
	}

	for _, tc := range tests {
		result := Encode(tc.input)
		if result != tc.expected {
			t.Errorf("Encode(%d) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestEncode_LargeNumbers(t *testing.T) {
	// Snowflake-sized IDs should produce non-empty strings
	result := Encode(1234567890)
	if result == "" {
		t.Error("Encode(1234567890) returned empty string")
	}
	if len(result) == 0 {
		t.Error("Expected non-empty result for large number")
	}
}

func TestEncode_Uniqueness(t *testing.T) {
	// Different inputs should produce different outputs
	results := make(map[string]int64)
	for i := int64(0); i < 1000; i++ {
		code := Encode(i)
		if prev, exists := results[code]; exists {
			t.Errorf("Encode(%d) and Encode(%d) both produced %q", prev, i, code)
		}
		results[code] = i
	}
}

func TestEncode_ConsistentResults(t *testing.T) {
	// Same input should always produce the same output
	for i := int64(0); i < 100; i++ {
		a := Encode(i)
		b := Encode(i)
		if a != b {
			t.Errorf("Encode(%d) produced inconsistent results: %q vs %q", i, a, b)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"abc", "cba"},
		{"hello", "olleh"},
	}

	for _, tc := range tests {
		result := reverse(tc.input)
		if result != tc.expected {
			t.Errorf("reverse(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(int64(i))
	}
}
