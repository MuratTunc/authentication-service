package errors

const (
	// Request Errors
	ErrInvalidRequest = "Invalid request body from UI"

	// Authentication & User Errors
	ErrUserNotFound    = "User-mail address not found! Please check your mail address or Signup"
	ErrInvalidPassword = "The password entered is incorrect. Invalid credentials"
	ErrTokenFailure    = "Failed to generate JWT token"
	ErrTokenMissing    = "Authorization token is missing"
	ErrTokenInvalid    = "Invalid or expired JWT token"

	// Database Errors
	ErrDatabaseConnection = "Failed to connect to the database"
	ErrDataNotFound       = "Requested data not found in the database"
	ErrDataInsertFailed   = "Failed to insert data into the database"

	// Server/Internal Errors
	ErrInternalServer = "Internal server error"
	ErrServiceDown    = "Dependent service is currently unavailable"
	ErrDatabaseInsert = "Could not insert bew user to DATABASE"
	ErrDatabaseQuery  = "In CheckUserExists function DatabaseQuery error"

	// Login Errors
	ErrInvalidCredentails  = "Invalid Login credentials"
	ErrFailedDecodingLogin = "Failed to decode login request: "

	// Logout Errors
	ErrAuthorizationHeader  = "Authorization header missing"
	ErrAuthorizationInvalid = "Invalid Authorization header format"
	ErrFailedToLogout       = "Failed to logout user"

	// JWT Errors
	ErrUnexpectedSigningMethodJWT = "unexpected signing method"
	ErrInvalidJWT                 = "InvalidJWTToken"
	ErrMissingUserIdJWT           = "User_id claim missing in token"
	ErrMissingUserIdIsNotStrJWT   = "User_id claim is not a string"
	ErrFailedToGenerateJWT        = "Failed to generate JWT token"
)
