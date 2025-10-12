package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtMaker struct {
	Secretkey string
}

func NewJwtMaker(secret string) Maker {
	return &JwtMaker{
		Secretkey: secret,
	}
}

func (m *JwtMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload := NewPayload(username, duration)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenstring, err := jwtToken.SignedString([]byte(m.Secretkey))
	if err != nil {
		return "", err
	}

	return tokenstring, nil
}

func (m *JwtMaker) VerifyToken(token string) (*Payload, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(m.Secretkey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyfunc)
	if err != nil {
		return nil, err
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || !jwtToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return payload, nil
}
