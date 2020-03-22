package security_test

import (
	"testing"

	"github.com/projects/secure-notes/internal/security"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateHashWithSalt(t *testing.T) {
	// when
	gotFirstHashSalt, firstErr := security.GenerateHashWithSalt("abc")
	gotSecondHashSalt, secondErr := security.GenerateHashWithSalt("abc")

	// then
	assert.NotEqual(t, gotFirstHashSalt, gotSecondHashSalt)
	assert.NoError(t, firstErr)
	assert.NoError(t, secondErr)
}
