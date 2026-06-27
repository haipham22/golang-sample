package echo

// ExportedTestAccessors provides access to internal state for testing.
// This follows the Go convention of using export_test.go for test helpers.

// ExportedSwaggerSettings is a wrapper around internal swaggerSettings for testing.
type ExportedSwaggerSettings struct {
	internal *swaggerSettings
}

// ExportGetSwaggerSettings returns a wrapper for swaggerSettings instance for testing.
func ExportGetSwaggerSettings() *ExportedSwaggerSettings {
	return &ExportedSwaggerSettings{internal: &swaggerSettings{}}
}

// ExportApplySwaggerOption applies an option to settings for testing.
func (s *ExportedSwaggerSettings) ExportApplySwaggerOption(opt SwaggerOption) {
	opt(s.internal)
}

// ExportGetSwaggerEnabled returns the enabled flag from settings.
func (s *ExportedSwaggerSettings) ExportGetSwaggerEnabled() bool {
	return s.internal.enabled
}

// ExportGetSwaggerPath returns the path from settings.
func (s *ExportedSwaggerSettings) ExportGetSwaggerPath() string {
	return s.internal.path
}

// ExportGetSwaggerInfo returns the info from settings.
func (s *ExportedSwaggerSettings) ExportGetSwaggerInfo() *SwaggerInfo {
	return s.internal.info
}

// ExportGetSwaggerAuth returns the auth from settings.
func (s *ExportedSwaggerSettings) ExportGetSwaggerAuth() *SwaggerAuth {
	return s.internal.auth
}
