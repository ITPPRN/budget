package models

// request

type RegisterKCReq struct {
	Username  string   `json:"username" example:"test1"`
	Password  string   `json:"password" example:"test1"`
	Email     string   `json:"email" example:"test@example.com"`
	FirstName string   `json:"first_name" example:"test1"`
	LastName  string   `json:"last_name" example:"test1"`
	Roles     []string `json:"roles" example:"[\"employee\", \"manager\"]"`
	// Role      string `json:"role" example:"employee"`
}

type LoginReq struct {
	Username string `json:"username" example:"test1"`
	Password string `json:"password" example:"test1"`
}

type ChangePasswordReq struct {
	NewPassword string `json:"new_password" example:"test1111"`
}

// res/////////////////////////////////////////////////////////
type UserInfo struct {
	UserId   string   `json:"userId"`
	UserName string   `json:"userName"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Role     []string `json:"role"`
}

type ResponseError struct {
	Message    string `json:"message"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

type ResponseData struct {
	Message    string      `json:"message"`
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}
