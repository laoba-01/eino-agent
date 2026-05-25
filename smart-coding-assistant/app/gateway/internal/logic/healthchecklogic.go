package logic

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthCheckLogic struct{}

func NewHealthCheckLogic() *HealthCheckLogic {
	return &HealthCheckLogic{}
}

func (l *HealthCheckLogic) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
