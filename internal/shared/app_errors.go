package shared

// AppError centralizes application error typing without HTTP coupling.
// Kind describes the class (bad_request, conflict, etc.), Code is a stable business code,
// Msg is a client-safe message, Err is an optional internal cause for wrapping.
type ErrorKind string

const (
	BadRequestKind ErrorKind = "bad_request"
	ConflictKind   ErrorKind = "conflict"
	NotFoundKind   ErrorKind = "not_found"
	InternalKind   ErrorKind = "internal"
)

type AppError struct {
	Kind ErrorKind
	Code string
	Msg  string
	Err  error
}

func (e *AppError) Error() string { return e.Msg }
func (e *AppError) Unwrap() error { return e.Err }

func NewBadRequest(code, msg string, cause error) *AppError {
	return &AppError{Kind: BadRequestKind, Code: code, Msg: msg, Err: cause}
}
func NewConflict(code, msg string, cause error) *AppError {
	return &AppError{Kind: ConflictKind, Code: code, Msg: msg, Err: cause}
}
func NewNotFound(code, msg string, cause error) *AppError {
	return &AppError{Kind: NotFoundKind, Code: code, Msg: msg, Err: cause}
}
func NewInternal(code, msg string, cause error) *AppError {
	return &AppError{Kind: InternalKind, Code: code, Msg: msg, Err: cause}
}
