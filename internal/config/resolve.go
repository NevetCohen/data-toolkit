package config

// Resolve applies layers in ascending precedence and never aliases slices from
// the base document or any patch.
func Resolve(base Document, layers ...Patch) (Document, error) {
	resolved := cloneDocument(base)
	for _, layer := range layers {
		applyPatch(&resolved, layer)
	}
	if err := Validate(resolved); err != nil {
		return Document{}, err
	}
	return resolved, nil
}

func cloneDocument(source Document) Document {
	result := source
	result.Aliases = append([]Alias(nil), source.Aliases...)
	result.SemanticTypes = append([]SemanticType(nil), source.SemanticTypes...)
	result.Styles = append([]Style(nil), source.Styles...)
	return result
}

func applyPatch(target *Document, patch Patch) {
	if patch.Display != nil {
		applyDisplay(&target.Display, *patch.Display)
	}
	if patch.Aliases != nil {
		target.Aliases = append([]Alias(nil), (*patch.Aliases)...)
	}
	if patch.SemanticTypes != nil {
		target.SemanticTypes = append([]SemanticType(nil), (*patch.SemanticTypes)...)
	}
	if patch.Styles != nil {
		target.Styles = append([]Style(nil), (*patch.Styles)...)
	}
	if patch.Output != nil {
		applyOutput(&target.Output, *patch.Output)
	}
	if patch.Cleaning != nil {
		applyCleaning(&target.Cleaning, *patch.Cleaning)
	}
	if patch.Exceptions != nil {
		if patch.Exceptions.Action != nil {
			target.Exceptions.Action = *patch.Exceptions.Action
		}
		if patch.Exceptions.Detail != nil {
			target.Exceptions.Detail = *patch.Exceptions.Detail
		}
	}
	if patch.Boolean != nil && patch.Boolean.NotScope != nil {
		target.Boolean.NotScope = *patch.Boolean.NotScope
	}
	if patch.Runtime != nil {
		applyRuntime(&target.Runtime, *patch.Runtime)
	}
	if patch.Interfaces != nil && patch.Interfaces.RecentWorkflows != nil {
		target.Interfaces.RecentWorkflows = *patch.Interfaces.RecentWorkflows
	}
}

func applyDisplay(target *Display, patch DisplayPatch) {
	if patch.NullDisplay != nil {
		target.NullDisplay = *patch.NullDisplay
	}
	if patch.DecimalPrecision != nil {
		target.DecimalPrecision = *patch.DecimalPrecision
	}
	if patch.DateFormat != nil {
		target.DateFormat = *patch.DateFormat
	}
	if patch.TimeFormat != nil {
		target.TimeFormat = *patch.TimeFormat
	}
	if patch.PhoneFormat != nil {
		target.PhoneFormat = *patch.PhoneFormat
	}
	if patch.TextEncoding != nil {
		target.TextEncoding = *patch.TextEncoding
	}
	if patch.UnicodeNormalization != nil {
		target.UnicodeNormalization = *patch.UnicodeNormalization
	}
}

func applyOutput(target *Output, patch OutputPatch) {
	if patch.Directory != nil {
		target.Directory = *patch.Directory
	}
	if patch.FilenameTemplate != nil {
		target.FilenameTemplate = *patch.FilenameTemplate
	}
	if patch.DefaultStyle != nil {
		target.DefaultStyle = *patch.DefaultStyle
	}
	if patch.TXTDelimiter != nil {
		target.TXTDelimiter = *patch.TXTDelimiter
	}
	if patch.Collision != nil {
		target.Collision = *patch.Collision
	}
}

func applyCleaning(target *Cleaning, patch CleaningPatch) {
	if patch.MixedTypeAction != nil {
		target.MixedTypeAction = *patch.MixedTypeAction
	}
	if patch.DeduplicateKeep != nil {
		target.DeduplicateKeep = *patch.DeduplicateKeep
	}
	if patch.AutoNormalizeFormats != nil {
		target.AutoNormalizeFormats = *patch.AutoNormalizeFormats
	}
	if patch.TrimDetectedWhitespace != nil {
		target.TrimDetectedWhitespace = *patch.TrimDetectedWhitespace
	}
}

func applyRuntime(target *Runtime, patch RuntimePatch) {
	if patch.MaximumMemoryBytes != nil {
		target.MaximumMemoryBytes = *patch.MaximumMemoryBytes
	}
	if patch.LockRetryCount != nil {
		target.LockRetryCount = *patch.LockRetryCount
	}
	if patch.LockRetryInterval != nil {
		target.LockRetryInterval = *patch.LockRetryInterval
	}
	if patch.LockTimeout != nil {
		target.LockTimeout = *patch.LockTimeout
	}
	if patch.TemporaryWorkspace != nil {
		target.TemporaryWorkspace = *patch.TemporaryWorkspace
	}
	if patch.FailedRunRetention != nil {
		target.FailedRunRetention = *patch.FailedRunRetention
	}
}
