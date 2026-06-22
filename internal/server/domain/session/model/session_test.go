package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashRefreshTokenDeterministic(t *testing.T) {
	t.Parallel()

	a := HashRefreshToken("token")
	b := HashRefreshToken("token")
	require.Equal(t, a, b)
	require.NotEqual(t, HashRefreshToken("other"), a)
}
