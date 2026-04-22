package upstreamhooks

import (
	"reflect"
	"testing"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single line no newline",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "single line with newline",
			input: "hello\n",
			want:  []string{"hello"},
		},
		{
			name:  "multiple lines",
			input: "foo\nbar\nbaz",
			want:  []string{"foo", "bar", "baz"},
		},
		{
			name:  "multiple lines trailing newline",
			input: "foo\nbar\n",
			want:  []string{"foo", "bar"},
		},
		{
			name:  "multiple trailing newlines",
			input: "foo\n\n\n",
			want:  []string{"foo"},
		},
		{
			name:  "CRLF line endings keeps embedded CR per line",
			input: "foo\r\nbar\r\n",
			want:  []string{"foo\r", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared in table above

			// Act
			got := splitLines(tt.input)

			// Assert
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitLines(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestZsplit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single entry",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "two entries",
			input: "foo\x00bar",
			want:  []string{"foo", "bar"},
		},
		{
			name:  "leading and trailing NUL stripped",
			input: "\x00foo\x00bar\x00",
			want:  []string{"foo", "bar"},
		},
		{
			name:  "only NUL bytes",
			input: "\x00\x00",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared in table above

			// Act
			got := zsplit(tt.input)

			// Assert
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("zsplit(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
