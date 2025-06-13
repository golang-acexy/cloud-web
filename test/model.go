package test

import "github.com/golang-acexy/cloud-web/webcloud"

// User 映射数据库
type User struct {
	ID        uint64 `json:"id"`
	ClassName string `json:"className"`
}

type AuthorityUser[ID uint64] struct {
	id uint64
}

func (a AuthorityUser[ID]) GetIdentityID() uint64 {
	return a.id
}

func (a AuthorityUser[ID]) GetPlatformID() webcloud.Platform {
	//TODO implement me
	panic("implement me")
}

type UserSDTO struct {
}

type UserMDTO struct {
}

type UserQDTO struct {
}

type UserDTO struct {
	User
}
