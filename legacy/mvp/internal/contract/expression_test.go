package contract

import (
	"errors"
	"testing"
)

func TestExpressionTreePreservesExplicitGrouping(t *testing.T) {
	workflow := validWorkflowForTest()
	comparison := *workflow.Operations[0].Condition
	workflow.Operations[0].Condition = &Expression{
		Operator: ExpressionAND,
		Arguments: []Expression{
			comparison,
			{
				Operator: ExpressionOR,
				Arguments: []Expression{
					comparison,
					comparison,
				},
			},
		},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
	condition := workflow.Operations[0].Condition
	if condition.Operator != ExpressionAND || condition.Arguments[1].Operator != ExpressionOR {
		t.Fatalf("grouping changed: %#v", condition)
	}
}

func TestExpressionTreeRejectsImplicitOrIncompleteGrouping(t *testing.T) {
	workflow := validWorkflowForTest()
	workflow.Operations[0].Condition = &Expression{
		Operator:  ExpressionXOR,
		Arguments: []Expression{*workflow.Operations[0].Condition},
	}

	err := ValidateWorkflow(workflow)
	var findings ValidationErrors
	if !errors.As(err, &findings) {
		t.Fatalf("error type = %T, want ValidationErrors: %v", err, err)
	}
	if findings[0].Path != "$.operations[0].condition.arguments" {
		t.Fatalf("path = %q, want grouped arguments path", findings[0].Path)
	}
}
