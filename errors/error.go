package errors

import "fmt"

const (
	NonFormattedLine = iota
	NonValidID
	TimeoutRequired
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func NewTimeoutRequired() Error {
	return Error{
		Code:    TimeoutRequired,
		Message: "no timeout specified",
	}
}

func NewNonFormattedLine(line []string) Error {
	return Error{
		Code:    NonFormattedLine,
		Message: fmt.Sprintf("line %v doesn't follow the expected pattern", line),
	}
}

func NewNonValidID(id string) Error {
	return Error{
		Code:    NonValidID,
		Message: fmt.Sprintf("id %v is not an integer", id),
	}
}
