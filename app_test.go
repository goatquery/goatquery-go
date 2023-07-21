package goatquery

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Base struct {
	Id uint `json:"id"`
}

type User struct {
	Base

	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}

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
}

func Test_EmptyQuery(t *testing.T) {
	query := Query{}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Top

func Test_QueryWithTop(t *testing.T) {
	query := Query{Top: 3}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Limit(query.Top).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithTopGreaterThanMaxTop(t *testing.T) {
	maxTop := 2
	query := Query{Top: 3}

	_, _, err := Apply(DB.Model(&User{}), query, &maxTop, nil)

	assert.NotNil(t, err)
}

func Test_QueryWithNilTopUsesMaxTop(t *testing.T) {
	maxTop := 2
	query := Query{}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, &maxTop, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Limit(maxTop).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Skip

func Test_QueryWithSkip(t *testing.T) {
	query := Query{Skip: 3}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Offset(query.Skip).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Count

func Test_QueryWithCount(t *testing.T) {
	query := Query{Count: true}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	var count int64
	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Count(&count).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Order by

func Test_QueryWithOrderby(t *testing.T) {
	query := Query{OrderBy: "firstname"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Order("firstname").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithOrderbyAsc(t *testing.T) {
	query := Query{OrderBy: "firstname asc"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Order("firstname asc").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithOrderbyDesc(t *testing.T) {
	query := Query{OrderBy: "firstname desc"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Order("firstname desc").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithOrderbyMultiple(t *testing.T) {
	query := Query{OrderBy: "firstname asc, lastname desc"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Order("firstname asc, lastname desc").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Select

func Test_QueryWithSelect(t *testing.T) {
	query := Query{Select: "firstname, lastname"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Select("firstname, lastname").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithSelectInvalidColumn(t *testing.T) {
	query := Query{Select: "firstname, invalid-col"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Select("firstname, invalid-col").Find(&[]User{})
	})

	err := DB.Raw(expectedSql).Find(&[]User{}).Error

	assert.Error(t, err)
	assert.Equal(t, expectedSql, sql)
}

// Search

func Test_QueryWithSearch(t *testing.T) {
	query := Query{Search: "Goat"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, func(db *gorm.DB, searchTerm string) *gorm.DB {
			t := fmt.Sprintf("%%%s%%", searchTerm)

			return db.Where("firstname like ? or lastname like ?", t, t)
		})
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname like \"%Goat%\" or lastname like \"%Goat%\"").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithSearchTermSpace(t *testing.T) {
	query := Query{Search: "Goat Query"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, func(db *gorm.DB, searchTerm string) *gorm.DB {
			t := fmt.Sprintf("%%%s%%", searchTerm)

			return db.Where("firstname like ? or lastname like ?", t, t)
		})
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname like \"%Goat Query%\" or lastname like \"%Goat Query%\"").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithSearchNilFunc(t *testing.T) {
	query := Query{Search: "Goat"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

// Filter

func Test_QueryWithFilterEquals(t *testing.T) {
	query := Query{Filter: "firstname eq 'goat'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname = 'goat'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterNotEquals(t *testing.T) {
	query := Query{Filter: "firstname ne 'goat'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname <> 'goat'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterEqualsAndEquals(t *testing.T) {
	query := Query{Filter: "firstname eq 'goat' and lastname eq 'query'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname = 'goat' and lastname = 'query'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterEqualsAndNotEquals(t *testing.T) {
	query := Query{Filter: "firstname eq 'goat' and lastname ne 'query'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname = 'goat' and lastname <> 'query'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterContains(t *testing.T) {
	query := Query{Filter: "firstname contains 'goat'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname like '%goat%'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterContainsAndEquals(t *testing.T) {
	query := Query{Filter: "firstname contains 'goat' and lastname eq 'query'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname like '%goat%' and lastname = 'query'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterContainsOrEquals(t *testing.T) {
	query := Query{Filter: "firstname contains 'goat' or lastname eq 'query'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname like '%goat%' or lastname = 'query'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterEqualsWithConjunction(t *testing.T) {
	query := Query{Filter: "firstname eq 'goatand'"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname = 'goatand'").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}

func Test_QueryWithFilterEqualsWithConjunctionAndSpaces(t *testing.T) {
	query := Query{Filter: "firstname eq ' and ' or lastname eq ' and or '"}

	sql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res, _, _ := Apply(tx.Model(&User{}), query, nil, nil)
		return res.Find(&[]User{})
	})

	expectedSql := DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("firstname = ' and ' or lastname = ' and or '").Find(&[]User{})
	})

	assert.Equal(t, expectedSql, sql)
}
