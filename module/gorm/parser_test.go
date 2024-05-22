package gorm

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Test_ParsingOrderBy(t *testing.T) {
	type Test struct {
		name string
	}

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	assert.NoError(t, err)

	expectedSql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&Test{}).Order("name").Find(&[]Test{})
	})

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		res := EvaluateOrderBy(tx, "name")
		return res.Find(&[]Test{})
	})

	assert.Equal(t, expectedSql, sql)
}
