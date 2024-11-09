package gorm

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/goatquery/goatquery-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Base struct {
	Age uint `gorm:"column:user_age"`
}

type User struct {
	Base

	UserId      uuid.UUID
	Firstname   string
	Balance     *float64
	DateOfBirth time.Time
	JsonProp    string `json:"random_json_name"`
}

var DB *gorm.DB

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func timeMustParse(value string) time.Time {
	val, err := time.Parse("2006-01-02 15:04:05", value)
	if err == nil {
		return val
	}

	val, err = time.Parse("2006-01-02", value)
	if err != nil {
		panic("unable to parse date time")
	}

	return val
}

var users = map[string]User{
	"John":  {Base: Base{Age: 2}, Firstname: "John", UserId: uuid.MustParse("58cdeca3-645b-457c-87aa-7d5f87734255"), DateOfBirth: timeMustParse("2004-01-31 23:59:59"), Balance: makePointer(1.50), JsonProp: "user_john"},
	"Jane":  {Base: Base{Age: 1}, Firstname: "Jane", UserId: uuid.MustParse("58cdeca3-645b-457c-87aa-7d5f87734255"), DateOfBirth: timeMustParse("2020-05-09 15:30:00"), Balance: makePointer(0.0)},
	"Apple": {Base: Base{Age: 2}, Firstname: "Apple", UserId: uuid.MustParse("58cdeca3-645b-457c-87aa-7d5f87734255"), DateOfBirth: timeMustParse("1980-12-31 00:00:01"), Balance: makePointer(1204050.98)},
	"Harry": {Base: Base{Age: 1}, Firstname: "Harry", UserId: uuid.MustParse("e4c7772b-8947-4e46-98ed-644b417d2a08"), DateOfBirth: timeMustParse("2002-08-01"), Balance: makePointer(0.5372958205929493)},
	"Doe":   {Base: Base{Age: 3}, Firstname: "Doe", UserId: uuid.MustParse("58cdeca3-645b-457c-87aa-7d5f87734255"), DateOfBirth: timeMustParse("2023-07-26 12:00:30"), Balance: nil},
	"Egg":   {Base: Base{Age: 3}, Firstname: "Egg", UserId: uuid.MustParse("58cdeca3-645b-457c-87aa-7d5f87734255"), DateOfBirth: timeMustParse("2000-01-01 00:00:00"), Balance: makePointer(1334534453453433.33435443343231235652)},
}

func setup() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("failed to connect database")
	}
	DB = db

	db.AutoMigrate(&User{})

	if err := db.Model(&User{}).Create([]User{
		users["John"],
		users["Jane"],
		users["Apple"],
		users["Harry"],
		users["Doe"],
		users["Egg"],
	}).Error; err != nil {
		panic("failed to seed")
	}
}

func Test_OrderBy(t *testing.T) {
	tests := []struct {
		input    string
		expected []User
	}{
		{"age desc, firstname asc", []User{
			users["Doe"],
			users["Egg"],
			users["Apple"],
			users["John"],
			users["Harry"],
			users["Jane"],
		}},
		{"age desc, firstname desc", []User{
			users["Egg"],
			users["Doe"],
			users["John"],
			users["Apple"],
			users["Jane"],
			users["Harry"],
		}},
		{"age desc", []User{
			users["Doe"],
			users["Egg"],
			users["John"],
			users["Apple"],
			users["Jane"],
			users["Harry"],
		}},
		{"Age asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"age", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"age asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"Age asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"aGE asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"AGe asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"aGE Asc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"age aSc", []User{
			users["Jane"],
			users["Harry"],
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"", []User{
			users["John"],
			users["Jane"],
			users["Apple"],
			users["Harry"],
			users["Doe"],
			users["Egg"],
		}},
	}

	for _, test := range tests {
		query := goatquery.Query{
			OrderBy: test.input,
		}

		res, _, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Equal(t, test.expected, output)
	}
}

func Test_Count(t *testing.T) {
	tests := []struct {
		input         bool
		expectedCount *int64
	}{
		{true, makePointer(int64(6))},
		{false, nil},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Count: test.input,
		}

		_, count, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, test.expectedCount, count)
	}
}

func Test_Search(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
	}{
		{"john", 1},
		{"JOHN", 1},
		{"j", 2},
		{"e", 4},
		{"eg", 1},
	}

	searchFunc := func(db *gorm.DB, searchTerm string) *gorm.DB {
		return db.Where("firstname like ?", fmt.Sprintf("%%%s%%", searchTerm)) // Escape % for LIKE
	}

	for _, test := range tests {
		query := goatquery.Query{
			Search: test.input,
		}

		res, _, err := Apply[User](DB, query, searchFunc, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Len(t, output, test.expectedCount)
	}
}

func Test_Skip(t *testing.T) {
	tests := []struct {
		input    int
		expected []User
	}{
		{1, []User{
			users["Jane"],
			users["Apple"],
			users["Harry"],
			users["Doe"],
			users["Egg"],
		}},
		{2, []User{
			users["Apple"],
			users["Harry"],
			users["Doe"],
			users["Egg"],
		}},
		{3, []User{
			users["Harry"],
			users["Doe"],
			users["Egg"],
		}},
		{4, []User{
			users["Doe"],
			users["Egg"],
		}},
		{5, []User{
			users["Egg"],
		}},
		{6, []User{}},
		{10_000, []User{}},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Skip: test.input,
		}

		res, _, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Equal(t, test.expected, output)
	}
}

func Test_Top(t *testing.T) {
	tests := []struct {
		input         int
		expectedCount int
	}{
		{-1, 6},
		{0, 6},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{100, 6},
		{100_000, 6},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Top: test.input,
		}

		res, _, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Len(t, output, test.expectedCount)
	}
}

func Test_TopWithMaxTop(t *testing.T) {
	tests := []struct {
		input         int
		expectedCount int
	}{
		{-1, 4},
		{0, 4},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Top: test.input,
		}

		options := goatquery.QueryOptions{
			MaxTop: 4,
		}

		res, _, err := Apply[User](DB, query, nil, &options)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Len(t, output, test.expectedCount)
	}
}

func Test_TopWithMaxTopReturnsError(t *testing.T) {
	tests := []int{
		5,
		100,
		100_000,
	}

	for _, test := range tests {
		query := goatquery.Query{
			Top: test,
		}

		options := goatquery.QueryOptions{
			MaxTop: 4,
		}

		_, _, err := Apply[User](DB, query, nil, &options)
		assert.Error(t, err)
	}
}

func Test_Filter(t *testing.T) {
	tests := []struct {
		input    string
		expected []User
	}{
		{"firstname eq 'John'", []User{
			users["John"],
		}},
		{"firstname eq 'Random'", []User{}},
		{"Age eq 1", []User{
			users["Jane"],
			users["Harry"],
		}},
		{"Age eq 0", []User{}},
		{"firstname eq 'John' and Age eq 2", []User{
			users["John"],
		}},
		{"firstname eq 'John' or Age eq 3", []User{
			users["John"],
			users["Doe"],
			users["Egg"],
		}},
		{"Age eq 1 and firstname eq 'Harry' or Age eq 2", []User{
			users["John"],
			users["Apple"],
			users["Harry"],
		}},
		{"Age eq 1 or Age eq 2 or firstname eq 'Egg'", []User{
			users["John"],
			users["Jane"],
			users["Apple"],
			users["Harry"],
			users["Egg"],
		}},
		{"Age ne 3", []User{
			users["John"],
			users["Jane"],
			users["Apple"],
			users["Harry"],
		}},
		{"firstname contains 'a'", []User{
			users["Jane"],
			users["Apple"],
			users["Harry"],
		}},
		{"Age ne 1 and firstname contains 'a'", []User{
			users["Apple"],
		}},
		{"Age ne 1 and firstname contains 'a' or firstname eq 'Apple'", []User{
			users["Apple"],
		}},
		{"firstname eq 'John' and Age eq 2 or Age eq 3", []User{
			users["John"],
			users["Doe"],
			users["Egg"],
		}},
		{"(firstname eq 'John' and Age eq 2) or Age eq 3", []User{
			users["John"],
			users["Doe"],
			users["Egg"],
		}},
		{"firstname eq 'John' and (Age eq 2 or Age eq 3)", []User{
			users["John"],
		}},
		{"(Firstname eq 'John' and Age eq 2 or Age eq 3)", []User{
			users["John"],
			users["Doe"],
			users["Egg"],
		}},
		{"(Firstname eq 'John') or (Age eq 3 and Firstname eq 'Egg') or Age eq 1 and (Age eq 2)", []User{
			users["John"],
			users["Egg"],
		}},
		{"UserId eq e4c7772b-8947-4e46-98ed-644b417d2a08", []User{
			users["Harry"],
		}},
		{"age lt 3", []User{
			users["John"],
			users["Jane"],
			users["Apple"],
			users["Harry"],
		}},
		{"age lt 1", []User{}},
		{"age lte 2", []User{
			users["John"],
			users["Jane"],
			users["Apple"],
			users["Harry"],
		}},
		{"age gt 1", []User{
			users["John"],
			users["Apple"],
			users["Doe"],
			users["Egg"],
		}},
		{"age gte 3", []User{
			users["Doe"],
			users["Egg"],
		}},
		{"age lt 3 and age gt 1", []User{
			users["John"],
			users["Apple"],
		}},
		{"balance eq 1.50f", []User{
			users["John"],
		}},
		{"balance gt 1f", []User{
			users["John"], users["Apple"], users["Egg"],
		}},
		{"balance gt 0.50f", []User{
			users["John"], users["Apple"], users["Harry"], users["Egg"],
		}},
		{"balance eq 0.5372958205929493f", []User{
			users["Harry"],
		}},
		{"balance eq 1334534453453433.33435443343231235652f", []User{
			users["Egg"],
		}},
		{"balance eq 1204050.98f", []User{
			users["Apple"],
		}},
		{"balance gt 2204050f", []User{
			users["Egg"],
		}},
		{"dateOfBirth eq 2000-01-01", []User{
			users["Egg"],
		}},
		{"dateOfBirth lt 2010-01-01", []User{
			users["John"], users["Apple"], users["Harry"], users["Egg"],
		}},
		{"dateOfBirth lte 2002-08-01", []User{
			users["Apple"], users["Harry"], users["Egg"],
		}},
		{"dateOfBirth gt 2000-08-01 and dateOfBirth lt 2023-01-01", []User{
			users["John"], users["Jane"], users["Harry"],
		}},
		{"dateOfBirth eq 2023-07-26T12:00:30Z", []User{
			users["Doe"],
		}},
		{"dateOfBirth gte 2000-01-01", []User{
			users["John"], users["Jane"], users["Harry"], users["Doe"], users["Egg"],
		}},
		{"dateOfBirth gte 2000-01-01 and dateOfBirth lte 2020-05-09T15:29:59Z", []User{
			users["John"], users["Harry"], users["Egg"],
		}},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Filter: test.input,
		}

		res, _, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Equal(t, test.expected, output)
	}
}

func Test_InvalidFilterReturnsError(t *testing.T) {
	input := `NonExistentProperty eq 'John'`

	query := goatquery.Query{
		Filter: input,
	}

	_, _, err := Apply[User](DB, query, nil, nil)
	assert.Error(t, err)
}

func Test_Filter_WithCustomJsonTag(t *testing.T) {
	tests := []struct {
		input    string
		expected []User
	}{
		{"random_json_name eq 'user_john'", []User{
			users["John"],
		}},
	}

	for _, test := range tests {

		query := goatquery.Query{
			Filter: test.input,
		}

		res, _, err := Apply[User](DB, query, nil, nil)
		assert.NoError(t, err)

		var output []User
		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Equal(t, test.expected, output)
	}
}

func makePointer[T any](v T) *T {
	return &v
}
