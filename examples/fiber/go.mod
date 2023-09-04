module github.com/goatquery/goatquery-go/examples/fiber

go 1.20

require (
	github.com/brianvoe/gofakeit/v6 v6.23.2
	github.com/gofiber/fiber/v2 v2.49.1
	github.com/google/uuid v1.3.1
	gorm.io/driver/sqlite v1.5.3
	gorm.io/gorm v1.25.4
)

replace goatquery => ../..

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/savsgio/dictpool v0.0.0-20221023140959-7bf2e61cea94 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.49.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	goatquery v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.12.0 // indirect
)
