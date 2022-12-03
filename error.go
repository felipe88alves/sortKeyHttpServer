package main

type apiError struct {
	Err    string
	Status int
}

func (e apiError) Error() string {
	return e.Err
}
