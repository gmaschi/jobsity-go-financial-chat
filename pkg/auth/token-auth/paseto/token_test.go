package paseto_test

import (
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth/paseto"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPasetoMaker(t *testing.T) {
	t.Run("Create Paseto Token", func(t *testing.T) {
		maker, err := paseto.NewMaker(random.String(32))
		require.NoError(t, err)

		username := random.String(8)
		duration := time.Minute
		issuedAt := time.Now()
		expiredAt := issuedAt.Add(duration)

		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, payload)

		require.NotZero(t, payload.ID)
		require.Equal(t, username, payload.Username)
		require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
		require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
	})

	t.Run("Expired Paseto Token", func(t *testing.T) {
		maker, err := paseto.NewMaker(random.String(32))
		require.NoError(t, err)

		username := random.String(8)
		duration := time.Minute

		token, err := maker.CreateToken(username, -duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.Error(t, err)
		require.EqualError(t, err, tokenauth.ErrExpiredToken.Error())
		require.Nil(t, payload)
	})
}
