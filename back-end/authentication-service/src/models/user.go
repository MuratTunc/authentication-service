package models

import "time"

type UserRequest struct {
	Username           string `json:"username"`
	MailAddress        string `json:"mail_address"`
	Password           string `json:"password"`
	PhoneNumber        string `json:"phone_number"`
	LanguagePreference string `json:"language_preference"`
	Role               string `json:"role"` // "Admin", "Sales Representative", "Customer"
}

type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	MailAddress string    `json:"mail_address"`
	Role        string    `json:"role"`
	Activated   bool      `json:"activated"`
	CreatedAt   time.Time `json:"created_at"`
}

type LoginRequest struct {
	MailAddress string `json:"mail_address"`
	Password    string `json:"password"`
}

type UpdateUserRequest struct {
	MailAddress string `json:"mail_address"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	Activated   *bool  `json:"activated"` // use pointer to allow optional field
}

type DeleteRequest struct {
	MailAddress string `json:"mail_address"`
}

type ReactivateUserRequest struct {
	MailAddress string `json:"mail_address"`
}

type DeactivateUserRequest struct {
	MailAddress string `json:"mail_address"`
}

type GenerateAuthCodeRequest struct {
	MailAddress string `json:"mail_address"`
}

type VerifyAuthCodeRequest struct {
	MailAddress        string `json:"mail_address"`
	AuthenticationCode string `json:"authentication_code"`
}

type CheckAuthCodeRequest struct {
	MailAddress        string `json:"mail_address"`
	AuthenticationCode string `json:"authentication_code"`
}

type ChangePasswordRequest struct {
	MailAddress string `json:"mail_address"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ForgotPasswordRequest struct {
	MailAddress string `json:"mail_address"`
}

type CheckMailAddressRequest struct {
	MailAddress string `json:"mail_address"`
}

type VerifyResetCodeRequest struct {
	MailAddress string `json:"mail_address"`
	ResetCode   string `json:"reset_code"`
}

type ResetPasswordRequest struct {
	MailAddress string `json:"mail_address"`
	NewPassword string `json:"new_password"`
}

type UserResponse struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	MailAddress string    `json:"mail_address"`
	Role        string    `json:"role"`
	Activated   bool      `json:"activated"`
	LoginStatus bool      `json:"login_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
