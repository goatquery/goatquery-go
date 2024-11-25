package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	goatquery "github.com/goatquery/goatquery-go"
	gqGorm "github.com/goatquery/goatquery-go/module/gorm"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	Id        uuid.UUID
	Firstname string
}

func main() {
	gofakeit.Seed(8675309)
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	dsn := postgresContainer.MustConnectionString(ctx)

	db, err := gorm.Open(pg.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	var users []User
	for range 1_000 {
		user := User{
			Id:        uuid.New(),
			Firstname: gofakeit.FirstName(),
		}

		users = append(users, user)
	}

	if err := db.Model(&User{}).Create(users).Error; err != nil {
		log.Fatalf("failed to create seeded users: %s", err)
	}

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		query := goatquery.Query{
			Filter: req.FormValue("filter"),
		}

		var users []User

		res, _, err := gqGorm.Apply[User](db, query, nil, nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, err.Error())
			return
		}

		if err := res.Find(&users).Error; err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	log.Println("Hosting http server on port :8080")

	http.ListenAndServe(":8080", nil)
}
