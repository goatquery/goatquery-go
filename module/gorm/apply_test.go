package gorm

import (
	"fmt"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/goatquery/goatquery-go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Base struct {
	Age uint `gorm:"column:user_age"`
}

type User struct {
	Base

	Firstname string
}

var DB *gorm.DB

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		panic("failed to connect database")
	}
	DB = db

	db.AutoMigrate(&User{})

	if err := db.Model(&User{}).Create([]User{
		{Base{2}, "John"},
		{Base{1}, "Jane"},
		{Base{2}, "Apple"},
		{Base{1}, "Harry"},
		{Base{3}, "Doe"},
		{Base{3}, "Egg"},
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
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
			{Base{2}, "Apple"},
			{Base{2}, "John"},
			{Base{1}, "Harry"},
			{Base{1}, "Jane"},
		}},
		// {"age desc, firstname desc", []User{
		// 	{Base{3}, "Egg"},
		// 	{Base{3}, "Doe"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// }},
		// {"age desc", []User{
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// }},
		// {"Age asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"age", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"age asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"Age asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"aGE asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"AGe asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"aGE Asc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"age aSc", []User{
		// 	{Base{1}, "Jane"},
		// 	{Base{1}, "Harry"},
		// 	{Base{2}, "John"},
		// 	{Base{2}, "Apple"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
		// {"", []User{
		// 	{Base{2}, "John"},
		// 	{Base{1}, "Jane"},
		// 	{Base{2}, "Apple"},
		// 	{Base{1}, "Harry"},
		// 	{Base{3}, "Doe"},
		// 	{Base{3}, "Egg"},
		// }},
	}

	for _, test := range tests {
		query := goatquery.Query{
			OrderBy: test.input,
		}

		var output []User
		res, _, err := Apply(DB, query, &output, nil, nil)
		assert.NoError(t, err)

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
		{true, makeIntPointer(6)},
		{false, nil},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Count: test.input,
		}

		_, count, err := Apply(DB, query, &User{}, nil, nil)
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

		var output []User
		res, _, err := Apply(DB, query, &output, searchFunc, nil)
		assert.NoError(t, err)

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
			{Base{1}, "Jane"},
			{Base{2}, "Apple"},
			{Base{2}, "John"},
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
		}},
		{2, []User{
			{Base{2}, "Apple"},
			{Base{2}, "John"},
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
		}},
		{3, []User{
			{Base{2}, "John"},
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
		}},
		{4, []User{
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
		}},
		{5, []User{
			{Base{3}, "Egg"},
		}},
		{6, []User{}},
		{10_000, []User{}},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Skip:    test.input,
			OrderBy: "age asc, firstname asc",
		}

		var output []User
		res, _, err := Apply(DB, query, &output, nil, nil)
		assert.NoError(t, err)

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

		var output []User
		res, _, err := Apply(DB, query, &output, nil, nil)
		assert.NoError(t, err)

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

		var output []User
		res, _, err := Apply(DB, query, &output, nil, &options)
		assert.NoError(t, err)

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

		_, _, err := Apply(DB, query, &User{}, nil, &options)
		assert.Error(t, err)
	}
}

func Test_Filter(t *testing.T) {
	tests := []struct {
		input    string
		expected []User
	}{
		{"firstname eq 'John'", []User{
			{Base{2}, "John"},
		}},
		{"firstname eq 'Random'", []User{}},
		{"Age eq 1", []User{
			{Base{1}, "Jane"},
			{Base{1}, "Harry"},
		}},
		{"Age eq 0", []User{}},
		{"firstname eq 'John' and Age eq 2", []User{
			{Base{2}, "John"},
		}},
		{"firstname eq 'John' or Age eq 3", []User{
			{Base{2}, "John"},
			{Base{3}, "Doe"},
			{Base{3}, "Egg"},
		}},
		{"Age eq 1 and firstName eq 'Harry' or Age eq 2", []User{
			{Base{2}, "John"},
			{Base{2}, "Apple"},
			{Base{1}, "Harry"},
		}},
		{"Age eq 1 or Age eq 2 or firstName eq 'Egg'", []User{
			{Base{2}, "John"},
			{Base{1}, "Jane"},
			{Base{2}, "Apple"},
			{Base{1}, "Harry"},
			{Base{3}, "Egg"},
		}},
		{"Age ne 3", []User{
			{Base{2}, "John"},
			{Base{1}, "Jane"},
			{Base{2}, "Apple"},
			{Base{1}, "Harry"},
		}},
		{"firstName contains 'a'", []User{
			{Base{1}, "Jane"},
			{Base{2}, "Apple"},
			{Base{1}, "Harry"},
		}},
		{"Age ne 1 and firstName contains 'a'", []User{
			{Base{2}, "Apple"},
		}},
		{"Age ne 1 and firstName contains 'a' or firstName eq 'Apple'", []User{
			{Base{2}, "Apple"},
		}},
	}

	for _, test := range tests {
		query := goatquery.Query{
			Filter: test.input,
		}

		var output []User
		res, _, err := Apply(DB, query, &output, nil, nil)
		assert.NoError(t, err)

		err = res.Find(&output).Error
		assert.NoError(t, err)

		assert.Equal(t, test.expected, output)
	}
}

func makeIntPointer(v int64) *int64 {
	return &v
}
