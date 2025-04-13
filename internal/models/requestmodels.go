package models

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateReceptionRequest struct {
	PVZId string `json:"pvzId"`
}

type AddProductRequest struct {
	Type  string `json:"type"`
	PVZId string `json:"pvzId"`
}
