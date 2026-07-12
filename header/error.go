package header

type Error uint8

const (
	ErrInvalid Error = iota
	ErrEmptyName
	ErrEmptyValue
)

func (err Error) Error() string {
	switch err {
	case ErrEmptyName:
		return "empty name"
	case ErrEmptyValue:
		return "empty value"
	default:
		return "unknown header error"
	}
}
