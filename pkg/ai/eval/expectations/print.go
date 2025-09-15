package expectations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type PrintExpectation struct{}

func Print() *PrintExpectation {
	return &PrintExpectation{}
}

func (e *PrintExpectation) String() string {
	return "print"
}

func (e *PrintExpectation) Eval(ctx context.Context, actual string) error {
	if gjson.Valid(actual) {
		var obj interface{}
		if err := json.Unmarshal([]byte(actual), &obj); err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}

		formatted, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}

		fmt.Println(string(formatted))
	} else {
		fmt.Println(actual)
	}
	return nil
}
