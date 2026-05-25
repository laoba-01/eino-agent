package logic

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"

	"github.com/redis/go-redis/v9"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *LoginLogic) Login(in *pb.LoginRequest) (*pb.LoginResponse, error) {
	usernameKey := "user:" + in.Username
	userJSON, err := l.svcCtx.Redis.Get(l.ctx, usernameKey).Result()
	if err != nil {
		if err == redis.Nil {
			return &pb.LoginResponse{Success: false, Message: "User not found"}, nil
		}
		return &pb.LoginResponse{Success: false, Message: "Database error"}, err
	}

	var user userInfo
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return &pb.LoginResponse{Success: false, Message: "Failed to parse user data"}, err
	}

	if !checkPassword(in.Password, user.Password) {
		return &pb.LoginResponse{Success: false, Message: "Invalid password"}, nil
	}

	cfg := l.svcCtx.Config
	tokenDuration := time.Duration(cfg.JWT.Expire) * time.Second
	token, err := generateToken(l.svcCtx.Redis, cfg.JWT.Secret, tokenDuration, user.UserID)
	if err != nil {
		return &pb.LoginResponse{Success: false, Message: "Failed to generate token"}, err
	}

	log.Printf("User logged in: %s (ID: %s)", in.Username, user.UserID)

	return &pb.LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		UserId:  user.UserID,
	}, nil
}
