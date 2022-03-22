package usersdb

import (
	"context"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/authenticators/passwords"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/random"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRandomUser(t *testing.T) User {
	randomPassword := random.String(8)
	hashedPassword, err := passwords.HashPassword(randomPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	arg := CreateUserParams{
		Username:       random.String(10),
		HashedPassword: hashedPassword,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	userRes, err := testQueries.GetUser(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, userRes)

	require.Equal(t, user.Username, userRes.Username)
	require.Equal(t, user.HashedPassword, userRes.HashedPassword)
	require.Equal(t, user.CreatedAt, userRes.CreatedAt)
	require.Equal(t, user.UpdatedAt, userRes.UpdatedAt)
}
