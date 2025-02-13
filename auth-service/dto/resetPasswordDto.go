package dto

type ResetPasswordDto struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
