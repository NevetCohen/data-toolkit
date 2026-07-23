package contract

// NormalizeWorkflow makes interface defaults explicit before semantic
// validation and planning. It mutates the supplied workflow in place.
func NormalizeWorkflow(workflow *Workflow) {
	defaultNotScope := workflow.Overrides.NotScope
	if defaultNotScope == "" {
		defaultNotScope = NotScopeBoth
	}

	for index := range workflow.Operations {
		operation := &workflow.Operations[index]
		if operation.Condition != nil {
			normalizeExpression(operation.Condition, defaultNotScope)
		}
		if operation.Split != nil {
			for branchIndex := range operation.Split.Branches {
				condition := operation.Split.Branches[branchIndex].Condition
				if condition != nil {
					normalizeExpression(condition, defaultNotScope)
				}
			}
		}
	}
}

func normalizeExpression(expression *Expression, defaultNotScope NotScope) {
	for index := range expression.Arguments {
		normalizeExpression(&expression.Arguments[index], defaultNotScope)
	}
	if expression.Argument != nil {
		normalizeExpression(expression.Argument, defaultNotScope)
	}

	if expression.Operator == ExpressionNOT && expression.NotScope == "" && expression.Argument != nil {
		if len(expressionSources(*expression.Argument)) > 1 {
			expression.NotScope = defaultNotScope
		}
	}
}

func expressionSources(expression Expression) map[string]struct{} {
	sources := make(map[string]struct{})
	collectExpressionSources(expression, sources)
	return sources
}

func collectExpressionSources(expression Expression, sources map[string]struct{}) {
	if expression.Comparison != nil {
		for _, operand := range []Operand{expression.Comparison.Left, expression.Comparison.Right} {
			if operand.Column != nil && operand.SourceID != "" {
				sources[operand.SourceID] = struct{}{}
			}
		}
	}
	for _, argument := range expression.Arguments {
		collectExpressionSources(argument, sources)
	}
	if expression.Argument != nil {
		collectExpressionSources(*expression.Argument, sources)
	}
}
