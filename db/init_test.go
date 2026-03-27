package db

import (
	"testing"
)

func TestSplitSQLStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // number of statements
	}{
		{
			name: "simple statements",
			input: `CREATE DATABASE test;
USE test;
CREATE TABLE users (id INT);`,
			expected: 3,
		},
		{
			name: "with comments",
			input: `-- This is a comment
CREATE DATABASE test;
-- Another comment
CREATE TABLE users (id INT);`,
			expected: 2,
		},
		{
			name: "multi-line statement",
			input: `CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);`,
			expected: 1,
		},
		{
			name: "empty lines",
			input: `CREATE DATABASE test;


CREATE TABLE users (id INT);`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSQLStatements(tt.input)
			
			// Filter out empty statements
			var nonEmpty []string
			for _, stmt := range result {
				if len(stmt) > 0 {
					nonEmpty = append(nonEmpty, stmt)
				}
			}
			
			if len(nonEmpty) != tt.expected {
				t.Errorf("Expected %d statements, got %d", tt.expected, len(nonEmpty))
				for i, stmt := range nonEmpty {
					t.Logf("Statement %d: %s", i+1, stmt)
				}
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "needs truncation",
			input:    "hello world this is long",
			maxLen:   10,
			expected: "hello worl...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
