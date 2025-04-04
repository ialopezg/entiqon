package builder

import (
	"fmt"
	"strings"
)

type ConditionType string

const (
	ConditionSimple ConditionType = "SIMPLE"
	ConditionAnd    ConditionType = "AND"
	ConditionOr     ConditionType = "OR"
)

type ConditionToken struct {
	Type      ConditionType
	Condition string
}

type SelectQueryBuilder struct {
	columns    []string
	from       string
	conditions []ConditionToken
	sorting    []string
	take       *int
	skip       *int
}

// Select sets the columns to retrieve
func (sb *SelectQueryBuilder) Select(columns ...string) *SelectQueryBuilder {
	sb.columns = columns
	return sb
}

// From sets the table to select from
func (sb *SelectQueryBuilder) From(from string) *SelectQueryBuilder {
	sb.from = from
	return sb
}

// Where initializes the WHERE conditions (resets any existing conditions)
func (sb *SelectQueryBuilder) Where(conditions ...string) *SelectQueryBuilder {
	sb.conditions = []ConditionToken{}
	sb.addCondition(ConditionSimple, conditions...)
	return sb
}

// AndWhere adds an AND condition
func (sb *SelectQueryBuilder) AndWhere(conditions ...string) *SelectQueryBuilder {
	sb.addCondition(ConditionAnd, conditions...)
	return sb
}

// OrWhere adds an OR condition
func (sb *SelectQueryBuilder) OrWhere(conditions ...string) *SelectQueryBuilder {
	sb.addCondition(ConditionOr, conditions...)
	return sb
}

// OrderBy adds an ORDER BY clause
func (sb *SelectQueryBuilder) OrderBy(column string) *SelectQueryBuilder {
	sb.sorting = append(sb.sorting, column)
	return sb
}

// Take sets the LIMIT (engine-agnostic equivalent)
func (sb *SelectQueryBuilder) Take(value int) *SelectQueryBuilder {
	sb.take = &value
	return sb
}

// Skip sets the OFFSET (engine-agnostic equivalent)
func (sb *SelectQueryBuilder) Skip(value int) *SelectQueryBuilder {
	sb.skip = &value
	return sb
}

// Build builds the SQL string
func (sb *SelectQueryBuilder) Build() (string, error) {
	if sb.from == "" {
		return "", fmt.Errorf("FROM clause is required")
	}

	columns := "*"
	if len(sb.columns) > 0 {
		columns = strings.Join(sb.columns, ", ")
	}

	tokens := []string{
		fmt.Sprintf("SELECT %s", columns),
		fmt.Sprintf("FROM %s", sb.from),
	}

	if len(sb.conditions) > 0 {
		var parts []string
		for _, condition := range sb.conditions {
			switch condition.Type {
			case ConditionSimple:
				parts = append(parts, condition.Condition)
			case ConditionAnd, ConditionOr:
				parts = append(parts, fmt.Sprintf("%s %s", condition.Type, condition.Condition))
			default:
				return "", fmt.Errorf("invalid condition type: %s", condition.Type)
			}
		}
		tokens = append(tokens, fmt.Sprintf("WHERE %s", strings.Join(parts, " ")))
	}

	if len(sb.sorting) > 0 {
		tokens = append(tokens, "ORDER BY "+strings.Join(sb.sorting, ", "))
	}

	if sb.take != nil {
		tokens = append(tokens, fmt.Sprintf("LIMIT %d", *sb.take))
	}
	if sb.skip != nil {
		tokens = append(tokens, fmt.Sprintf("OFFSET %d", *sb.skip))
	}

	return strings.Join(tokens, " "), nil
}

func (sb *SelectQueryBuilder) addCondition(conditionType ConditionType, conditions ...string) {
	if len(conditions) == 0 {
		return
	}
	if len(conditions) == 1 {
		sb.conditions = append(sb.conditions, ConditionToken{
			Type:      conditionType,
			Condition: conditions[0],
		})
		return
	}
	// Join and wrap in parentheses
	joiner := string(conditionType)
	group := "(" + strings.Join(conditions, fmt.Sprintf(" %s ", joiner)) + ")"
	sb.conditions = append(sb.conditions, ConditionToken{
		Type:      conditionType,
		Condition: group,
	})
}
