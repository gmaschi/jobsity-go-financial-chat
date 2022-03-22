package chatfactory

import (
	"fmt"
	"github.com/gin-gonic/gin"
	chatcontroller "github.com/gmaschi/jobsity-go-financial-chat/internal/controllers/chat"
	authmiddleware "github.com/gmaschi/jobsity-go-financial-chat/internal/controllers/middlewares/auth"
	usercontroller "github.com/gmaschi/jobsity-go-financial-chat/internal/controllers/user"
	usersdb "github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/postgresql/users"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth/paseto"
	"github.com/gmaschi/jobsity-go-financial-chat/pkg/config/env"
	"log"
	"os"
	"path/filepath"
)

type (
	Factory struct {
		store       usersdb.Store
		chatHandler chatHandler
		TokenAuth   tokenauth.Maker
		Config      env.Config
		Router      *gin.Engine
	}

	chatHandler struct {
		userController *usercontroller.Controller
		chatController *chatcontroller.Controller
	}
)

// New creates a new factory
func New(config env.Config, store usersdb.Store) (*Factory, error) {
	tokenMaker, err := paseto.NewMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	factory := &Factory{
		store: store,
		chatHandler: chatHandler{
			userController: usercontroller.New(store),
			chatController: chatcontroller.New(),
		},
		TokenAuth: tokenMaker,
		Config:    config,
	}

	router := gin.Default()

	factory.setupTemplates(router)
	factory.setupRoutes(router)

	factory.Router = router
	return factory, nil
}

func (f *Factory) setupRoutes(router *gin.Engine) {
	users := router.Group("/users")
	{
		users.POST("", f.chatHandler.userController.Create)
		users.GET("/login", f.chatHandler.userController.LoginForm)
		users.POST("/login", f.chatHandler.userController.Login)
	}

	authChatRoutes := router.Group("/chat").Use(authmiddleware.AuthMiddleware(f.TokenAuth))
	{
		authChatRoutes.GET("", f.chatHandler.chatController.Home)
		authChatRoutes.GET("/:roomId", f.chatHandler.chatController.Room)
		authChatRoutes.GET("/ws/:roomId", f.chatHandler.chatController.ChatMessages)
	}
}

func (f *Factory) setupTemplates(router *gin.Engine) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	templatesPath := filepath.Join(wd, "public", "templates")

	router.LoadHTMLFiles(
		filepath.Join(templatesPath, "login.html"),
		filepath.Join(templatesPath, "chat-lobby.html"),
		filepath.Join(templatesPath, "chat-room.html"),
	)
}

func (f *Factory) Start(address string) error {
	return f.Router.Run(address)
}
