package test

import (
	"errors"
	"sync"

	"github.com/golang-acexy/cloud-web/webcloud"
)

var errServiceFailure = errors.New("service failure")

type UserBizService struct {
	lock           sync.Mutex
	lastSave       UserSDTO
	lastCondition  map[string]any
	lastUpdate     map[string]any
	lastTimeRanges []webcloud.TimeRange
}

func (u *UserBizService) reset() {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.lastSave = UserSDTO{}
	u.lastCondition = nil
	u.lastUpdate = nil
	u.lastTimeRanges = nil
}

func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func (u *UserBizService) recordCondition(condition map[string]any) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.lastCondition = cloneMap(condition)
}

func (u *UserBizService) MaxQuerySize() int { return 500 }

func (u *UserBizService) DefaultOrderBy() string { return "id desc" }

func (u *UserBizService) DefaultTimeRangeField() string { return "created_at" }

func (u *UserBizService) AllowedTimeRangeFields() []string {
	return []string{"created_at", "updated_at"}
}

func (u *UserBizService) Save(save *UserSDTO) (uint64, error) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.lastSave = *save
	return 1, nil
}

func (u *UserBizService) SaveWithoutZeroFields(save *UserSDTO) (uint64, error) {
	return u.Save(save)
}

func (u *UserBizService) SaveBatch(saves []*UserSDTO) ([]uint64, error) {
	ids := make([]uint64, len(saves))
	for index := range saves {
		ids[index] = uint64(index + 1)
	}
	return ids, nil
}

func (u *UserBizService) BaseQueryByID(condition map[string]any) (*UserDTO, error) {
	u.recordCondition(condition)
	if condition["id"] == uint64(500) {
		return nil, errServiceFailure
	}
	if condition["id"] == uint64(404) {
		return nil, nil
	}
	return &UserDTO{User: User{ID: condition["id"].(uint64), ClassName: "one"}}, nil
}

func (u *UserBizService) BaseQueryOne(condition map[string]any) (*UserDTO, error) {
	u.recordCondition(condition)
	if condition["name"] == "missing" {
		return nil, nil
	}
	return &UserDTO{User: User{ID: 1, ClassName: "one"}}, nil
}

func (u *UserBizService) BaseQuery(condition map[string]any) ([]*UserDTO, error) {
	u.recordCondition(condition)
	if condition["name"] == "failure" {
		return nil, errServiceFailure
	}
	if condition["name"] == "empty" {
		return []*UserDTO{}, nil
	}
	return []*UserDTO{{User: User{ID: 1, ClassName: "list"}}}, nil
}

func (u *UserBizService) BaseQueryPage(condition map[string]any, timeRanges []webcloud.TimeRange, pager *webcloud.Pager[UserDTO]) error {
	u.recordCondition(condition)
	u.lock.Lock()
	u.lastTimeRanges = append([]webcloud.TimeRange(nil), timeRanges...)
	u.lock.Unlock()
	pager.Total = 1
	pager.Records = []*UserDTO{{User: User{ID: 1, ClassName: "page"}}}
	return nil
}

func (u *UserBizService) BaseModifyByID(update, condition map[string]any) (int64, error) {
	u.lock.Lock()
	u.lastUpdate = cloneMap(update)
	u.lastCondition = cloneMap(condition)
	u.lock.Unlock()
	if condition["id"] == uint64(500) {
		return 0, errServiceFailure
	}
	if condition["id"] == uint64(404) {
		return 0, nil
	}
	return 1, nil
}

func (u *UserBizService) BaseRemoveByID(condition map[string]any) (int64, error) {
	u.recordCondition(condition)
	if condition["id"] == uint64(500) {
		return 0, errServiceFailure
	}
	if condition["id"] == uint64(404) {
		return 0, nil
	}
	return 1, nil
}

func (u *UserBizService) QueryByID(id uint64) (*UserDTO, error) {
	return &UserDTO{User: User{ID: id}}, nil
}

func (u *UserBizService) QueryByIDs(ids []uint64) ([]*UserDTO, error) {
	result := make([]*UserDTO, 0, len(ids))
	for _, id := range ids {
		result = append(result, &UserDTO{User: User{ID: id}})
	}
	return result, nil
}

func (u *UserBizService) ExistsByID(id uint64) (bool, error) {
	return id != 0, nil
}

func (u *UserBizService) QueryOneByCond(condition UserQDTO) (*UserDTO, error) {
	return &UserDTO{User: User{ID: condition.UserID}}, nil
}

func (u *UserBizService) QueryByCond(condition UserQDTO) ([]*UserDTO, error) {
	return []*UserDTO{{User: User{ID: condition.UserID}}}, nil
}

func (u *UserBizService) CountByCond(condition UserQDTO) (int64, error) {
	return int64(condition.UserID), nil
}

func (u *UserBizService) QueryPage(pager webcloud.PagerDTO[UserQDTO]) (webcloud.Pager[UserDTO], error) {
	return webcloud.Pager[UserDTO]{Number: pager.Number, Size: pager.Size}, nil
}

func (u *UserBizService) ModifyByID(id uint64, updated *UserMDTO) (int64, error) {
	return 1, nil
}

func (u *UserBizService) ModifyByIDWithoutZeroFields(id uint64, updated *UserMDTO) (int64, error) {
	return 1, nil
}

func (u *UserBizService) ModifyByIDWithMap(id uint64, updated map[string]any) (int64, error) {
	return 1, nil
}

func (u *UserBizService) RemoveByID(id uint64) (int64, error) { return 1, nil }

func (u *UserBizService) RemoveByIDs(ids []uint64) (int64, error) {
	return int64(len(ids)), nil
}

func (u *UserBizService) RemoveByCond(condition UserQDTO) (int64, error) { return 1, nil }

func (u *UserBizService) RemoveByMap(condition map[string]any) (int64, error) {
	return 1, nil
}
