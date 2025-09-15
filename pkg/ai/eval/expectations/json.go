package expectations

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"

	_ "embed"
)

type JSONExpectation struct {
	Path     string
	Expected Assertion
}

type Assertion interface {
	Validate(value any) error
}

func JSONPath(path string) *JSONExpectation {
	return &JSONExpectation{
		Path: path,
	}
}

func (e *JSONExpectation) Eval(_ context.Context, actual string) error {
	if e.Expected == nil {
		return fmt.Errorf("no assertion given for path %s", e.Path)
	}

	value := gjson.Get(actual, e.Path)
	err := e.Expected.Validate(value.Value())

	if err != nil {
		return fmt.Errorf("path %s %v", e.Path, err)
	}

	return nil
}

func (e *JSONExpectation) String() string {
	return fmt.Sprintf("%s %v", e.Path, e.Expected)
}

func (e *JSONExpectation) Eq(right any) *JSONExpectation {
	e.Expected = &Eq_{Right: right}
	return e
}

func (e *JSONExpectation) Contains(right any) *JSONExpectation {
	e.Expected = &Contains_{Right: right}
	return e
}

func (e *JSONExpectation) AnyContains(right any) *JSONExpectation {
	e.Expected = &AnyContains_{Right: right}
	return e
}

func (e *JSONExpectation) Exists() *JSONExpectation {
	e.Expected = &Exists_{}
	return e
}

func (e *JSONExpectation) DoesNotExist() *JSONExpectation {
	e.Expected = &DoesNotExist_{}
	return e
}

func (e *JSONExpectation) Print() *JSONExpectation {
	e.Expected = &Print_{}
	return e
}

func (e *JSONExpectation) JSONPath(path string) *JSONExpectation {
	return &JSONExpectation{Path: path}
}

// EQUALS

type Eq_ struct {
	Right any
}

func (e *Eq_) Validate(left any) error {
	if left == e.Right {
		return nil
	}

	return fmt.Errorf("expected '%v', got '%v'", e.Right, left)
}

func (e *Eq_) String() string {
	return fmt.Sprintf("equals %v", e.Right)
}

// PRINT

type Print_ struct {
}

func (e *Print_) Validate(value any) error {
	log.Println(value)

	return nil
}

func (e *Print_) String() string {
	return "print"
}

// CONTAINS

type Contains_ struct {
	Right any
}

func (e *Contains_) Validate(left any) error {
	log.Printf("%+v of %T", left, left)
	switch l := left.(type) {
	case []interface{}:
		for _, item := range l {
			if item == e.Right {
				return nil
			}
		}

		return fmt.Errorf("list does not contain '%v'", e.Right)

	case string:
		if !strings.Contains(l, fmt.Sprintf("%v", e.Right)) {
			return fmt.Errorf("string does not contain '%v'", e.Right)
		}

		return nil

	default:
		return fmt.Errorf("contains not supported for type %T", left)
	}
}

func (e *Contains_) String() string {
	return fmt.Sprintf("contains %v", e.Right)
}

type AnyContains_ struct {
	Right any
}

func (e *AnyContains_) Validate(left any) error {
	switch l := left.(type) {
	case []interface{}:
		for _, item := range l {
			switch item := item.(type) {
			case string:
				if strings.Contains(item, fmt.Sprintf("%v", e.Right)) {
					return nil
				}
			default:
				return fmt.Errorf("any contains not supported for type []%T", item)
			}
		}

		return fmt.Errorf("list does not contain '%v'", e.Right)

	default:
		return fmt.Errorf("any contains not supported for type %T", left)
	}
}

func (e *AnyContains_) String() string {
	return fmt.Sprintf("any contains %v", e.Right)
}

// EXISTS

type Exists_ struct {
}

func (e *Exists_) Validate(value any) error {
	if value == nil {
		return fmt.Errorf("expected value, got nil")
	}

	return nil
}

func (e *Exists_) String() string {
	return "exists"
}

var Exists = &Exists_{}

var DoesNotExist = &DoesNotExist_{}

type DoesNotExist_ struct {
}

func (e *DoesNotExist_) Validate(value any) error {
	if value == nil {
		return nil
	}

	return fmt.Errorf("expected nil value, got '%v'", value)
}

func (e *DoesNotExist_) String() string {
	return "does not exist"
}
