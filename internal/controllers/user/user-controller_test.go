package usercontroller_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	chatfactory "github.com/gmaschi/jobsity-go-financial-chat/internal/factories/chat"
	usermodel "github.com/gmaschi/jobsity-go-financial-chat/internal/models/user"
	mockeduserstore "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/mocks/postgresql/users"
	usersdb "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/postgresql/users"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/authenticators/passwords"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/random"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
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

type eqCreateUserParamsMatcher struct {
	arg      usersdb.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(usersdb.CreateUserParams)
	if !ok {
		return false
	}

	err := passwords.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg usersdb.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreate(t *testing.T) {
	user, randomPassword := randomUser(t)

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockeduserstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"username": user.Username,
				"password": randomPassword,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				arg := usersdb.CreateUserParams{
					Username: user.Username,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, randomPassword)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchCreate(t, recorder.Body, user)
			},
		},
		{
			name: "InvalidUsername",
			body: map[string]interface{}{
				"username": "invalid-username#",
				"password": randomPassword,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPassword",
			body: map[string]interface{}{
				"username": user.Username,
				"password": "abc",
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"username": user.Username,
				"password": randomPassword,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(usersdb.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ExistingUsername",
			body: map[string]interface{}{
				"username": user.Username,
				"password": randomPassword,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(usersdb.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockeduserstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			config, err := env.NewConfig()
			require.NoError(t, err)

			server, err := chatfactory.New(config, store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestLogin(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          map[string]interface{}
		buildStubs    func(store *mockeduserstore.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: map[string]interface{}{
				"username": "invalid-user#1",
				"password": password,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: map[string]interface{}{
				"username": "NotFound",
				"password": password,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(usersdb.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IncorrectPassword",
			body: map[string]interface{}{
				"username": user.Username,
				"password": "incorrect",
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]interface{}{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockeduserstore.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(usersdb.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockeduserstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			config, err := env.NewConfig()
			require.NoError(t, err)

			server, err := chatfactory.New(config, store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/login"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (usersdb.User, string) {
	randomPassword := random.String(8)
	hashedPassword, err := passwords.HashPassword(randomPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	now := time.Now().UTC()
	user := usersdb.User{
		Username:       random.String(10),
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return user, randomPassword
}

func requireBodyMatchCreate(t *testing.T, body *bytes.Buffer, user usersdb.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var expectedUserModel, gotUser usermodel.CreateResponse
	jsonModelAuthor, err := json.Marshal(&user)
	require.NoError(t, err)
	err = json.Unmarshal(jsonModelAuthor, &expectedUserModel)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, expectedUserModel, gotUser)

	require.Equal(t, expectedUserModel.Username, gotUser.Username)
	require.Equal(t, expectedUserModel.CreatedAt, gotUser.CreatedAt)
	require.Empty(t, gotUser.UpdatedAt)
	require.Empty(t, gotUser.HashedPassword)
}
