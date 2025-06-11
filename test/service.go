package test

import "fmt"

type UserService[T User, ID uint64] struct {
}

func (u UserService[T, ID]) Save(t *T) (ID, error) {
	return 0, nil
}
func (u UserService[T, ID]) QueryById(id uint64, result *User) (int64, error) {
	fmt.Println(id)
	return 0, nil
}
func (u UserService[T, ID]) ModifyById(id ID, update map[string]any) (int64, error) {
	return 0, nil
}
