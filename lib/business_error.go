package lib

type BusinessError struct {
	code int
	msg  string
	err  error
}

func NewBusinessError(code int, msg string) *BusinessError {
	return &BusinessError{
		code: code,
		msg:  msg,
	}
}

func IsBusinssError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

func (be *BusinessError) Error() string {
	return be.msg
}

func (be *BusinessError) Code() int {
	return be.code
}

func (be *BusinessError) Is(target error) bool {
	t, ok := target.(*BusinessError)
	if !ok {
		return false
	}
	return be.code == t.code
}

func (be *BusinessError) As(target interface{}) bool {
	switch t := target.(type) {
	case **BusinessError:
		*t = be
		return true
	case *BusinessError:
		*t = *be
		return true
	default:
		return false
	}
}

func (be *BusinessError) Wrap(err error) *BusinessError {
	return &BusinessError{
		code: be.code,
		msg:  be.msg,
		err:  err,
	}
}

func (be *BusinessError) Unwrap() error {
	return be.err
}
