module github.com/goatquery/goatquery-go/examples/http

go 1.21

require (
	github.com/brianvoe/gofakeit/v6 v6.23.2
	github.com/google/uuid v1.5.0
	goatquery v0.0.0-00010101000000-000000000000
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

replace goatquery => ../..

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
)
