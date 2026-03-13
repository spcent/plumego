package sharding

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SQLType represents the type of SQL statement
type SQLType int

const (
	SQLTypeUnknown SQLType = iota
	SQLTypeSelect
	SQLTypeInsert
	SQLTypeUpdate
	SQLTypeDelete
)

// String returns the string representation of SQLType
func (t SQLType) String() string {
	switch t {
	case SQLTypeSelect:
		return "SELECT"
	case SQLTypeInsert:
		return "INSERT"
	case SQLTypeUpdate:
		return "UPDATE"
	case SQLTypeDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// ParsedSQL contains parsed SQL information
type ParsedSQL struct {
	Type        SQLType
	TableName   string
	WhereClause string
	Columns     []string          // For INSERT
	Values      []string          // For INSERT VALUES
	Conditions  map[string]string // Parsed WHERE conditions (column -> placeholder)
}

var (
	// ErrUnsupportedSQL is returned when the SQL statement is not supported
	ErrUnsupportedSQL = errors.New("unsupported SQL statement")

	// ErrNoTable is returned when no table name is found
	ErrNoTable = errors.New("no table name found")
)

// SQLParser parses SQL statements to extract table names and conditions
type SQLParser struct {
	// Precompiled regex patterns
	selectPattern *regexp.Regexp
	insertPattern *regexp.Regexp
	updatePattern *regexp.Regexp
	deletePattern *regexp.Regexp
	wherePattern  *regexp.Regexp
}

// NewSQLParser creates a new SQL parser
func NewSQLParser() *SQLParser {
	return &SQLParser{
		// Match: SELECT ... FROM table_name
		selectPattern: regexp.MustCompile(`(?i)^\s*SELECT\s+.*?\s+FROM\s+` + "`?" + `([a-zA-Z_][a-zA-Z0-9_]*)` + "`?" + `\s*`),

		// Match: INSERT INTO table_name
		insertPattern: regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+` + "`?" + `([a-zA-Z_][a-zA-Z0-9_]*)` + "`?" + `\s*`),

		// Match: UPDATE table_name SET
		updatePattern: regexp.MustCompile(`(?i)^\s*UPDATE\s+` + "`?" + `([a-zA-Z_][a-zA-Z0-9_]*)` + "`?" + `\s+SET\s+`),

		// Match: DELETE FROM table_name
		deletePattern: regexp.MustCompile(`(?i)^\s*DELETE\s+FROM\s+` + "`?" + `([a-zA-Z_][a-zA-Z0-9_]*)` + "`?" + `\s*`),

		// Match: WHERE clause
		wherePattern: regexp.MustCompile(`(?i)\s+WHERE\s+(.+?)(?:\s+ORDER\s+BY|\s+LIMIT|\s+GROUP\s+BY|\s*$)`),
	}
}

// Parse parses a SQL statement and extracts relevant information
func (p *SQLParser) Parse(sql string) (*ParsedSQL, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, ErrUnsupportedSQL
	}

	// Determine SQL type and extract table name
	sqlType, tableName, err := p.parseTypeAndTable(sql)
	if err != nil {
		return nil, err
	}

	parsed := &ParsedSQL{
		Type:       sqlType,
		TableName:  tableName,
		Conditions: make(map[string]string),
	}

	// Extract WHERE clause if present
	if matches := p.wherePattern.FindStringSubmatch(sql); len(matches) > 1 {
		parsed.WhereClause = strings.TrimSpace(matches[1])
		parsed.Conditions = p.parseWhereConditions(parsed.WhereClause)
	}

	// Special handling for INSERT
	if sqlType == SQLTypeInsert {
		p.parseInsertColumnsAndValues(sql, parsed)
	}

	return parsed, nil
}

// parseTypeAndTable determines the SQL type and extracts the table name
func (p *SQLParser) parseTypeAndTable(sql string) (SQLType, string, error) {
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))

	// Try SELECT
	if strings.HasPrefix(sqlUpper, "SELECT") {
		if matches := p.selectPattern.FindStringSubmatch(sql); len(matches) > 1 {
			return SQLTypeSelect, matches[1], nil
		}
	}

	// Try INSERT
	if strings.HasPrefix(sqlUpper, "INSERT") {
		if matches := p.insertPattern.FindStringSubmatch(sql); len(matches) > 1 {
			return SQLTypeInsert, matches[1], nil
		}
	}

	// Try UPDATE
	if strings.HasPrefix(sqlUpper, "UPDATE") {
		if matches := p.updatePattern.FindStringSubmatch(sql); len(matches) > 1 {
			return SQLTypeUpdate, matches[1], nil
		}
	}

	// Try DELETE
	if strings.HasPrefix(sqlUpper, "DELETE") {
		if matches := p.deletePattern.FindStringSubmatch(sql); len(matches) > 1 {
			return SQLTypeDelete, matches[1], nil
		}
	}

	return SQLTypeUnknown, "", ErrUnsupportedSQL
}

// parseWhereConditions parses WHERE clause and extracts column conditions
// Handles patterns like: column = ?, column IN (?), column > ?
func (p *SQLParser) parseWhereConditions(whereClause string) map[string]string {
	conditions := make(map[string]string)

	// Remove backticks and normalize spaces
	whereClause = strings.ReplaceAll(whereClause, "`", "")
	whereClause = strings.TrimSpace(whereClause)

	// Split by AND/OR (simplified parsing)
	parts := splitConditions(whereClause)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Match: column = ? or column IN (?)
		// Look for patterns: column [operator] placeholder
		if idx := strings.Index(part, "="); idx > 0 {
			column := strings.TrimSpace(part[:idx])
			// Extract just the column name (remove any table alias)
			if dotIdx := strings.LastIndex(column, "."); dotIdx > 0 {
				column = column[dotIdx+1:]
			}
			conditions[column] = "="
		} else if strings.Contains(strings.ToUpper(part), " IN ") {
			// Handle IN clause
			inIdx := strings.Index(strings.ToUpper(part), " IN ")
			if inIdx > 0 {
				column := strings.TrimSpace(part[:inIdx])
				if dotIdx := strings.LastIndex(column, "."); dotIdx > 0 {
					column = column[dotIdx+1:]
				}
				conditions[column] = "IN"
			}
		} else {
			// Try other operators: >, <, >=, <=, !=
			for _, op := range []string{">=", "<=", "!=", "<>", ">", "<"} {
				if idx := strings.Index(part, op); idx > 0 {
					column := strings.TrimSpace(part[:idx])
					if dotIdx := strings.LastIndex(column, "."); dotIdx > 0 {
						column = column[dotIdx+1:]
					}
					conditions[column] = op
					break
				}
			}
		}
	}

	return conditions
}

// splitConditions splits WHERE clause by AND/OR operators
func splitConditions(whereClause string) []string {
	var parts []string
	var current strings.Builder
	var inParens int

	words := strings.Fields(whereClause)

	for i, word := range words {
		wordUpper := strings.ToUpper(word)

		// Track parentheses depth
		inParens += strings.Count(word, "(") - strings.Count(word, ")")

		// Split on AND/OR only when not inside parentheses
		if inParens == 0 && (wordUpper == "AND" || wordUpper == "OR") {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(words[i])
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseInsertColumnsAndValues extracts column names and values from INSERT statement
func (p *SQLParser) parseInsertColumnsAndValues(sql string, parsed *ParsedSQL) {
	// Match: INSERT INTO table (col1, col2) VALUES (?, ?)
	columnsPattern := regexp.MustCompile(`(?i)\((.*?)\)\s*VALUES\s*\((.*?)\)`)

	if matches := columnsPattern.FindStringSubmatch(sql); len(matches) > 2 {
		// Extract column names
		colStr := matches[1]
		cols := strings.Split(colStr, ",")
		for _, col := range cols {
			col = strings.TrimSpace(col)
			col = strings.Trim(col, "`")
			if col != "" {
				parsed.Columns = append(parsed.Columns, col)
			}
		}

		// Extract value placeholders
		valStr := matches[2]
		vals := strings.Split(valStr, ",")
		for _, val := range vals {
			val = strings.TrimSpace(val)
			if val != "" {
				parsed.Values = append(parsed.Values, val)
			}
		}
	}
}

// ExtractTableName is a convenience method to quickly extract just the table name
func (p *SQLParser) ExtractTableName(sql string) (string, error) {
	parsed, err := p.Parse(sql)
	if err != nil {
		return "", err
	}
	return parsed.TableName, nil
}

// IsWrite returns true if the SQL statement is a write operation
func (p *ParsedSQL) IsWrite() bool {
	return p.Type == SQLTypeInsert || p.Type == SQLTypeUpdate || p.Type == SQLTypeDelete
}

// IsRead returns true if the SQL statement is a read operation
func (p *ParsedSQL) IsRead() bool {
	return p.Type == SQLTypeSelect
}

// HasWhere returns true if the SQL has a WHERE clause
func (p *ParsedSQL) HasWhere() bool {
	return p.WhereClause != ""
}

// GetCondition returns the operator for a given column in WHERE clause
func (p *ParsedSQL) GetCondition(column string) (string, bool) {
	op, ok := p.Conditions[column]
	return op, ok
}

// String returns a string representation of the parsed SQL
func (p *ParsedSQL) String() string {
	return fmt.Sprintf("Type=%s, Table=%s, Where=%q, Columns=%v, Conditions=%v",
		p.Type, p.TableName, p.WhereClause, p.Columns, p.Conditions)
}
