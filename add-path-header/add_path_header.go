package add_path_header

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

const defaultHeaderName = "x-path-group"

var (
	// uuidPattern matches standard UUID format: 8-4-4-4-12 hex digits
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// numericPattern matches pure numeric IDs
	numericPattern = regexp.MustCompile(`^\d+$`)
	// slugPattern matches alphanumeric chars, dashes, and underscores
	slugPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Config holds the plugin configuration.
type Config struct {
	HeaderName string `json:"headerName,omitempty"`
}

// CreateConfig returns the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName: defaultHeaderName,
	}
}

// AddPathHeader is the middleware plugin that injects the request path into a header.
type AddPathHeader struct {
	next       http.Handler
	headerName string
	name       string
}

// New creates a new AddPathHeader middleware plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	headerName := config.HeaderName
	if headerName == "" {
		headerName = defaultHeaderName
	}

	return &AddPathHeader{
		next:       next,
		headerName: headerName,
		name:       name,
	}, nil
}

// extractPathGroup normalizes a path by replacing ID segments with *
func extractPathGroup(path string) string {
	if path == "" || path == "/" {
		return path
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	result := make([]string, 0, len(segments))

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// Check if segment is a UUID
		if uuidPattern.MatchString(segment) {
			result = append(result, "*")
			continue
		}

		// Check if segment is pure numeric
		if numericPattern.MatchString(segment) {
			result = append(result, "*")
			continue
		}

		// Check if segment is an alphanumeric slug with digits and separators
		if slugPattern.MatchString(segment) {
			hasDigit := false
			hasSeparator := false
			for _, r := range segment {
				if r >= '0' && r <= '9' {
					hasDigit = true
				}
				if r == '-' || r == '_' {
					hasSeparator = true
				}
			}
			if hasDigit && hasSeparator {
				result = append(result, "*")
				continue
			}
		}

		// Keep the segment as-is
		result = append(result, segment)
	}

	return "/" + strings.Join(result, "/")
}

func (a *AddPathHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	pathGroup := extractPathGroup(req.URL.Path)
	req.Header.Set(a.headerName, pathGroup)
	a.next.ServeHTTP(rw, req)
}
