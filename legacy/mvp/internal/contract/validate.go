package contract

import "fmt"

// ValidateWorkflow applies semantic checks that cannot be expressed cleanly in
// JSON Schema and also protects programmatically constructed workflows.
func ValidateWorkflow(workflow Workflow) error {
	findings := make(ValidationErrors, 0)
	if workflow.SchemaVersion != CurrentSchemaVersion {
		addFinding(&findings, "$.schema_version", "unsupported schema version %q", workflow.SchemaVersion)
	}
	if workflow.ID == "" {
		addFinding(&findings, "$.id", "workflow id is required")
	}

	sourceIDs := make(map[string]struct{}, len(workflow.Sources))
	for index, source := range workflow.Sources {
		path := fmt.Sprintf("$.sources[%d]", index)
		if source.ID == "" {
			addFinding(&findings, path+".id", "source id is required")
		} else if _, exists := sourceIDs[source.ID]; exists {
			addFinding(&findings, path+".id", "duplicate source id %q", source.ID)
		} else {
			sourceIDs[source.ID] = struct{}{}
		}
		if !supportedFormat(source.Format) {
			addFinding(&findings, path+".format", "unsupported source format %q", source.Format)
		}
	}

	operationIDs := make(map[string]struct{}, len(workflow.Operations))
	for index := range workflow.Operations {
		operation := &workflow.Operations[index]
		path := fmt.Sprintf("$.operations[%d]", index)
		if operation.ID == "" {
			addFinding(&findings, path+".id", "operation id is required")
		} else if _, exists := operationIDs[operation.ID]; exists {
			addFinding(&findings, path+".id", "duplicate operation id %q", operation.ID)
		} else {
			operationIDs[operation.ID] = struct{}{}
		}
		validateOperation(operation, path, &findings)
	}

	if len(workflow.Outputs) == 0 && len(workflow.Queries) == 0 {
		addFinding(&findings, "$", "at least one output or query is required")
	}
	for index, output := range workflow.Outputs {
		path := fmt.Sprintf("$.outputs[%d]", index)
		if !supportedFormat(output.Format) {
			addFinding(&findings, path+".format", "unsupported output format %q", output.Format)
		}
		if (output.Location == "") == (output.LocationPolicy == nil) {
			addFinding(&findings, path, "declare exactly one of location or location_policy")
		}
		if len(output.Projection) == 0 {
			addFinding(&findings, path+".projection", "output projection is required")
		}
		if output.Overrides.Format != "" && !supportedFormat(output.Overrides.Format) {
			addFinding(&findings, path+".overrides.format", "unsupported output format %q", output.Overrides.Format)
		}
		if output.Overrides.Collision != "" && !supportedCollision(output.Overrides.Collision) {
			addFinding(&findings, path+".overrides.collision", "unsupported collision action %q", output.Overrides.Collision)
		}
	}
	for index, query := range workflow.Queries {
		path := fmt.Sprintf("$.queries[%d]", index)
		if !supportedQueryKind(query.Kind) {
			addFinding(&findings, path+".kind", "unsupported query kind %q", query.Kind)
		}
		if !supportedQueryShape(query.Shape) {
			addFinding(&findings, path+".shape", "unsupported query shape %q", query.Shape)
		} else if !queryShapeMatchesKind(query.Kind, query.Shape) {
			addFinding(&findings, path+".shape", "query kind %q requires shape %q", query.Kind, requiredQueryShape(query.Kind))
		}
	}
	if !supportedExceptionAction(workflow.Exception.Action) {
		addFinding(&findings, "$.exception.action", "unsupported exception action %q", workflow.Exception.Action)
	}

	if len(findings) > 0 {
		return findings
	}
	return nil
}

func validateOperation(operation *Operation, path string, findings *ValidationErrors) {
	if len(operation.Inputs) == 0 {
		addFinding(findings, path+".inputs", "operation requires at least one input")
	}

	switch operation.Kind {
	case OperationFilter, OperationDeleteRows:
		if operation.Condition == nil {
			addFinding(findings, path+".condition", "%s operation requires a condition", operation.Kind)
		}
	case OperationRename:
		if len(operation.Rename) == 0 {
			addFinding(findings, path+".rename", "rename operation requires at least one field")
		}
	case OperationNormalize:
		if len(operation.Columns) == 0 {
			addFinding(findings, path+".columns", "normalize operation requires at least one column")
		}
		if operation.Normalize == nil {
			addFinding(findings, path+".normalize", "normalize operation requires normalize options")
		}
	case OperationTrim:
		if len(operation.Columns) == 0 {
			addFinding(findings, path+".columns", "trim operation requires at least one column")
		}
	case OperationDeduplicate:
		if operation.Deduplicate == nil {
			addFinding(findings, path+".deduplicate", "deduplicate operation requires deduplicate options")
		} else if operation.Deduplicate.Keep != "" && !supportedKeepPolicy(operation.Deduplicate.Keep) {
			addFinding(findings, path+".deduplicate.keep", "unsupported keep policy %q", operation.Deduplicate.Keep)
		}
	case OperationDeleteEmptyRows:
	case OperationSort:
		if len(operation.Sort) == 0 {
			addFinding(findings, path+".sort", "sort operation requires at least one sort field")
		}
	case OperationSplit:
		if operation.Split == nil || len(operation.Split.Branches) == 0 {
			addFinding(findings, path+".split", "split operation requires at least one branch")
		}
	case OperationConcatenate, OperationTransformText:
		if operation.Text == nil {
			addFinding(findings, path+".text", "%s operation requires text options", operation.Kind)
		}
	case OperationRegexTransform:
		if operation.Regex == nil {
			addFinding(findings, path+".regex", "regex_transform operation requires regex options")
		}
	default:
		addFinding(findings, path+".kind", "unsupported operation kind %q", operation.Kind)
	}

	if operation.Condition != nil {
		validateExpression(*operation.Condition, path+".condition", findings)
	}
	if operation.Split != nil {
		for index, branch := range operation.Split.Branches {
			if branch.Condition != nil {
				validateExpression(*branch.Condition, fmt.Sprintf("%s.split.branches[%d].condition", path, index), findings)
			}
		}
	}
}

func validateExpression(expression Expression, path string, findings *ValidationErrors) {
	switch expression.Operator {
	case ExpressionCompare:
		if expression.Comparison == nil {
			addFinding(findings, path+".comparison", "compare expression requires comparison")
		}
	case ExpressionAND, ExpressionOR, ExpressionXOR:
		if len(expression.Arguments) < 2 {
			addFinding(findings, path+".arguments", "%s expression requires at least two grouped arguments", expression.Operator)
		}
		for index, argument := range expression.Arguments {
			validateExpression(argument, fmt.Sprintf("%s.arguments[%d]", path, index), findings)
		}
	case ExpressionNOT:
		if expression.Argument == nil {
			addFinding(findings, path+".argument", "not expression requires exactly one argument")
		} else {
			validateExpression(*expression.Argument, path+".argument", findings)
			if len(expressionSources(*expression.Argument)) > 1 && expression.NotScope == "" {
				addFinding(findings, path+".not_scope", "cross-source NOT requires an explicit complement scope")
			}
		}
		if expression.NotScope != "" && !supportedNotScope(expression.NotScope) {
			addFinding(findings, path+".not_scope", "unsupported NOT scope %q", expression.NotScope)
		}
	default:
		addFinding(findings, path+".operator", "unsupported expression operator %q", expression.Operator)
	}
}

func addFinding(findings *ValidationErrors, path, message string, arguments ...any) {
	*findings = append(*findings, PathError{Path: path, Message: fmt.Sprintf(message, arguments...)})
}

func supportedFormat(value DataFormat) bool {
	switch value {
	case FormatJSON, FormatCSV, FormatExcel, FormatTXT, FormatGoogleSheets:
		return true
	default:
		return false
	}
}

func supportedKeepPolicy(value KeepPolicy) bool {
	switch value {
	case KeepFirst, KeepLast, KeepError:
		return true
	default:
		return false
	}
}

func supportedCollision(value CollisionAction) bool {
	switch value {
	case CollisionBlock, CollisionOverwrite, CollisionAlternateName:
		return true
	default:
		return false
	}
}

func supportedNotScope(value NotScope) bool {
	switch value {
	case NotScopeSource1, NotScopeSource2, NotScopeBoth:
		return true
	default:
		return false
	}
}

func supportedQueryKind(value QueryKind) bool {
	switch value {
	case QueryRows, QueryCount, QueryDistinctValues:
		return true
	default:
		return false
	}
}

func supportedQueryShape(value QueryShape) bool {
	switch value {
	case QueryShapeRows, QueryShapeScalar, QueryShapeValues:
		return true
	default:
		return false
	}
}

func queryShapeMatchesKind(kind QueryKind, shape QueryShape) bool {
	return requiredQueryShape(kind) == shape
}

func requiredQueryShape(kind QueryKind) QueryShape {
	switch kind {
	case QueryRows:
		return QueryShapeRows
	case QueryCount:
		return QueryShapeScalar
	case QueryDistinctValues:
		return QueryShapeValues
	default:
		return ""
	}
}

func supportedExceptionAction(value ExceptionAction) bool {
	switch value {
	case ExceptionBlock, ExceptionReport, ExceptionIgnore:
		return true
	default:
		return false
	}
}
