package events

import (
	"encoding/json"
	"fmt"

	"github.com/getsynq/ai/pkg/ai/eval"
	"github.com/getsynq/ai/pkg/ai/eval/expectations"
	"github.com/getsynq/ai/pkg/ai/workflows"
)

type LoggingSuiteEvents struct{}

func NewLoggingSuiteEvents() eval.SuiteEvents {
	return &LoggingSuiteEvents{}
}

func (e *LoggingSuiteEvents) OnSuiteStart(s *eval.Suite) {
	fmt.Printf("")
	fmt.Printf("running suite with %d cases\n", len(s.Cases))
	fmt.Println("")
}

func (e *LoggingSuiteEvents) OnSuiteEnd(s *eval.Suite) {
	// fmt.Printf("  = total=%d, ok=%d, error=%d, fatal=%d | expectations total=%d, ok=%d, error=%d | score=%.1f\n",
	// 	s.Total.CasesTotal, s.Total.CasesOk, s.Total.CasesErrors, s.Total.CasesFatal,
	// 	s.Total.ExpectationsTotal, s.Total.ExpectationsOk, s.Total.ExpectationsErrors,
	// 	float64(s.Total.ExpectationsOk)/float64(s.Total.ExpectationsTotal)*100,
	// )

	fmt.Printf("  ~ usage input=%d, output=%d, total=%d\n",
		s.Usage.PromptTokens, s.Usage.CompletionTokens, s.Usage.TotalTokens,
	)
}

func (e *LoggingSuiteEvents) OnSuiteError(error error) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnCaseStart(variant workflows.Task, case_ *eval.Case) {
	// fmt.Printf("%s | case '%s'\n", variant.Model, case_)
}

func (e *LoggingSuiteEvents) OnCaseEnd(variant workflows.Task, case_ *eval.Case, errors []error) {
	fmt.Printf("%s | case '%s' = total=%d, ok=%d, error=%d\n", variant.Name(), case_, len(case_.Expectations), len(case_.Expectations)-len(errors), len(errors))
	// fmt.Println("")
}

func (e *LoggingSuiteEvents) OnCaseError(variant workflows.Task, case_ *eval.Case, err error) {
	fmt.Printf("%s |  — %v ... [\033[31mERR\033[0m]\n", variant.Name(), err)
	// fmt.Println("")
}

func (e *LoggingSuiteEvents) OnExpectationStart(variant workflows.Task, case_ *eval.Case, expectation expectations.Expectation) {
	// @TODO
}

func (e *LoggingSuiteEvents) OnExpectationEnd(variant workflows.Task, case_ *eval.Case, expectation expectations.Expectation, err error) {
	// fmt.Printf("  — %v ... [\033[32mOK\033[0m]\n", expectation)
}

func (e *LoggingSuiteEvents) OnExpectationError(variant workflows.Task, case_ *eval.Case, actual string, expectation expectations.Expectation, err error) {
	fmt.Printf("  — %v ... [\033[31mERR\033[0m]\n", expectation)

	var directive map[string]any
	if err = json.Unmarshal([]byte(actual), &directive); err != nil {
		fmt.Printf("failed unmarshalling: %v", actual)
	}

	if payload, err := json.MarshalIndent(directive, "", "  "); err != nil {
		fmt.Printf("failed marshalling: %v", err)
	} else {
		fmt.Println("actual output was ```")
		fmt.Println(string(payload))
		fmt.Println("```")
	}
}
