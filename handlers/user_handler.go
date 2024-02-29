package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// User struct to represent a user
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Database to store users (simulated in-memory database)
var database = struct {
	users map[int]User
	mu    sync.RWMutex
}{users: make(map[int]User)}

// Function to generate a unique ID for a new user
func generateID() int {
	database.mu.Lock()
	defer database.mu.Unlock()
	return len(database.users) + 1
}

// Function to create a new user
func CreateUser(username, password string) (User, error) {
	newUser := User{
		ID:       generateID(),
		Username: username,
		Password: password,
	}

	database.mu.Lock()
	defer database.mu.Unlock()
	database.users[newUser.ID] = newUser

	return newUser, nil
}

// Function to get paginated users
func getPaginatedUsers(page, pageSize int) []User {
	database.mu.RLock()
	defer database.mu.RUnlock()

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	users := make([]User, 0, len(database.users))
	for _, user := range database.users {
		users = append(users, user)
	}

	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > len(users) {
		endIndex = len(users)
	}

	return users[startIndex:endIndex]
}

// Function to get a user by ID
func getUserByID(userID int) (User, bool) {
	database.mu.RLock()
	defer database.mu.RUnlock()

	user, exists := database.users[userID]
	return user, exists
}

// Function to update a user
func updateUser(userID int, updatedUser User) error {
	database.mu.Lock()
	defer database.mu.Unlock()

	if _, exists := database.users[userID]; !exists {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	database.users[userID] = updatedUser
	return nil
}

// Function to delete a user
func deleteUser(userID int) error {
	database.mu.Lock()
	defer database.mu.Unlock()

	if _, exists := database.users[userID]; !exists {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	delete(database.users, userID)
	return nil
}

// Function to check if provided credentials are valid
func IsValidCredentials(username, password string) bool {
	database.mu.RLock()
	defer database.mu.RUnlock()

	for _, user := range database.users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

// Handler for user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Placeholder implementation for user login
	// You might want to implement your actual login logic here
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login successful!")
}

// Handler for creating a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use Goroutine to prevent race condition
	go func() {
		createdUser, err := CreateUser(newUser.Username, newUser.Password)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdUser)
	}()
}

// Handler for getting paginated users
func GetPaginatedUsersHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	// Use Goroutine to allow concurrent paginated user retrieval
	go func() {
		users := getPaginatedUsers(page, pageSize)
		json.NewEncoder(w).Encode(users)
	}()
}

// Handler for getting a user by ID
func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
	user, exists := getUserByID(userID)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// Handler for updating a user
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
	var updatedUser User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use Goroutine to prevent race condition
	go func() {
		err := updateUser(userID, updatedUser)
		if err != nil {
			log.Printf("Error updating user: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}()
}

// Handler for deleting a user
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)

	// Use Goroutine to prevent race condition
	go func() {
		err := deleteUser(userID)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}()
}

// Function to extract user ID from request path
func getUserIDFromRequest(r *http.Request) int {
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
	}
	return userID
}
