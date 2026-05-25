package logic

import (
	"context"
	"encoding/json"
	"log"

	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"
)

type userInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterLogic) Register(in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	usernameKey := "user:" + in.Username
	exists, err := l.svcCtx.Redis.Exists(l.ctx, usernameKey).Result()
	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: "Database error"}, err
	}
	if exists > 0 {
		return &pb.RegisterResponse{Success: false, Message: "Username already exists"}, nil
	}

	hashedPassword, err := hashPassword(in.Password)
	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: "Failed to hash password"}, err
	}

	userID := generateUserID()
	user := userInfo{
		UserID:   userID,
		Username: in.Username,
		Password: hashedPassword,
		Email:    in.Email,
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: "Failed to create user"}, err
	}

	err = l.svcCtx.Redis.Set(l.ctx, "user:"+in.Username, userJSON, 0).Err()
	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: "Failed to store user"}, err
	}

	log.Printf("User registered: %s (ID: %s)", in.Username, userID)

	return &pb.RegisterResponse{
		Success: true,
		Message: "Registration successful",
		UserId:  userID,
	}, nil
}
