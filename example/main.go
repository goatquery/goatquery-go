package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	goatquery "github.com/goatquery/goatquery-go"
	goatgorm "github.com/goatquery/goatquery-go/module/gorm"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Models ---

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Firstname       string     `json:"first_name"`
	Lastname        string     `json:"lastname"`
	Age             int        `json:"age"`
	Gender          string     `json:"gender"`
	IsDeleted       bool       `json:"isDeleted"`
	IsEmailVerified bool       `json:"isEmailVerified"`
	Test            float64    `json:"test"`
	NullableInt     *int       `json:"nullableInt"`
	DateOfBirth     time.Time  `json:"dateOfBirth"`
	ManagerID       *uuid.UUID `gorm:"type:uuid" json:"-"`
	Manager         *User      `json:"manager"`
	CompanyID       *uuid.UUID `gorm:"type:uuid" json:"-"`
	Company         *Company   `json:"company"`
	Addresses       []Address  `json:"addresses"`
	Orders          []Order    `json:"orders"`
	Tags            []string   `gorm:"type:jsonb;serializer:json" json:"tags"`
}

type Company struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string    `json:"name"`
	Department string    `json:"department"`
}

type Address struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid" json:"-"`
	AddressLine1 string    `json:"addressLine1"`
	CityID       uuid.UUID `gorm:"type:uuid" json:"-"`
	City         City      `json:"city"`
}

type City struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name    string    `json:"name"`
	Country string    `json:"country"`
}

type Order struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID   `gorm:"type:uuid" json:"-"`
	OrderNumber string      `json:"orderNumber"`
	OrderDate   time.Time   `json:"orderDate"`
	Total       float64     `json:"total"`
	Status      string      `json:"status"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid" json:"-"`
	ProductID uuid.UUID `gorm:"type:uuid" json:"-"`
	Product   Product   `json:"product"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unitPrice"`
}

type Product struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
}

func main() {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("goatquery_example"),
		tcpostgres.WithUsername("example"),
		tcpostgres.WithPassword("example"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Migrate
	db.AutoMigrate(&Company{}, &City{}, &Product{}, &User{}, &Address{}, &Order{}, &OrderItem{})

	// Seed if empty
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		seed(db)
	}

	mux := http.NewServeMux()

	// GET /users — full-featured queryable endpoint
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		query := parseQuery(r)
		options := &goatquery.QueryOptions{MaxTop: 50}

		searchFunc := func(db *gorm.DB, term string) *gorm.DB {
			like := "%" + term + "%"
			return db.Where("firstname LIKE ? OR lastname LIKE ?", like, like)
		}

		tx := db.Model(&User{}).
			Where("users.is_deleted = ?", false).
			Preload("Company").
			Preload("Addresses.City").
			Preload("Manager.Manager").
			Preload("Orders.Items.Product")

		result, count, err := goatgorm.Apply[User](tx, query, searchFunc, options)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var users []User
		if err := result.Find(&users).Error; err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, goatquery.PagedResponse[User]{Value: users, Count: count})
	})

	// GET /products — simpler endpoint
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		query := parseQuery(r)
		options := &goatquery.QueryOptions{MaxTop: 100}

		result, count, err := goatgorm.Apply[Product](db.Model(&Product{}), query, nil, options)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var products []Product
		if err := result.Find(&products).Error; err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, goatquery.PagedResponse[Product]{Value: products, Count: count})
	})

	// GET /orders — orders endpoint with search
	mux.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {
		query := parseQuery(r)
		options := &goatquery.QueryOptions{MaxTop: 50}

		searchFunc := func(db *gorm.DB, term string) *gorm.DB {
			like := "%" + term + "%"
			return db.Where("order_number LIKE ?", like)
		}

		tx := db.Model(&Order{}).Preload("Items.Product")

		result, count, err := goatgorm.Apply[Order](tx, query, searchFunc, options)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		var orders []Order
		if err := result.Find(&orders).Error; err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, goatquery.PagedResponse[Order]{Value: orders, Count: count})
	})

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// parseQuery extracts GoatQuery parameters from the request URL query string.
func parseQuery(r *http.Request) goatquery.Query {
	q := r.URL.Query()
	query := goatquery.Query{
		Filter:  q.Get("filter"),
		OrderBy: q.Get("orderby"),
		Search:  q.Get("search"),
	}

	if v := q.Get("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			query.Top = &n
		}
	}
	if v := q.Get("skip"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			query.Skip = &n
		}
	}
	if v := q.Get("count"); v == "true" {
		b := true
		query.Count = &b
	}
	return query
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// --- Seed data ---

func seed(db *gorm.DB) {
	fake := gofakeit.New(42)

	// Products
	products := make([]Product, 50)
	for i := range products {
		products[i] = Product{
			ID:       uuid.New(),
			Name:     fake.ProductName(),
			Category: fake.RandomString([]string{"Electronics", "Clothing", "Books", "Home", "Sports"}),
			Price:    fake.Price(5, 500),
		}
	}
	db.Create(&products)

	// Companies
	companies := make([]Company, 20)
	for i := range companies {
		companies[i] = Company{
			ID:         uuid.New(),
			Name:       fake.Company(),
			Department: fake.RandomString([]string{"Engineering", "Sales", "Marketing", "Support", "Finance"}),
		}
	}
	db.Create(&companies)

	// Cities
	cities := make([]City, 50)
	for i := range cities {
		cities[i] = City{
			ID:      uuid.New(),
			Name:    fake.City(),
			Country: fake.Country(),
		}
	}
	db.Create(&cities)

	// Users
	statuses := []string{"Pending", "Shipped", "Delivered", "Cancelled"}
	tagValues := []string{"vip", "new", "premium", "trial", "enterprise", "beta"}

	for i := 0; i < 1000; i++ {
		companyID := companies[fake.Number(0, len(companies)-1)].ID

		var nullableInt *int
		if fake.Bool() {
			v := fake.Number(1, 100)
			nullableInt = &v
		}

		// Generate random tags (0-5)
		numTags := fake.Number(0, 5)
		tags := make([]string, numTags)
		for j := range tags {
			tags[j] = fake.RandomString(tagValues)
		}

		user := User{
			ID:              uuid.New(),
			Firstname:       fake.FirstName(),
			Lastname:        fake.LastName(),
			Age:             fake.Number(18, 78),
			Gender:          fake.Gender(),
			IsDeleted:       fake.Bool(),
			IsEmailVerified: fake.Bool(),
			Test:            fake.Float64(),
			NullableInt:     nullableInt,
			DateOfBirth:     fake.DateRange(time.Now().AddDate(-68, 0, 0), time.Now().AddDate(-18, 0, 0)),
			CompanyID:       &companyID,
			Tags:            tags,
		}
		db.Create(&user)

		// Manager (60% chance, up to 2 levels deep)
		if fake.Float64Range(0, 1) < 0.6 {
			manager := createManager(fake, 3)
			db.Create(&manager)
			db.Model(&user).Update("manager_id", manager.ID)
		}

		// Addresses (1-3)
		numAddrs := fake.Number(1, 3)
		for j := 0; j < numAddrs; j++ {
			db.Create(&Address{
				ID:           uuid.New(),
				UserID:       user.ID,
				AddressLine1: fake.StreetNumber() + " " + fake.StreetName(),
				CityID:       cities[fake.Number(0, len(cities)-1)].ID,
			})
		}

		// Orders (0-5)
		numOrders := fake.Number(0, 5)
		for j := 0; j < numOrders; j++ {
			order := Order{
				ID:          uuid.New(),
				UserID:      user.ID,
				OrderNumber: fmt.Sprintf("ORD-%s-%s", fake.DigitN(4), fake.DigitN(4)),
				OrderDate:   fake.DateRange(time.Now().AddDate(-2, 0, 0), time.Now()),
				Status:      fake.RandomString(statuses),
			}

			// Items (1-5)
			numItems := fake.Number(1, 5)
			var total float64
			for k := 0; k < numItems; k++ {
				prod := products[fake.Number(0, len(products)-1)]
				qty := fake.Number(1, 10)
				item := OrderItem{
					ID:        uuid.New(),
					ProductID: prod.ID,
					Quantity:  qty,
					UnitPrice: prod.Price,
				}
				order.Items = append(order.Items, item)
				total += prod.Price * float64(qty)
			}
			order.Total = total
			db.Create(&order)
		}
	}

	fmt.Println("Seeded 1,000 users with orders, products, addresses, and companies")
}

func createManager(fake *gofakeit.Faker, depth int) User {
	manager := User{
		ID:              uuid.New(),
		Firstname:       fake.FirstName(),
		Lastname:        fake.LastName(),
		Age:             fake.Number(30, 60),
		Gender:          fake.RandomString([]string{"Male", "Female", "Other"}),
		IsDeleted:       false,
		IsEmailVerified: true,
		Test:            fake.Float64(),
		DateOfBirth:     fake.DateRange(time.Now().AddDate(-60, 0, 0), time.Now().AddDate(-30, 0, 0)),
	}

	if depth > 1 && fake.Float64Range(0, 1) < 0.4 {
		nested := createManager(fake, depth-1)
		manager.Manager = &nested
	}

	return manager
}
