package featureflags

// featureError represents an error related to feature flag operations
type featureError struct {
	Message string
}

func (e *featureError) Error() string {
	return e.Message
}

func newFeatureError(message string) *featureError {
	return &featureError{Message: message}
}
