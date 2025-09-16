package expectations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
)

type JudgeExpectation struct {
	Name        string
	Instruction string
	Model       ai.ModelId
	llm         ai.LLM

	judgement func(ctx context.Context, response *ai.TextMessage) error
}

func NewJudge[T any](name string, parent ai.LLM, model ai.ModelId, instruction string, judgement func(ctx context.Context, response *ai.TextMessage) error) *JudgeExpectation {
	directiveSchema := openai.NewOpenAISchemaGenerator().MustGenerate(new(T))

	return &JudgeExpectation{
		Name:        name,
		Instruction: instruction,
		Model:       model,
		llm: structured.NewLLM(
			directiveSchema, parent,
		),
		judgement: judgement,
	}
}

func (e *JudgeExpectation) Eval(ctx context.Context, actual string) error {
	req := ai.NewLLMRequest(
		ai.WithModel(e.Model),
		ai.WithSystem(e.Instruction),
		ai.WithHistory(ai.NewHistory(ai.NewUserMessage(actual))),
		ai.WithTemperature(0.0),
		ai.WithMaxCompletionTokens(1000),
	)

	response, err := e.llm.Invoke(ctx, req)
	if err != nil {
		return err
	}

	if len(response.Messages) == 0 {
		return fmt.Errorf("no response from judge")
	}

	lastMessage, ok := response.Messages[len(response.Messages)-1].(*ai.TextMessage)
	if !ok {
		return fmt.Errorf("no text message in response from judge")
	}

	if e.judgement == nil {
		return fmt.Errorf("no judgement function given for judge %s", e.Name)
	}

	return e.judgement(ctx, lastMessage)
}

func (e *JudgeExpectation) String() string {
	return e.Name
}

type ScoringJudgeExpectation struct {
	judge *JudgeExpectation
}

// SIMPLE SCORING

type ScoringJudgeVerdict struct {
	// Score that indicates to what degree you agree with the statement. 0 means you disagree completely, 100 means you agree completely.
	Score int `json:"score" jsonschema:"required,minimum=0,maximum=100"`

	// Reason for the score.
	Reason string `json:"reason" jsonschema:"required"`
}

func NewScoringJudge(name string, parent ai.LLM, model ai.ModelId, instruction string, threshold int) *ScoringJudgeExpectation {
	judgement := func(ctx context.Context, response *ai.TextMessage) error {
		var verdict ScoringJudgeVerdict
		if err := json.Unmarshal([]byte(response.Content), &verdict); err != nil {
			return err
		}

		log.Println(verdict)

		if verdict.Score < 0 || verdict.Score > 100 {
			return fmt.Errorf("score must be between 0 and 100")
		}

		if verdict.Reason == "" {
			return fmt.Errorf("reason is required")
		}

		if verdict.Score < threshold {
			return fmt.Errorf("score must be greater than 50")
		}

		return nil
	}

	return &ScoringJudgeExpectation{judge: NewJudge[ScoringJudgeVerdict](name, parent, model, instruction, judgement)}
}

func (e *ScoringJudgeExpectation) Eval(ctx context.Context, actual string) error {
	return e.judge.Eval(ctx, actual)
}

func (e *ScoringJudgeExpectation) String() string {
	return e.judge.String()
}
