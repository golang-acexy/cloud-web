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

func (a AuthorityUser[ID]) GetPlatform() webcloud.Platform {
	return "test"
}

type UserSDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserMDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserQDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserDTO struct {
	User
}
