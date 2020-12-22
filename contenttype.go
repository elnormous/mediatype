package contenttype

import (
	"errors"
	"net/http"
	"strings"
)

var InvalidContentTypeError = errors.New("Invalid content type")
var InvalidParameterError = errors.New("Invalid parameter")

func isWhiteSpaceChar(c byte) bool {
	// RFC 7230, 3.2.3. Whitespace
	return c == 0x09 || c == 0x20 // HTAB or SP
}

func isDigitChar(c byte) bool {
	// RFC 5234, Appendix B.1. Core Rules
	return c >= 0x30 && c <= 0x39
}

func isAlphaChar(c byte) bool {
	// RFC 5234, Appendix B.1. Core Rules
	return (c >= 0x41 && c <= 0x5A) || (c >= 0x61 && c <= 0x7A)
}

func isTokenChar(c byte) bool {
	// RFC 7230, 3.2.6. Field Value Components
	return c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' || c == '*' ||
		c == '+' || c == '-' || c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~' ||
		isDigitChar(c) ||
		isAlphaChar(c)
}

func isVisibleChar(c byte) bool {
	// RFC 5234, Appendix B.1. Core Rules
	return c >= 0x21 && c <= 0x7E
}

func isObsoleteTextChar(c byte) bool {
	// RFC 7230, 3.2.6. Field Value Components
	return c >= 0x80 && c <= 0xFF
}

func isQuotedTextChar(c byte) bool {
	// RFC 7230, 3.2.6. Field Value Components
	return c == 0x09 || c == 0x20 || // HTAB or SP
		c == 0x21 ||
		(c >= 0x23 && c <= 0x5B) ||
		(c >= 0x5D && c <= 0x7E) ||
		isObsoleteTextChar(c)
}

func isQuotedPairChar(c byte) bool {
	// RFC 7230, 3.2.6. Field Value Components
	return c == 0x09 || c == 0x20 || // HTAB or SP
		isVisibleChar(c) ||
		isObsoleteTextChar(c)
}

func consumeToken(s string) (token, remaining string, consumed bool) {
	for i := 0; i < len(s); i++ {
		if !isTokenChar(s[i]) {
			return s[:i], s[i:], i > 0
		}
	}

	return s, "", true
}

func consumeQuotedString(s string) (token, remaining string, consumed bool) {
	if len(s) == 0 || s[0] != '"' {
		return "", s, false
	}

	var stringBuilder strings.Builder

	for index := 1; index < len(s); index++ {
		if s[index] == '"' {
			return stringBuilder.String(), s[index+1:], true
		}

		if s[index] == '\\' {
			index++
			if len(s) <= index || !isQuotedPairChar(s[index]) {
				return "", s, false
			}

			stringBuilder.WriteByte(s[index])
		} else {
			if !isQuotedTextChar(s[index]) {
				return "", s, false
			}
			stringBuilder.WriteByte(s[index])
		}
	}

	return "", s, false
}

func skipWhiteSpaces(s string) string {
	for i := 0; i < len(s); i++ {
		if !isWhiteSpaceChar(s[i]) {
			return s[i:]
		}
	}

	return ""
}

func GetMediaType(request *http.Request) (string, map[string]string, error) {
	// RFC 7231, 3.1.1.1. Media Type
	contentTypes := request.Header.Values("Content-Type")

	if len(contentTypes) == 0 {
		return "", map[string]string{}, nil
	}

	s := skipWhiteSpaces(contentTypes[0])

	var ok bool
	var supertype string
	supertype, s, ok = consumeToken(s)
	if !ok {
		return "", nil, InvalidContentTypeError
	}

	if len(s) == 0 || s[0] != '/' {
		return "", nil, InvalidContentTypeError
	}

	s = s[1:] // skip the slash

	var subtype string
	subtype, s, ok = consumeToken(s)
	if !ok {
		return "", nil, InvalidContentTypeError
	}

	s = skipWhiteSpaces(s)

	parameters := make(map[string]string)

	for len(s) != 0 {
		if s[0] != ';' {
			return "", nil, InvalidParameterError
		}

		s = s[1:] // skip the semicolon
		s = skipWhiteSpaces(s)

		var key string
		key, s, ok = consumeToken(s)
		if !ok {
			return "", nil, InvalidParameterError
		}

		if len(s) == 0 || s[0] != '=' {
			return "", nil, InvalidParameterError
		}

		s = s[1:] // skip the equal sign

		var value string
		if len(s) != 0 && s[0] == '"' { // opening quote
			value, s, ok = consumeQuotedString(s)

			if !ok {
				return "", nil, InvalidParameterError
			}

		} else {
			value, s, ok = consumeToken(s)
			if !ok {
				return "", nil, InvalidParameterError
			}
		}

		parameters[strings.ToLower(key)] = strings.ToLower(value)

		s = skipWhiteSpaces(s)
	}

	var stringBuilder strings.Builder
	stringBuilder.WriteString(strings.ToLower(supertype))
	stringBuilder.WriteByte('/')
	stringBuilder.WriteString(strings.ToLower(subtype))

	return stringBuilder.String(), parameters, nil
}
