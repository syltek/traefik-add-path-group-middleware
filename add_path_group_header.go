package traefik_add_path_group_middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

const defaultHeaderName = "x-path-group"

// ID type labels
const (
	labelUUID      = "uuid"
	labelNumericID = "numeric_id"
	labelISODate   = "iso_date"
	labelULID      = "ulid"
	labelCUID      = "cuid"
	labelCUID2     = "cuid2"
	labelNanoID    = "nanoid"
	labelFile      = "file"
	labelSlug      = "slug"
)

var (
	// uuidPattern matches standard UUID format: 8-4-4-4-12 hex digits
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// numericPattern matches pure numeric IDs
	numericPattern = regexp.MustCompile(`^\d+$`)
	// isoDatePattern matches ISO 8601 date/datetime formats:
	// - Date only: YYYY-MM-DD (e.g., 2026-02-26)
	// - Datetime: YYYY-MM-DD[Tt]HH:MM:SS (e.g., 2026-02-26T00:01:55 or 2026-02-26t00:01:55)
	// - With timezone: YYYY-MM-DD[Tt]HH:MM:SSZ or YYYY-MM-DD[Tt]HH:MM:SS[+-]HH:MM
	// - With milliseconds: YYYY-MM-DD[Tt]HH:MM:SS[.SSS][Z|[+-]HH:MM]
	isoDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}([Tt]\d{2}:\d{2}:\d{2}(\.\d{1,9})?([Zz]|[+-]\d{2}:\d{2})?)?$`)
	// ulidPattern matches ULID format: exactly 26 chars, Crockford Base32 (excludes I, L, O, U)
	ulidPattern = regexp.MustCompile(`^[0-9A-HJ-NP-TV-Za-hj-np-tv-z]{26}$`)
	// cuidPattern matches CUID (v1) format: exactly 25 chars, starts with 'c', lowercase alphanumeric
	cuidPattern = regexp.MustCompile(`^c[a-z0-9]{24}$`)
	// cuid2Pattern matches CUID2 format: exactly 24 chars, starts with lowercase letter
	cuid2Pattern = regexp.MustCompile(`^[a-z][a-z0-9]{23}$`)
	// nanoidPattern matches NanoID format: URL-safe alphabet with at least one digit.
	// Length is checked separately (len == 21) since RE2 doesn't support lookaheads.
	// ~97% of random 21-char NanoIDs contain at least one digit.
	nanoidPattern = regexp.MustCompile(`^[A-Za-z0-9_-]*[0-9][A-Za-z0-9_-]*$`)
	// filePattern matches file segments ending with a file extension (e.g., .html, .css, .js, .png)
	// Matches segments that contain at least one character before a dot, followed by 1-15 alphanumeric characters
	filePattern = regexp.MustCompile(`^.+\.\w{1,15}$`)
	// slugPattern matches alphanumeric chars, dashes, and underscores
	slugPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	// prefixPattern matches alphanumeric prefix (for prefixed IDs)
	prefixPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// Config holds the plugin configuration
type Config struct {
	HeaderName string `json:"headerName,omitempty"`
}

// CreateConfig returns the default plugin configuration
func CreateConfig() *Config {
	return &Config{
		HeaderName: defaultHeaderName,
	}
}

// AddPathHeader is the middleware plugin that injects the request path into a header
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

// identifyIDType identifies the type of ID in a segment, checking patterns in order of specificity.
// Returns the ID type label if matched, empty string otherwise.
// Also handles prefixed IDs (e.g., "prefix:uuid", "prefix_nanoid").
func identifyIDType(segment string) string {
	if segment == "" {
		return ""
	}

	// 1. Check UUID (unique dash structure, 36 chars)
	if uuidPattern.MatchString(segment) {
		return labelUUID
	}

	// 2. Check Numeric (digits only, unambiguous)
	if numericPattern.MatchString(segment) {
		return labelNumericID
	}

	// 3. Check ISO Date/Datetime (YYYY-MM-DD with optional time and timezone)
	if isoDatePattern.MatchString(segment) {
		return labelISODate
	}

	// 4. Check ULID (26 chars, specific charset)
	if ulidPattern.MatchString(segment) {
		return labelULID
	}

	// 5. Check CUID (25 chars, starts with 'c')
	if cuidPattern.MatchString(segment) {
		return labelCUID
	}

	// 6. Check CUID2 (24 chars, starts with lowercase)
	if cuid2Pattern.MatchString(segment) {
		return labelCUID2
	}

	// 7. Check NanoID (21 chars, broader charset, must contain a digit)
	if len(segment) == 21 && nanoidPattern.MatchString(segment) {
		return labelNanoID
	}

	// 8. Check File (segments ending with file extension like .html, .css, .js, .png)
	if filePattern.MatchString(segment) {
		return labelFile
	}

	// 9. Try prefix extraction (check for prefix:ID or prefix_ID)
	// Try colon separator first (unambiguous)
	if idx := strings.Index(segment, ":"); idx > 0 {
		prefix := segment[:idx]
		suffix := segment[idx+1:]
		if prefixPattern.MatchString(prefix) && suffix != "" {
			if label := identifyIDType(suffix); label != "" {
				return label
			}
		}
	}

	// Try underscore separator (can appear in NanoID, but we already checked full segment)
	// For underscore, treat as prefixed ID if:
	// - Suffix matches non-numeric ID patterns (UUID, ULID, CUID, CUID2, NanoID, ISO Date), OR
	// - Suffix is numeric with 3+ digits (longer numeric IDs are more likely to be prefixed)
	// Shorter numeric suffixes (1-2 digits) are more likely to be slugs like "user_42"
	if idx := strings.Index(segment, "_"); idx > 0 {
		prefix := segment[:idx]
		suffix := segment[idx+1:]
		if prefixPattern.MatchString(prefix) && suffix != "" {
			// Check if suffix matches a non-numeric ID pattern
			if uuidPattern.MatchString(suffix) ||
				isoDatePattern.MatchString(suffix) ||
				ulidPattern.MatchString(suffix) ||
				cuidPattern.MatchString(suffix) ||
				cuid2Pattern.MatchString(suffix) ||
				(len(suffix) == 21 && nanoidPattern.MatchString(suffix)) {
				// Recursively identify the ID type
				if label := identifyIDType(suffix); label != "" {
					return label
				}
			} else if numericPattern.MatchString(suffix) && len(suffix) >= 3 {
				// Numeric suffix with 3+ digits - treat as prefixed numeric ID
				return labelNumericID
			}
		}
	}

	// 10. Check slug (alphanumeric with digits and separators)
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
			return labelSlug
		}
	}

	return ""
}

// extractPathGroup normalizes a path by replacing ID segments with their type labels
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

		if label := identifyIDType(segment); label != "" {
			result = append(result, label)
		} else {
			result = append(result, segment)
		}
	}

	return "/" + strings.Join(result, "/")
}

func (a *AddPathHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	pathGroup := extractPathGroup(req.URL.Path)
	req.Header.Set(a.headerName, pathGroup)
	a.next.ServeHTTP(rw, req)
}
