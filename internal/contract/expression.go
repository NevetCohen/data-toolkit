package contract

// Expression is a recursively grouped Boolean expression. Exactly one shape
// is valid for each operator: comparison, n-ary arguments, or unary argument.
type Expression struct {
	Operator   ExpressionOperator `json:"operator" yaml:"operator"`
	Comparison *Comparison        `json:"comparison,omitempty" yaml:"comparison,omitempty"`
	Arguments  []Expression       `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Argument   *Expression        `json:"argument,omitempty" yaml:"argument,omitempty"`
	NotScope   NotScope           `json:"not_scope,omitempty" yaml:"not_scope,omitempty"`
}

type ExpressionOperator string

const (
	ExpressionCompare ExpressionOperator = "compare"
	ExpressionAND     ExpressionOperator = "and"
	ExpressionOR      ExpressionOperator = "or"
	ExpressionXOR     ExpressionOperator = "xor"
	ExpressionNOT     ExpressionOperator = "not"
)

type Comparison struct {
	Left     Operand            `json:"left" yaml:"left"`
	Operator ComparisonOperator `json:"operator" yaml:"operator"`
	Right    Operand            `json:"right" yaml:"right"`
}

type ComparisonOperator string

const (
	CompareEqual              ComparisonOperator = "eq"
	CompareNotEqual           ComparisonOperator = "ne"
	CompareLessThan           ComparisonOperator = "lt"
	CompareLessThanOrEqual    ComparisonOperator = "lte"
	CompareGreaterThan        ComparisonOperator = "gt"
	CompareGreaterThanOrEqual ComparisonOperator = "gte"
	CompareContains           ComparisonOperator = "contains"
	CompareMatches            ComparisonOperator = "matches"
)

// Operand contains exactly one column reference or one typed literal.
type Operand struct {
	SourceID string         `json:"source_id,omitempty" yaml:"source_id,omitempty"`
	Column   *ColumnRef     `json:"column,omitempty" yaml:"column,omitempty"`
	Literal  *ScalarLiteral `json:"literal,omitempty" yaml:"literal,omitempty"`
}

type ScalarLiteral struct {
	Kind  LiteralKind `json:"kind" yaml:"kind"`
	Value string      `json:"value,omitempty" yaml:"value,omitempty"`
}

type LiteralKind string

const (
	LiteralNull    LiteralKind = "null"
	LiteralString  LiteralKind = "string"
	LiteralBoolean LiteralKind = "boolean"
	LiteralInteger LiteralKind = "integer"
	LiteralDecimal LiteralKind = "decimal"
	LiteralDate    LiteralKind = "date"
	LiteralTime    LiteralKind = "time"
)

type NotScope string

const (
	NotScopeSource1 NotScope = "source_1"
	NotScopeSource2 NotScope = "source_2"
	NotScopeBoth    NotScope = "both"
)
