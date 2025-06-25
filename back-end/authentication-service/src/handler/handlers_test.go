package handler_test

import (
	"authentication-service/src/config"
	"authentication-service/src/handler"
	"authentication-service/src/logger"
	"authentication-service/src/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealthCheckHandler(t *testing.T) {
	// Create a minimal config.Config
	cfg := &config.Config{
		Logger: logger.NewLogger(logger.INFO),
		// Add other fields if needed, e.g., DB
	}

	h := handler.Handler{
		App: cfg,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheckHandler(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d but got %d", http.StatusOK, res.StatusCode)
	}
}

func TestRegisterUserHandler_Success(t *testing.T) {
	// Step 1: Set up SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %s", err)
	}
	defer db.Close()

	// Step 2: Set up logger and config
	cfg := &config.Config{
		Logger: logger.NewLogger(logger.INFO),
		DB:     db,
	}
	h := &handler.Handler{App: cfg}

	// Step 3: Mock expectations
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username=\$1 OR mail_address=\$2\)`).
		WithArgs("testuser", "test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("testuser", "test@example.com", sqlmock.AnyArg(), "user", "1234567890", "en").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Step 4: Create request
	user := models.UserRequest{
		Username:           "testuser",
		MailAddress:        "test@example.com",
		Password:           "password123",
		Role:               "user",
		PhoneNumber:        "1234567890",
		LanguagePreference: "en",
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Step 5: Call the handler
	h.RegisterUserHandler(w, req)

	// Step 6: Assert results
	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
	}
}

func TestRegisterUserHandler_UserExists(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cfg := &config.Config{
		Logger: logger.NewLogger(logger.INFO),
		DB:     db,
	}
	h := &handler.Handler{App: cfg}

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username=\$1 OR mail_address=\$2\)`).
		WithArgs("testuser", "test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	user := models.UserRequest{
		Username:    "testuser",
		MailAddress: "test@example.com",
		Password:    "password123",
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	if w.Result().StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict, got %d", w.Result().StatusCode)
	}
}

func TestRegisterUserHandler_InsertFails(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	cfg := &config.Config{
		Logger: logger.NewLogger(logger.INFO),
		DB:     db,
	}
	h := &handler.Handler{App: cfg}

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE username=\$1 OR mail_address=\$2\)`).
		WithArgs("testuser", "test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs("testuser", "test@example.com", sqlmock.AnyArg(), "user", "1234567890", "en").
		WillReturnError(fmt.Errorf("insert failed"))

	user := models.UserRequest{
		Username:           "testuser",
		MailAddress:        "test@example.com",
		Password:           "password123",
		Role:               "user",
		PhoneNumber:        "1234567890",
		LanguagePreference: "en",
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 InternalServerError, got %d", w.Result().StatusCode)
	}
}

func TestRegisterUserHandler_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		Logger: logger.NewLogger(logger.INFO),
		DB:     nil, // DB won’t be used here
	}
	h := &handler.Handler{App: cfg}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{invalid}`)))
	w := httptest.NewRecorder()

	h.RegisterUserHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 BadRequest, got %d", w.Result().StatusCode)
	}
}
