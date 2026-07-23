package contract

import "testing"

func TestNormalizeCrossSourceNOTDefaultsToBoth(t *testing.T) {
	workflow := crossSourceNOTWorkflow(NotScope(""))

	NormalizeWorkflow(&workflow)

	condition := workflow.Operations[0].Condition
	if condition.NotScope != NotScopeBoth {
		t.Fatalf("not_scope = %q, want %q", condition.NotScope, NotScopeBoth)
	}
	if err := ValidateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeCrossSourceNOTPreservesExplicitScope(t *testing.T) {
	workflow := crossSourceNOTWorkflow(NotScopeSource2)

	NormalizeWorkflow(&workflow)

	if got := workflow.Operations[0].Condition.NotScope; got != NotScopeSource2 {
		t.Fatalf("not_scope = %q, want %q", got, NotScopeSource2)
	}
}

func TestNormalizeSingleSourceNOTDoesNotInventComplementUniverse(t *testing.T) {
	workflow := validWorkflowForTest()
	comparison := workflow.Operations[0].Condition
	workflow.Operations[0].Condition = &Expression{Operator: ExpressionNOT, Argument: comparison}

	NormalizeWorkflow(&workflow)

	if got := workflow.Operations[0].Condition.NotScope; got != "" {
		t.Fatalf("not_scope = %q, want empty for a single-source expression", got)
	}
}

func crossSourceNOTWorkflow(scope NotScope) Workflow {
	workflow := validWorkflowForTest()
	workflow.Sources = append(workflow.Sources, Source{ID: "source-2", Format: FormatCSV, Location: "input-2.csv"})
	comparison := Expression{
		Operator: ExpressionCompare,
		Comparison: &Comparison{
			Left:     Operand{SourceID: "source", Column: &ColumnRef{Header: "A"}},
			Operator: CompareEqual,
			Right:    Operand{SourceID: "source-2", Column: &ColumnRef{Header: "B"}},
		},
	}
	workflow.Operations[0].Inputs = []string{"source", "source-2"}
	workflow.Operations[0].Condition = &Expression{
		Operator: ExpressionNOT,
		Argument: &Expression{
			Operator:  ExpressionOR,
			Arguments: []Expression{comparison, comparison},
		},
		NotScope: scope,
	}
	return workflow
}
