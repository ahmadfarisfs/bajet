package model

type ApplicationError struct {
	Message string
}

func (e *ApplicationError) Error() string {
	return e.Message
}

var (
	ErrDuplicateCategoryName = &ApplicationError{Message: "Category name already exists"}
	ErrDuplicateCategoryIcon = &ApplicationError{Message: "Category icon already exists"}
)
