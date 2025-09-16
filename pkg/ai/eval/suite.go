package eval

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/eval/expectations"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"

	_ "embed"
)

type Suite struct {
	Cases  []*Case
	Tasks  []workflows.Task
	Usage  *ai.LLMUsage
	events SuiteEvents

	llm ai.LLM
	ctx context.Context
}

type SuiteEvents interface {
	OnSuiteStart(suite *Suite)
	OnSuiteEnd(suite *Suite)
	OnSuiteError(error error)

	OnCaseStart(task workflows.Task, case_ *Case)
	OnCaseEnd(task workflows.Task, case_ *Case, errors []error)
	OnCaseError(task workflows.Task, case_ *Case, error error)

	OnExpectationStart(task workflows.Task, case_ *Case, expectation expectations.Expectation)
	OnExpectationEnd(task workflows.Task, case_ *Case, expectation expectations.Expectation, err error)
	OnExpectationError(task workflows.Task, case_ *Case, actual string, expectation expectations.Expectation, error error)
}

type SuiteResult struct {
	ExpectationsTotal  int
	ExpectationsOk     int
	ExpectationsErrors int

	CasesTotal  int
	CasesOk     int
	CasesErrors int
	CasesFatal  int
}

func NewSuiteResult() *SuiteResult {
	return &SuiteResult{
		ExpectationsTotal:  0,
		ExpectationsOk:     0,
		ExpectationsErrors: 0,

		CasesTotal:  0,
		CasesOk:     0,
		CasesErrors: 0,
		CasesFatal:  0,
	}
}

func NewSuite(ctx context.Context, llm ai.LLM, events SuiteEvents, cases []*Case, tasks []workflows.Task) *Suite {
	return &Suite{
		ctx:    ctx,
		llm:    llm,
		Cases:  cases,
		Tasks:  tasks,
		Usage:  ai.NewLLMUsage(0, 0, 0),
		events: events,
	}
}

func (s *Suite) Run() error {
	s.events.OnSuiteStart(s)
	for _, task := range s.Tasks {
		for _, q := range s.Cases {

			s.events.OnCaseStart(task, q)

			// Run the agent
			response, err := task.Invoke(s.ctx, s.llm, q.Input)

			if err != nil {
				s.events.OnCaseError(task, q, err)
				continue
			}

			s.Usage.Add(response.Usage)

			lastMessage, ok := response.Messages[len(response.Messages)-1].(*ai.TextMessage)
			if !ok {
				continue
			}

			q.Eval(s.ctx, s.events, task, lastMessage.Content)
		}
	}

	s.events.OnSuiteEnd(s)

	return nil
}

func (s *SuiteResult) Result(result *CaseResult) {
	s.CasesTotal += 1

	if result.Errors > 0 {
		s.CasesErrors += 1
	} else {
		s.CasesOk += 1
	}

	s.ExpectationsTotal += result.Total
	s.ExpectationsOk += result.Ok
	s.ExpectationsErrors += result.Errors
}

func (s *SuiteResult) FatalResult() {
	s.CasesFatal += 1
}
