package test

import (
	"github.com/golang-acexy/cloud-web/webcloud"
)

type UserService[T User, ID uint64] struct {
}

func (u UserService[T, ID]) Save(t *T) (ID, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserService[T, ID]) QueryByID(id ID, result *T) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserService[T, ID]) QueryOne(condition map[string]any, result *T) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserService[T, ID]) QueryByPager(condition map[string]any, pager *webcloud.Pager[T]) {
	//TODO implement me
	panic("implement me")
}

func (u UserService[T, ID]) ModifyByID(id ID, update map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserService[T, ID]) RemoveByID(id ID) (int64, error) {
	//TODO implement me
	panic("implement me")
}
