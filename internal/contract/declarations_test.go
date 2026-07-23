package contract

import "testing"

func TestOutputsResolveIndependentProjectionsAndOverrides(t *testing.T) {
	workflow := validWorkflowForTest()
	workflow.Outputs = append(workflow.Outputs, Output{
		ID:       "out-2",
		Input:    "filter",
		Format:   FormatCSV,
		Location: "out-2.csv",
		Projection: []ProjectionField{
			{Column: ColumnRef{Header: "A"}, OutputAs: "Renamed A"},
		},
		Overrides: OutputOverrides{
			Filename:  "second.csv",
			Delimiter: ";",
			Collision: CollisionAlternateName,
		},
	})
	workflow.Outputs[0].Overrides = OutputOverrides{Collision: CollisionBlock}

	if err := ValidateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
	if workflow.Outputs[0].Overrides.Collision == workflow.Outputs[1].Overrides.Collision {
		t.Fatal("per-output collision overrides were not independent")
	}
}

func TestBothScopeDeclarationCarriesBranchProjection(t *testing.T) {
	workflow := crossSourceNOTWorkflow(NotScopeBoth)
	workflow.Outputs[0].BranchProjection = &BranchProjection{
		Source1: []ProjectionField{{Column: ColumnRef{Header: "A"}, OutputAs: "value"}},
		Source2: []ProjectionField{{Column: ColumnRef{Header: "B"}, OutputAs: "value"}},
	}

	if err := ValidateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
}

func TestQueryDeclarationRequiresDeterministicShape(t *testing.T) {
	workflow := validWorkflowForTest()
	workflow.Outputs = nil
	workflow.Queries = []Query{{ID: "count", Input: "filter", Kind: QueryCount, Shape: QueryShapeRows}}

	if err := ValidateWorkflow(workflow); err == nil {
		t.Fatal("expected query kind and shape mismatch")
	}
	workflow.Queries[0].Shape = QueryShapeScalar
	if err := ValidateWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
}
