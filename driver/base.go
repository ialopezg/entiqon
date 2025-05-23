// Code generated by ENTIQON.
// File: internal/core/driver/dialect_base.go
// Description: Provides BaseDialect implementation of the Dialect interface.
// Since: v1.4.0

package driver

import (
	"fmt"
	"strconv"
	"strings"
)

// BaseDialect provides a foundational implementation of the Dialect interface.
// It can be embedded and selectively overridden by specific dialect structs.
//
// Since: v1.4.0
type BaseDialect struct {
	// EnableAliasing indicates whether the dialect supports table aliases
	// in clauses such as DELETE FROM, UPDATE FROM, or SELECT FROM.
	//
	// When true, builders will include alias expressions using RenderFrom.
	//
	// Since: v1.4.0
	EnableAliasing bool

	// EnableReturning specifies whether the dialect supports SQL RETURNING clauses,
	// such as `INSERT ... RETURNING id` or `UPDATE ... RETURNING *`.
	//
	// This flag is evaluated by the SupportsReturning method.
	// Commonly enabled in dialects like PostgreSQL.
	//
	// Since: v1.4.0
	EnableReturning bool

	// EnableUpsert specifies whether the dialect supports native UPSERT syntax,
	// such as `INSERT ... ON CONFLICT DO UPDATE` or `INSERT ... ON DUPLICATE KEY UPDATE`.
	//
	// This flag is evaluated by the SupportsUpsert method.
	// Typically enabled in PostgreSQL and MySQL dialects.
	//
	// Since: v1.4.0
	EnableUpsert bool

	// Name holds the dialect identifier (e.g., "postgres", "mysql").
	Name string

	// Quotation defines the quoting style used for identifiers.
	Quotation QuotationType

	// PlaceholderSymbol is an optional function that generates argument placeholders (e.g., $1, ?, :GetName).
	PlaceholderSymbol PlaceholderSymbol
}

// GetName returns the dialect Name.
// If unset, "base" is returned.
//
// Updated: v1.4.0
func (b *BaseDialect) GetName() string {
	if b.Name == "" {
		return "base"
	}
	return b.Name
}

// QuoteType returns the configured identifier Quotation style.
//
// Since: v1.4.0
func (b *BaseDialect) QuoteType() QuotationType {
	return b.Quotation
}

// QuoteIdentifier returns the given identifier with dialect-appropriate quoting.
// Defaults to no quoting unless the Quotation style is configured.
//
// Updated: v1.4.0
func (b *BaseDialect) QuoteIdentifier(identifier string) string {
	switch b.Quotation {
	case QuoteDouble:
		return `"` + identifier + `"`
	case QuoteBacktick:
		return "`" + identifier + "`"
	case QuoteBracket:
		return "[" + identifier + "]"
	default:
		return identifier
	}
}

// QuoteLiteral returns a printable literal string for debugging/logging purposes only.
// ⚠️ DO NOT use this for building real queries — use placeholders instead.
//
// Updated: v1.4.0
func (b *BaseDialect) QuoteLiteral(value any) string {
	switch v := value.(type) {
	case string:
		return "'" + v + "'"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// Placeholder returns a positional argument placeholder string.
// If a placeholder function is set, it delegates to that.
// Otherwise, returns a generic "?".
//
// Updated: v1.4.0
func (b *BaseDialect) Placeholder(index int) string {
	if !b.PlaceholderSymbol.IsValid() {
		return "?"
	}
	if b.PlaceholderSymbol == PlaceholderQuestion {
		return "?"
	}
	// Always fallback to dynamic prefix behavior
	return fmt.Sprintf("%s%d", b.PlaceholderSymbol, index)
}

// BuildLimitOffset returns the dialect-compatible LIMIT and OFFSET clause.
// Supports any combination of positive limit/offset values.
// Returns an empty string if neither is defined.
//
// Updated: v1.4.0
func (b *BaseDialect) BuildLimitOffset(limit, offset int) string {
	switch {
	case limit >= 0 && offset >= 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit >= 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset >= 0:
		return fmt.Sprintf("OFFSET %d", offset)
	default:
		return ""
	}
}

// RenderFrom returns a dialect-safe FROM clause expression.
// Table Name is quoted; alias is not quoted.
//
// Example: `"users" u`, `\`orders\` o`, `[logs] l`
//
// Updated: v1.4.0
func (b *BaseDialect) RenderFrom(table string, alias string) string {
	quoted := b.QuoteIdentifier(table)
	if alias != "" && b.EnableAliasing {
		return fmt.Sprintf("%s %s", quoted, alias)
	}
	return quoted
}

// SupportsReturning returns true if the dialect supports RETURNING clauses.
//
// This delegates to the EnableReturning field in BaseDialect, but may be overridden
// in custom dialect implementations to enforce computed behavior.
//
// Since: v1.4.0
func (b *BaseDialect) SupportsReturning() bool {
	return b.EnableReturning
}

// SupportsUpsert returns true if the dialect supports native UPSERT syntax.
//
// This default implementation reads the EnableUpsert field in BaseDialect.
// Custom dialects may override this method to implement dynamic or computed support.
//
// Since: v1.4.0
func (b *BaseDialect) SupportsUpsert() bool {
	return b.EnableUpsert
}

// Validate ensures the BaseDialect is correctly configured.
//
// A dialect is considered valid if:
//   - PlaceholderSymbol is set to a non-empty value
//   - Name is not empty or whitespace
//
// The Quotation style is allowed to be QuoteNone if the dialect does not require identifier quoting.
//
// Returns:
//   - nil if the dialect is valid
//   - an error if the configuration is incomplete or invalid
//
// Since: v1.4.0
func (b *BaseDialect) Validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("BaseDialect: name is not set")
	}
	if !b.PlaceholderSymbol.IsValid() {
		return fmt.Errorf("BaseDialect: placeholder symbol is not set")
	}
	return nil
}
