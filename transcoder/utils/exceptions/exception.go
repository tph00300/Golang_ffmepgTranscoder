package exceptions

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
func New(text string) error {
	return &errorString{text}
}
