package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// User 用户结构体
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// NewServer 创建并返回一个 http.Handler
func NewServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/users/", userHandler)
	mux.HandleFunc("/echo", echoHandler)

	return mux
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, World!"))
}

// userHandler 根据路径 /users/{id} 返回用户信息
func userHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	user := User{ID: 1, Name: "Alice", Age: 30}
	if id != "1" {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// echoHandler 回显 POST body
func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"echo": "received",
	})
}