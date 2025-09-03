## Go URL Shortener

Simple URL shortener service with in-memory storage.

### Run

```bash
go run .
```

Environment variables:
- `ADDR` (default `:8080`)
- `DOMAIN` (default `http://localhost:8080`)

### Step-by-step (junior-friendly)

1) Start server
```bash
go run .
```
2) Create short link (Postman)
- POST `http://localhost:8080/api/shorten`
- Body (JSON): `{ "url": "https://example.com" }`

3) Follow redirect
- GET `http://localhost:8080/{code}` from the response

Structure:
- `internal/storage`: in-memory map (thread-safe)
- `internal/service`: business rules (validate URL, generate code, save)
- `internal/api`: HTTP handlers (JSON in/out, redirect)
- `main.go`: wires everything and starts server

### API

- POST `/api/shorten`
  - Request JSON: `{ "url": "https://example.com", "custom_alias": "docs" }`
  - Responses:
    - 201 `{ "code": "abc123", "short": "http://localhost:8080/abc123" }`
    - 400 invalid json or invalid url
    - 409 alias taken

- GET `/{code}`
  - 302 Redirect to original URL
  - 404 if not found

### Test

```bash
go test ./...
```


