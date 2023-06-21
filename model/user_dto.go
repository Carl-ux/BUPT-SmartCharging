package model

// DTO（Data Transfer Object）是一种数据传输对象，通常用于在不同层之间传输数据
type UserDTO struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// Path: model/user_dto.go
func ToUserDto(user User) UserDTO {
	return UserDTO{
		Name:  user.Name,
		Admin: user.Admin,
	}
}
