package chatcontroller_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	chatfactory "github.com/gmaschi/jobsity-go-financial-chat/internal/factories/chat"
	mockeduserstore "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/mocks/postgresql/users"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "..", "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestHome(t *testing.T) {
	testCases := []struct {
		name          string
		body          map[string]interface{}
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "No Authorization",
			body:      nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockeduserstore.NewMockStore(ctrl)

			config, err := env.NewConfig()
			require.NoError(t, err)

			server, err := chatfactory.New(config, store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := "/chat"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.TokenAuth)
			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestRoom(t *testing.T) {
	testCases := []struct {
		name          string
		roomID        string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			roomID: "1",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "Bad Request",
			roomID: "invalid-room-ID",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "No Authorization",
			roomID:    "1",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockeduserstore.NewMockStore(ctrl)

			config, err := env.NewConfig()
			require.NoError(t, err)

			server, err := chatfactory.New(config, store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/chat/%s", tc.roomID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, req, server.TokenAuth)
			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestChatMessages(t *testing.T) {
	testCases := []struct {
		name          string
		roomID        string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		//{
		//	name:   "OK",
		//	roomID: "1",
		//	setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
		//		addAuthorization(t, request, tokenMaker, "user", time.Minute)
		//	},
		//	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		//		require.Equal(t, http.StatusOK, recorder.Code)
		//	},
		//},
		{
			name:   "Bad Request",
			roomID: "invalid-room-ID",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {
				addAuthorization(t, request, tokenMaker, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "No Authorization",
			roomID:    "1",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokenauth.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockeduserstore.NewMockStore(ctrl)

			config, err := env.NewConfig()
			require.NoError(t, err)

			server, err := chatfactory.New(config, store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/chat/ws/%s", tc.roomID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
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
