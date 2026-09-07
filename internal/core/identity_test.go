package core

import "testing"

func TestIdentityValidation(t *testing.T) {
	const validDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	cases := []struct {
		name      string
		wantValid bool
		validate  func() error
	}{
		{name: "schema version v1", wantValid: true, validate: func() error { return V1.Validate() }},
		{name: "schema version unknown", validate: func() error { return SchemaVersion("v2").Validate() }},
		{name: "request ID UUID", wantValid: true, validate: func() error { return RequestID("123e4567-e89b-12d3-a456-426614174000").Validate() }},
		{name: "request ID malformed", validate: func() error { return RequestID("not-a-uuid").Validate() }},
		{name: "run ID UUID", wantValid: true, validate: func() error { return RunID("123e4567-e89b-12d3-a456-426614174000").Validate() }},
		{name: "run ID malformed", validate: func() error { return RunID("not-a-uuid").Validate() }},
		{name: "receipt ID digest", wantValid: true, validate: func() error { return ReceiptID(validDigest).Validate() }},
		{name: "receipt ID uppercase digest", validate: func() error {
			return ReceiptID("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").Validate()
		}},
		{name: "data type ID dotted", wantValid: true, validate: func() error { return DataTypeID("custom.type-v1").Validate() }},
		{name: "data type ID uppercase", validate: func() error { return DataTypeID("Custom.Type").Validate() }},
		{name: "operation ID dotted", wantValid: true, validate: func() error { return OperationID("row.filter").Validate() }},
		{name: "operation ID empty segment", validate: func() error { return OperationID("row..filter").Validate() }},
		{name: "source ID stable", wantValid: true, validate: func() error { return SourceID("source-1").Validate() }},
		{name: "source ID empty", validate: func() error { return SourceID(" ").Validate() }},
		{name: "sheet ID stable", wantValid: true, validate: func() error { return SheetID("sheet-1").Validate() }},
		{name: "sheet ID empty", validate: func() error { return SheetID(" ").Validate() }},
		{name: "table ID stable", wantValid: true, validate: func() error { return TableID("table-1").Validate() }},
		{name: "table ID empty", validate: func() error { return TableID(" ").Validate() }},
		{name: "column ID stable", wantValid: true, validate: func() error { return ColumnID("column-1").Validate() }},
		{name: "column ID empty", validate: func() error { return ColumnID(" ").Validate() }},
		{name: "row ID stable", wantValid: true, validate: func() error { return RowID("row-1").Validate() }},
		{name: "row ID empty", validate: func() error { return RowID(" ").Validate() }},
		{name: "SHA-256 digest", wantValid: true, validate: func() error { return SHA256Digest(validDigest).Validate() }},
		{name: "SHA-256 digest short", validate: func() error { return SHA256Digest("abc").Validate() }},
		{name: "format CSV", wantValid: true, validate: func() error { return CSVFormat.Validate() }},
		{name: "format JSON", wantValid: true, validate: func() error { return JSONFormat.Validate() }},
		{name: "format Excel", wantValid: true, validate: func() error { return ExcelFormat.Validate() }},
		{name: "format Google Sheets bridge", validate: func() error { return FormatID("google_sheets").Validate() }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.validate()
			if testCase.wantValid && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if !testCase.wantValid && err == nil {
				t.Fatal("Validate() error = nil, want rejection")
			}
		})
	}
}

func TestEnumValidation(t *testing.T) {
	cases := []struct {
		name      string
		wantValid bool
		validate  func() error
	}{
		{name: "type authority declared", wantValid: true, validate: func() error { return TypeAuthorityDeclared.Validate() }},
		{name: "type authority inferred", wantValid: true, validate: func() error { return TypeAuthorityInferred.Validate() }},
		{name: "type authority unknown", validate: func() error { return TypeAuthority("unknown").Validate() }},
		{name: "finding severity info", wantValid: true, validate: func() error { return FindingSeverityInfo.Validate() }},
		{name: "finding severity warning", wantValid: true, validate: func() error { return FindingSeverityWarning.Validate() }},
		{name: "finding severity error", wantValid: true, validate: func() error { return FindingSeverityError.Validate() }},
		{name: "finding severity unknown", validate: func() error { return FindingSeverity("unknown").Validate() }},
		{name: "terminal status succeeded", wantValid: true, validate: func() error { return TerminalStatusSucceeded.Validate() }},
		{name: "terminal status failed", wantValid: true, validate: func() error { return TerminalStatusFailed.Validate() }},
		{name: "terminal status cancelled", wantValid: true, validate: func() error { return TerminalStatusCancelled.Validate() }},
		{name: "terminal status blocked", wantValid: true, validate: func() error { return TerminalStatusBlocked.Validate() }},
		{name: "terminal status unknown", validate: func() error { return TerminalStatus("unknown").Validate() }},
		{name: "operation owner reader engine", wantValid: true, validate: func() error { return OperationOwnerReaderEngine.Validate() }},
		{name: "operation owner logical engine", wantValid: true, validate: func() error { return OperationOwnerLogicalEngine.Validate() }},
		{name: "operation owner writer engine", wantValid: true, validate: func() error { return OperationOwnerWriterEngine.Validate() }},
		{name: "operation owner unknown", validate: func() error { return OperationOwner("unknown").Validate() }},
		{name: "operation exposure API", wantValid: true, validate: func() error { return OperationExposureAPI.Validate() }},
		{name: "operation exposure workflow", wantValid: true, validate: func() error { return OperationExposureWorkflow.Validate() }},
		{name: "operation exposure lifecycle", wantValid: true, validate: func() error { return OperationExposureLifecycle.Validate() }},
		{name: "operation exposure unknown", validate: func() error { return OperationExposure("unknown").Validate() }},
		{name: "streaming mode streaming", wantValid: true, validate: func() error { return StreamingModeStreaming.Validate() }},
		{name: "streaming mode bounded spool", wantValid: true, validate: func() error { return StreamingModeBoundedSpool.Validate() }},
		{name: "streaming mode unknown", validate: func() error { return StreamingMode("unknown").Validate() }},
		{name: "collision policy block", wantValid: true, validate: func() error { return CollisionPolicyBlock.Validate() }},
		{name: "collision policy overwrite", wantValid: true, validate: func() error { return CollisionPolicyOverwrite.Validate() }},
		{name: "collision policy alternate name", wantValid: true, validate: func() error { return CollisionPolicyAlternateName.Validate() }},
		{name: "collision policy unknown", validate: func() error { return CollisionPolicy("unknown").Validate() }},
		{name: "remote collision policy block", wantValid: true, validate: func() error { return RemoteCollisionPolicyBlock.Validate() }},
		{name: "remote collision policy alternate name", wantValid: true, validate: func() error { return RemoteCollisionPolicyAlternateName.Validate() }},
		{name: "remote collision policy overwrite", validate: func() error { return RemoteCollisionPolicy("overwrite").Validate() }},
		{name: "exception action block", wantValid: true, validate: func() error { return ExceptionActionBlock.Validate() }},
		{name: "exception action report", wantValid: true, validate: func() error { return ExceptionActionReport.Validate() }},
		{name: "exception action ignore", wantValid: true, validate: func() error { return ExceptionActionIgnore.Validate() }},
		{name: "exception action unknown", validate: func() error { return ExceptionAction("unknown").Validate() }},
		{name: "exception detail summary", wantValid: true, validate: func() error { return ExceptionDetailSummary.Validate() }},
		{name: "exception detail full", wantValid: true, validate: func() error { return ExceptionDetailFull.Validate() }},
		{name: "exception detail unknown", validate: func() error { return ExceptionDetail("unknown").Validate() }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.validate()
			if testCase.wantValid && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if !testCase.wantValid && err == nil {
				t.Fatal("Validate() error = nil, want rejection")
			}
		})
	}
}
