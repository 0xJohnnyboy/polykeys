package errors

// ErrorCode represents a Polykeys error code
type ErrorCode string

const (
	// General errors (000-099)
	ErrCodeUnknown ErrorCode = "PK_000"

	// Layout switching errors (100-199)
	ErrCodeLayoutNotFound        ErrorCode = "PK_100"
	ErrCodeLayoutEnableFailed    ErrorCode = "PK_101"
	ErrCodeLayoutSelectFailed    ErrorCode = "PK_102"
	ErrCodeLayoutInvalidOS       ErrorCode = "PK_103"
	ErrCodeLayoutStringFailed    ErrorCode = "PK_104"
	ErrCodeLayoutInvalidIdentifier ErrorCode = "PK_105"

	// Device detection errors (200-299)
	ErrCodeDeviceNotFound        ErrorCode = "PK_200"
	ErrCodeDeviceDetectionFailed ErrorCode = "PK_201"
	ErrCodeDeviceScanFailed      ErrorCode = "PK_202"

	// Configuration errors (300-399)
	ErrCodeConfigLoadFailed  ErrorCode = "PK_300"
	ErrCodeConfigParseFailed ErrorCode = "PK_301"
	ErrCodeConfigSaveFailed  ErrorCode = "PK_302"
	ErrCodeConfigNotFound    ErrorCode = "PK_303"

	// Use case errors (400-499)
	ErrCodeMappingNotFound   ErrorCode = "PK_400"
	ErrCodeMappingExists     ErrorCode = "PK_401"
	ErrCodeInvalidMapping    ErrorCode = "PK_402"

	// Repository errors (500-599)
	ErrCodeRepositoryFailed  ErrorCode = "PK_500"
	ErrCodeRepositoryNotFound ErrorCode = "PK_501"
)
