package workflows

import (
	"context"
	"fmt"

	"github.com/getsynq/ai/pkg/ai"
)

type ChainTask struct {
	before Task
	after  Task

	lastOnly bool

	name string
}

func NewChainTask(before, after Task, lastOnly bool) *ChainTask {
	return &ChainTask{
		before:   before,
		after:    after,
		lastOnly: lastOnly,
	}
}

func (c *ChainTask) Name() string {
	if c.name != "" {
		return c.name
	}

	return c.before.Name() + " > " + c.after.Name()
}

func (c *ChainTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	before, err := c.before.Invoke(ctx, llm, history)
	if err != nil {
		return nil, err
	}

	if c.lastOnly {
		if last := before.Messages.Last(); last != nil {
			before.Messages = ai.NewHistory(last)
		} else {
			return nil, fmt.Errorf("last message is nil")
		}
	}

	after, err := c.after.Invoke(ctx, llm, before.Messages)
	if err != nil {
		return nil, err
	}

	// Add usage to get a total
	after.Usage.Add(before.Usage)

	return after, nil
}

func (c *ChainTask) Clone() Task {
	return &ChainTask{
		before: c.before.Clone(),
		after:  c.after.Clone(),
	}
}

func (c *ChainTask) WithName(name string) Task {
	c.name = name

	return c
}

func (c *ChainTask) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	return &ChainTask{
		before: c.before.WithRequestOpts(opts...),
		after:  c.after.WithRequestOpts(opts...),
	}
}

func (c *ChainTask) Then(task Task) Task {
	return NewChainTask(c, task, false)
}

func (c *ChainTask) Pipe(task Task) Task {
	return NewChainTask(c, task, true)
}
