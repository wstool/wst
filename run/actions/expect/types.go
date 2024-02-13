package expect

type MatchType string

const (
	MatchTypeExact  MatchType = "exact"
	MatchTypeRegexp MatchType = "regexp"
)

type OrderType string

const (
	OrderTypeFixed  OrderType = "fixed"
	OrderTypeRandom OrderType = "random"
)

type OutputType string

const (
	OutputTypeStdout OutputType = "stdout"
	OutputTypeStderr OutputType = "stderr"
	OutputTypeAny    OutputType = "any"
)
