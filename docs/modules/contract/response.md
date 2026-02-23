# Response Helpers

> **Package**: `github.com/spcent/plumego/contract`

Response helpers provide convenient methods for sending various types of HTTP responses.

---

## Table of Contents

- [JSON Responses](#json-responses)
- [XML Responses](#xml-responses)
- [Text Responses](#text-responses)
- [HTML Responses](#html-responses)
- [File Responses](#file-responses)
- [Stream Responses](#stream-responses)
- [Redirects](#redirects)
- [Status Codes](#status-codes)

---

## JSON Responses

### Basic JSON

```go
app.GetCtx("/users", func(ctx *plumego.Context) {
    users := []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    }

    ctx.JSON(http.StatusOK, users)
})
```

### With Status Chaining

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    var user User
    ctx.Bind(&user)

    // Create user...
    ctx.Status(http.StatusCreated).JSON(user)
})
```

### JSON Pretty Print

```go
ctx.JSONPretty(http.StatusOK, data, "  ")
```

---

## XML Responses

```go
type Response struct {
    XMLName xml.Name `xml:"response"`
    Status  string   `xml:"status"`
    Data    []User   `xml:"users>user"`
}

app.GetCtx("/users.xml", func(ctx *plumego.Context) {
    resp := Response{
        Status: "ok",
        Data:   users,
    }

    ctx.XML(http.StatusOK, resp)
})
```

---

## Text Responses

### Plain String

```go
ctx.String(http.StatusOK, "Hello World")
```

### Formatted String

```go
ctx.String(http.StatusOK, "User ID: %s, Name: %s", user.ID, user.Name)
```

---

## HTML Responses

```go
app.GetCtx("/", func(ctx *plumego.Context) {
    html := `
    <!DOCTYPE html>
    <html>
    <head><title>Home</title></head>
    <body><h1>Welcome</h1></body>
    </html>
    `
    ctx.HTML(http.StatusOK, html)
})
```

---

## File Responses

### Serve File

```go
app.GetCtx("/download/:filename", func(ctx *plumego.Context) {
    filename := ctx.Param("filename")
    path := "./files/" + filename

    ctx.File(path)
})
```

### Attachment (Force Download)

```go
app.GetCtx("/reports/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    path := "./reports/" + id + ".pdf"

    ctx.Attachment(path, "report.pdf")
})
```

### Inline Content

```go
ctx.Inline(path, "document.pdf")
```

---

## Stream Responses

### SSE (Server-Sent Events)

```go
app.GetCtx("/events", func(ctx *plumego.Context) {
    ctx.SetHeader("Content-Type", "text/event-stream")
    ctx.SetHeader("Cache-Control", "no-cache")
    ctx.SetHeader("Connection", "keep-alive")

    flusher, _ := ctx.Writer.(http.Flusher)

    for i := 0; i < 10; i++ {
        fmt.Fprintf(ctx.Writer, "data: Event %d\n\n", i)
        flusher.Flush()
        time.Sleep(1 * time.Second)
    }
})
```

### Custom Stream

```go
ctx.Stream(http.StatusOK, "text/event-stream", func(w io.Writer) error {
    for event := range eventChannel {
        if _, err := fmt.Fprintf(w, "data: %s\n\n", event); err != nil {
            return err
        }
        if f, ok := w.(http.Flusher); ok {
            f.Flush()
        }
    }
    return nil
})
```

---

## Redirects

### Temporary Redirect (302)

```go
ctx.Redirect(http.StatusFound, "/login")
```

### Permanent Redirect (301)

```go
ctx.Redirect(http.StatusMovedPermanently, "/new-location")
```

---

## Status Codes

### Set Status

```go
ctx.Status(http.StatusCreated)
ctx.JSON(user)
```

### No Content

```go
ctx.NoContent(http.StatusNoContent)
```

### Common Status Codes

```go
http.StatusOK                   // 200
http.StatusCreated              // 201
http.StatusAccepted             // 202
http.StatusNoContent            // 204
http.StatusMovedPermanently     // 301
http.StatusFound                // 302
http.StatusBadRequest           // 400
http.StatusUnauthorized         // 401
http.StatusForbidden            // 403
http.StatusNotFound             // 404
http.StatusConflict             // 409
http.StatusInternalServerError  // 500
```

---

**Next**: [Request Parsing](request.md)
