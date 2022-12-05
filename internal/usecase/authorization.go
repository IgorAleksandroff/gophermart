package usecase

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/golang-jwt/jwt/v4"
)

//go:generate mockery --name Authorization
//go:generate mockery --name AuthorizationRepository

const (
	salt     = "hjjrhjqw134617ajfhajs"
	key      = "qlkjk#4#%35FSFJlja#4253KSFjH"
	tokenTTL = time.Hour
)

var ErrUserLogin = errors.New("invalid password or login")

type authService struct {
	repo UserRepository
}

type Authorization interface {
	CreateUser(ctx context.Context, user entity.User) error
	GenerateToken(ctx context.Context, username, password string) (string, error)
	ParseToken(token string) (string, error)
}

type UserRepository interface {
	SaveUser(ctx context.Context, user entity.User) error
	GetUser(ctx context.Context, login string) (entity.User, error)
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserLogin string `json:"login"`
}

func NewAuthorization(repo UserRepository) *authService {
	return &authService{repo: repo}
}

func (s *authService) CreateUser(ctx context.Context, user entity.User) error {
	user.Password = generatePasswordHash(user.Password)
	return s.repo.SaveUser(ctx, user)
}

func (s *authService) GenerateToken(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetUser(ctx, username)
	if err != nil {
		return "", err
	}
	if user.Password != generatePasswordHash(password) {
		return "", ErrUserLogin
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(tokenTTL)},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
		UserLogin: user.Login,
	})

	return token.SignedString([]byte(key))
}

func (s *authService) ParseToken(accessToken string) (string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(key), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return "", errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserLogin, nil
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
