package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterNewUserInput struct {
	Name         string
	ProfileImage *os.File
	Email        string
	Password     string
}

type LoginNRegisterOutput struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
}

func RegisterNewUser(in *RegisterNewUserInput) (*LoginNRegisterOutput, error) {
	tx := DB()
	user := &User{}
	if err := tx.Model(user).
		Where("email = ?", in.Email).
		First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if !user.ID.IsEmpty() {
		return nil, fiber.NewError(fiber.StatusBadRequest, "user is already registered")
	}
	user.ID = NewUUIDv4()
	user.Name = in.Name
	user.Email = in.Email
	user.ProfileImageIcon = StringVar(fmt.Sprintf("https://api.dicebear.com/9.x/pixel-art/svg?seed=%s", in.Email))
	pass, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(pass)
	token, err := createToken(user.ID.String())
	if err != nil {
		return nil, err
	}
	if err := tx.Create(user).Error; err != nil {
		return nil, err
	}
	return &LoginNRegisterOutput{
		User:        user,
		AccessToken: token,
	}, nil

}

type LoginUserInput struct {
	Email    string
	Password string
}

func LoginUser(in *LoginUserInput) (*LoginNRegisterOutput, error) {
	tx := DB()
	user := &User{}
	if err := tx.Model(user).
		Where("email = ?", in.Email).
		First(user).Error; err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user.ID.IsEmpty() {
		return nil, fiber.NewError(fiber.StatusBadRequest, "user is not registered")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "password is incorrect")
	}
	token, err := createToken(user.ID.String())
	if err != nil {
		return nil, err
	}
	return &LoginNRegisterOutput{
		User:        user,
		AccessToken: token,
	}, nil
}

func GetUserWithToken(tokenString string) (*User, error) {
	token, err := verifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return nil, err
	}
	tx := DB()
	user := &User{}
	if err := tx.Model(user).
		Where("id = ?", sub).
		First(user).Error; err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user.ID.IsEmpty() {
		return nil, fiber.NewError(fiber.StatusBadRequest, "user is not registered")
	}
	return user, nil
}

func createToken(username string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,                                  // Subject (user identifier)
		"iss": "spock",                                   // Issuer
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                         // Issued at
	})
	jwtToken := os.Getenv("JWT_SECRET")
	if jwtToken == "" {
		return "", errors.New("JWT_SECRET env variable not set")
	}
	tokenString, err := claims.SignedString([]byte(jwtToken))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	jwtToken := os.Getenv("JWT_SECRET")
	if jwtToken == "" {
		return nil, errors.New("JWT_SECRET env variable not set")
	}
	// Parse the token with the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtToken), nil
	})
	// Check for verification errors
	if err != nil {
		return nil, err
	}
	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	// Return the verified token
	return token, nil
}
