package bwgame

type NotPortedError struct {
	Name string
}

func (e *NotPortedError) Error() string {
	return "not ported: " + e.Name
}

func notPorted(cppName string) error {
	return &NotPortedError{Name: cppName}
}
