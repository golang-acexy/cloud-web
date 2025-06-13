package test

import "github.com/golang-acexy/cloud-web/webcloud"

type UserBizService[ID webcloud.IDType, S, M, Q, D any] struct {
}

func (u UserBizService[ID, S, M, Q, D]) Save(save *S) (ID, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryByID(condition map[string]any, result *D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryOne(condition map[string]any, result *D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) Query(condition map[string]any, result *[]*D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryByPager(condition map[string]any, pager *webcloud.Pager[D]) error {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) ModifyByID(condition map[string]any, update map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) RemoveByID(condition map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}
