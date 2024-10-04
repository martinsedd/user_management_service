package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"regexp"
	"time"
)

type RegistrationRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	DateOfBirth string `json:"date_of_birth"`
}

func RegisterUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegistrationRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if validationError := validateRegistrationInput(req); validationError != nil {
			http.Error(w, validationError.Error(), http.StatusBadRequest)
			return
		}

		if db == nil {
			log.Println("Database connection is nil")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", req.Email).Scan(&exists)
		if err != nil {
			log.Printf("Error querying database for email: %v\n", err)
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if exists {
			http.Error(w, "Email already in use", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userID := uuid.NewString()
		_, err = db.Exec(
			"INSERT INTO users (id, first_name, last_name, email, phone_number, password, date_of_birth, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			userID, req.FirstName, req.LastName, req.Email, req.PhoneNumber, hashedPassword, req.DateOfBirth, true, time.Now(), time.Now())
		if err != nil {
			http.Error(w, "Failed to insert user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
		if err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func validateRegistrationInput(req RegistrationRequest) error {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.PhoneNumber == "" || req.Password == "" || req.DateOfBirth == "" {
		return fmt.Errorf("all fields are required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	phoneRegex := regexp.MustCompile(`^\+1\(\d{3}\)\d{3}-\d{4}$`)
	if !phoneRegex.MatchString(req.PhoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return fmt.Errorf("invalid date of birth format, expected YYYY-MM-DD")
	}

	if !isAdult(dob) {
		return fmt.Errorf("user must be at least 18 years old")
	}

	return nil
}

func isAdult(dob time.Time) bool {
	now := time.Now()
	age := now.Year() - dob.Year()

	if now.YearDay() < dob.YearDay() {
		age--
	}

	return age >= 18
}
