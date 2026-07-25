package auth

import (
	"net/url"
	"strings"
)

// SafeReturnTo validates a return_to URL to prevent open redirect attacks.
// It only allows relative paths that start with "/".
// Any absolute URL, protocol-relative URL, or malformed input defaults to "/".
func SafeReturnTo(rawURL string) string {
	if rawURL == "" {
		return "/"
	}

	// Reject anything with a scheme (e.g. https://, javascript:, data:)
	if strings.Contains(rawURL, "://") {
		return "/"
	}

	// Reject protocol-relative URLs (//evil.com)
	if strings.HasPrefix(rawURL, "//") {
		return "/"
	}

	// Must start with /
	if !strings.HasPrefix(rawURL, "/") {
		return "/"
	}

	// Parse and reject if a host is present (catches edge cases)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "/"
	}
	if parsed.Host != "" {
		return "/"
	}
	if parsed.Scheme != "" {
		return "/"
	}

	// Reject backslash tricks (e.g. /\evil.com — some browsers treat \ as /)
	if strings.ContainsRune(rawURL, '\\') {
		return "/"
	}

	return rawURL
}
