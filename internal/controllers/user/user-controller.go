package usercontroller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	usermodel "github.com/gmaschi/jobsity-go-financial-chat/internal/models/user"
	usersdb "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/postgresql/users"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth/paseto"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/authenticators/passwords"
	parseerrors "github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/parse-errors"
	"github.com/lib/pq"
	"net/http"
	"time"
)

type Controller struct {
	store usersdb.Store
}

// New creates a reference to a Controller
func New(store usersdb.Store) *Controller {
	return &Controller{
		store: store,
	}
}

// Create handles the request to create a new user
func (c *Controller) Create(ctx *gin.Context) {
	var req usermodel.CreateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseerrors.ErrorResponse(err))
		return
	}

	hashedPassword, err := passwords.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseerrors.ErrorResponse(err))
		return
	}

	createArgs := usersdb.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
	}

	user, err := c.store.CreateUser(ctx, createArgs)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, parseerrors.ErrorResponse(pqError))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, parseerrors.ErrorResponse(err))
		return
	}

	userResponse := usermodel.CreateResponse(user)
	ctx.JSON(http.StatusCreated, userResponse)
}

// LoginForm displays the form for a user to log in
func (c *Controller) LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", nil)
}

// Login handles the request to log a user in
func (c *Controller) Login(ctx *gin.Context) {
	var req usermodel.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseerrors.ErrorResponse(err))
		return
	}

	user, err := c.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseerrors.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseerrors.ErrorResponse(err))
		return
	}

	err = passwords.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, parseerrors.ErrorResponse(err))
		return
	}

	config, err := env.NewConfig()

	maker, err := paseto.NewMaker(config.TokenSymmetricKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseerrors.ErrorResponse(err))
		return
	}

	token, err := maker.CreateToken(user.Username, time.Duration(config.TokenDuration)*time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseerrors.ErrorResponse(err))
		return
	}

	res := usermodel.LoginResponse{
		AccessToken: token,
		Username:    user.Username,
	}

	ctx.SetCookie("auth-financial-chat", token, 10000, "/", "localhost", true, true)
	ctx.JSON(http.StatusOK, res)
}
