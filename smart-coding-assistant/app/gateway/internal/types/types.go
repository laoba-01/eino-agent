package types

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserId  string `json:"user_id,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	UserId  string `json:"user_id,omitempty"`
}

type LogoutResponse struct {
	Success bool `json:"success"`
}

type ChatRequest struct {
	Message string            `json:"message"`
	Context map[string]string `json:"context"`
}

type ChatResponse struct {
	Response   string            `json:"response"`
	IsFinished bool              `json:"is_finished"`
	Context    map[string]string `json:"context"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
