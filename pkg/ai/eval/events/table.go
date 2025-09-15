package events

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/getsynq/ai/pkg/ai/eval"
	"github.com/getsynq/ai/pkg/ai/eval/expectations"
	"github.com/getsynq/ai/pkg/ai/workflows"
)

// TableSuiteEvents creates a table showing test results across variants.
// Usage example:
//   suite := eval.NewSuite(
//       events.NewTableSuiteEvents(), // Use table events instead of logging events
//       cases,
//       model,
//       variants,
//   )
//   suite.Run(ctx)

// TestResult represents the result of a single test (case + expectation)
type TestResult struct {
	Case        *eval.Case
	Expectation expectations.Expectation
	Passed      bool
	Error       error
}

// VariantResult tracks all test results for a single variant
type VariantResult struct {
	Variant workflows.Task
	Results map[string]*TestResult // key: caseAlias + "|" + expectationString
	Summary *eval.SuiteResult
}

// TableSuiteEvents creates a table showing test results across variants
type TableSuiteEvents struct {
	variants map[string]*VariantResult // key: variant.Name()
	cases    []*eval.Case
}

func NewTableSuiteEvents() eval.SuiteEvents {
	return &TableSuiteEvents{
		variants: make(map[string]*VariantResult),
		cases:    make([]*eval.Case, 0),
	}
}

func (e *TableSuiteEvents) OnSuiteStart(s *eval.Suite) {
	e.cases = s.Cases
	// Initialize variant results
	for _, variant := range s.Tasks {
		e.variants[variant.Name()] = &VariantResult{
			Variant: variant,
			Results: make(map[string]*TestResult),
			Summary: eval.NewSuiteResult(),
		}
	}
}

func (e *TableSuiteEvents) OnSuiteEnd(s *eval.Suite) {
	e.printTable()
	e.printSummary()
	e.printUsage(s.Usage)
}

func (e *TableSuiteEvents) OnSuiteError(error error) {
	// Handle suite-level errors if needed
}

func (e *TableSuiteEvents) OnCaseStart(variant workflows.Task, case_ *eval.Case) {
	// No action needed for case start

	fmt.Printf("%s | case '%s' \n", variant.Name(), case_)
}

func (e *TableSuiteEvents) OnCaseEnd(variant workflows.Task, case_ *eval.Case, errors []error) {
	fmt.Printf("  = [\033[32mOK\033[0m] total=%d, ok=%d, error=%d\n", len(case_.Expectations), len(case_.Expectations)-len(errors), len(errors))

	variantResult := e.variants[variant.Name()]
	if variantResult == nil {
		return
	}

	// Update variant summary
	result := &eval.CaseResult{
		Total:  len(case_.Expectations),
		Ok:     len(case_.Expectations) - len(errors),
		Errors: len(errors),
	}
	variantResult.Summary.Result(result)
}

func (e *TableSuiteEvents) OnCaseError(variant workflows.Task, case_ *eval.Case, err error) {
	fmt.Printf("  = [\033[31mERR\033[0m] %v\n", err)

	variantResult := e.variants[variant.Name()]
	if variantResult == nil {
		return
	}

	// Mark all expectations for this case as failed
	for _, expectation := range case_.Expectations {
		key := e.getTestKey(case_, expectation)
		variantResult.Results[key] = &TestResult{
			Case:        case_,
			Expectation: expectation,
			Passed:      false,
			Error:       err,
		}
	}

	// Update variant summary
	variantResult.Summary.FatalResult()
}

func (e *TableSuiteEvents) OnExpectationStart(variant workflows.Task, case_ *eval.Case, expectation expectations.Expectation) {
	fmt.Printf("%s | expectation '%s' ", variant.Name(), expectation.String())
}

func (e *TableSuiteEvents) OnExpectationEnd(variant workflows.Task, case_ *eval.Case, expectation expectations.Expectation, err error) {
	fmt.Printf("[\033[32mOK\033[0m]\n")

	variantResult := e.variants[variant.Name()]
	if variantResult == nil {
		return
	}

	key := e.getTestKey(case_, expectation)
	variantResult.Results[key] = &TestResult{
		Case:        case_,
		Expectation: expectation,
		Passed:      err == nil,
		Error:       err,
	}
}

func (e *TableSuiteEvents) OnExpectationError(variant workflows.Task, case_ *eval.Case, actual string, expectation expectations.Expectation, err error) {
	fmt.Printf("[\033[31mERR\033[0m] %v\n", err)

	variantResult := e.variants[variant.Name()]
	if variantResult == nil {
		return
	}

	key := e.getTestKey(case_, expectation)
	variantResult.Results[key] = &TestResult{
		Case:        case_,
		Expectation: expectation,
		Passed:      false,
		Error:       err,
	}
}

// getTestKey creates a unique key for a test (case + expectation)
func (e *TableSuiteEvents) getTestKey(case_ *eval.Case, expectation expectations.Expectation) string {
	return fmt.Sprintf("%s|%s", case_.String(), expectation.String())
}

// printTable creates and prints the results table
func (e *TableSuiteEvents) printTable() {
	if len(e.variants) == 0 {
		return
	}

	fmt.Println("\n=== Results ===")

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Get ordered list of variants for consistent column ordering
	var orderedVariants []*VariantResult
	for _, variant := range e.variants {
		orderedVariants = append(orderedVariants, variant)
	}

	// Print header
	header := "Test Case"
	for _, variant := range orderedVariants {
		header += "\t" + string(variant.Variant.Name())
	}
	fmt.Fprintln(w, header)

	// Print separator
	separator := strings.Repeat("-", 20)
	for range orderedVariants {
		separator += "\t" + strings.Repeat("-", 10)
	}
	fmt.Fprintln(w, separator)

	// Print test results
	for _, case_ := range e.cases {
		for _, expectation := range case_.Expectations {
			row := fmt.Sprintf("%s | %s", case_.String(), expectation.String())

			for _, variant := range orderedVariants {
				key := e.getTestKey(case_, expectation)
				result := variant.Results[key]

				if result == nil {
					row += "\tFATAL"
				} else if result.Passed {
					row += "\tPASS"
				} else {
					row += "\tFAIL"
				}
			}

			fmt.Fprintln(w, row)
		}
	}

	w.Flush()
	fmt.Println()
}

// printSummary prints the summary per variant
func (e *TableSuiteEvents) printSummary() {
	fmt.Println("=== Summary ===")

	for _, variant := range e.variants {
		summary := variant.Summary
		score := float64(0)
		if summary.ExpectationsTotal > 0 {
			score = float64(summary.ExpectationsOk) / float64(summary.ExpectationsTotal) * 100
		}

		fmt.Printf("%s:", variant.Variant.Name())
		fmt.Printf(", cases total=%d, ok=%d, error=%d, fatal=%d",
			summary.CasesTotal, summary.CasesOk, summary.CasesErrors, summary.CasesFatal)
		fmt.Printf(", expectations total=%d, ok=%d, error=%d",
			summary.ExpectationsTotal, summary.ExpectationsOk, summary.ExpectationsErrors)
		fmt.Printf(", score %.1f%%\n", score)
	}
}

func (e *TableSuiteEvents) printUsage(usage *ai.LLMUsage) {
	fmt.Println()
	fmt.Println("=== Usage ===")
	fmt.Printf("%s\n", usage)
}
