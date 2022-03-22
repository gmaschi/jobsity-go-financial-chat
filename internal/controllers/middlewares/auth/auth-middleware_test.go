package authmiddleware_test

import (
	"github.com/gin-gonic/gin"
	authmiddleware "github.com/gmaschi/jobsity-go-financial-chat/internal/controllers/middlewares/auth"
	chatfactory "github.com/gmaschi/jobsity-go-financial-chat/internal/factories/chat"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := env.NewConfig()
			require.NoError(t, err)
			server, err := chatfactory.New(config, nil)
			require.NoError(t, err)

			authPath := "/auth"

			server.Router.GET(
				authPath,
				authmiddleware.AuthMiddleware(server.TokenAuth),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, map[string]interface{}{})
				},
			)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.TokenAuth)
			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker tokenauth.Maker,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	cookie := &http.Cookie{
		Name:     "auth-financial-chat",
		Value:    token,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   int(duration),
		Secure:   true,
		HttpOnly: true,
	}

	request.AddCookie(cookie)
}
