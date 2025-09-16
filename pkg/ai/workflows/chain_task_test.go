package workflows

import (
	"context"
	"testing"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChainTaskTestSuite struct {
	suite.Suite
	*require.Assertions
}

func TestChainTaskSuite(t *testing.T) {
	suite.Run(t, new(ChainTaskTestSuite))
}

func (s *ChainTaskTestSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *ChainTaskTestSuite) TestConcat() {
	ctx := context.Background()
	mockLLM := &MockLLM{}

	task1 := NewPreloadTask("one", func(ctx context.Context) (ai.History, error) {
		return ai.History{ai.NewSystemMessage("One")}, nil
	})

	task2 := NewPreloadTask("two", func(ctx context.Context) (ai.History, error) {
		return ai.History{ai.NewSystemMessage("Two")}, nil
	})

	task3 := NewPreloadTask("three", func(ctx context.Context) (ai.History, error) {
		return ai.History{ai.NewSystemMessage("Three")}, nil
	})

	s.Run("Then", func() {
		result := task1.Then(task2)
		res, err := result.Invoke(ctx, mockLLM, ai.History{
			ai.NewSystemMessage("Input"),
		})

		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(3, len(res.Messages))
		s.Require().Equal(ai.NewSystemMessage("Input"), res.Messages[0])
		s.Equal(ai.NewSystemMessage("One"), res.Messages[1])
		s.Equal(ai.NewSystemMessage("Two"), res.Messages[2])
	})

	s.Run("Pipe", func() {
		result := task1.Pipe(task2).Pipe(task3)
		res, err := result.Invoke(ctx, mockLLM, ai.History{})

		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(2, len(res.Messages))
		s.Require().Equal(ai.NewSystemMessage("Two"), res.Messages[0])
		s.Require().Equal(ai.NewSystemMessage("Three"), res.Messages[1])
	})
}
