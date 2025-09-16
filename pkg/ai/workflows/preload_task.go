package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

type PreloadTaskCallback func(ctx context.Context) (ai.History, error)

func NewPreloadTask(name string, callback PreloadTaskCallback) Task {
	return NewLazyTask(name, func(ctx context.Context, _ ai.LLM, _ ai.History) (*ai.LLMResponse, error) {
		history, err := callback(ctx)
		if err != nil {
			return nil, err
		}
		return ai.NewLLMResponse(history...), nil
	})
}
