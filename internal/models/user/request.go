package usermodel

type (
	CreateRequest struct {
		Username string `json:"username" binding:"required,alphanum"`
		Password string `json:"password" binding:"required,min=6"`
	}

	GetRequest struct {
		Username string `uri:"username" binding:"required,alphanum"`
	}

	LoginRequest struct {
		Username string `json:"username" binding:"required,alphanum"`
		Password string `json:"password" binding:"required,min=6"`
	}
)
