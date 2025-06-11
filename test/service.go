package test

import "fmt"

type UserService[T User, ID uint64] struct {
}

func (u UserService[T, ID]) QueryById(id uint64, result *User) int64 {
	fmt.Println(id)
	return 0
}
