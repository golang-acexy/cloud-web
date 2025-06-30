package test

import "github.com/golang-acexy/cloud-web/webcloud"

type UserBizService[ID webcloud.IDType, S, M, Q, D any] struct {
}

func (u UserBizService[ID, S, M, Q, D]) MaxQueryCount() int {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) DefaultOrderBySQL() string {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) Save(save *S) (ID, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseQueryByID(condition map[string]any, result *D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseQueryOne(condition map[string]any, result *D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseQuery(condition map[string]any, result *[]*D) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseQueryByPager(condition map[string]any, pager webcloud.Pager[D]) error {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseModifyByID(update, condition map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) BaseRemoveByID(condition map[string]any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryByID(id ID) *D {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryOneByCond(condition *Q) *D {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryByCond(condition *Q) []*D {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) QueryByPager(pager webcloud.PagerDTO[Q]) webcloud.Pager[D] {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) ModifyByID(updated *M) bool {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) ModifyByIDExcludeZeroField(updated *M) bool {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) ModifyByIdUseMap(updated map[string]any, id ID) bool {
	//TODO implement me
	panic("implement me")
}

func (u UserBizService[ID, S, M, Q, D]) RemoveByID(id ID) bool {
	//TODO implement me
	panic("implement me")
}
