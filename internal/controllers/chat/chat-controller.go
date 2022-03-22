package chatcontroller

import (
	"github.com/gin-gonic/gin"
	authmiddleware "github.com/gmaschi/jobsity-go-financial-chat/internal/controllers/middlewares/auth"
	chatmodel "github.com/gmaschi/jobsity-go-financial-chat/internal/models/chat"
	chatmessages "github.com/gmaschi/jobsity-go-financial-chat/internal/services/chat-messages"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	parseerrors "github.com/gmaschi/jobsity-go-financial-chat/pkg/tools/parse-errors"
	"net/http"
)

type Controller struct {
}

const (
	usersLoginRoute = "/users/login"
)

// New creates a reference to a Controller
func New() *Controller {
	return &Controller{}
}

// Home creates a handler to handle the home chat lobby page
func (c *Controller) Home(ctx *gin.Context) {
	authPayload, ok := ctx.Get(authmiddleware.AuthorizationPayloadKey)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	_, ok = authPayload.(*tokenauth.Payload)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	ctx.HTML(http.StatusOK, "chat-lobby.html", nil)
}

// Room creates a handler to handle the room chat page
func (c *Controller) Room(ctx *gin.Context) {
	var req chatmodel.GetRoom

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseerrors.ErrorResponse(err))
		return
	}

	authPayload, ok := ctx.Get(authmiddleware.AuthorizationPayloadKey)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	tokenAuthPayload, ok := authPayload.(*tokenauth.Payload)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	_ = tokenAuthPayload

	ctx.HTML(http.StatusOK, "chat-room.html", nil)
}

// ChatMessages handles the route for incoming messages on each room
func (c *Controller) ChatMessages(ctx *gin.Context) {
	var req chatmodel.GetRoom

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseerrors.ErrorResponse(err))
		return
	}

	authPayload, ok := ctx.Get(authmiddleware.AuthorizationPayloadKey)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	tokenAuthPayload, ok := authPayload.(*tokenauth.Payload)
	if !ok {
		ctx.Redirect(http.StatusTemporaryRedirect, usersLoginRoute)
		return
	}

	username := tokenAuthPayload.Username

	chatmessages.ServeWs(ctx.Writer, ctx.Request, req.RoomID, username)
}
