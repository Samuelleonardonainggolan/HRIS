package models

// UserResponse - ringkasan data user untuk response
type UserResponse struct {
    ID         string `json:"id"`
    NIK        string `json:"nik"`
    Email      string `json:"email"`
    FullName   string `json:"full_name"`
    Role       string `json:"role"`
    Department string `json:"department"`
    Position   string `json:"position"`
    Avatar     string `json:"avatar,omitempty"`
}
