package errors

import "errors"

// ── AppError — стандартная ошибка приложения с HTTP-статусом ─────────────────

// AppError содержит код ошибки, сообщение и HTTP-статус для маппинга в хендлере.
type AppError struct {
	Code    error
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError создаёт AppError с HTTP-статусом, кодом и сообщением.
func NewAppError(status int, code error, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// AsAppError пробует извлечь *AppError из error.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ── HTTP code errors ────────────────────────────────────────────────────────

var (
	ErrCodeBadRequest         = errors.New("BAD_REQUEST")
	ErrCodeInternalServer     = errors.New("INTERNAL_SERVER")
	ErrCodeNotFound           = errors.New("NOT_FOUND")
	ErrCodeUnauthorized       = errors.New("UNAUTHORIZED")
	ErrCodeForbidden          = errors.New("FORBIDDEN")
	ErrCodeConflict           = errors.New("CONFLICT")
	ErrCodePreconditionFailed = errors.New("PRECONDITION_FAILED")
	ErrCodeNotImplemented     = errors.New("NOT_IMPLEMENTED")
	ErrCodeBadGateway         = errors.New("BAD_GATEWAY")
	ErrCodeServiceUnavailable = errors.New("SERVICE_UNAVAILABLE")
	ErrCodeGatewayTimeout     = errors.New("GATEWAY_TIMEOUT")
	ErrCodeUnauthenticated    = errors.New("UNAUTHENTICATED")
)

// ── Auth errors ─────────────────────────────────────────────────────────────

var (
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrEmailAlreadyExists = errors.New("EMAIL_ALREADY_EXISTS")
	ErrTokenExpired       = errors.New("TOKEN_EXPIRED")
	ErrTokenInvalid       = errors.New("TOKEN_INVALID")
	ErrJWTNullInHeader    = errors.New("JWT_NULL_IN_HEADER")
	ErrJWTInvalid         = errors.New("JWT_INVALID")
	ErrJWTUnexpected      = errors.New("JWT_UNEXPECTED")
	ErrJWTParseFailed     = errors.New("JWT_PARSE_FAILED")
	ErrRefreshTokenHash   = errors.New("REFRESH_TOKEN_HASH")
)

// ── User errors ─────────────────────────────────────────────────────────────

var (
	ErrUserNotFound   = errors.New("USER_NOT_FOUND")
	ErrUserIsInactive = errors.New("USER_IS_INACTIVE")
	ErrEmailNotUnique = errors.New("EMAIL_NOT_UNIQUE")
)

// ── School errors ──────────────────────────────────────────────────────────

var (
	ErrSchoolNotFound = errors.New("SCHOOL_NOT_FOUND")
)

// ── Class errors ───────────────────────────────────────────────────────────

var (
	ErrClassNotFound      = errors.New("CLASS_NOT_FOUND")
	ErrClassAlreadyExists = errors.New("CLASS_ALREADY_EXISTS")
)

// ── Interest errors ────────────────────────────────────────────────────────

var (
	ErrInterestNotFound      = errors.New("INTEREST_NOT_FOUND")
	ErrInterestAlreadyExists = errors.New("INTEREST_ALREADY_EXISTS")
)

// ── App lifecycle errors ────────────────────────────────────────────────────

var (
	ErrInitConfig   = errors.New("init config error")
	ErrInitLogger   = errors.New("init logger error")
	ErrCloseDB      = errors.New("close db error")
	ErrCloseCache   = errors.New("close redis error")
	ErrCloseS3      = errors.New("close s3 error")
	ErrShutdownHTTP = errors.New("shutdown http server error")
)

// ── Validation errors ───────────────────────────────────────────────────────

var (
	ErrValidationFailed  = errors.New("validation failed")
	ErrEmptinessParam    = errors.New("emptiness param")
	ErrCheckPort         = errors.New("invalid port")
	ErrCheckMore         = errors.New("the parameter is less than the required value")
	ErrCheckLevel        = errors.New("invalid level")
	ErrLackVersion       = errors.New("lack of numbers")
	ErrMustBeNumber      = errors.New("must be a int number")
	ErrNoSupportInstance = errors.New("no support Instance")
	ErrEmptinessInstance = errors.New("emptiness instance")
	ErrEmptinessEnv      = errors.New("emptiness Env")
	ErrNoSupportEnv      = errors.New("no support Env")
	ErrPassLen           = errors.New("password length error")
	ErrPassInvalidChar   = errors.New("char invalid")
	ErrPassCountUppers   = errors.New("count uppers")
	ErrPassCountLowers   = errors.New("count lowers")
	ErrPassCountNumbers  = errors.New("count numbers")
	ErrPassCountSpecial  = errors.New("count special")
)

// ── Env parsing errors ──────────────────────────────────────────────────────

var (
	ErrEnvParseError     = errors.New("error parsing environment variable")
	ErrEnvKeyNotFound    = errors.New("key not found in environment variable")
	ErrEnvNotSupportType = errors.New("environment variable not supported")
)

// ── Storage type errors ─────────────────────────────────────────────────────

var (
	ErrTypeCache = errors.New("type cache error")
	ErrTypeDb    = errors.New("type db error")
	ErrTypeS3    = errors.New("type s3 error")
)

// ── PostgreSQL errors ───────────────────────────────────────────────────────

var (
	ErrPgParseConfig = errors.New("failed to parse config")
	ErrPgCreatePool  = errors.New("failed to create pool")
	ErrPgPing        = errors.New("failed to ping")
)

// ── Redis errors ────────────────────────────────────────────────────────────

var (
	ErrRedisConnectionFailed = errors.New("redis connection failed")
	ErrCacheMiss             = errors.New("cache miss")
)

// ── MinIO errors ────────────────────────────────────────────────────────────

var (
	ErrMinioConnect   = errors.New("failed to connect to minio")
	ErrBucketCheck    = errors.New("failed to check bucket existence")
	ErrBucketCreate   = errors.New("failed to create bucket")
	ErrSetRangeOffset = errors.New("failed to set range offset")
	ErrGetObject      = errors.New("failed to get object from minio")
	ErrStatObject     = errors.New("failed to get object info")
	ErrRemoveObject   = errors.New("failed to remove object")
	ErrListObjects    = errors.New("failed to list objects")
	ErrPutObject      = errors.New("failed to put object")
)
