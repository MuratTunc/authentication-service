package handler

import (
	"authentication-service/src/config"
	errors "authentication-service/src/error"
	"authentication-service/src/models"
	mailer "authentication-service/src/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	App *config.Config
}

// NewHandler creates a new Handler instance
func NewHandler(app *config.Config) *Handler {
	app.Logger.Info("✅ Handler initialized successfully")
	return &Handler{App: app}
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now() // Start the timer for request processing duration

	errorOccurred := false // Change this to simulate an error

	if errorOccurred {
		http.Error(w, errors.ErrInternalServer, http.StatusInternalServerError) // Using the error constant
		h.App.Logger.Error(errors.ErrInternalServer)
		return
	}

	// If everything is fine, return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication Service is up!"))

	// Calculate the time taken for the request to be processed
	duration := time.Since(startTime)

	// Log the success after the HTTP response
	h.App.Logger.Info("HealthCheck passed Duration: " + duration.String())
}

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	input, err := parseRegisterRequest(r)
	if err != nil {
		http.Error(w, errors.ErrInvalidRequest, http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode register request: " + err.Error())
		return
	}

	// 🔠 Normalize email
	input.MailAddress = strings.ToLower(strings.TrimSpace(input.MailAddress))

	// Check if user exists before inserting
	exists, err := checkUserExists(h.App.DB, input.Username, input.MailAddress)
	if err != nil {
		http.Error(w, errors.ErrDatabaseQuery, http.StatusInternalServerError)
		h.App.Logger.Error("Failed to check existing user: " + err.Error())
		return
	}
	if exists {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	// 🔐 Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to hash password: " + err.Error())
		return
	}
	input.Password = string(hashedPassword)

	// ✅ Insert the user with hashed password
	if err := insertUser(h.App.DB, input); err != nil {
		http.Error(w, errors.ErrDatabaseInsert, http.StatusInternalServerError)
		h.App.Logger.Error("Failed to insert user: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
	h.App.Logger.Info("User registered successfully in " + time.Since(startTime).String())
}

func (h *Handler) GetLastUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	query := `SELECT id, username, mail_address, role, activated, created_at FROM users ORDER BY created_at DESC LIMIT 1`

	row := h.App.DB.QueryRow(query)

	var user models.User

	err := row.Scan(&user.ID, &user.Username, &user.MailAddress, &user.Role, &user.Activated, &user.CreatedAt)
	if err != nil {
		h.App.Logger.Error("Failed to fetch last user: " + err.Error())
		http.Error(w, "Failed to retrieve last user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.App.Logger.Error("Failed to encode user response: " + err.Error())
	}

	h.App.Logger.Info("Last user fetched successfully in " + time.Since(startTime).String())
}

func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	input, err := parseLoginRequest(r)
	if err != nil {
		http.Error(w, errors.ErrInvalidRequest, http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode login request: " + err.Error())
		return
	}

	userID, hashedPwd, role, err := fetchUserCredentials(h.App.DB, input.MailAddress)
	if err != nil {
		handleLoginDBError(err, input.MailAddress, w, h.App.Logger)
		return
	}

	if !validatePassword(hashedPwd, input.Password) {
		http.Error(w, errors.ErrInvalidRequest, http.StatusUnauthorized)
		h.App.Logger.Error("Invalid password for user: " + input.MailAddress)
		return
	}

	tokenString, err := generateJWT(userID, role, h.App.JWTSecret, h.App.JWTExpiration)
	if err != nil {
		http.Error(w, errors.ErrFailedToGenerateJWT, http.StatusInternalServerError)
		h.App.Logger.Error("Failed to sign JWT: " + err.Error())
		return
	}

	h.App.Logger.Info("✅ User logged in successfully in " + time.Since(startTime).String())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"role":  role,
	})
}

func (h *Handler) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get the token from Authorization header (Bearer token)
	userID, _, err := GetValidatedUserIDRole(r, h.App.DB, h.App.JWTSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("Logout failed: " + err.Error())
		return
	}

	// Update login_status to false
	err = updateLoginStatus(h.App.DB, userID, false)
	if err != nil {
		http.Error(w, errors.ErrFailedToLogout, http.StatusInternalServerError)
		h.App.Logger.Error("Logout DB update failed: " + err.Error())
		return
	}

	h.App.Logger.Info("User logged out successfully in " + time.Since(startTime).String())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

func (h *Handler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get the token from Authorization header (Bearer token)
	userID, _, err := GetValidatedUserIDRole(r, h.App.DB, h.App.JWTSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("Logout failed: " + err.Error())
		return
	}

	// Fetch user role
	role, err := fetchUserRole(h.App.DB, userID)
	if err != nil {
		http.Error(w, "Failed to fetch user role", http.StatusInternalServerError)
		h.App.Logger.Error("DB error fetching role: " + err.Error())
		return
	}

	// Generate new token including role
	newToken, err := generateJWT(userID, role, h.App.JWTSecret, h.App.JWTExpiration)
	if err != nil {
		http.Error(w, errors.ErrTokenFailure, http.StatusInternalServerError)
		h.App.Logger.Error("Failed to generate new JWT token: " + err.Error())
		return
	}

	h.App.Logger.Info("Token refreshed successfully in " + time.Since(startTime).String())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": newToken})
}

func (h *Handler) GetUserByMailAddressHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get mail address from query parameters
	mail := r.URL.Query().Get("mail_address")
	if mail == "" {
		http.Error(w, "Missing mail_address query parameter", http.StatusBadRequest)
		return
	}

	query := `SELECT id, username, mail_address, role, activated, created_at FROM users WHERE mail_address = $1 LIMIT 1`
	row := h.App.DB.QueryRow(query, mail)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.MailAddress, &user.Role, &user.Activated, &user.CreatedAt)
	if err != nil {
		h.App.Logger.Error("User not found: " + err.Error())
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.App.Logger.Error("Failed to encode user response: " + err.Error())
	}

	h.App.Logger.Info("User fetched by mail_address successfully in " + time.Since(startTime).String())
}

func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Validate JWT and get userID + role from token
	userID, role, err := GetValidatedUserIDRole(r, h.App.DB, h.App.JWTSecret)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("UpdateUser failed: " + err.Error())
		return
	}

	// Parse request body
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.App.Logger.Error("Failed to decode request body: " + err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.MailAddress == "" {
		http.Error(w, "mail_address is required", http.StatusBadRequest)
		return
	}

	// Authorization logic:
	// Admin can update any user
	// Non-admin user can only update their own data (mail_address must match)
	if role != "Admin" {
		// Fetch mail address of logged-in user by userID from DB
		var loggedInMail string
		err = h.App.DB.QueryRow("SELECT mail_address FROM users WHERE id = $1", userID).Scan(&loggedInMail)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.App.Logger.Error("Failed to fetch logged-in user mail: " + err.Error())
			return
		}

		if loggedInMail != req.MailAddress {
			http.Error(w, "Forbidden: cannot update other user's data", http.StatusForbidden)
			h.App.Logger.Warn("User " + loggedInMail + " attempted unauthorized update on " + req.MailAddress)
			return
		}
	}

	// Build and execute update query
	queryResult := BuildUpdateUserQuery(req.Username, req.Role, req.Activated, req.MailAddress)
	if queryResult.HasError {
		http.Error(w, queryResult.ErrorMsg, http.StatusBadRequest)
		return
	}

	result, err := h.App.DB.Exec(queryResult.Query, queryResult.Args...)
	if err != nil {
		h.App.Logger.Error("Failed to execute update: " + err.Error())
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No user found with the given mail address", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User updated successfully"))
	h.App.Logger.Info("User updated successfully in " + time.Since(startTime).String())
}

func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Validate JWT and get requesting user's role WITHOUT DB check
	_, role, err := GetValidatedUserRoleOnly(r, h.App.JWTSecret)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	if role != "Admin" {
		http.Error(w, "Forbidden: only admin can delete users", http.StatusForbidden)
		return
	}

	var req models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.MailAddress == "" {
		http.Error(w, "Missing mail_address", http.StatusBadRequest)
		return
	}

	// Delete user by mail_address directly, no pre-check
	query := `DELETE FROM users WHERE mail_address = $1`
	res, err := h.App.DB.Exec(query, req.MailAddress)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking deletion result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}

func (h *Handler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get the token from Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, errors.ErrAuthorizationHeader, http.StatusUnauthorized)
		h.App.Logger.Error("Logout failed: missing Authorization header")
		return
	}

	var tokenStr string
	fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)
	if tokenStr == "" {
		http.Error(w, errors.ErrAuthorizationInvalid, http.StatusUnauthorized)
		h.App.Logger.Error("Logout failed: invalid Authorization header format")
		return
	}

	// Parse user ID from token
	userIDStr, err := parseUserIDFromJWT(tokenStr, h.App.JWTSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		h.App.Logger.Error("JWT parse error: " + err.Error())
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		h.App.Logger.Error("User ID conversion error: " + err.Error())
		return
	}

	// 🔎 Step 2: Validate user ID
	isValid, err := isUserIDValid(h.App.DB, userID)
	if err != nil {
		http.Error(w, "Error validating user ID", http.StatusInternalServerError)
		h.App.Logger.Error("User validation error: " + err.Error())
		return
	}
	if !isValid {
		http.Error(w, "User not found", http.StatusUnauthorized)
		h.App.Logger.Warn("Password change attempt with invalid user ID")
		return
	}

	// 🔄 Step 3: Parse request body
	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode change password request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(req.MailAddress)

	// Check if mail address matches the authenticated user
	var dbMailAddress, hashedPassword string
	query := `SELECT mail_address, password FROM users WHERE id = $1`
	err = h.App.DB.QueryRow(query, userID).Scan(&dbMailAddress, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		h.App.Logger.Error("Failed to fetch user for password change: " + err.Error())
		return
	}

	if dbMailAddress != req.MailAddress {
		http.Error(w, "Email mismatch", http.StatusUnauthorized)
		h.App.Logger.Warn("Email in request does not match JWT user")
		return
	}

	// 🔐 Step 4: Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Old password is incorrect", http.StatusUnauthorized)
		h.App.Logger.Warn("Incorrect old password attempt for user ID " + userIDStr)
		return
	}

	// 🔐 Step 5: Hash new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		h.App.Logger.Error("Password hashing error: " + err.Error())
		return
	}

	// 📝 Step 6: Update password
	updateQuery := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = h.App.DB.Exec(updateQuery, newHashedPassword, userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to update password in DB: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password changed successfully"))
	h.App.Logger.Info("Password changed for user ID " + userIDStr + " in " + time.Since(startTime).String())
}

func (h *Handler) SendMailResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode forgot password request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(strings.TrimSpace(req.MailAddress))
	if !strings.Contains(req.MailAddress, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	var userID int
	err := h.App.DB.QueryRow(`SELECT id FROM users WHERE mail_address = $1`, req.MailAddress).Scan(&userID)

	if err == sql.ErrNoRows {
		// User not found, respond with generic success message (no info leak)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("If the email exists, a reset link/code has been sent."))
		h.App.Logger.Info("Forgot password requested for non-existing email: " + req.MailAddress)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		h.App.Logger.Error("Database query failed: " + err.Error())
		return
	}

	// User found, generate reset code and update DB
	resetCode, err := generateResetCode()
	if err != nil {
		h.App.Logger.Error("Error generating reset code: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = h.UpdateResetCode(req.MailAddress, resetCode)
	if err != nil {
		h.App.Logger.Error("Failed to update reset code: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func() {
		err := mailer.SendPasswordResetMail(req.MailAddress, resetCode, h.App)
		if err != nil {
			h.App.Logger.Error("Failed to send password reset email to: " + req.MailAddress + " " + err.Error())
		} else {
			h.App.Logger.Info("Password reset email sent to " + req.MailAddress)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("If the email exists, a reset link/code has been sent."))
	h.App.Logger.Info("Forgot password process completed in " + time.Since(startTime).String())
}

func (h *Handler) VerifyResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.VerifyResetCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode verify reset code request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(strings.TrimSpace(req.MailAddress))
	req.ResetCode = strings.TrimSpace(req.ResetCode)

	if req.MailAddress == "" || req.ResetCode == "" {
		http.Error(w, "mail_address and reset_code are required", http.StatusBadRequest)
		return
	}

	// Check if user exists and resetcode matches
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE mail_address = $1 AND resetcode = $2)`
	err := h.App.DB.QueryRow(checkQuery, req.MailAddress, req.ResetCode).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to query for reset code verification: " + err.Error())
		return
	}

	if !exists {
		http.Error(w, "Invalid mail address or reset code", http.StatusUnauthorized)
		return
	}

	// Update reset_verified = true
	updateQuery := `UPDATE users SET reset_verified = true, updated_at = CURRENT_TIMESTAMP WHERE mail_address = $1`
	_, err = h.App.DB.Exec(updateQuery, req.MailAddress)
	if err != nil {
		http.Error(w, "Failed to update reset_verified flag", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to update reset_verified: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset code verified successfully"))
	h.App.Logger.Info("Reset code verified for " + req.MailAddress + " in " + time.Since(startTime).String())
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode reset password request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(req.MailAddress)

	// Check if reset_verified is true
	var resetVerified bool
	query := `SELECT reset_verified FROM users WHERE mail_address = $1`
	err := h.App.DB.QueryRow(query, req.MailAddress).Scan(&resetVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		h.App.Logger.Error("Failed to fetch reset_verified status: " + err.Error())
		return
	}

	if !resetVerified {
		http.Error(w, "Reset code not verified", http.StatusUnauthorized)
		h.App.Logger.Warn("Password reset attempted without verification for: " + req.MailAddress)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to hash new password: " + err.Error())
		return
	}

	// Update password and reset reset_verified to false
	updateQuery := `UPDATE users SET password = $1, reset_verified = false, updated_at = CURRENT_TIMESTAMP WHERE mail_address = $2`
	_, err = h.App.DB.Exec(updateQuery, hashedPassword, req.MailAddress)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to update password in DB: " + err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password reset successfully! User can use new password."))
	h.App.Logger.Info("Password reset successfully for " + req.MailAddress + " in " + time.Since(startTime).String())
}

func (h *Handler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 🔐 Extract claims from JWT
	claims, err := ExtractClaimsFromRequest(r, h.App.JWTSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("Failed to extract JWT claims: " + err.Error())
		return
	}

	// 🛡️ Check if user role is Admin
	role, ok := claims["role"].(string)
	if !ok || role != "Admin" {
		http.Error(w, "Access denied: Admins only", http.StatusForbidden)
		h.App.Logger.Warn("Unauthorized access attempt to list users")
		return
	}

	// Query all users from the DB
	rows, err := h.App.DB.Query(`SELECT id, username, mail_address, role, activated, login_status, created_at, updated_at FROM users`)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to query users: " + err.Error())
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var u models.UserResponse
		err := rows.Scan(&u.ID, &u.Username, &u.MailAddress, &u.Role, &u.Activated, &u.LoginStatus, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			http.Error(w, "Error scanning user data", http.StatusInternalServerError)
			h.App.Logger.Error("Failed to scan user row: " + err.Error())
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
	h.App.Logger.Info("✅ Admin listed all users in " + time.Since(startTime).String())
}

func (h *Handler) DeactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// ✅ Get userID and role from JWT WITHOUT DB check
	userID, role, err := GetValidatedUserRoleOnly(r, h.App.JWTSecret)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("Deactivation failed: " + err.Error())
		return
	}

	var req models.DeactivateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MailAddress == "" {
		http.Error(w, "Invalid request body or missing mail_address", http.StatusBadRequest)
		h.App.Logger.Error("Deactivation failed: invalid JSON or missing mail_address - " + err.Error())
		return
	}

	if role != "Admin" {
		http.Error(w, "Forbidden: only Admins can deactivate users", http.StatusForbidden)
		h.App.Logger.Error(fmt.Sprintf("User %d (role: %s) unauthorized to deactivate user with mail %s", userID, role, req.MailAddress))
		return
	}

	query := `UPDATE users SET activated = false, updated_at = CURRENT_TIMESTAMP WHERE mail_address = $1`
	res, err := h.App.DB.Exec(query, req.MailAddress)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		h.App.Logger.Error(fmt.Sprintf("DB error during deactivation of user %s: %v", req.MailAddress, err))
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		h.App.Logger.Error(fmt.Sprintf("Deactivation failed: user with mail %s not found", req.MailAddress))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deactivated successfully"))
	h.App.Logger.Info(fmt.Sprintf("User %d (role: %s) deactivated user with mail %s in %v", userID, role, req.MailAddress, time.Since(startTime)))
}

func (h *Handler) ReactivateUserHandler(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	// ✅ Get userID and role from JWT WITHOUT DB check
	_, role, err := GetValidatedUserRoleOnly(r, h.App.JWTSecret)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		h.App.Logger.Error("Deactivation failed: " + err.Error())
		return
	}

	if role != "Admin" {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Parse request body for mail_address
	var req models.ReactivateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MailAddress == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Perform reactivation
	query := `UPDATE users SET activated = true, updated_at = CURRENT_TIMESTAMP WHERE mail_address = $1`
	res, err := h.App.DB.Exec(query, req.MailAddress)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "No user found to reactivate", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User reactivated successfully!"))
	h.App.Logger.Info(" User reactivated successfully " + req.MailAddress + " in " + time.Since(startTime).String())
}

func (h *Handler) CheckMailExistHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.CheckMailAddressRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Invalid request body or missing mail_address: " + err.Error())
		return
	}

	// Query user existence
	query := `SELECT 1 FROM users WHERE mail_address = $1 LIMIT 1`
	row := h.App.DB.QueryRow(query, req.MailAddress)

	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		// User does not exist
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Email does not exist in DB."))
		h.App.Logger.Info("Checked non-existent email: " + req.MailAddress + " in " + time.Since(startTime).String())
		return
	} else if err != nil {
		// DB error
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		h.App.Logger.Error("DB error while checking email " + req.MailAddress + ": " + err.Error())
		return
	}

	// User exists
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email exists in DB."))
	h.App.Logger.Info("Checked existing email: " + req.MailAddress + " in " + time.Since(startTime).String())
}

func (h *Handler) GenerateAuthCodeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.GenerateAuthCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode verify mail address request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(strings.TrimSpace(req.MailAddress))
	if !strings.Contains(req.MailAddress, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Insert new mail address if not exists
	_, err := h.App.DB.Exec(`
		INSERT INTO users (mail_address, activated, created_at, updated_at)
		VALUES ($1, false, NOW(), NOW())
		ON CONFLICT (mail_address) DO NOTHING
	`, req.MailAddress)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		h.App.Logger.Error("Failed to insert user: " + err.Error())
		return
	}

	authCode, err := generateAuthCode()
	if err != nil {
		h.App.Logger.Error("Error generating auth code: " + err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = h.UpdateAuthenticationCode(req.MailAddress, authCode)
	if err != nil {
		h.App.Logger.Error("Failed to update authentication code: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func() {
		err := mailer.SendAuthenticationCode(req.MailAddress, authCode, h.App)
		if err != nil {
			h.App.Logger.Error("Failed to send verification email to: " + req.MailAddress + " " + err.Error())
		} else {
			h.App.Logger.Info("Verification email sent to " + req.MailAddress)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("If the email exists, a verification code has been sent."))
	h.App.Logger.Info("Mail verification process completed in " + time.Since(startTime).String())
}

func (h *Handler) VerifyAuthCodeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req models.VerifyAuthCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		h.App.Logger.Error("Failed to decode auth code check request: " + err.Error())
		return
	}

	req.MailAddress = strings.ToLower(strings.TrimSpace(req.MailAddress))
	req.AuthenticationCode = strings.TrimSpace(req.AuthenticationCode)

	if req.MailAddress == "" || req.AuthenticationCode == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	var storedCode string
	err := h.App.DB.QueryRow(
		`SELECT authentication_code FROM users WHERE mail_address = $1`,
		req.MailAddress,
	).Scan(&storedCode)

	if err == sql.ErrNoRows {
		http.Error(w, "Email not found", http.StatusNotFound)
		h.App.Logger.Warn("CheckAuthCode: mail not found " + req.MailAddress)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		h.App.Logger.Error("CheckAuthCode: DB error - " + err.Error())
		return
	}

	if storedCode != req.AuthenticationCode {
		http.Error(w, "Authentication code mismatch", http.StatusUnauthorized)
		h.App.Logger.Info("CheckAuthCode: code mismatch for " + req.MailAddress)
		return
	}

	// ✅ Mark user as activated
	_, err = h.App.DB.Exec(`UPDATE users SET activated = true WHERE mail_address = $1`, req.MailAddress)
	if err != nil {
		http.Error(w, "Failed to update user activation", http.StatusInternalServerError)
		h.App.Logger.Error("CheckAuthCode: failed to update activated status - " + err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication code verified and user activated."))
	h.App.Logger.Info("CheckAuthCode: verified and activated " + req.MailAddress + " in " + time.Since(startTime).String())
}
