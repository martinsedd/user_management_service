package user

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestRegisterUserHandler_InvalidPayload(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := strings.NewReader(`{invalid_json}`)
	req := httptest.NewRequest(http.MethodPost, "/register", reqBody)
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid request payload") {
		t.Errorf("Expected error message for invalid payload, got %s", w.Body.String())
	}
}

func TestRegisterUserHandler_DatabaseNil(t *testing.T) {
	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validPassword",
		DateOfBirth: "1990-01-01",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal server error") {
		t.Errorf("Expected error message for database connection nil, got %s", w.Body.String())
	}
}

func TestRegisterUserHandler_InvalidEmail(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "invalid-email",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validpassword",
		DateOfBirth: "1990-01-01",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid email format") {
		t.Errorf("Expected invalid email format error, got %s", w.Body.String())
	}
}

// Test for invalid phone number format
func TestRegisterUserHandler_InvalidPhoneNumber(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "invalid-phone",
		Password:    "validpassword",
		DateOfBirth: "1990-01-01",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid phone number format") {
		t.Errorf("Expected invalid phone number format error, got %s", w.Body.String())
	}
}

// Test for invalid password length
func TestRegisterUserHandler_InvalidPasswordLength(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "short",
		DateOfBirth: "1990-01-01",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "password must be at least 8 characters long") {
		t.Errorf("Expected password length error, got %s", w.Body.String())
	}
}

// Test for invalid date of birth format
func TestRegisterUserHandler_InvalidDateOfBirth(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validpassword",
		DateOfBirth: "invalid-date",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "invalid date of birth format") {
		t.Errorf("Expected invalid date of birth format error, got %s", w.Body.String())
	}
}

// Test for user not being an adult
func TestRegisterUserHandler_UserNotAdult(t *testing.T) {
	db, _, _ := sqlmock.New()

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validpassword",
		DateOfBirth: "2010-01-01", // User too young
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "user must be at least 18 years old") {
		t.Errorf("Expected age validation error, got %s", w.Body.String())
	}
}

func TestRegisterUserHandler_EmailAlreadyExists(t *testing.T) {
	db, mock, _ := sqlmock.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)")).
		WithArgs("johndoe@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validPassword",
		DateOfBirth: "1990-01-01",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Email already in use") {
		t.Errorf("Expected error message for email already in use, got %s", w.Body.String())
	}
}

func TestRegisterUserHandler_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	// Removed defer db.Close() as it's unnecessary with sqlmock

	// Use regexp.QuoteMeta for flexible matching and mock email does not exist
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)")).
		WithArgs("johndoe@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false)) // Ensure email does not exist

	// Mock successful insert
	//goland:noinspection GoConvertStringLiterals
	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), "John", "Doe", "johndoe@example.com", "+1(123)456-7890", sqlmock.AnyArg(), "1990-01-01", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	reqBody := RegistrationRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "johndoe@example.com",
		PhoneNumber: "+1(123)456-7890",
		Password:    "validpassword",
		DateOfBirth: "1990-01-01",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := RegisterUserHandler(db)
	handler.ServeHTTP(w, req)

	// Expect status 201 (Created) instead of 200
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Expect success message
	if !strings.Contains(w.Body.String(), "User created successfully") {
		t.Errorf("Expected success message, got %s", w.Body.String())
	}
}
