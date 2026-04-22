package upstreamhooks

import (
	"testing"
)

func TestTrimBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no trailing newlines unchanged",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "trailing LF stripped",
			input: "hello\n",
			want:  "hello",
		},
		{
			name:  "trailing CRLF stripped",
			input: "hello\r\n",
			want:  "hello",
		},
		{
			name:  "multiple trailing newlines stripped",
			input: "hello\n\n\r\n",
			want:  "hello",
		},
		{
			name:  "internal newlines preserved",
			input: "hel\nlo\n",
			want:  "hel\nlo",
		},
		{
			name:  "empty string unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "only newlines becomes empty",
			input: "\n\r\n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: input prepared above

			// Act
			got := trimBytes(tt.input)

			// Assert
			if got != tt.want {
				t.Errorf("trimBytes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
