package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/agent"
)

// AgentTask is a task that uses an agent to invoke the LLM.
// Agents have tools that they can call to perform the job.
// Agent will continue execution in the loop, calling tools
// until it completes the task.

type AgentTask struct {
	Name_     string
	Request   *ai.LLMRequest
	AgentOpts []agent.AgentOpts
}

func NewAgentTask(name string, request *ai.LLMRequest, agentOpts ...agent.AgentOpts) Task {
	return &AgentTask{
		Name_:     name,
		Request:   request,
		AgentOpts: agentOpts,
	}
}

func (t *AgentTask) Name() string {
	return t.Name_
}

func (t *AgentTask) Clone() Task {
	return NewAgentTask(t.Name_, t.Request.Clone(), t.AgentOpts...)
}

func (t *AgentTask) Invoke(ctx context.Context, llm ai.LLM, history ai.History) (*ai.LLMResponse, error) {
	// Add hook to save the agent progress
	hook := NewAgentStorageHook(t.Name_)
	t.AgentOpts = append(t.AgentOpts, agent.WithEvents(hook))

	request := t.Request.Clone(ai.WithAddedHistory(history))

	if response, ok := loadAgentTask(ctx, t.Name_); ok {
		if response.Terminal {
			return response.Response, nil
		} else {
			// Initiate agent with the stored history
			request = t.Request.Clone(ai.WithHistory(response.Response.Messages))
		}
	}

	agent := agent.NewAgent(llm, t.AgentOpts...)
	response, err := agent.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (t *AgentTask) WithRequestOpts(opts ...ai.LLMRequestOpts) Task {
	new := t.Clone().(*AgentTask)
	for _, opt := range opts {
		opt(new.Request)
	}

	return new
}

func (t *AgentTask) WithName(name string) Task {
	new := t.Clone().(*AgentTask)
	new.Name_ = name
	return new
}

func (t *AgentTask) Then(task Task) Task {
	return NewChainTask(t, task, false)
}

func (t *AgentTask) Pipe(task Task) Task {
	return NewChainTask(t, task, true)
}
