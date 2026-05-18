# GoatQuery

A Go library for parsing query parameters and applying them to GORM queries. Enables database-level filtering, sorting, and pagination from HTTP query strings.

## Installation

```bash
go get github.com/goatquery/goatquery-go

# For GORM integration
go get github.com/goatquery/goatquery-go/module/gorm
```

## Quick Start

```go
import (
    goatquery "github.com/goatquery/goatquery-go"
    goatgorm "github.com/goatquery/goatquery-go/module/gorm"
)

// Basic filtering
query := goatquery.Query{Filter: "age gt 18 and isActive eq true"}
db, count, err := goatgorm.Apply[User](db, query, nil, nil)

// Lambda expressions for collection filtering
query := goatquery.Query{Filter: "addresses/any(x: x/city eq 'London')"}
db, _, err := goatgorm.Apply[User](db, query, nil, nil)

// Complex nested filtering
query := goatquery.Query{
    Filter: "isActive eq true and orders/any(o: o/items/any(i: i/price gt 1000))",
}
db, _, err := goatgorm.Apply[User](db, query, nil, nil)

// Pagination with count
top := 10
skip := 20
count := true
query := goatquery.Query{
    Filter:  "age gt 18",
    OrderBy: "lastName asc",
    Top:     &top,
    Skip:    &skip,
    Count:   &count,
}
db, totalCount, err := goatgorm.Apply[User](db, query, nil, nil)
```

## Supported Syntax

```
GET /api/users?filter=age gt 18 and isActive eq true
GET /api/users?filter=addresses/any(x: x/city eq 'London')
GET /api/users?filter=tags/any(x: x eq 'premium')
GET /api/users?orderby=lastName asc, firstName desc
GET /api/users?orderby=company/name asc
GET /api/users?top=10&skip=20&count=true
GET /api/users?search=john
```

## Filtering

### Operators

- **Comparison**: `eq`, `ne`, `gt`, `gte`, `lt`, `lte`
- **Logical**: `and`, `or`
- **String**: `contains`

### Data Types

| Type | Example |
|------|---------|
| String | `'value'`, `'it\'s escaped'`, `'back\\slash'` |
| Integer | `42`, `99999999999`, `-7` |
| Double | `3.14`, `1.0`, `-2.5` |
| Boolean | `true`, `false` |
| DateTime | `2023-12-25T10:30:00Z` |
| DateTimeOffset | `2023-12-25T10:30:00+05:00` |
| Date | `2023-12-25` (expanded to full-day range) |
| GUID | `123e4567-e89b-12d3-a456-426614174000` |
| Null | `null` |

String literals support backslash escaping: `\'` for a literal single quote, `\\` for a literal backslash.

### Property Path Navigation

Access nested properties using forward slash (`/`) syntax:

```
filter=company/name eq 'TechCorp'
filter=profile/address/city eq 'London'
filter=user/addresses/any(x: x/country/name eq 'UK')
```

### Lambda Expressions

Filter collections using `any()` and `all()`:

```
// any() - true if at least one element matches
addresses/any(x: x/city eq 'London')

// all() - true if all elements match (requires non-empty collection)
addresses/all(x: x/isVerified eq true)

// Nested lambda expressions
orders/any(o: o/items/any(i: i/price gt 100))

// Primitive array filtering
tags/any(x: x eq 'premium')
scores/any(x: x gt 80)

// Root property access inside lambda body
orders/any(o: o/total gt 100 and status eq 'Active')
```

### Date-Only Range Expansion

Date-only literals (e.g., `2023-12-25`) are automatically expanded to full-day range comparisons against `time.Time` columns:

| Operator | Expansion |
|----------|-----------|
| `eq` | `>= 2023-12-25T00:00:00` AND `< 2023-12-26T00:00:00` |
| `ne` | `< 2023-12-25T00:00:00` OR `>= 2023-12-26T00:00:00` |
| `lt` | `< 2023-12-25T00:00:00` |
| `lte` | `< 2023-12-26T00:00:00` |
| `gt` | `>= 2023-12-26T00:00:00` |
| `gte` | `>= 2023-12-25T00:00:00` |

### GUID String Fallback

When a GUID literal (e.g., `123e4567-e89b-12d3-a456-426614174000`) is compared against a `string` column, it is automatically treated as a string value instead of raising a type error.

Lambda expressions are evaluated as `EXISTS` / `NOT EXISTS` subqueries against the related table. Primitive collections (jsonb arrays) use Postgres `@>` containment. GORM's schema parser resolves relationship metadata (foreign keys, join tables) automatically.

### Null Safety

String comparisons (`eq`, `ne`, `contains`) are null-safe. When a string column is `NULL` in the database:

- `eq` and `contains` **exclude** null rows (null does not equal any value)
- `ne` **includes** null rows (null is "not equal" to any value)

### Examples

```
age gt 18
firstName eq 'John' and isActive ne false
name contains 'smith'
createdAt gte 2023-01-01T00:00:00Z
createdAt eq 2023-06-15
balance gt -100
addresses/any(x: x/city eq 'London' and x/isActive eq true)
age gt 25 and orders/any(o: o/total gt 100)
tags/any(x: x eq 'premium')
orders/any(o: o/total gt 50 and status eq 'Active')
```

## Ordering

Sort by one or more properties, including nested properties:

```
GET /api/users?orderby=lastName asc
GET /api/users?orderby=lastName asc, firstName desc
GET /api/users?orderby=company/name asc
GET /api/users?orderby=company/name asc, age desc
```

Default direction is ascending when omitted.

## Property Mapping

Properties in query strings are resolved using struct field json tags, matched **case-insensitively**:

```go
type User struct {
    FirstName string `json:"first_name"`
    Age       int    `json:"age"`
}
```

```
filter=first_name eq 'John' and age gt 18
```

If no json tag is present, the Go struct field name is used.

## Configuration

Configure query behaviour through `QueryOptions`:

```go
options := &goatquery.QueryOptions{
    MaxTop:                  100, // Maximum allowed top value (0 = no limit)
    MaxPropertyMappingDepth: 5,   // Max depth for nested property resolution (default: 5)
}

db, count, err := goatgorm.Apply[User](db, query, nil, options)
```

## Search

Implement custom search logic by passing a `SearchFunc`:

```go
searchFunc := func(db *gorm.DB, searchTerm string) *gorm.DB {
    term := "%" + searchTerm + "%"
    return db.Where("first_name ILIKE ? OR last_name ILIKE ?", term, term)
}

db, count, err := goatgorm.Apply[User](db, query, searchFunc, nil)
```

## Error Handling

`Apply` returns a standard Go `error`:

```go
db, count, err := goatgorm.Apply[User](db, query, nil, nil)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

var users []User
db.Find(&users)
```

## HTTP Handler Example

```go
func GetUsers(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()

    query := goatquery.Query{
        Filter:  q.Get("filter"),
        OrderBy: q.Get("orderby"),
        Search:  q.Get("search"),
    }

    if v := q.Get("top"); v != "" {
        top, _ := strconv.Atoi(v)
        query.Top = &top
    }
    if v := q.Get("skip"); v != "" {
        skip, _ := strconv.Atoi(v)
        query.Skip = &skip
    }
    if v := q.Get("count"); v != "" {
        count, _ := strconv.ParseBool(v)
        query.Count = &count
    }

    db, totalCount, err := goatgorm.Apply[User](db, query, nil, nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var users []User
    db.Find(&users)

    resp := goatquery.PagedResponse[User]{Value: users, Count: totalCount}
    json.NewEncoder(w).Encode(resp)
}
```

## Development

### Test

```bash
# Core module (lexer, parser)
go test ./... -v

# GORM module (requires Docker for Testcontainers + Postgres)
cd module/gorm && go test ./... -v
```

**Requires**: Go 1.26+