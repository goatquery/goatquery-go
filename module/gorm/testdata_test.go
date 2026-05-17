package gorm

import (
	"time"

	"github.com/google/uuid"
)

// --- Test Models ---

// Gender is a custom string type to test filtering on named string types.
type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

type Company struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name       string
	Department string
}

type City struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name    string
	Country string
}

type Address struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	AddressLine1 string
	CityID       uuid.UUID `gorm:"type:uuid"`
	City         City
	UserID       uuid.UUID `gorm:"type:uuid"`
}

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Age             int
	Firstname       string
	Lastname        *string `json:"last_name"`
	Gender          *Gender
	BalanceDecimal  *float64
	BalanceDouble   *float64
	BalanceFloat    *float64
	DateOfBirth     *time.Time
	IsEmailVerified bool
	LargeNumber     *int64
	CompanyID       *uuid.UUID `gorm:"type:uuid"`
	Company         *Company
	ManagerID       *uuid.UUID `gorm:"type:uuid"`
	Manager         *User
	Addresses       []Address
	Tags            []string `gorm:"type:jsonb;serializer:json"`
}

// --- Seed Data ---

var (
	user01ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	user02ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	user03ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	user04ID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	user05ID = uuid.MustParse("55555555-5555-5555-5555-555555555555")

	company1ID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	company2ID = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	company3ID = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	company5ID = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	city1ID = uuid.MustParse("11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	city2ID = uuid.MustParse("22222222-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	city3ID = uuid.MustParse("33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	city4ID = uuid.MustParse("44444444-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	addr1ID = uuid.MustParse("11111111-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	addr2ID = uuid.MustParse("22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	addr3ID = uuid.MustParse("33333333-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	addr4ID = uuid.MustParse("44444444-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
)

func timeMustParse(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t, err = time.Parse(time.DateOnly, value)
		if err != nil {
			panic("unable to parse time: " + value)
		}
	}
	return t
}

func ptr[T any](v T) *T {
	return &v
}

var seedCompanies = []Company{
	{ID: company1ID, Name: "TechCorp", Department: "Engineering"},
	{ID: company2ID, Name: "DataSoft", Department: "Development"},
	{ID: company3ID, Name: "Tech Solutions", Department: "Sales"},
	{ID: company5ID, Name: "WebCorp", Department: "Marketing"},
}

var seedCities = []City{
	{ID: city1ID, Name: "New York", Country: "USA"},
	{ID: city2ID, Name: "Chicago", Country: "USA"},
	{ID: city3ID, Name: "Seattle", Country: "USA"},
	{ID: city4ID, Name: "Miami", Country: "USA"},
}

var seedAddresses = []Address{
	{ID: addr1ID, AddressLine1: "123 Main St", CityID: city1ID, UserID: user01ID},
	{ID: addr2ID, AddressLine1: "456 Oak Ave", CityID: city2ID, UserID: user01ID},
	{ID: addr3ID, AddressLine1: "789 Pine St", CityID: city3ID, UserID: user02ID},
	{ID: addr4ID, AddressLine1: "999 Broadway", CityID: city4ID, UserID: user05ID},
}

var seedUsers = []User{
	{
		ID:              user01ID,
		Age:             25,
		Firstname:       "User01",
		Lastname:        ptr("Smith"),
		Gender:          ptr(GenderMale),
		BalanceDecimal:  ptr(1500.75),
		BalanceDouble:   ptr(2500.50),
		BalanceFloat:    ptr(3500.25),
		DateOfBirth:     ptr(timeMustParse("1998-03-15T10:30:00Z")),
		IsEmailVerified: true,
		LargeNumber:     ptr(int64(99999999999)),
		CompanyID:       &company1ID,
		ManagerID:       nil,
		Tags:            []string{"vip", "premium"},
	},
	{
		ID:              user02ID,
		Age:             30,
		Firstname:       "User02",
		Lastname:        ptr("Jones"),
		Gender:          ptr(GenderFemale),
		BalanceDecimal:  ptr(500.00),
		BalanceDouble:   nil,
		BalanceFloat:    ptr(750.50),
		DateOfBirth:     ptr(timeMustParse("1993-07-20T14:00:00Z")),
		IsEmailVerified: false,
		LargeNumber:     ptr(int64(42)),
		CompanyID:       &company2ID,
		ManagerID:       &user01ID,
		Tags:            []string{"premium"},
	},
	{
		ID:              user03ID,
		Age:             30,
		Firstname:       "User03",
		Lastname:        nil,
		Gender:          nil,
		BalanceDecimal:  nil,
		BalanceDouble:   ptr(1000.00),
		BalanceFloat:    nil,
		DateOfBirth:     ptr(timeMustParse("1993-11-10T09:15:00Z")),
		IsEmailVerified: true,
		LargeNumber:     nil,
		CompanyID:       &company3ID,
		ManagerID:       &user02ID,
		Tags:            []string{},
	},
	{
		ID:              user04ID,
		Age:             35,
		Firstname:       "User04",
		Lastname:        nil,
		Gender:          ptr(GenderMale),
		BalanceDecimal:  nil,
		BalanceDouble:   nil,
		BalanceFloat:    nil,
		DateOfBirth:     nil,
		IsEmailVerified: false,
		LargeNumber:     nil,
		CompanyID:       nil,
		ManagerID:       &user02ID,
		Tags:            []string{},
	},
	{
		ID:              user05ID,
		Age:             25,
		Firstname:       "User05",
		Lastname:        ptr("Williams"),
		Gender:          ptr(GenderFemale),
		BalanceDecimal:  ptr(0.00),
		BalanceDouble:   ptr(0.00),
		BalanceFloat:    nil,
		DateOfBirth:     ptr(timeMustParse("1998-12-25T18:45:00Z")),
		IsEmailVerified: true,
		LargeNumber:     nil,
		CompanyID:       &company5ID,
		ManagerID:       nil,
		Tags:            []string{"standard"},
	},
}
