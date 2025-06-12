package test

// User 映射数据库
type User struct {
	ID        uint64 `json:"id"`
	ClassName string `json:"className"`
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
