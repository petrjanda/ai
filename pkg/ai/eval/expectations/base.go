package expectations

import (
	"context"
	_ "embed"
)

type Expectation interface {
	Eval(ctx context.Context, actual string) error
	String() string
}
