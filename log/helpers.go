package log

// Debug logs a message at DebugLevel.
func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

// Info logs a message at InfoLevel.
func Info(args ...any) {
	defaultLogger.Info(args...)
}

// Warn logs a message at WarnLevel.
func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

// Error logs a message at ErrorLevel.
func Error(args ...any) {
	defaultLogger.Error(args...)
}

// Fatal logs a message at FatalLevel and exits.
func Fatal(args ...any) {
	defaultLogger.Fatal(args...)
}

// Debugf logs a formatted message at DebugLevel.
func Debugf(template string, args ...any) {
	defaultLogger.Debugf(template, args...)
}

// Infof logs a formatted message at InfoLevel.
func Infof(template string, args ...any) {
	defaultLogger.Infof(template, args...)
}

// Warnf logs a formatted message at WarnLevel.
func Warnf(template string, args ...any) {
	defaultLogger.Warnf(template, args...)
}

// Errorf logs a formatted message at ErrorLevel.
func Errorf(template string, args ...any) {
	defaultLogger.Errorf(template, args...)
}

// Fatalf logs a formatted message at FatalLevel and exits.
func Fatalf(template string, args ...any) {
	defaultLogger.Fatalf(template, args...)
}

// Debugw logs a message with key-value pairs at DebugLevel.
func Debugw(msg string, keysAndValues ...any) {
	defaultLogger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with key-value pairs at InfoLevel.
func Infow(msg string, keysAndValues ...any) {
	defaultLogger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with key-value pairs at WarnLevel.
func Warnw(msg string, keysAndValues ...any) {
	defaultLogger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with key-value pairs at ErrorLevel.
func Errorw(msg string, keysAndValues ...any) {
	defaultLogger.Errorw(msg, keysAndValues...)
}

// Fatalw logs a message with key-value pairs at FatalLevel and exits.
func Fatalw(msg string, keysAndValues ...any) {
	defaultLogger.Fatalw(msg, keysAndValues...)
}
