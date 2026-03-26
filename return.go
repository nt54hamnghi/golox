package main

import "fmt"

type ReturnThis struct {
	Value Object
}

func (rt ReturnThis) Error() string {
	return fmt.Sprintf("%v", rt)
}
