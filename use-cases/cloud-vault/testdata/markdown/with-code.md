# Document with Code Blocks

This document contains various code examples to test code block rendering and syntax highlighting.

## Go Example

Here's a simple Go function:

```go
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    for i := 0; i < 10; i++ {
        fmt.Printf("fib(%d) = %d\n", i, fibonacci(i))
    }
}
```

## JavaScript Example

A JavaScript class example:

```javascript
class EventEmitter {
  constructor() {
    this.listeners = new Map();
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
    return this;
  }

  emit(event, ...args) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(cb => cb(...args));
    }
  }
}
```

## Inline Code

You can also use `inline code` within paragraphs. For example, the `fmt.Println()` function prints to stdout.

## SQL Example

```sql
SELECT 
    d.id,
    d.title,
    COUNT(dv.id) as version_count
FROM documents d
LEFT JOIN document_versions dv ON dv.document_id = d.id
WHERE d.status = 'active'
GROUP BY d.id, d.title
HAVING version_count > 5
ORDER BY version_count DESC;
```

## Conclusion

Code blocks are essential for technical documentation and should be properly formatted and highlighted.
