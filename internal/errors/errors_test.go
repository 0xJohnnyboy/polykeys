package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeLayoutNotFound, "layout not found")

	if err.Code != ErrCodeLayoutNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeLayoutNotFound, err.Code)
	}

	if err.Message != "layout not found" {
		t.Errorf("expected message 'layout not found', got '%s'", err.Message)
	}

	expected := "[PK_100] layout not found"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestWrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := Wrap(ErrCodeConfigLoadFailed, "failed to load config", underlying)

	if err.Code != ErrCodeConfigLoadFailed {
		t.Errorf("expected code %s, got %s", ErrCodeConfigLoadFailed, err.Code)
	}

	if err.Err != underlying {
		t.Errorf("expected wrapped error to be stored")
	}

	expected := "[PK_300] failed to load config: underlying error"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestUnwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := Wrap(ErrCodeDeviceDetectionFailed, "detection failed", underlying)

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("expected unwrapped error to match underlying error")
	}
}

func TestWithDetails(t *testing.T) {
	err := New(ErrCodeLayoutSelectFailed, "failed to select layout")
	details := map[string]interface{}{
		"layout": "US QWERTY",
		"os":     "darwin",
	}

	err = WithDetails(err, details)

	if err.Details == nil {
		t.Error("expected details to be set")
	}

	if err.Details["layout"] != "US QWERTY" {
		t.Errorf("expected layout detail to be 'US QWERTY', got '%v'", err.Details["layout"])
	}

	if err.Details["os"] != "darwin" {
		t.Errorf("expected os detail to be 'darwin', got '%v'", err.Details["os"])
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "PolykeysError",
			err:      New(ErrCodeLayoutNotFound, "not found"),
			expected: ErrCodeLayoutNotFound,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: ErrCodeUnknown,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetCode(tt.err)
			if code != tt.expected {
				t.Errorf("expected code %s, got %s", tt.expected, code)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify error codes are unique
	codes := []ErrorCode{
		ErrCodeUnknown,
		ErrCodeLayoutNotFound,
		ErrCodeLayoutEnableFailed,
		ErrCodeLayoutSelectFailed,
		ErrCodeLayoutInvalidOS,
		ErrCodeLayoutStringFailed,
		ErrCodeLayoutInvalidIdentifier,
		ErrCodeDeviceNotFound,
		ErrCodeDeviceDetectionFailed,
		ErrCodeDeviceScanFailed,
		ErrCodeConfigLoadFailed,
		ErrCodeConfigParseFailed,
		ErrCodeConfigSaveFailed,
		ErrCodeConfigNotFound,
		ErrCodeMappingNotFound,
		ErrCodeMappingExists,
		ErrCodeInvalidMapping,
		ErrCodeRepositoryFailed,
		ErrCodeRepositoryNotFound,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("duplicate error code: %s", code)
		}
		seen[code] = true
	}
}

func TestErrorCodeRanges(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		prefix   string
		rangeMin int
		rangeMax int
	}{
		{"General", ErrCodeUnknown, "PK_0", 0, 99},
		{"Layout", ErrCodeLayoutNotFound, "PK_1", 100, 199},
		{"Device", ErrCodeDeviceNotFound, "PK_2", 200, 299},
		{"Config", ErrCodeConfigLoadFailed, "PK_3", 300, 399},
		{"Mapping", ErrCodeMappingNotFound, "PK_4", 400, 499},
		{"Repository", ErrCodeRepositoryFailed, "PK_5", 500, 599},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.code)[:4] != tt.prefix {
				t.Errorf("expected code %s to start with %s", tt.code, tt.prefix)
			}
		})
	}
}
