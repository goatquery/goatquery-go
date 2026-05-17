package gorm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goatquery "github.com/goatquery/goatquery-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("goatquery_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start postgres container: %v", err))
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}

	DB, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Migrate
	if err := DB.AutoMigrate(&City{}, &Company{}, &User{}, &Address{}); err != nil {
		panic(fmt.Sprintf("failed to migrate: %v", err))
	}

	// Seed
	DB.Create(&seedCities)
	DB.Create(&seedCompanies)
	DB.Create(&seedUsers)
	DB.Create(&seedAddresses)

	code := m.Run()

	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}

// --- Top Tests ---

func Test_Top(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{5, 5},
		{100, 5},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("top_%d", test.input), func(t *testing.T) {
			query := goatquery.Query{Top: ptr(test.input)}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
		})
	}
}

func Test_TopNil(t *testing.T) {
	query := goatquery.Query{}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 5)
}

func Test_TopWithMaxTop(t *testing.T) {
	tests := []struct {
		name     string
		top      *int
		expected int
	}{
		{"nil uses maxTop", nil, 4},
		{"zero uses maxTop", ptr(0), 4},
		{"explicit top", ptr(2), 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Top: test.top}
			options := goatquery.QueryOptions{MaxTop: 4}

			res, _, err := Apply[User](DB, query, nil, &options)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
		})
	}
}

func Test_TopExceedsMaxTopReturnsError(t *testing.T) {
	query := goatquery.Query{Top: ptr(100)}
	options := goatquery.QueryOptions{MaxTop: 4}

	_, _, err := Apply[User](DB, query, nil, &options)
	assert.Error(t, err)
}

func Test_TopZero(t *testing.T) {
	// Top=0 with MaxTop configured should apply MaxTop
	query := goatquery.Query{Top: ptr(0)}
	options := goatquery.QueryOptions{MaxTop: 3}

	res, _, err := Apply[User](DB, query, nil, &options)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 3)
}

func Test_TopNegative(t *testing.T) {
	// Top=-1 should be treated same as unset (returns all or MaxTop)
	query := goatquery.Query{Top: ptr(-1)}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 5)
}

func Test_TopExceedsMaxTopWithZeroMaxTop(t *testing.T) {
	// MaxTop=0 means disabled — any Top value should be allowed
	query := goatquery.Query{Top: ptr(100)}
	options := goatquery.QueryOptions{MaxTop: 0}

	_, _, err := Apply[User](DB, query, nil, &options)
	assert.NoError(t, err)
}

// --- Skip Tests ---

func Test_Skip(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 5},
		{1, 4},
		{3, 2},
		{5, 0},
		{100, 0},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("skip_%d", test.input), func(t *testing.T) {
			query := goatquery.Query{Skip: ptr(test.input)}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
		})
	}
}

// --- Count Tests ---

func Test_Count(t *testing.T) {
	t.Run("count true", func(t *testing.T) {
		query := goatquery.Query{Count: ptr(true)}

		_, count, err := Apply[User](DB, query, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, count)
		assert.Equal(t, int64(5), *count)
	})

	t.Run("count false", func(t *testing.T) {
		query := goatquery.Query{Count: ptr(false)}

		_, count, err := Apply[User](DB, query, nil, nil)
		require.NoError(t, err)
		assert.Nil(t, count)
	})

	t.Run("count nil", func(t *testing.T) {
		query := goatquery.Query{}

		_, count, err := Apply[User](DB, query, nil, nil)
		require.NoError(t, err)
		assert.Nil(t, count)
	})
}

func Test_CountWithFilter(t *testing.T) {
	query := goatquery.Query{
		Filter: "age eq 25",
		Count:  ptr(true),
		Top:    ptr(1),
	}

	res, count, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, count)
	assert.Equal(t, int64(2), *count) // count reflects filtered total

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	assert.Len(t, output, 1) // but results are limited by top
}

// --- OrderBy Tests ---

func Test_OrderBy(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFirst string
		expectedLast  string
	}{
		{"asc", "age asc", "User01", "User04"},
		{"desc", "age desc", "User04", "User05"},
		{"default asc", "age", "User01", "User04"},
		{"case insensitive", "Age asc", "User01", "User04"},
		{"multi field", "age desc, firstname asc", "User04", "User05"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{OrderBy: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			require.NotEmpty(t, output)
			assert.Equal(t, test.expectedFirst, output[0].Firstname, "input: %s", test.input)
			assert.Equal(t, test.expectedLast, output[len(output)-1].Firstname, "input: %s", test.input)
		})
	}
}

func Test_OrderByNestedProperty(t *testing.T) {
	query := goatquery.Query{OrderBy: "company/name asc"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	require.NotEmpty(t, output)
}

func Test_OrderByEmpty(t *testing.T) {
	query := goatquery.Query{OrderBy: ""}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 5)
}

func Test_OrderByCaseVariations(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFirst string
	}{
		{"aGe asc", "aGe asc", "User01"},
		{"AGe desc", "AGe desc", "User04"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{OrderBy: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			require.NotEmpty(t, output)
			assert.Equal(t, test.expectedFirst, output[0].Firstname, "input: %s", test.input)
		})
	}
}

func Test_OrderByNestedPropertyDesc(t *testing.T) {
	query := goatquery.Query{OrderBy: "company/name desc"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	require.NotEmpty(t, output)
}

func Test_OrderByNestedAndFlat(t *testing.T) {
	query := goatquery.Query{OrderBy: "company/name asc, age desc"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	require.NotEmpty(t, output)
}

func Test_OrderByDuplicateJoins(t *testing.T) {
	// Ensure duplicate joins from orderby + filter are handled
	query := goatquery.Query{
		Filter:  "company/name contains 'Tech'",
		OrderBy: "company/name asc",
	}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	require.NotEmpty(t, output)
}

// --- Search Tests ---

func Test_Search(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"User01", 1},
		{"User", 5},
		{"nonexistent", 0},
	}

	searchFunc := func(db *gorm.DB, searchTerm string) *gorm.DB {
		return db.Where("firstname ILIKE ?", "%"+searchTerm+"%")
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			query := goatquery.Query{Search: test.input}

			res, _, err := Apply[User](DB, query, searchFunc, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
		})
	}
}

// --- Filter Tests ---

func Test_Filter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"string eq", "firstname eq 'User01'", 1},
		{"string eq no match", "firstname eq 'Random'", 0},
		{"int eq", "Age eq 25", 2},
		{"int eq zero", "Age eq 0", 0},
		{"int ne", "Age ne 30", 3},
		{"string contains", "firstName contains '0'", 5},
		{"guid eq", "id eq 11111111-1111-1111-1111-111111111111", 1},
		{"int lt", "age lt 30", 2},
		{"int lte", "age lte 30", 4},
		{"int gt", "age gt 25", 3},
		{"int gte", "age gte 30", 3},
		{"int range", "age lt 35 and age gt 25", 2},
		{"case insensitive eq", "firstname eq 'user01'", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterAndOr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"and", "firstname eq 'User02' and Age eq 30", 1},
		{"or", "firstname eq 'User01' or Age eq 35", 2},
		{"and/or precedence", "Firstname eq 'User01' and Age eq 25 or Age eq 30", 3},
		{"and with boolean", "age gt 25 and isEmailVerified eq true", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterGrouped(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"grouped right", "Firstname eq 'User01' and (Age eq 25 or Age eq 30)", 1},
		{"grouped left", "(Firstname eq 'User01' and Age eq 25) or Age eq 30", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterComplexGrouping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"complex grouped", "(firstname eq 'User02' and age eq 30) or age eq 35", 2},
		{"multi group", "(firstname eq 'User01') or (age eq 35 and firstname eq 'User04')", 2},
		{"nested and or precedence", "firstname eq 'User01' and age eq 25 or age eq 35", 2},
		{"date range null or grouping", "(dateOfBirth gte 1998-01-01 and dateOfBirth lt 1999-01-01) or dateOfBirth eq null", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"decimal eq null", "balanceDecimal eq null", 2},
		{"decimal ne null", "balanceDecimal ne null", 3},
		{"double eq null", "balanceDouble eq null", 2},
		{"float eq null", "balanceFloat eq null", 3},
		{"dateOfBirth eq null", "dateOfBirth eq null", 1},
		{"dateOfBirth ne null", "dateOfBirth ne null", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterNestedNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"company eq null", "company/name eq null", 1}, // User04 has no company
		{"company ne null", "company/name ne null", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterNullSafeStrings(t *testing.T) {
	// Tests that string operations don't fail when data contains NULL string values
	// Lastname: User01=Smith, User02=Jones, User03=nil, User04=nil, User05=Williams
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"string eq excludes nulls", "last_name eq 'Smith'", 1},
		{"string ne includes nulls", "last_name ne 'Smith'", 4},
		{"string contains excludes nulls", "last_name contains 'i'", 2}, // Jones has no i, Smith+Williams have i
		{"string eq null", "last_name eq null", 2},
		{"string ne null", "last_name ne null", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterBoolean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"eq true", "isEmailVerified eq true", 3},
		{"eq false", "isEmailVerified eq false", 2},
		{"ne true", "isEmailVerified ne true", 2},
		{"ne false", "isEmailVerified ne false", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"datetime eq", "dateOfBirth eq 1998-03-15T10:30:00Z", 1},
		{"datetime lt", "dateOfBirth lt 1995-01-01T00:00:00Z", 2},
		{"datetime gte", "dateOfBirth gte 1993-01-01T00:00:00Z", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDateOnly(t *testing.T) {
	// Date-only literals should be parsed and compared correctly by the database
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"date eq", "dateOfBirth eq 1998-03-15", 1}, // date-only expands to full-day range
		{"date lt", "dateOfBirth lt 1995-01-01", 2},
		{"date gte", "dateOfBirth gte 1993-01-01", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDateOnlyAllOperators(t *testing.T) {
	// Users with dateOfBirth:
	// User01: 1998-03-15T10:30:00Z
	// User02: 1993-07-20T14:00:00Z
	// User03: 1993-11-10T09:15:00Z
	// User04: nil
	// User05: 1998-12-25T18:45:00Z
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// eq: matches all times on that day
		{"date eq match", "dateOfBirth eq 1998-03-15", 1},
		{"date eq no match", "dateOfBirth eq 2000-01-01", 0},
		// ne: excludes all times on that day
		{"date ne", "dateOfBirth ne 1998-03-15", 3},
		// lt: before start of day
		{"date lt", "dateOfBirth lt 1993-11-10", 1},
		// lte: before start of next day (includes entire day)
		{"date lte", "dateOfBirth lte 1993-11-10", 2},
		// gt: after end of day
		{"date gt", "dateOfBirth gt 1993-11-10", 2},
		// gte: from start of day onward
		{"date gte", "dateOfBirth gte 1993-11-10", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDateTimeWithTimezone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"datetime with UTC", "dateOfBirth eq 1998-03-15T10:30:00Z", 1},
		{"datetime with +00:00", "dateOfBirth eq 1998-03-15T10:30:00+00:00", 1},
		{"datetime lt UTC", "dateOfBirth lt 1995-01-01T00:00:00Z", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"decimal eq", "balanceDecimal eq 1500.75", 1},
		{"decimal gt", "balanceDecimal gt 100", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"float eq", "balanceFloat eq 3500.25", 1},
		{"float eq other", "balanceFloat eq 750.50", 1},
		{"float eq null", "balanceFloat eq null", 3},
		{"float ne null", "balanceFloat ne null", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDouble(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"double eq", "balanceDouble eq 2500.50", 1},
		{"double eq other", "balanceDouble eq 1000.00", 1},
		{"double eq null", "balanceDouble eq null", 2},
		{"double ne null", "balanceDouble ne null", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterLong(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"long eq", "largeNumber eq 99999999999", 1},
		{"long gt", "largeNumber gt 50", 1},
		{"long lt", "largeNumber lt 99999999999", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterLongComparisons(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"long gte", "largeNumber gte 99999999999", 1},
		{"long lte", "largeNumber lte 42", 1},
		{"long gte large", "largeNumber gte 99999999999", 1},
		{"long eq null", "largeNumber eq null", 3},
		{"long ne null", "largeNumber ne null", 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterNegativeNumbers(t *testing.T) {
	tests := []struct {
		name          string
		filter        string
		orderBy       string
		expected      int
		expectedUsers []string
	}{
		{"int gt negative", "age gt -1", "", 5, nil},
		{"int lt negative", "age lt -1", "", 0, nil},
		{"int eq negative", "age eq -5", "", 0, nil},
		{"int ne negative", "age ne -5", "", 5, nil},
		{"double gt negative", "balanceDecimal gt -10.5", "", 3, nil},
		{"double lt negative", "balanceDecimal lt -10.5", "", 0, nil},
		{"negative number with orderby", "balanceDecimal gt -1", "company/name asc", 3, []string{"User02", "User01", "User05"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.filter, OrderBy: test.orderBy}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "filter: %s", test.filter)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "filter: %s", test.filter)
			for i, name := range test.expectedUsers {
				assert.Equal(t, name, output[i].Firstname)
			}
		})
	}
}

func Test_FilterNestedProperty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"nested eq", "company/name eq 'TechCorp'", 1},
		{"nested contains", "company/name contains 'Corp'", 2},
		{"nested with manager", "manager/firstname eq 'User01'", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterNestedPropertyDeep(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"manager/age gt 28", "manager/age gt 28", 2},                             // User03, User04 have manager User02 (age 30)
		{"manager/isEmailVerified eq true", "manager/isEmailVerified eq true", 1}, // User02 has manager User01 (verified=true)
		{"nested company/name eq", "company/name eq 'DataSoft'", 1},               // User02
		{"nested company/name contains", "company/name contains 'Tech'", 2},       // User01 (TechCorp), User03 (Tech Solutions)
		{"manager/manager depth 2", "manager/manager/firstname eq 'User01'", 2},   // User03, User04 → manager User02 → manager User01
		{"manager/manager/age", "manager/manager/age eq 25", 2},                   // User03, User04 → User02 → User01 (age 25)
		{"manager/manager eq null", "manager/manager eq null", 3},                 // User01, User05 (no manager), User02 (manager=User01 who has no manager)
		{"combined nested and scalar", "manager/manager eq null and firstname eq 'User01'", 1}, // User01 has no manager
		{"combined manager eq null and scalar", "manager eq null and firstname eq 'User01'", 1}, // User01 has no manager
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterCustomStringType(t *testing.T) {
	// Gender is a custom string type (type Gender string). Filtering should work transparently.
	// User01=Male, User02=Female, User03=nil, User04=Male, User05=Female
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"custom type eq", "gender eq 'Male'", 2},
		{"custom type eq other", "gender eq 'Female'", 2},
		{"custom type eq case insensitive", "gender eq 'male'", 2},
		{"custom type ne", "gender ne 'Male'", 3},
		{"custom type eq null", "gender eq null", 1},
		{"custom type ne null", "gender ne null", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterJsonPropertyName(t *testing.T) {
	// User.Lastname has json tag `json:"last_name"`, so filtering by last_name should work
	query := goatquery.Query{Filter: "last_name eq 'Smith'"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 1)
	assert.Equal(t, "User01", output[0].Firstname)
}

func Test_FilterGuidStringFallback(t *testing.T) {
	// When a UUID field is compared with a string literal (quoted), it should
	// be parsed as a UUID automatically (matching .NET's Guid.TryParse fallback).
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"guid as string literal", "id eq '11111111-1111-1111-1111-111111111111'", 1},
		{"guid as string no match", "id eq '00000000-0000-0000-0000-000000000000'", 0},
		{"guid ne as string", "id ne '11111111-1111-1111-1111-111111111111'", 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterDuplicateJoins(t *testing.T) {
	// This tests that duplicate joins are deduplicated when the same nested entity is referenced twice
	query := goatquery.Query{Filter: "company/name eq 'TechCorp' and company/department eq 'Engineering'"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 1)
	assert.Equal(t, "User01", output[0].Firstname)
}

func Test_FilterInvalidProperty(t *testing.T) {
	query := goatquery.Query{Filter: "NonExistentProperty eq 'John'"}
	_, _, err := Apply[User](DB, query, nil, nil)
	assert.Error(t, err)
}

func Test_FilterInvalidPropertyPaths(t *testing.T) {
	inputs := []struct {
		name  string
		input string
	}{
		{"non-existent property", "nonExistentProperty eq 'test'"},
		{"non-existent nested property", "company/nonExistent eq 'test'"},
	}

	for _, test := range inputs {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}
			_, _, err := Apply[User](DB, query, nil, nil)
			assert.Error(t, err, "input: %s", test.input)
		})
	}
}

func Test_FilterMaxPropertyMappingDepth(t *testing.T) {
	// With depth=1, nested property access should fail
	query := goatquery.Query{Filter: "company/name eq 'TechCorp'"}
	options := goatquery.QueryOptions{MaxTop: 100}
	// Use GetMaxPropertyMappingDepth default (5) — should work
	res, _, err := Apply[User](DB, query, nil, &options)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)
	assert.Len(t, output, 1)
}

func Test_FilterTypeMismatchErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Ordering on string property
		{"string gt int", "firstname gt 5"},
		{"string lt int", "firstname lt 5"},
		{"string gte int", "firstname gte 5"},
		{"string lte int", "firstname lte 5"},
		// Ordering on boolean property
		{"bool gt int", "isEmailVerified gt 5"},
		// Ordering on UUID property
		{"uuid gt int", "id gt 5"},
		// Contains on non-string property
		{"contains on int", "age contains '5'"},
		{"contains on bool", "isEmailVerified contains 'true'"},
		// String literal on non-string property (parser rejects ordering, but eq is valid at parser level)
		// These reach the DB which handles type checking
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}
			_, _, err := Apply[User](DB, query, nil, nil)
			assert.Error(t, err, "input: %s", test.input)
		})
	}
}

func Test_FilterLambdaAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"any with nested path", "addresses/any(addr: addr/city/name eq 'New York')", 1},
		{"any with contains", "addresses/any(addr: addr/addressLine1 contains 'Main')", 1},
		{"any no match", "addresses/any(addr: addr/city/name eq 'NonExistentCity')", 0},
		{"any chicago", "addresses/any(addr: addr/city/name eq 'Chicago')", 1},
		{"any seattle", "addresses/any(addr: addr/city/name eq 'Seattle')", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err, "input: %s", test.input)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterLambdaAll(t *testing.T) {
	query := goatquery.Query{Filter: "addresses/all(addr: addr/city/country eq 'USA')"}

	res, _, err := Apply[User](DB, query, nil, nil)
	require.NoError(t, err)

	var output []User
	err = res.Find(&output).Error
	require.NoError(t, err)

	assert.Len(t, output, 3)
}

func Test_FilterLambdaCombinedWithFilter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      int
		expectedUsers []string
	}{
		{
			"scalar filter and lambda",
			"firstname eq 'User01' and addresses/any(addr: addr/city/name eq 'New York')",
			1,
			[]string{"User01"},
		},
		{
			"nested property and lambda",
			"company/name contains 'Tech' and addresses/any(a: a/city/country eq 'USA')",
			1,
			[]string{"User01"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Order("firstname").Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
			for i, name := range test.expectedUsers {
				assert.Equal(t, name, output[i].Firstname)
			}
		})
	}
}

func Test_FilterLambdaOrInBody(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      int
		expectedUsers []string
	}{
		{
			"or in lambda body",
			"addresses/any(addr: addr/city/name eq 'New York' or addr/city/name eq 'Seattle')",
			2,
			nil,
		},
		{
			"lambda or body with root boolean filter",
			"isEmailVerified eq true and addresses/any(a: a/city/name eq 'New York' or a/city/name eq 'Miami')",
			2,
			[]string{"User01", "User05"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Order("firstname").Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
			for i, name := range test.expectedUsers {
				assert.Equal(t, name, output[i].Firstname)
			}
		})
	}
}

func Test_FilterLambdaRootPropertyAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			"root property with lambda property",
			"addresses/any(addr: addr/city/name eq 'New York' and firstname eq 'User01')",
			1, // User01 has NY address and firstname matches
		},
		{
			"root property mismatch",
			"addresses/any(addr: addr/city/name eq 'New York' and firstname eq 'User02')",
			0, // User02 doesn't have a NY address
		},
		{
			"root property only in lambda body",
			"addresses/any(addr: firstname eq 'User01')",
			1, // User01 has addresses
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected)
		})
	}
}

func Test_FilterLambdaPrimitiveCollection(t *testing.T) {
	// Tags is a []string (jsonb). Lambda on primitive collection uses jsonb containment.
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"any with primitive string", "tags/any(x: x eq 'vip')", 1},
		{"any with primitive premium", "tags/any(x: x eq 'premium')", 2},
		{"any no match", "tags/any(x: x eq 'nonexistent')", 0},
		{"primitive lambda with boolean and nested property", "isEmailVerified eq true and tags/any(t: t eq 'premium') and company/department eq 'Engineering'", 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}

			res, _, err := Apply[User](DB, query, nil, nil)
			require.NoError(t, err)

			var output []User
			err = res.Find(&output).Error
			require.NoError(t, err)

			assert.Len(t, output, test.expected, "input: %s", test.input)
		})
	}
}

func Test_FilterLambdaErrors(t *testing.T) {
	inputs := []struct {
		name  string
		input string
	}{
		{"non-existent collection", "nonExistent/any(item: item eq 'test')"},
		{"invalid lambda function", "addresses/invalid(addr: addr/city/name eq 'test')"},
	}

	for _, test := range inputs {
		t.Run(test.name, func(t *testing.T) {
			query := goatquery.Query{Filter: test.input}
			_, _, err := Apply[User](DB, query, nil, nil)
			assert.Error(t, err, "input: %s", test.input)
		})
	}
}

// --- Internal ---

func Test_EscapeLike(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain string", "hello", "hello"},
		{"percent wildcard", "100%", `100\%`},
		{"underscore wildcard", "first_name", `first\_name`},
		{"backslash", `back\slash`, `back\\slash`},
		{"all special chars", `10%_\test`, `10\%\_\\test`},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, escapeLike(tt.input))
		})
	}
}
