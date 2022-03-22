package paseto

import (
	"fmt"
	tokenauth "github.com/gmaschi/jobsity-go-financial-chat/pkg/auth/token-auth"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
	"time"
)

// Maker is a PASETO token maker
type Maker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewMaker creates a new Maker
func NewMaker(symmetricKey string) (tokenauth.Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: key must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &Maker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return maker, nil
}

func (m *Maker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := tokenauth.NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	return m.paseto.Encrypt(m.symmetricKey, payload, nil)
}

func (m *Maker) VerifyToken(token string) (*tokenauth.Payload, error) {
	payload := &tokenauth.Payload{}

	err := m.paseto.Decrypt(token, m.symmetricKey, payload, nil)
	if err != nil {
		return nil, err
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}
