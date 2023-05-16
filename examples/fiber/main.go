package main

import (
	"errors"
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofiber/fiber/v2"
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

	app := fiber.New()

	app.Get("/users", getUsers)

	app.Listen(":8080")
}

func getUsers(c *fiber.Ctx) error {
	query := goatquery.Query{
		Top:     c.QueryInt("top", 0),
		Skip:    c.QueryInt("skip", 0),
		Count:   c.QueryBool("count", false),
		OrderBy: c.Query("orderby"),
		Select:  c.Query("select"),
		Search:  c.Query("search"),
		Filter:  c.Query("filter"),
	}

	var users []UserDto
	res, count, err := goatquery.Apply(GetAllUsers(DB), query, nil, nil)
	if err != nil {
		return c.Status(400).JSON(goatquery.QueryErrorResponse{Status: 400, Message: err.Error()})
	}

	if err := res.Find(&users).Error; err != nil {
		return c.Status(400).JSON(goatquery.QueryErrorResponse{Status: 400, Message: err.Error()})
	}

	response := goatquery.BuildResponseFastHttp(users, query, c.Response(), count)

	return c.JSON(response)
}

func GetAllUsers(db *gorm.DB) *gorm.DB {
	return db.Model(&User{}).Where("is_deleted <> ?", true)
}

func UserDtoSearch(db *gorm.DB, searchTerm string) *gorm.DB {
	t := fmt.Sprintf("%%%s%%", searchTerm)

	return db.Where("firstname like ? or lastname like ?", t, t)
}
