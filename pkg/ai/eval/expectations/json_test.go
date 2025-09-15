package expectations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONExpectations(t *testing.T) {
	jsonData := `{
		"tests": [
			{
				"explanation": "Validate that started_at timestamps are not in the far future (more than 1 day from now) to catch potential data entry errors or system clock issues",
				"min_max": {
					"column_name": "started_at", 
					"max_value": 1
				}
			},
			{
				"explanation": "Validate that started_at timestamps are not unreasonably old (before 2020) to ensure data quality and catch potential migration issues",
				"min_max": {
					"column_name": "started_at",
					"min_value": 18262
				}
			},
			{
				"business_rule": {
					"sql_expression": "ended_at != toDateTime('1900-01-01 00:00:00') AND toDate(ended_at) < toDate('2020-01-01')"
				},
				"explanation": "Validate that ended_at timestamps are either the default placeholder value (1900-01-01) for ongoing issues or a reasonable date after 2020"
			},
			{
				"business_rule": {
					"sql_expression": "ended_at != toDateTime('1900-01-01 00:00:00') AND ended_at <= started_at"
				},
				"explanation": "Validate that when ended_at is not the default placeholder value, it occurs after started_at to ensure proper temporal ordering"
			}
		]
	}`

	t.Run("should find ended_at in sql expressions", func(t *testing.T) {
		exp := JSONPath("tests.#.business_rule.sql_expression").AnyContains("ended_at")
		err := exp.Eval(context.Background(), jsonData)
		require.NoError(t, err)
	})

	t.Run("should not find created_at in sql expressions", func(t *testing.T) {
		exp := JSONPath("tests.#.business_rule.sql_expression").AnyContains("created_at")
		err := exp.Eval(context.Background(), jsonData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "list does not contain 'created_at'")
	})
}
