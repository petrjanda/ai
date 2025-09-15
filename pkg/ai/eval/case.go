package eval

import (
	"context"

	_ "embed"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/getsynq/ai/pkg/ai/eval/expectations"
	"github.com/getsynq/ai/pkg/ai/workflows"
)

type Case struct {
	Input        ai.History
	Expectations []expectations.Expectation
	Alias        string
}

func NewCase(input ai.History, alias string, opts ...CaseOption) *Case {
	c := &Case{
		Input: input,
		Alias: alias,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type CaseOption func(*Case)

func WithAlias(alias string) CaseOption {
	return func(c *Case) {
		c.Alias = alias
	}
}

func (c *Case) Expect(expectation expectations.Expectation) *Case {
	c.Expectations = append(c.Expectations, expectation)
	return c
}

func (c *Case) String() string {
	return c.Alias
}

type CaseResult struct {
	Total  int
	Ok     int
	Errors int
}

func (c *Case) Eval(ctx context.Context, events SuiteEvents, variant workflows.Task, actual string) *CaseResult {
	errors := []error{}

	for _, expectation := range c.Expectations {
		events.OnExpectationStart(variant, c, expectation)
		if err := expectation.Eval(ctx, actual); err != nil {
			events.OnExpectationError(variant, c, actual, expectation, err)
			errors = append(errors, err)
		} else {
			events.OnExpectationEnd(variant, c, expectation, nil)
		}
	}

	events.OnCaseEnd(variant, c, errors)

	return &CaseResult{
		Total:  len(c.Expectations),
		Ok:     len(c.Expectations) - len(errors),
		Errors: len(errors),
	}
}
