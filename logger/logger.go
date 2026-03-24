package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Custom log levels (slog uses int, lower = more severe)
const (
	LevelTrace = slog.Level(-8)
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  // 0
	LevelWarn  = slog.LevelWarn  // 4
	LevelError = slog.LevelError // 8
	LevelFatal = slog.Level(12)
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorBoldRed = "\033[1;31m"
)

// Config holds logger configuration
type Config struct {
	Level                   string
	Format                  string
	EnableFileLogging       bool
	EnableConsoleLogging    bool
	EnableAccessFileLogging bool
	LogDirectory            string
	MaxFileSize             int
	MaxFiles                int
}

var (
	// L is the main logger instance
	L *slog.Logger

	// config holds the current configuration
	config Config

	// writers for cleanup
	combinedWriter *lumberjack.Logger
	errorWriter    *lumberjack.Logger
	accessWriter   *lumberjack.Logger

	// accessLogger for HTTP access logs
	accessLogger *slog.Logger

	// mu protects config and loggers
	mu sync.RWMutex
)

// Stream provides io.Writer interface for middleware integration
type Stream struct{}

func (s *Stream) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}

	// Try to parse as JSON for structured access logs
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &meta); err == nil {
		statusCode := 0
		if sc, ok := meta["statusCode"].(float64); ok {
			statusCode = int(sc)
		}
		level := LevelInfo
		if statusCode >= 500 {
			level = LevelError
		} else if statusCode >= 400 {
			level = LevelWarn
		}
		accessLog(level, "HTTP request completed", attrsFromMap(meta)...)
	} else {
		accessLog(LevelInfo, msg)
	}
	return len(p), nil
}

// LogStream returns a Stream for middleware integration
var LogStream = &Stream{}

func parseBoolean(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseFileSize(size string) int {
	if size == "" {
		return 20 * 1024 * 1024
	}

	re := regexp.MustCompile(`^(\d+)([kmgKMG]?)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(size))
	if matches == nil {
		return 20 * 1024 * 1024
	}

	amount, _ := strconv.Atoi(matches[1])
	switch strings.ToLower(matches[2]) {
	case "k":
		return amount * 1024
	case "m":
		return amount * 1024 * 1024
	case "g":
		return amount * 1024 * 1024 * 1024
	default:
		return amount
	}
}

func parseMaxFiles(value string) int {
	if value == "" {
		return 14
	}
	// Remove 'd' suffix if present (for compatibility with JS "14d" format)
	value = strings.TrimSuffix(strings.TrimSpace(value), "d")
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return n
	}
	return 14
}

func levelFromString(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info", "":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

func levelToString(level slog.Level) string {
	switch {
	case level <= LevelTrace:
		return "TRACE"
	case level <= LevelDebug:
		return "DEBUG"
	case level <= LevelInfo:
		return "INFO"
	case level <= LevelWarn:
		return "WARN"
	case level <= LevelError:
		return "ERROR"
	default:
		return "FATAL"
	}
}

func levelColor(level slog.Level) string {
	switch {
	case level <= LevelTrace:
		return colorMagenta
	case level <= LevelDebug:
		return colorBlue
	case level <= LevelInfo:
		return colorGreen
	case level <= LevelWarn:
		return colorYellow
	case level <= LevelError:
		return colorRed
	default:
		return colorBoldRed
	}
}

// multiHandler implements slog.Handler to write to multiple outputs
type multiHandler struct {
	handlers []slog.Handler
	mu       sync.RWMutex
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return newMultiHandler(newHandlers...)
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return newMultiHandler(newHandlers...)
}

// textHandler implements colorized text output
type textHandler struct {
	w       io.Writer
	level   slog.Level
	colorize bool
	mu      sync.Mutex
	attrs   []slog.Attr
	groups  []string
}

func newTextHandler(w io.Writer, level slog.Level, colorize bool) *textHandler {
	return &textHandler{
		w:        w,
		level:    level,
		colorize: colorize,
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timestamp := r.Time.Format("2006-01-02 15:04:05")
	levelStr := levelToString(r.Level)

	var sb strings.Builder
	if h.colorize {
		color := levelColor(r.Level)
		sb.WriteString(fmt.Sprintf("%s %s[%s]%s: %s", timestamp, color, levelStr, colorReset, r.Message))
	} else {
		sb.WriteString(fmt.Sprintf("%s [%s]: %s", timestamp, levelStr, r.Message))
	}

	// Collect attributes
	attrs := make(map[string]interface{})
	for _, attr := range h.attrs {
		attrs[attr.Key] = attr.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	if len(attrs) > 0 {
		if jsonBytes, err := json.Marshal(attrs); err == nil {
			sb.WriteString(" ")
			sb.Write(jsonBytes)
		}
	}

	sb.WriteString("\n")
	_, err := h.w.Write([]byte(sb.String()))
	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &textHandler{
		w:        h.w,
		level:    h.level,
		colorize: h.colorize,
		attrs:    make([]slog.Attr, len(h.attrs)+len(attrs)),
		groups:   h.groups,
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	return newHandler
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	newHandler := &textHandler{
		w:        h.w,
		level:    h.level,
		colorize: h.colorize,
		attrs:    h.attrs,
		groups:   append(h.groups, name),
	}
	return newHandler
}

// levelFilterHandler wraps a handler to filter by minimum level
type levelFilterHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func newLevelFilterHandler(h slog.Handler, minLevel slog.Level) *levelFilterHandler {
	return &levelFilterHandler{handler: h, minLevel: minLevel}
}

func (h *levelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel && h.handler.Enabled(ctx, level)
}

func (h *levelFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.minLevel {
		return h.handler.Handle(ctx, r)
	}
	return nil
}

func (h *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelFilterHandler{handler: h.handler.WithAttrs(attrs), minLevel: h.minLevel}
}

func (h *levelFilterHandler) WithGroup(name string) slog.Handler {
	return &levelFilterHandler{handler: h.handler.WithGroup(name), minLevel: h.minLevel}
}

// stderrHandler routes error/fatal to stderr
type stderrHandler struct {
	stdoutHandler slog.Handler
	stderrHandler slog.Handler
}

func newStderrHandler(stdoutH, stderrH slog.Handler) *stderrHandler {
	return &stderrHandler{stdoutHandler: stdoutH, stderrHandler: stderrH}
}

func (h *stderrHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdoutHandler.Enabled(ctx, level) || h.stderrHandler.Enabled(ctx, level)
}

func (h *stderrHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= LevelError {
		return h.stderrHandler.Handle(ctx, r)
	}
	return h.stdoutHandler.Handle(ctx, r)
}

func (h *stderrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &stderrHandler{
		stdoutHandler: h.stdoutHandler.WithAttrs(attrs),
		stderrHandler: h.stderrHandler.WithAttrs(attrs),
	}
}

func (h *stderrHandler) WithGroup(name string) slog.Handler {
	return &stderrHandler{
		stdoutHandler: h.stdoutHandler.WithGroup(name),
		stderrHandler: h.stderrHandler.WithGroup(name),
	}
}

func loadConfig() Config {
	return Config{
		Level:                   strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))),
		Format:                  strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT"))),
		EnableFileLogging:       parseBoolean(os.Getenv("ENABLE_FILE_LOGGING"), true),
		EnableConsoleLogging:    parseBoolean(os.Getenv("ENABLE_CONSOLE_LOGGING"), true),
		EnableAccessFileLogging: parseBoolean(os.Getenv("ENABLE_ACCESS_FILE_LOGGING"), true),
		LogDirectory:            os.Getenv("LOG_DIRECTORY"),
		MaxFileSize:             parseFileSize(os.Getenv("LOG_MAX_FILE_SIZE")),
		MaxFiles:                parseMaxFiles(os.Getenv("LOG_MAX_FILES")),
	}
}

func createFileHandler(filename string, level slog.Level, format string, maxSize, maxFiles int) (slog.Handler, *lumberjack.Logger) {
	lj := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize / (1024 * 1024), // Convert bytes to MB
		MaxBackups: maxFiles,
		Compress:   false,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(lj, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					return slog.String(slog.LevelKey, levelToString(a.Value.Any().(slog.Level)))
				}
				return a
			},
		})
	} else {
		handler = newTextHandler(lj, level, false)
	}
	return handler, lj
}

// Init initializes the logger with configuration from environment variables
func Init() {
	mu.Lock()
	defer mu.Unlock()

	config = loadConfig()

	if config.Level == "" {
		config.Level = "info"
	}
	if config.Format == "" {
		config.Format = "text"
	}
	if config.LogDirectory == "" {
		config.LogDirectory = "logs"
	}

	level := levelFromString(config.Level)
	var handlers []slog.Handler
	var accessHandlers []slog.Handler

	// Console handler
	if config.EnableConsoleLogging {
		var consoleHandler slog.Handler
		if config.Format == "json" {
			stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: level,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey {
						return slog.String(slog.LevelKey, levelToString(a.Value.Any().(slog.Level)))
					}
					return a
				},
			})
			stderrHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.LevelKey {
						return slog.String(slog.LevelKey, levelToString(a.Value.Any().(slog.Level)))
					}
					return a
				},
			})
			consoleHandler = newStderrHandler(stdoutHandler, stderrHandler)
		} else {
			stdoutHandler := newTextHandler(os.Stdout, level, true)
			stderrHandler := newTextHandler(os.Stderr, level, true)
			consoleHandler = newStderrHandler(stdoutHandler, stderrHandler)
		}
		handlers = append(handlers, consoleHandler)
		accessHandlers = append(accessHandlers, consoleHandler)
	}

	// File handlers
	if config.EnableFileLogging {
		// Create log directory
		logDir := config.LogDirectory
		if !filepath.IsAbs(logDir) {
			if cwd, err := os.Getwd(); err == nil {
				logDir = filepath.Join(cwd, logDir)
			}
		}

		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory \"%s\": %v\n", logDir, err)
			config.EnableFileLogging = false
		} else {
			// Combined log (all levels)
			combinedPath := filepath.Join(logDir, "combined.log")
			combinedHandler, cw := createFileHandler(combinedPath, level, config.Format, config.MaxFileSize, config.MaxFiles)
			combinedWriter = cw
			handlers = append(handlers, combinedHandler)

			// Error log (error and fatal only)
			errorPath := filepath.Join(logDir, "error.log")
			errorHandler, ew := createFileHandler(errorPath, LevelError, config.Format, config.MaxFileSize, config.MaxFiles)
			errorWriter = ew
			handlers = append(handlers, newLevelFilterHandler(errorHandler, LevelError))

			// Access log
			if config.EnableAccessFileLogging {
				accessPath := filepath.Join(logDir, "access.log")
				accessHandler, aw := createFileHandler(accessPath, level, config.Format, config.MaxFileSize, config.MaxFiles)
				accessWriter = aw
				accessHandlers = append(accessHandlers, accessHandler)
			}
		}
	}

	// Create main logger
	if len(handlers) == 0 {
		// Fallback to stdout if no handlers configured
		L = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	} else if len(handlers) == 1 {
		L = slog.New(handlers[0])
	} else {
		L = slog.New(newMultiHandler(handlers...))
	}

	// Create access logger
	if len(accessHandlers) == 0 {
		accessLogger = L
	} else if len(accessHandlers) == 1 {
		accessLogger = slog.New(accessHandlers[0])
	} else {
		accessLogger = slog.New(newMultiHandler(accessHandlers...))
	}

	slog.SetDefault(L)
}

// Close cleans up logger resources
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if combinedWriter != nil {
		combinedWriter.Close()
	}
	if errorWriter != nil {
		errorWriter.Close()
	}
	if accessWriter != nil {
		accessWriter.Close()
	}
}

// GetConfig returns the current logger configuration
func GetConfig() Config {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// Helper functions for logging at different levels
func Trace(msg string, args ...any) {
	L.Log(context.Background(), LevelTrace, msg, args...)
}

func Debug(msg string, args ...any) {
	L.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	L.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	L.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	L.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	L.Log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

// accessLog logs to the access logger
func accessLog(level slog.Level, msg string, args ...any) {
	mu.RLock()
	logger := accessLogger
	mu.RUnlock()
	if logger != nil {
		logger.Log(context.Background(), level, msg, args...)
	}
}

// AccessLog logs an HTTP access entry
func AccessLog(level slog.Level, msg string, args ...any) {
	accessLog(level, msg, args...)
}

// AccessLogInfo logs an access entry at info level
func AccessLogInfo(msg string, args ...any) {
	accessLog(LevelInfo, msg, args...)
}

func attrsFromMap(m map[string]interface{}) []any {
	args := make([]any, 0, len(m)*2)
	for k, v := range m {
		args = append(args, k, v)
	}
	return args
}

// WithFields returns a logger with additional fields
func WithFields(args ...any) *slog.Logger {
	return L.With(args...)
}

// GetLevel returns the current log level
func GetLevel() string {
	mu.RLock()
	defer mu.RUnlock()
	return config.Level
}

// Ensure Stream implements io.Writer
var _ io.Writer = (*Stream)(nil)

// init ensures logger is initialized with defaults if Init() is not called
func init() {
	// Initialize with defaults - can be overridden by calling Init()
	L = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: LevelInfo}))
	accessLogger = L
	config = Config{
		Level:  "info",
		Format: "json",
	}
}

// Attr creates a slog.Attr for convenience
func Attr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// TimeAttr creates a time attribute
func TimeAttr(key string, t time.Time) slog.Attr {
	return slog.Time(key, t)
}

// DurationAttr creates a duration attribute
func DurationAttr(key string, d time.Duration) slog.Attr {
	return slog.Duration(key, d)
}

// ErrAttr creates an error attribute
func ErrAttr(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("error", err.Error())
}
