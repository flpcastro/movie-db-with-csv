package errors

import (
	"fmt"
	"sync"
)

type List struct {
	errors []error
	mutex  sync.Mutex
}

func (l *List) Len() int {
	return len(l.errors)
}

func (l *List) AddError(err error) {
	l.mutex.Lock()
	l.errors = append(l.errors, err)
	l.mutex.Unlock()
}

func (l *List) Error() string {
	var message string
	for _, err := range l.errors {
		message += fmt.Sprintf("%s\n", err)
	}

	return message
}
