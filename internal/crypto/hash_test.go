package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		key  string
	}{
		{
			name: "simple",
			body: []byte("hello"),
			key:  "secret",
		},
		{
			name: "empty body",
			body: []byte{},
			key:  "secret",
		},
		{
			name: "empty key",
			body: []byte("hello"),
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := HashSHA256(tt.body, tt.key)
			hash2 := HashSHA256(tt.body, tt.key)

			assert.Equal(t, hash1, hash2)
			assert.Len(t, hash1, 64)
		})
	}
}

func TestHashSHA256_DifferentInput(t *testing.T) {
	hash1 := HashSHA256([]byte("hello"), "secret")
	hash2 := HashSHA256([]byte("world"), "secret")
	hash3 := HashSHA256([]byte("hello"), "another-secret")

	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}
