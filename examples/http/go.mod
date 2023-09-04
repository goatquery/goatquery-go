module github.com/goatquery/goatquery-go/examples/http

go 1.20

require (
	github.com/brianvoe/gofakeit/v6 v6.23.2
	github.com/google/uuid v1.3.1
	goatquery v0.0.0-00010101000000-000000000000
	gorm.io/driver/sqlite v1.5.3
	gorm.io/gorm v1.25.4
)

replace goatquery => ../..

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.47.0 // indirect
)
