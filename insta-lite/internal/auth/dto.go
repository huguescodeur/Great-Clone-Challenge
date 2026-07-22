package auth

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3" example:"huguesgoli"`
	Email    string `json:"email" validate:"required,email" example:"hugues@gmail.com"`
	FullName string `json:"fullName" validate:"required" example:"Goli yao Hugues"`
	Password string `json:"password" validate:"required,min=6" example:"pass1234"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}
