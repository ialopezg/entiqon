// Code generated by ENTIQON.
// File: internal/build/token/condition.go
// Description: Defines a minimal SQL Condition structure for use in builders.
// Since: v1.6.0

package token

import (
	"fmt"
	"strings"

	"github.com/entiqon/entiqon/driver"
)

// Condition represents a simple SQL condition like "column = value" or "column IN (?)".
//
// This version is dialect-agnostic. It stores only the raw column name, operator,
// and value(s). Placeholder generation and quoting are resolved by the builder layer.
//
// Example:
//
//	cond := NewCondition("id", 1)
//	fmt.Println(cond.Column)    // "id"
//	fmt.Println(cond.Operator)  // "="
//	fmt.Println(cond.Value)     // 1
//	fmt.Println(cond.IsValid()) // true
type Condition struct {
	Type     ConditionType     // Clause type: Simple (WHERE), AND, OR
	Column   *Column           // Column definition
	Operator ConditionOperator // SQL operator (e.g. "=", "IN")
	Value    any               // Value(s) to compare against
	Error    error             // Validation error, if any
}

// NewCondition creates a new SQL condition using the given column and value.
//
// The operator defaults to "=" for simple equality comparisons.
//
// This function is dialect-agnostic. It does not format placeholders or quote
// identifiers — it only stores structural intent, which must be resolved
// later by the builder or renderer.
//
// # Examples
//
//	Simple equality:
//
//		cond := NewCondition("status", "active")
//		fmt.Println(cond.Column)   // "status"
//		fmt.Println(cond.Operator) // "="
//		fmt.Println(cond.Value)    // "active"
//
//	Fully qualified column:
//
//		cond := NewCondition("u.id", 123)
//		fmt.Println(cond.Column)   // "u.id"
//		fmt.Println(cond.Operator) // "="
//		fmt.Println(cond.Value)    // 123
//
//	This condition can later be rendered as:
//	  "u.id = ?" or `"u"."id" = $1` depending on dialect.
func NewCondition(column string, args ...any) *Condition {
	return resolveCondition(ConditionTypeSimple, column, args...)
}

func NewConditionAnd(column string, args ...any) *Condition {
	return resolveCondition(ConditionTypeAnd, column, args...)
}

func NewConditionOr(column string, args ...any) *Condition {
	return resolveCondition(ConditionTypeOr, column, args...)
}

// NewConditionWith creates a condition using the given clause type and operator.
//
// Example:
//
//	c := NewConditionWith(ConditionTypeSimple, "created_at", GreaterThan, "2024-01-01")
//	fmt.Println(c.Type)     // ConditionTypeSimple
//	fmt.Println(c.Operator) // ">"
func NewConditionWith(kind ConditionType, name string, operator ConditionOperator, values ...any) *Condition {
	var value any
	if len(values) == 1 {
		value = values[0]
	} else {
		value = values
	}

	return &Condition{
		Type:     kind,
		Column:   NewColumn(name),
		Operator: operator,
		Value:    value,
	}
}

// IsValid returns true if the condition is structurally valid and does not contain an error.
//
// It performs internal validation on first call, ensuring:
//   - Error is nil (early exit if already set)
//   - Column is not empty
//   - Operator is one of the known and supported operators
//
// If any validation fails, it sets the first encountered error
// and avoids overwriting previously existing errors.
func (c *Condition) IsValid() bool {
	if c.Error != nil {
		return false
	}

	if c.Column == nil || !c.Column.IsValid() {
		c.SetError(fmt.Errorf("column definition is required"))
		return false
	}
	if !c.Operator.IsValid() {
		c.SetError(fmt.Errorf("invalid or unsupported operator: %q", c.Operator))
		return false
	}
	return true
}

// Render builds the SQL fragment and parameter list for the condition.
//
// It uses the provided ParamBinder to generate dialect-aware placeholders.
//
// Example:
//
//	c := NewCondition("status", "IN", []any{"active", "banned"})
//	binder := &util.ParamBinder{Dialect: postgres}
//	sql, args := c.Render(binder)
//	// sql  = "status IN ($1, $2)"
//	// args = ["active", "banned"]
func (c *Condition) Render(d driver.Dialect) (string, []any) {
	if c == nil || c.Column == nil || !c.Column.IsValid() {
		return "", nil
	}

	d.ResetPlaceholders()
	col := c.Column.Render(d)

	switch c.Operator {
	case IsNull, IsNotNull:
		return fmt.Sprintf("%s %s", col, c.Operator), nil

	case In, NotIn:
		values, ok := c.Value.([]any)
		if !ok || len(values) == 0 {
			return fmt.Sprintf("%s %s ()", col, c.Operator), nil
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = d.NextPlaceholder()
		}
		return fmt.Sprintf("%s %s (%s)", col, c.Operator, strings.Join(placeholders, ", ")), values

	case Between:
		values, ok := c.Value.([]any)
		if !ok || len(values) != 2 {
			return fmt.Sprintf("%s BETWEEN ? AND ?", col), nil
		}
		return fmt.Sprintf("%s BETWEEN %s AND %s", col, d.NextPlaceholder(), d.NextPlaceholder()), values

	default:
		return fmt.Sprintf("%s %s %s", col, c.Operator, d.NextPlaceholder()), []any{c.Value}

	}
}

// SetError attaches a validation error to the condition.
//
// Returns the same Condition pointer for fluent chaining.
//
// Example:
//
//	c := NewCondition("age", 25).SetError(fmt.Errorf("must be >= 0"))
//	fmt.Println(c.IsValid())     // false
//	fmt.Println(c.Error.Error()) // "must be >= 0"
func (c *Condition) SetError(err error) *Condition {
	c.Error = err
	return c
}

// String returns a string view of the condition for inspection or debugging.
//
// Output examples:
//
//	"status = ?"
//	"created BETWEEN ? AND ?"
//	"id IN (?)"
func (c *Condition) String() string {
	if c == nil {
		return "Condition(nil)"
	}

	name := "<nil>"
	qualified := false
	valid := false

	if c.Column != nil {
		qualified = c.Column.IsQualified()
		name = c.Column.GetName()
		valid = c.Column.IsValid()
	}

	return fmt.Sprintf(
		`Condition(%q) [qualified: %v, column: %v, value: %v, errored: %v]`,
		name,
		qualified,
		valid,
		c.Value,
		c.Error != nil,
	)
}

func resolveCondition(kind ConditionType, column string, args ...any) *Condition {
	switch len(args) {
	case 1:
		return NewConditionWith(kind, column, Equal, args[0])

	case 2:
		op, err := ParseConditionOperator(args[0])
		if err != nil {
			return invalidCondition(kind, column, err)
		}
		return NewConditionWith(kind, column, op, args[1])

	default:
		return invalidCondition(kind, column, fmt.Errorf(
			"too many arguments: expected 1 or 2, got %d", len(args),
		))

	}
}

func invalidCondition(kind ConditionType, column string, err error) *Condition {
	c := &Condition{
		Type:   kind,
		Column: NewColumn(column),
	}
	return c.SetError(err)
}
