package expect

import (
	"github.com/bukka/wst/instances/runtime"
)

type OutputAction struct {
	Order          OrderType
	Match          MatchType
	Type           OutputType
	Messages       []string
	RenderTemplate bool
}

type OrderType string

const (
	OrderTypeFixed  OrderType = "fixed"
	OrderTypeRandom OrderType = "random"
)

type MatchType string

const (
	MatchTypeExact  MatchType = "exact"
	MatchTypeRegexp MatchType = "regexp"
)

type OutputType string

const (
	OutputTypeStdout OutputType = "stdout"
	OutputTypeStderr OutputType = "stderr"
	OutputTypeAny    OutputType = "any"
)

func (a OutputAction) Execute(runData *runtime.Data) error {
	// implementation here
	// use runData.Store(key, value) to store data.
	// and value, ok := runData.Load(key) to retrieve data.
	return nil
}
