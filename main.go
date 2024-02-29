// main.go

package main

import (
	"log"
	"net/http"
	"restfull_API_2/handlers"
	middleware "restfull_API_2/middlerware"

	"github.com/gorilla/mux"
)

func main() {
	// Create some initial users for testing
	handlers.CreateUser("user1", "password1")
	handlers.CreateUser("user2", "password2")

	// Setup routes
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Protected routes (require authentication)
	protectedRouter := r.PathPrefix("/api").Subrouter()

	// Apply authentication middleware to protected routes
	protectedRouter.Use(middleware.AuthenticateMiddleware)

	// Other protected routes
	protectedRouter.HandleFunc("/users", handlers.CreateUserHandler).Methods("POST")
	protectedRouter.HandleFunc("/users/paginated", handlers.GetPaginatedUsersHandler).Methods("GET")
	protectedRouter.HandleFunc("/users/{id:[0-9]+}", handlers.GetUserByIDHandler).Methods("GET")
	protectedRouter.HandleFunc("/users/{id:[0-9]+}", handlers.UpdateUserHandler).Methods("PUT")
	protectedRouter.HandleFunc("/users/{id:[0-9]+}", handlers.DeleteUserHandler).Methods("DELETE")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
