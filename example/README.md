# GoatQuery Go Example

A sample net/http API demonstrating GoatQuery with GORM and Postgres.

## Prerequisites

- Go 1.26+
- Docker (for Postgres via Testcontainers)

## Running

```bash
cd example
go run .
```

The server starts a Postgres container automatically and seeds 1,000 users with orders, products, addresses, and companies.

## Endpoints

| Endpoint | Description |
|---|---|
| `GET /users` | Users with company, addresses, manager, orders, tags. Search across name. MaxTop: 50 |
| `GET /products` | Products. No search. MaxTop: 100 |
| `GET /orders` | Orders with items/products. Search by order number. MaxTop: 50 |

## Query Parameters

| Parameter | Description |
|---|---|
| `filter` | Filter expression |
| `orderby` | Order by expression |
| `top` | Number of results to return |
| `skip` | Number of results to skip |
| `count` | Set to `true` to include total count |
| `search` | Free-text search (endpoint-specific) |

## Sample Queries

### Basic Filtering

```bash
# Exact match
curl 'localhost:8080/users?filter=first_name eq '\''Alice'\'''

# Not equal
curl 'localhost:8080/users?filter=age ne 25'

# Comparison operators
curl 'localhost:8080/users?filter=age gt 30'
curl 'localhost:8080/users?filter=age gte 40'
curl 'localhost:8080/users?filter=age lt 25'
curl 'localhost:8080/users?filter=age lte 20'

# String contains
curl 'localhost:8080/users?filter=lastname contains '\''son'\'''

# Boolean
curl 'localhost:8080/users?filter=isEmailVerified eq true'

# Null checks
curl 'localhost:8080/users?filter=nullableInt eq null'
curl 'localhost:8080/users?filter=nullableInt ne null'
```

### Logical Operators

```bash
# AND
curl 'localhost:8080/users?filter=age gt 30 and isEmailVerified eq true'

# OR
curl 'localhost:8080/users?filter=first_name eq '\''Alice'\'' or first_name eq '\''Bob'\'''

# Grouped expressions
curl 'localhost:8080/users?filter=(age gt 30 and age lt 50) or isEmailVerified eq true'
```

### Nested Property Navigation

```bash
# Navigate relationships with /
curl 'localhost:8080/users?filter=company/name eq '\''Acme Corp'\'''
curl 'localhost:8080/users?filter=company/department contains '\''Engineer'\'''
```

### Lambda Expressions (Collections)

```bash
# any() — at least one element matches (primitive collection)
curl 'localhost:8080/users?filter=tags/any(x: x eq '\''vip'\'')'

# any() — relationship navigation
curl 'localhost:8080/users?filter=addresses/any(a: a/city/name eq '\''New York'\'')'

# all() — every element matches
curl 'localhost:8080/users?filter=addresses/all(a: a/addressLine1 ne '\'''\'')'

# Orders with specific status
curl 'localhost:8080/users?filter=orders/any(o: o/status eq '\''Delivered'\'')'

# Orders above a total amount
curl 'localhost:8080/users?filter=orders/any(o: o/total gt 500)'
```

### Nested Lambdas

```bash
# Users with any order containing an Electronics product
curl 'localhost:8080/users?filter=orders/any(o: o/items/any(i: i/product/category eq '\''Electronics'\''))'

# Users where all orders have at least one item
curl 'localhost:8080/users?filter=orders/all(o: o/items/any(i: i/quantity gt 0))'
```

### Ordering

```bash
curl 'localhost:8080/users?orderby=age asc'
curl 'localhost:8080/users?orderby=age desc'
curl 'localhost:8080/users?orderby=lastname asc, age desc'
curl 'localhost:8080/users?orderby=company/name asc'
```

### Pagination

```bash
curl 'localhost:8080/users?top=5'
curl 'localhost:8080/users?top=5&skip=10'
curl 'localhost:8080/users?top=5&skip=10&count=true'
```

### Search

```bash
curl 'localhost:8080/users?search=alice'
curl 'localhost:8080/orders?search=ORD-0042'
```

### Combined Queries

```bash
# Filter + order + paginate + count
curl 'localhost:8080/users?filter=age gt 25&orderby=age desc&top=5&count=true'

# Search + filter + pagination
curl 'localhost:8080/users?search=smith&filter=isEmailVerified eq true&top=10&count=true'

# Products by category, ordered by price
curl 'localhost:8080/products?filter=category eq '\''Electronics'\''&orderby=price desc&top=10&count=true'
```
