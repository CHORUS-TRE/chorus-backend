package service

type ResourceAlreadyExistsErr struct{}

func (e *ResourceAlreadyExistsErr) Error() string {
	return "resource already exists"
}

type InvalidParametersErr struct{}

func (e *InvalidParametersErr) Error() string {
	return "invalid parameters"
}
