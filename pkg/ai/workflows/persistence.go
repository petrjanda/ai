package workflows

import (
	"context"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
)

type StorageProvider interface {
	Storage(ctx context.Context, workflowId string) Storage
}

type Storage interface {
	Store(ctx context.Context, itemId string, data any) error
	Load(ctx context.Context, itemId string) (any, error)
}

type MemoryStorageProvider struct {
	state map[string]*MemoryStorage
}

func NewMemoryStorageProvider() *MemoryStorageProvider {
	return &MemoryStorageProvider{state: make(map[string]*MemoryStorage)}
}

func (m *MemoryStorageProvider) Storage(id string) Storage {
	if _, ok := m.state[id]; !ok {
		m.state[id] = NewMemoryStorage(id)
	}

	return m.state[id]
}

type MemoryStorage struct {
	id    string
	state map[string]any
}

func NewMemoryStorage(id string) *MemoryStorage {
	return &MemoryStorage{id: id, state: make(map[string]any)}
}

func (m *MemoryStorage) Store(ctx context.Context, id string, data any) error {
	m.state[id] = data
	return nil
}

func (m *MemoryStorage) Load(ctx context.Context, id string) (any, error) {
	return m.state[id], nil
}

type storageKey struct{}

func WithStorage(ctx context.Context, s Storage) context.Context {
	return context.WithValue(ctx, storageKey{}, s)
}

func StorageFrom(ctx context.Context) (Storage, bool) {
	s, ok := ctx.Value(storageKey{}).(Storage)
	if !ok {
		return nil, false
	}

	return s, true
}

//
// TASK PERSISTENCE
//

func saveTask(ctx context.Context, id string, response *ai.LLMResponse) (*ai.LLMResponse, error) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return response, nil
	}

	return response, storage.Store(ctx, id, response)
}

func loadTask(ctx context.Context, id string) (*ai.LLMResponse, bool) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return nil, false
	}

	response, err := storage.Load(ctx, id)
	if err != nil {
		return nil, false
	}

	if response == nil {
		return nil, false
	}

	return response.(*ai.LLMResponse), true
}

//
// AGENT TASK PERSISTENCE
//

type AgentTaskState struct {
	Response *ai.LLMResponse
	Terminal bool
}

func NewAgentTaskState(response *ai.LLMResponse, terminal bool) *AgentTaskState {
	return &AgentTaskState{
		Response: response,
		Terminal: terminal,
	}
}

func saveAgentTask(ctx context.Context, id string, response *ai.LLMResponse, terminal bool) (*ai.LLMResponse, error) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return response, nil
	}

	state := NewAgentTaskState(response, terminal)
	return response, storage.Store(ctx, id, state)
}

func loadAgentTask(ctx context.Context, id string) (*AgentTaskState, bool) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return nil, false
	}

	response, err := storage.Load(ctx, id)
	if err != nil {
		return nil, false
	}

	if response == nil {
		return nil, false
	}

	state, ok := response.(*AgentTaskState)
	if !ok {
		return nil, false
	}

	return state, true
}

// Agent async storage hook

type AgentStorageHook struct {
	*ai.NoopAgentEvents
	id string
}

func NewAgentStorageHook(id string) *AgentStorageHook {
	return &AgentStorageHook{NoopAgentEvents: ai.NewNoopAgentEvents(), id: id}
}

func (h *AgentStorageHook) OnResponse(ctx context.Context, request *ai.LLMRequest, response *ai.LLMResponse, terminal bool) {
	saved := ai.
		NewLLMResponse(append(request.History, response.Messages...)...). // store entire history of agent state (current request + current response)
		SetUsage(response.Usage)                                          // remember past usage

	saveAgentTask(ctx, h.id, saved, terminal)
}

//
// WORK PERSISTENCE
//

func loadWork[T any](ctx context.Context, id string) (*T, bool) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return nil, false
	}

	response, err := storage.Load(ctx, id)
	if err != nil {
		return nil, false
	}

	if response == nil {
		return nil, false
	}

	return response.(*T), true
}

func saveWork[T any](ctx context.Context, id string, response *T) (*T, error) {
	storage, ok := StorageFrom(ctx)
	if !ok {
		return response, nil
	}

	return response, storage.Store(ctx, id, response)
}
