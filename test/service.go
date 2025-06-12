package test

import "github.com/golang-acexy/cloud-web/webcloud"

type UserBizService[ID webcloud.IDType, S, M, Q, T any] struct {
}

func (u UserBizService[ID, S, M, Q, T]) Save(save map[string]any) (ID, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, T]) QueryByID(condition map[string]any, result *T) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, T]) QueryOne(condition map[string]any, result *T) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, T]) QueryByPager(condition map[string]any, pager *webcloud.Pager[T]) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, T]) ModifyByID(condition map[string]any, update map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, T]) RemoveByID(condition map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}
