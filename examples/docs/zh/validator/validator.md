# Validator 扩展功能示例

本文档展示了 validator 包扩展后的所有验证规则和使用方法。

## 新增验证规则

### 1. 数字验证规则

#### Int - 整数验证
```go
type User struct {
    Age int `validate:"int"`
}

// 支持的类型：int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64
// 也支持字符串格式的整数
```

#### Float - 浮点数验证
```go
type Product struct {
    Price float64 `validate:"float"`
}

// 支持的类型：float32, float64
// 也支持字符串格式的浮点数
```

#### Range - 范围验证
```go
type Age struct {
    Value int `validate:"range=1,100"`
}

// 或者使用 Range 函数
rule := validator.Range(1, 100)
```

#### MinFloat / MaxFloat - 浮点数范围验证
```go
type Price struct {
    Value float64 `validate:"minFloat=0.01,maxFloat=9999.99"`
}

// 或者使用函数
rule := validator.MinFloat(0.01)
rule := validator.MaxFloat(9999.99)
rule := validator.RangeFloat(0.01, 9999.99)
```

### 2. 布尔值验证规则

#### Bool - 布尔值验证
```go
type Settings struct {
    Enabled bool `validate:"bool"`
}

// 支持：true, false, "true", "false", "1", "0"
```

### 3. UUID 验证规则

#### UUID - UUID格式验证
```go
type User struct {
    ID string `validate:"uuid"`
}

// 支持标准UUID格式：550e8400-e29b-41d4-a716-446655440000
```

### 4. IP地址验证规则

#### IPv4 - IPv4地址验证
```go
type Server struct {
    IP string `validate:"ipv4"`
}

// 支持：192.168.1.1, 127.0.0.1 等
```

#### IPv6 - IPv6地址验证
```go
type Server struct {
    IP string `validate:"ipv6"`
}

// 支持：2001:db8::1, 2001:0db8:0000:0000:0000:0000:0000:0001 等
```

#### IP - IP地址验证（IPv4或IPv6）
```go
type Server struct {
    IP string `validate:"ip"`
}

// 支持IPv4和IPv6
```

### 5. MAC地址验证规则

#### MAC - MAC地址验证
```go
type Network struct {
    MAC string `validate:"mac"`
}

// 支持：00:1A:2B:3C:4D:5E 或 00-1A-2B-3C-4D-5E
```

### 6. 十六进制验证规则

#### Hex - 十六进制字符串验证
```go
type Data struct {
    Hex string `validate:"hex"`
}

// 支持：abc123, ABC123, AbC123 等
```

### 7. Base64验证规则

#### Base64 - Base64编码验证
```go
type File struct {
    Content string `validate:"base64"`
}

// 支持标准Base64和URL编码（无填充）
```

### 8. 日期时间验证规则

#### Date - 日期验证
```go
type Event struct {
    Date string `validate:"date"`
}

// 格式：YYYY-MM-DD (例如：2024-01-15)
```

#### Time - 时间验证
```go
type Schedule struct {
    Time string `validate:"time"`
}

// 格式：HH:MM:SS 或 HH:MM (例如：12:34:56 或 12:34)
```

#### DateTime - 日期时间验证
```go
type Log struct {
    Timestamp string `validate:"datetime"`
}

// 支持多种格式：
// - 2024-01-15 12:34:56
// - 2024-01-15T12:34:56Z
// - RFC3339, RFC3339Nano 等
```

#### AfterDate / BeforeDate - 日期范围验证
```go
type DateRange struct {
    Start string `validate:"afterDate=2024-01-01"`
    End   string `validate:"beforeDate=2024-12-31"`
}

// 或者使用函数
rule := validator.AfterDate("2024-01-01")
rule := validator.BeforeDate("2024-12-31")
```

### 9. 数组/切片验证规则

#### Array - 数组/切片验证
```go
type Data struct {
    Items []string `validate:"array"`
}
```

#### MinItems / MaxItems - 数组长度验证
```go
type Data struct {
    Items []string `validate:"minItems=1,maxItems=10"`
}

// 或者使用函数
rule := validator.MinItems(1)
rule := validator.MaxItems(10)
```

#### Unique - 数组元素唯一性验证
```go
type Data struct {
    Items []string `validate:"unique"`
}
```

### 10. Map验证规则

#### MinMapKeys / MaxMapKeys - Map键数量验证
```go
type Data struct {
    Map map[string]string `validate:"minMapKeys=1,maxMapKeys=10"`
}

// 或者使用函数
rule := validator.MinMapKeys(1)
rule := validator.MaxMapKeys(10)
```

### 11. 对象验证规则

#### Object - 对象验证
```go
type Data struct {
    Value any `validate:"object"`
}

// 支持struct和map
```

### 12. 字符串验证规则

#### EmailList - 邮件列表验证
```go
type Contact struct {
    Emails string `validate:"emailList"`
}

// 支持逗号分隔的多个邮箱：test@example.com,admin@example.com
```

#### URLList - URL列表验证
```go
type Links struct {
    URLs string `validate:"urlList"`
}

// 支持逗号分隔的多个URL：https://example.com,https://google.com
```

#### OneOf - 枚举值验证
```go
type User struct {
    Role string `validate:"oneOf=admin,user,moderator"`
}

// 或者使用函数
rule := validator.OneOf("admin", "user", "moderator")
```

#### NotEmpty - 非空字符串验证
```go
type User struct {
    Name string `validate:"notEmpty"`
}
```

#### NotZero - 非零值验证
```go
type User struct {
    ID int `validate:"notZero"`
}
```

#### Contains - 子字符串验证
```go
type Text struct {
    Content string `validate:"contains=hello"`
}

// 或者使用函数
rule := validator.Contains("hello")
```

#### HasPrefix / HasSuffix - 前缀/后缀验证
```go
type Text struct {
    Content string `validate:"hasPrefix=http://,hasSuffix=.html"`
}

// 或者使用函数
rule := validator.HasPrefix("http://")
rule := validator.HasSuffix(".html")
```

#### CaseInsensitive - 大小写不敏感匹配
```go
type Text struct {
    Content string `validate:"caseInsensitive=^[a-z]+$"`
}

// 或者使用函数
rule := validator.CaseInsensitive("^[a-z]+$")
```

#### MinLengthBytes / MaxLengthBytes - 字节长度验证
```go
type Data struct {
    Content string `validate:"minLengthBytes=10,maxLengthBytes=100"`
}

// 或者使用函数
rule := validator.MinLengthBytes(10)
rule := validator.MaxLengthBytes(100)
```

### 13. 自定义验证规则

#### CustomRule - 自定义验证规则
```go
// 创建自定义验证规则
rule := validator.CustomRule("even", func(value any) bool {
    if i, ok := value.(int); ok {
        return i%2 == 0
    }
    return false
})
```

### 14. 组合验证规则

#### Optional - 可选验证
```go
type User struct {
    Email string `validate:"optional,email"`
}

// 或者使用函数
rule := validator.Optional(validator.Email())
```

#### WithMessage - 自定义错误消息
```go
// 或者使用函数
rule := validator.WithMessage(validator.Min(18), "年龄必须大于等于18岁")
```

#### WithCode - 自定义错误代码
```go
// 或者使用函数
rule := validator.WithCode(validator.Min(18), "AGE_TOO_SMALL")
```

## 使用示例

### 1. 基本使用

```go
package main

import (
    "fmt"
    "github.com/spcent/plumego/validator"
)

type User struct {
    Name     string `validate:"required,notEmpty"`
    Age      int    `validate:"int,range=1,100"`
    Email    string `validate:"email"`
    UUID     string `validate:"uuid"`
    IP       string `validate:"ip"`
    Role     string `validate:"oneOf=admin,user,moderator"`
    Birthday string `validate:"date,afterDate=1900-01-01"`
}

func main() {
    user := User{
        Name:     "John Doe",
        Age:      25,
        Email:    "john@example.com",
        UUID:     "550e8400-e29b-41d4-a716-446655440000",
        IP:       "192.168.1.1",
        Role:     "admin",
        Birthday: "1999-01-01",
    }

    err := validator.Validate(user)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### 2. 使用自定义验证规则

```go
package main

import (
    "fmt"
    "github.com/spcent/plumego/validator"
)

type Product struct {
    Price    float64 `validate:"minFloat=0.01"`
    Quantity int     `validate:"min=1"`
    Discount float64 `validate:"rangeFloat=0.0,0.5"`
}

func main() {
    product := Product{
        Price:    99.99,
        Quantity: 10,
        Discount: 0.1,
    }

    err := validator.Validate(product)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### 3. 使用组合验证规则

```go
package main

import (
    "fmt"
    "github.com/spcent/plumego/validator"
)

type Config struct {
    APIKey   string `validate:"required,notEmpty,minLength=32"`
    Timeout  int    `validate:"int,range=1,300"`
    Endpoint string `validate:"url,optional"`
}

func main() {
    config := Config{
        APIKey:   "abcdefghijklmnopqrstuvwxyz123456",
        Timeout:  60,
        Endpoint: "https://api.example.com",
    }

    err := validator.Validate(config)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

### 4. 使用自定义错误消息

```go
package main

import (
    "fmt"
    "github.com/spcent/plumego/validator"
)

type User struct {
    Age int `validate:"min=18"`
}

func main() {
    user := User{Age: 15}

    // 创建带自定义错误消息的验证器
    rule := validator.WithMessage(validator.Min(18), "年龄必须大于等于18岁")
    err := rule.Validate(user.Age)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err.Message)
    }
}
```

### 5. 使用自定义验证规则

```go
package main

import (
    "fmt"
    "github.com/spcent/plumego/validator"
)

type Product struct {
    Price float64
}

func main() {
    // 创建自定义验证规则：价格必须是偶数
    evenPriceRule := validator.CustomRule("evenPrice", func(value any) bool {
        if f, ok := value.(float64); ok {
            return int(f)%2 == 0
        }
        return false
    })

    product := Product{Price: 99.0}
    err := evenPriceRule.Validate(product.Price)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err.Message)
    }
}
```

### 6. 使用HTTP请求验证

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/spcent/plumego/validator"
)

type CreateUserRequest struct {
    Name  string `validate:"required,notEmpty"`
    Email string `validate:"email"`
    Age   int    `validate:"int,range=1,100"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := validator.BindJSON(r, &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 处理请求...
    fmt.Fprintf(w, "User created successfully")
}

func main() {
    http.HandleFunc("/users", createUserHandler)
    http.ListenAndServe(":8080", nil)
}
```

## 验证规则注册表

所有验证规则都已注册到默认规则注册表中，可以直接在struct tag中使用：

```go
type User struct {
    // 基本规则
    Name string `validate:"required"`
    Age  int    `validate:"min=18"`
    
    // 新增规则
    UUID     string `validate:"uuid"`
    IP       string `validate:"ip"`
    MAC      string `validate:"mac"`
    Hex      string `validate:"hex"`
    Base64   string `validate:"base64"`
    Date     string `validate:"date"`
    Time     string `validate:"time"`
    DateTime string `validate:"datetime"`
    
    // 数组规则
    Tags []string `validate:"minItems=1,maxItems=10,unique"`
    
    // Map规则
    Metadata map[string]string `validate:"minMapKeys=1,maxMapKeys=5"`
    
    // 字符串规则
    Role     string `validate:"oneOf=admin,user,moderator"`
    Email    string `validate:"emailList"`
    URLs     string `validate:"urlList"`
    Content  string `validate:"contains=hello,hasPrefix=http://,hasSuffix=.html"`
}
```

## 性能优化

1. **规则缓存**：验证规则在首次解析后会被缓存，提高后续验证性能
2. **线程安全**：使用读写锁保护规则注册表和缓存
3. **零依赖**：仅使用Go标准库，无外部依赖

## 错误处理

所有验证错误都返回 `ValidationError` 或 `FieldErrors`：

```go
type ValidationError struct {
    Field   string `json:"field"`
    Code    string `json:"code"`
    Message string `json:"message"`
    Value   any    `json:"value,omitempty"`
}

type FieldErrors struct {
    errors []ValidationError
}
```

错误信息可以轻松转换为JSON响应：

```go
if err := validator.Validate(data); err != nil {
    if fieldErrs, ok := err.(validator.FieldErrors); ok {
        // 返回详细的字段错误信息
        for _, fieldErr := range fieldErrs.Errors() {
            fmt.Printf("Field: %s, Code: %s, Message: %s\n", 
                fieldErr.Field, fieldErr.Code, fieldErr.Message)
        }
    }
}
```

## 总结

扩展后的validator包提供了全面的验证规则，涵盖了：

- ✅ 数字验证（整数、浮点数、范围）
- ✅ 布尔值验证
- ✅ UUID验证
- ✅ IP地址验证（IPv4、IPv6）
- ✅ MAC地址验证
- ✅ 十六进制验证
- ✅ Base64验证
- ✅ 日期时间验证
- ✅ 数组/切片验证
- ✅ Map验证
- ✅ 对象验证
- ✅ 字符串验证（邮件列表、URL列表、枚举、子字符串等）
- ✅ 自定义验证规则
- ✅ 组合验证规则

所有规则都保持了代码的简洁、优雅和高效，符合Go语言的最佳实践。
