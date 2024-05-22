package gorm

import "gorm.io/gorm"

func EvaluateOrderBy(db *gorm.DB, orderby string) *gorm.DB {
	if orderby != "" {
		db = db.Order(orderby)
	}

	return db
}
