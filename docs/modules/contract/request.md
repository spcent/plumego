# Request Parsing

> **Package**: `github.com/spcent/plumego/contract`

Request parsing methods for binding and validating incoming data.

---

## Table of Contents

- [Request Binding](#request-binding)
- [JSON Binding](#json-binding)
- [XML Binding](#xml-binding)
- [Form Binding](#form-binding)
- [Query Binding](#query-binding)
- [File Upload](#file-upload)
- [Validation](#validation)

---

## Request Binding

### Auto-Detect Content Type

```go
type CreateUserRequest struct {
    Name  string `json:"name" xml:"name" form:"name"`
    Email string `json:"email" xml:"email" form:"email"`
}

app.PostCtx("/users", func(ctx *plumego.Context) {
    var req CreateUserRequest

    if err := ctx.Bind(&req); err != nil {
        ctx.Error(http.StatusBadRequest, err.Error())
        return
    }

    // Use req...
})
```

`Bind()` automatically detects content type:
- `application/json` → JSON binding
- `application/xml` → XML binding
- `application/x-www-form-urlencoded` → Form binding
- `multipart/form-data` → Form binding

---

## JSON Binding

### Basic JSON

```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

app.PostCtx("/login", func(ctx *plumego.Context) {
    var req LoginRequest

    if err := ctx.BindJSON(&req); err != nil {
        ctx.Error(400, "Invalid JSON")
        return
    }

    // Authenticate...
})
```

### Nested JSON

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Address Address `json:"address"`
}

// Request body:
// {
//   "name": "Alice",
//   "email": "alice@example.com",
//   "address": {
//     "street": "123 Main St",
//     "city": "NYC",
//     "country": "USA"
//   }
// }
```

---

## XML Binding

```go
type XMLRequest struct {
    XMLName xml.Name `xml:"request"`
    Name    string   `xml:"name"`
    Email   string   `xml:"email"`
}

app.PostCtx("/data", func(ctx *plumego.Context) {
    var req XMLRequest

    if err := ctx.BindXML(&req); err != nil {
        ctx.Error(400, "Invalid XML")
        return
    }

    // Use req...
})
```

---

## Form Binding

### URL-Encoded Form

```go
type ContactForm struct {
    Name    string `form:"name"`
    Email   string `form:"email"`
    Message string `form:"message"`
}

app.PostCtx("/contact", func(ctx *plumego.Context) {
    var form ContactForm

    if err := ctx.BindForm(&form); err != nil {
        ctx.Error(400, "Invalid form data")
        return
    }

    // Process form...
})
```

### Multipart Form

```go
app.PostCtx("/profile", func(ctx *plumego.Context) {
    // Parse multipart form
    if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
        ctx.Error(400, "Failed to parse form")
        return
    }

    name := ctx.Request.FormValue("name")
    email := ctx.Request.FormValue("email")

    // Handle file upload (see File Upload section)
})
```

---

## Query Binding

Bind query parameters to struct:

```go
type SearchParams struct {
    Q      string `form:"q"`
    Page   int    `form:"page"`
    Limit  int    `form:"limit"`
    SortBy string `form:"sort_by"`
}

app.GetCtx("/search", func(ctx *plumego.Context) {
    var params SearchParams

    if err := ctx.BindQuery(&params); err != nil {
        ctx.Error(400, "Invalid query parameters")
        return
    }

    // Use params...
    results := search(params.Q, params.Page, params.Limit)
    ctx.JSON(200, results)
})
```

---

## File Upload

### Single File

```go
app.PostCtx("/upload", func(ctx *plumego.Context) {
    file, header, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.Error(400, "No file uploaded")
        return
    }
    defer file.Close()

    // Validate file size
    if header.Size > 10<<20 {
        ctx.Error(400, "File too large (max 10MB)")
        return
    }

    // Save file
    filename := filepath.Base(header.Filename)
    dst, err := os.Create("./uploads/" + filename)
    if err != nil {
        ctx.Error(500, "Failed to save file")
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        ctx.Error(500, "Failed to save file")
        return
    }

    ctx.JSON(200, map[string]interface{}{
        "filename": filename,
        "size":     header.Size,
    })
})
```

### Multiple Files

```go
app.PostCtx("/upload-multiple", func(ctx *plumego.Context) {
    if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
        ctx.Error(400, "Failed to parse form")
        return
    }

    files := ctx.Request.MultipartForm.File["files"]
    uploaded := []string{}

    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            continue
        }
        defer file.Close()

        // Save file...
        filename := filepath.Base(fileHeader.Filename)
        dst, _ := os.Create("./uploads/" + filename)
        defer dst.Close()
        io.Copy(dst, file)

        uploaded = append(uploaded, filename)
    }

    ctx.JSON(200, map[string]interface{}{
        "uploaded": uploaded,
    })
})
```

---

## Validation

### Struct Tags

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=120"`
}

app.PostCtx("/users", func(ctx *plumego.Context) {
    var req CreateUserRequest

    if err := ctx.Bind(&req); err != nil {
        ctx.Error(400, err.Error())
        return
    }

    // Validation happens automatically if validator is configured
})
```

### Manual Validation

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    var req CreateUserRequest

    if err := ctx.Bind(&req); err != nil {
        ctx.Error(400, err.Error())
        return
    }

    // Manual validation
    if req.Name == "" {
        ctx.Error(400, "Name is required")
        return
    }

    if !isValidEmail(req.Email) {
        ctx.Error(400, "Invalid email format")
        return
    }

    // Create user...
})
```

---

## Complete Example

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
}

func main() {
    app := core.New()

    // Create user with validation
    app.PostCtx("/users", func(ctx *plumego.Context) {
        var req CreateUserRequest

        // Bind and validate
        if err := ctx.Bind(&req); err != nil {
            ctx.Error(http.StatusBadRequest, err.Error())
            return
        }

        // Additional business validation
        if userExists(req.Email) {
            ctx.Error(http.StatusConflict, "Email already registered")
            return
        }

        // Create user
        user := createUser(req)
        ctx.Status(http.StatusCreated).JSON(user)
    })

    // Upload profile picture
    app.PostCtx("/users/:id/avatar", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        file, header, err := ctx.Request.FormFile("avatar")
        if err != nil {
            ctx.Error(400, "No file uploaded")
            return
        }
        defer file.Close()

        // Validate file type
        contentType := header.Header.Get("Content-Type")
        if !strings.HasPrefix(contentType, "image/") {
            ctx.Error(400, "Only images allowed")
            return
        }

        // Save avatar
        filename := id + filepath.Ext(header.Filename)
        dst, _ := os.Create("./avatars/" + filename)
        defer dst.Close()
        io.Copy(dst, file)

        ctx.JSON(200, map[string]string{
            "avatar": filename,
        })
    })

    app.Boot()
}
```

---

**Next**: [Protocol Adapters](protocol/)
