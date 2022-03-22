package usermodel

import "time"

type (
	CreateResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"-"`
	}

	GetResponse struct {
		Username       string    `json:"username"`
		HashedPassword string    `json:"-"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	LoginResponse struct {
		AccessToken    string `json:"access_token"`
		Username       string `json:"username"`
		HashedPassword string `json:"-"`
	}
)
