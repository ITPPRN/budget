package models

// request

type UserRegisReq struct{
	
}

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
    OldPassword     string `json:"old_password" validate:"required" example:"old_secret123"`
    NewPassword     string `json:"new_password" validate:"required,min=6" example:"new_secret123"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword" example:"new_secret123"`
}

type AdminResetPasswordReq struct {
	NewPassword string `json:"new_password" validate:"required,min=6"`
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

type UserRes struct{

}