package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-faker/faker/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func main() {
	if err := openDB(); err != nil {
		panic(err)
	}
	defer closeDB()
	log.Println("Seeding users")
	users := []map[string]interface{}{}
	for i := 0; i < 200; i++ {
		users = append(users, createUser())
	}
	err := db.Table("users").Create(users).Error
	if err != nil {
		panic(err)
	}
	log.Println("Users seeded")
}

func createUser() map[string]interface{} {
	pass, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	name := faker.FirstName()
	if time.Now().Unix()%3 == 0 {
		name += " " + faker.LastName()
	}
	email := faker.Email()
	return map[string]interface{}{
		"id":            faker.UUIDDigit(),
		"name":          name,
		"profile_image": fmt.Sprintf("https://api.dicebear.com/9.x/pixel-art/svg?seed=%s", email),
		"email":         email,
		"password":      string(pass),
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}
}

func openDB() error {
	if db != nil {
		return nil
	}
	dsn := getSecretEnv("POSTGRESQL_URL")
	if dsn == "" {
		return errors.New("POSTGRESQL_URL env variable not set")
	}
	postgres, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		CreateBatchSize: 100,
	})
	if err != nil {
		return errors.New("Failed to connect to Postgres: " + err.Error())
	}
	db = postgres
	return nil
}

func closeDB() error {
	db, _ := db.DB()
	return db.Close()
}

func getSecretEnv(key string) string {
	s := os.Getenv(key)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(d)
}
