package config

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"

	"data-toolkit/internal/contract"
)

const redactedValue = "[REDACTED]"

// CredentialRequest contains only a non-secret profile reference and the
// least-privilege scopes required by the adapter.
type CredentialRequest struct {
	Profile string
	Scopes  []string
}

// CredentialProvider keeps token loading and refresh outside ordinary
// configuration. Adapters receive an authorized opaque HTTP client.
type CredentialProvider interface {
	GoogleHTTPClient(context.Context, CredentialRequest) (*http.Client, error)
}

// Redactor removes credential material from strings, errors, and report
// structures. Secret values are supplied by the external provider at runtime.
type Redactor struct {
	secrets []string
}

func NewRedactor(secretValues ...string) Redactor {
	unique := make(map[string]struct{}, len(secretValues))
	for _, secret := range secretValues {
		if secret != "" {
			unique[secret] = struct{}{}
		}
	}
	secrets := make([]string, 0, len(unique))
	for secret := range unique {
		secrets = append(secrets, secret)
	}
	sort.Slice(secrets, func(left, right int) bool {
		return len(secrets[left]) > len(secrets[right])
	})
	return Redactor{secrets: secrets}
}

func (redactor Redactor) RedactString(value string) string {
	for _, secret := range redactor.secrets {
		value = strings.ReplaceAll(value, secret, redactedValue)
	}
	return value
}

func (redactor Redactor) RedactError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(redactor.RedactString(err.Error()))
}

func (redactor Redactor) RedactFinalResult(result contract.FinalResult) contract.FinalResult {
	result.ResolvedSettings = redactMap(redactor, result.ResolvedSettings)
	for index := range result.QueryAnswers {
		result.QueryAnswers[index].Value = redactAny(redactor, result.QueryAnswers[index].Value)
	}
	for index := range result.Warnings {
		result.Warnings[index] = redactDiagnostic(redactor, result.Warnings[index])
	}
	for index := range result.Errors {
		result.Errors[index] = redactDiagnostic(redactor, result.Errors[index])
	}
	for index := range result.Exceptions {
		result.Exceptions[index].Details = redactMap(redactor, result.Exceptions[index].Details)
	}
	return result
}

func redactDiagnostic(redactor Redactor, diagnostic contract.Diagnostic) contract.Diagnostic {
	diagnostic.Message = redactor.RedactString(diagnostic.Message)
	diagnostic.Details = redactMap(redactor, diagnostic.Details)
	return diagnostic
}

func redactMap(redactor Redactor, source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	target := make(map[string]any, len(source))
	for key, value := range source {
		target[key] = redactAny(redactor, value)
	}
	return target
}

func redactAny(redactor Redactor, value any) any {
	switch typed := value.(type) {
	case string:
		return redactor.RedactString(typed)
	case error:
		return redactor.RedactString(typed.Error())
	case map[string]any:
		return redactMap(redactor, typed)
	case []any:
		result := make([]any, len(typed))
		for index := range typed {
			result[index] = redactAny(redactor, typed[index])
		}
		return result
	default:
		return value
	}
}
