
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

const (
	port          = ":50054"
	jwtSecret     = "your-secret-key-change-in-production"
	tokenDuration = 24 * time.Hour
)

var rdb *redis.Client

type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AuthServer struct {
	UnimplementedAuthServiceServer
}

func generateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)[:16]
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	// Store token in Redis
	ctx := context.Background()
	err = rdb.Set(ctx, "token:"+tokenString, userID, tokenDuration).Err()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateToken(tokenString string) (string, bool) {
	ctx := context.Background()

	// Check if token exists in Redis
	userID, err := rdb.Get(ctx, "token:"+tokenString).Result()
	if err != nil {
		return "", false
	}

	// Parse and validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return "", false
			}
		}
		if uid, ok := claims["user_id"].(string); ok {
			return uid, true
		}
	}

	return userID, true
}

func (s *AuthServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Check if username already exists
	usernameKey := "user:" + req.Username
	exists, err := rdb.Exists(ctx, usernameKey).Result()
	if err != nil {
		return &RegisterResponse{
			Success: false,
			Message: "Database error",
		}, err
	}
	if exists > 0 {
		return &RegisterResponse{
			Success: false,
			Message: "Username already exists",
		}, nil
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return &RegisterResponse{
			Success: false,
			Message: "Failed to hash password",
		}, err
	}

	// Create user
	userID := generateUserID()
	user := UserInfo{
		UserID:   userID,
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return &RegisterResponse{
			Success: false,
			Message: "Failed to create user",
		}, err
	}

	// Store user in Redis
	err = rdb.Set(ctx, "user:"+req.Username, userJSON, 0).Err()
	if err != nil {
		return &RegisterResponse{
			Success: false,
			Message: "Failed to store user",
		}, err
	}

	log.Printf("User registered: %s (ID: %s)", req.Username, userID)

	return &RegisterResponse{
		Success: true,
		Message: "Registration successful",
		UserID:  userID,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user from Redis
	usernameKey := "user:" + req.Username
	userJSON, err := rdb.Get(ctx, usernameKey).Result()
	if err != nil {
		if err == redis.Nil {
			return &LoginResponse{
				Success: false,
				Message: "User not found",
			}, nil
		}
		return &LoginResponse{
			Success: false,
			Message: "Database error",
		}, err
	}

	// Parse user
	var user UserInfo
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Failed to parse user data",
		}, err
	}

	// Check password
	if !checkPassword(req.Password, user.Password) {
		return &LoginResponse{
			Success: false,
			Message: "Invalid password",
		}, nil
	}

	// Generate token
	token, err := generateToken(user.UserID)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		}, err
	}

	log.Printf("User logged in: %s (ID: %s)", req.Username, user.UserID)

	return &LoginResponse{
		Success: true,
		Message:  "Login successful",
		Token:    token,
		UserID:   user.UserID,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	userID, valid := validateToken(req.Token)
	if !valid {
		return &ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token",
		}, nil
	}

	return &ValidateTokenResponse{
		Valid:   true,
		UserID:  userID,
		Message: "Token is valid",
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	err := rdb.Del(ctx, "token:"+req.Token).Err()
	if err != nil {
		return &LogoutResponse{
			Success: false,
			Message: "Failed to logout",
		}, err
	}

	return &LogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}

func main() {
	// Connect to Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Start gRPC server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterAuthServiceServer(s, &AuthServer{})

	log.Printf("Auth Service listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}