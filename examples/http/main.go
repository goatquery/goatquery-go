package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"goatquery"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	AvatarUrl string    `json:"avatarUrl"`
	IsDeleted bool      `json:"-"`
}

type UserDto struct {
	Id        uuid.UUID `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	AvatarUrl string    `json:"-"`
}

var DB *gorm.DB

func main() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("failed to connect database")
	}
	DB = db

	db.AutoMigrate(&User{})

	gofakeit.Seed(123)

	if err := db.First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		items := []User{}

		for i := 1; i < 1_000; i++ {
			person := gofakeit.Person()

			item := User{
				Id:        uuid.New(),
				Firstname: person.FirstName,
				Lastname:  person.LastName,
				Email:     person.Contact.Email,
				AvatarUrl: gofakeit.ImageURL(64, 64),
				IsDeleted: gofakeit.Bool(),
			}

			items = append(items, item)
		}

		if err := db.CreateInBatches(items, 1_000).Error; err != nil {
			fmt.Printf("error occured while seeding data: %v\n", err)
			return
		}

		fmt.Printf("Seeded data...\n")
	}

	http.HandleFunc("/users", getUsers)

	fmt.Printf("server starting on port :8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query, err := getQuery(r.URL.Query())
	if err != nil {
		w.Write(createQueryError(err.Error()))
		return
	}

	var users []UserDto
	res, count, err := goatquery.Apply(GetAllUsers(DB), *query, nil, nil)
	if err != nil {
		w.Write(createQueryError(err.Error()))
		return
	}
	if err := res.Find(&users).Error; err != nil {
		w.Write(createQueryError(err.Error()))
		return
	}

	response := goatquery.BuildPagedResponse(users, *query, count)

	json, err := json.Marshal(response)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	w.Write(json)
}

func GetAllUsers(db *gorm.DB) *gorm.DB {
	return db.Model(&User{}).Where("is_deleted <> ?", true)
}

func getQuery(query url.Values) (*goatquery.Query, error) {
	topQuery, err := strconv.Atoi(query.Get("top"))
	if err != nil && query.Has("top") {
		return nil, err
	}

	skipQuery, err := strconv.Atoi(query.Get("skip"))
	if err != nil && query.Has("skip") {
		return nil, err
	}

	countQuery, err := strconv.ParseBool(query.Get("count"))
	if err != nil && query.Has("count") {
		return nil, err
	}

	result := goatquery.Query{
		Top:     topQuery,
		Skip:    skipQuery,
		Count:   countQuery,
		OrderBy: query.Get("orderby"),
		Select:  query.Get("select"),
		Search:  query.Get("search"),
		Filter:  query.Get("filter"),
	}

	return &result, nil
}

func createQueryError(message string) []byte {
	response := goatquery.QueryErrorResponse{
		Status:  400,
		Message: message,
	}

	json, err := json.Marshal(response)
	if err != nil {
		return []byte{}
	}

	return json
}
