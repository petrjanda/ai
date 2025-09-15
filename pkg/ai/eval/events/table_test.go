package events

// import (
// 	"errors"
// 	"testing"

// 	"github.com/getsynq/ai/pkg/ai"
// 	"github.com/getsynq/ai/pkg/ai/eval"
// 	"github.com/getsynq/ai/pkg/ai/eval/expectations"
// 	"github.com/getsynq/ai/pkg/ai/workflows"
// )

// func TestTableSuiteEvents(t *testing.T) {
// 	// Create mock cases
// 	cases := []*eval.Case{
// 		eval.NewCase(ai.NewHistory(ai.NewUserMessage("Test case 1")), "Test case 1").
// 			Expect(expectations.JSONPath("field1").Eq("value1")).
// 			Expect(expectations.JSONPath("field2").Exists()),
// 		eval.NewCase(ai.NewHistory(ai.NewUserMessage("Test case 2")), "Test case 2").
// 			Expect(expectations.JSONPath("field3").Eq("value3")),
// 	}

// 	// Create mock variants
// 	variants := []workflows.Task{
// 		workflows.NewTask("gpt-3.5-turbo", ai.NewLLMRequest(ai.WithModel("gpt-3.5-turbo"), ai.WithSystem("Test prompt"))),
// 		workflows.NewTask("gpt-4", ai.NewLLMRequest(ai.WithModel("gpt-4"), ai.WithSystem("Test prompt"))),
// 	}

// 	// Create table events
// 	tableEvents := NewTableSuiteEvents()

// 	// Create suite
// 	suite := eval.NewSuite(tableEvents, cases, variants)

// 	// Test OnSuiteStart
// 	tableEvents.OnSuiteStart(suite)

// 	// Simulate some test results
// 	for _, variant := range variants {
// 		for _, case_ := range cases {
// 			// Simulate case start
// 			tableEvents.OnCaseStart(variant, case_)

// 			// Simulate expectation results
// 			caseErrors := []error{}
// 			for _, expectation := range case_.Expectations {
// 				// Simulate some passing and some failing
// 				if variant.Name() == "gpt-4" {
// 					tableEvents.OnExpectationEnd(variant, case_, expectation, nil) // PASS
// 				} else {
// 					err := errors.New("test expectation failed")
// 					tableEvents.OnExpectationError(variant, case_, "", expectation, err) // FAIL
// 					caseErrors = append(caseErrors, err)
// 				}
// 			}
// 			tableEvents.OnCaseEnd(variant, case_, caseErrors)
// 		}
// 	}

// 	// Test OnSuiteEnd (this will print the table)
// 	tableEvents.OnSuiteEnd(suite)

// 	// Verify that we have results for all variants
// 	tableEventsImpl := tableEvents.(*TableSuiteEvents)
// 	if len(tableEventsImpl.variants) != 2 {
// 		t.Errorf("Expected 2 variants, got %d", len(tableEventsImpl.variants))
// 	}

// 	// Verify that we have results for all test combinations
// 	expectedTests := 3 // 2 expectations in case 1 + 1 expectation in case 2
// 	for _, variant := range tableEventsImpl.variants {
// 		if len(variant.Results) != expectedTests {
// 			t.Errorf("Expected %d test results for variant %s, got %d",
// 				expectedTests, variant.Variant.Name, len(variant.Results))
// 		}
// 	}
// }
