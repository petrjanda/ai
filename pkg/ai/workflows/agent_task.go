package workflows

import (
	"context"

	"github.com/getsynq/ai/pkg/ai"
	"github.com/getsynq/ai/pkg/ai/agent"
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
	agent := agent.NewAgent(llm, t.AgentOpts...)
	request := t.Request.Clone(ai.WithAddedHistory(history))

	return agent.Invoke(ctx, request)
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
