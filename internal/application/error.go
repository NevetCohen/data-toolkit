package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/locator"
	"data-toolkit/internal/report"
)

// ErrorLocator is the API's single locator union. Exactly one of Source or
// Output may be set; JSON exposes the selected locator directly.
type ErrorLocator struct {
	Source *locator.SourceLocator
	Output *locator.OutputLocator
}

func SourceErrorLocator(value locator.SourceLocator) *ErrorLocator {
	return &ErrorLocator{Source: &value}
}

func OutputErrorLocator(value locator.OutputLocator) *ErrorLocator {
	return &ErrorLocator{Output: &value}
}

func (value ErrorLocator) Validate() error {
	if (value.Source == nil) == (value.Output == nil) {
		return errors.New("error locator must contain exactly one source or output locator")
	}
	if value.Source != nil && value.Source.SourceID == "" {
		return errors.New("source error locator source ID is required")
	}
	if value.Output != nil && value.Output.OutputPath == "" {
		return errors.New("output error locator output path is required")
	}
	return nil
}

func (value ErrorLocator) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	if value.Source != nil {
		return json.Marshal(value.Source)
	}
	return json.Marshal(value.Output)
}

func (value *ErrorLocator) UnmarshalJSON(data []byte) error {
	if value == nil {
		return errors.New("cannot unmarshal error locator into nil receiver")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return errors.New("error locator must be an object")
	}
	_, hasSource := fields["source_id"]
	_, hasOutput := fields["output_path"]
	if hasSource == hasOutput {
		return errors.New("error locator must identify exactly one source or output variant")
	}
	var source locator.SourceLocator
	if hasSource {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&source); err != nil || source.SourceID == "" {
			return errors.New("invalid source error locator")
		}
		*value = ErrorLocator{Source: &source}
		return nil
	}
	var output locator.OutputLocator
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&output); err != nil || output.OutputPath == "" {
		return errors.New("error locator must be a source or output locator")
	}
	*value = ErrorLocator{Output: &output}
	return nil
}

// RunFailureContext identifies a run associated with an API error. ReceiptID
// is set only by the post-append path; it is omitted before receipt append.
type RunFailureContext struct {
	RunID     core.RunID          `json:"run_id"`
	Status    core.TerminalStatus `json:"status"`
	ReceiptID *core.ReceiptID     `json:"receipt_id,omitempty"`
}

func (value RunFailureContext) Validate() error {
	if err := value.RunID.Validate(); err != nil {
		return err
	}
	if err := value.Status.Validate(); err != nil {
		return err
	}
	if value.ReceiptID != nil {
		if err := value.ReceiptID.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// APIError is the stable, product-level error returned by the Application API.
// Its field set is intentionally closed at exactly eight JSON fields.
type APIError struct {
	Code       string               `json:"code"`
	Message    string               `json:"message"`
	Component  string               `json:"component"`
	FieldPath  *string              `json:"field_path,omitempty"`
	Locator    *ErrorLocator        `json:"locator,omitempty"`
	Retryable  bool                 `json:"retryable"`
	Details    []report.ErrorDetail `json:"details"`
	RunContext *RunFailureContext   `json:"run_context,omitempty"`
}

func (value APIError) Validate() error {
	if value.Code == "" {
		return errors.New("API error code is required")
	}
	if value.Message == "" {
		return errors.New("API error message is required")
	}
	switch value.Component {
	case "infrastructure", "reader_engine", "logical_engine", "writer_engine", "orchestrator", "cli", "codex_adapter":
	default:
		return fmt.Errorf("API error component %q is invalid", value.Component)
	}
	if value.Locator != nil {
		if err := value.Locator.Validate(); err != nil {
			return err
		}
	}
	if value.RunContext != nil {
		if err := value.RunContext.Validate(); err != nil {
			return err
		}
	}
	if value.Details == nil {
		return errors.New("API error details are required")
	}
	for index := 1; index < len(value.Details); index++ {
		if value.Details[index-1].Key >= value.Details[index].Key {
			return errors.New("API error details must be sorted by key")
		}
	}
	return nil
}

// MarshalJSON validates the closed error contract before serialization.
func (value APIError) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type apiError APIError
	return json.Marshal(apiError(value))
}

// UnmarshalJSON rejects unknown APIError fields through a strict decoder.
func (value *APIError) UnmarshalJSON(data []byte) error {
	type apiError APIError
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded apiError
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := APIError(decoded).Validate(); err != nil {
		return err
	}
	*value = APIError(decoded)
	return nil
}
