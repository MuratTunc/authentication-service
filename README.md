# authentication-service

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-ready microservice for user authentication using **Go**, **PostgreSQL**, **Docker**, and **JWT**.  
Supports role-based access, secure password storage, and scalable deployment via Docker Compose or Kubernetes.

---

## 📦 Project Structure & Flow

```bash
                [ .env file ]
                    ↓
             [ docker-compose.yml ]
                    ↓
           ┌----------------------┐
           |     Build Phase      |
           |----------------------|
           | 1. Read Dockerfile   |
           | 2. Build Go binary   |
           | 3. Package image     |
           └----------------------┘
                    ↓
         [ Docker Image per service ]
                    ↓
         [ docker-compose up --build ]
                    ↓
         [ Containers running binaries ]
```



## ✅ Authentication Flow (Best Practice)
```bash

[ UI / Frontend ]
     ↓
POST /auth-service/login
Body: { mailAddress, password }

     ↓
[ Authentication Service ]
- Lookup user by mailAddress in DB
- Check password using bcrypt
- If valid:
    - Generate signed JWT
    - Token includes: { user_id, role, mail, exp }
    - Return: { success: true, token, user info }

     ↓
[ UI stores JWT securely ]
- localStorage or HttpOnly cookie
- Add JWT to all future requests:
  Authorization: Bearer <JWT>
```

## 🗃️ Database Design
### 📂 Related Files:
```bash
/src/sql/init_users_table.sql  
/src/database/connection.go  
/src/models/user.go
```

### users Table Schema

| Column               | Data Type       | Constraints / Description                                                   |
|----------------------|----------------|------------------------------------------------------------------------------|
| id                   | SERIAL          | Primary Key                                                                 |
| username             | VARCHAR(100)    | Not Null, Unique                                                            |
| mail_address         | VARCHAR(255)    | Not Null, Unique                                                            |
| password             | TEXT            | Not Null                                                                    |
| role                 | VARCHAR(50)     | Not Null, must be one of: `'Admin'`, `'Sales Representative'`, `'Customer'` |
| phone_number         | VARCHAR(20)     | Optional                                                                    |
| language_preference  | VARCHAR(10)     | Default `'en'`                                                              |
| resetcode            | VARCHAR(20)     | Optional                                                                    |
| reset_verified       | BOOLEAN         | Not Null, Default `false`                                                   |
| authentication_code  | VARCHAR(20)     | Optional                                                                    |
| activated            | BOOLEAN         | Not Null, Default `false`                                                   |
| login_status         | BOOLEAN         | Not Null, Default `false`                                                   |
| created_at           | TIMESTAMP       | Not Null, Default `CURRENT_TIMESTAMP`                                       |
| updated_at           | TIMESTAMP       | Not Null, Default `CURRENT_TIMESTAMP`                                       |


### 🧑‍💼 User Roles

| Role                | Permissions                                                                 |
|---------------------|------------------------------------------------------------------------------|
| `Admin`             | Full access: manage users, system settings, view/edit all data               |
| `Sales Representative` | Access customer info, update leads/orders, generate sales reports            |
| `Customer`          | View own profile/orders, update account settings, limited access to features |

##  🔐 JWT-Protected Endpoints

Some endpoints in the authentication service **require a valid JWT token** to ensure that the caller is authenticated and authorized. Whether you need to send a token depends on the **endpoint's access level** and **user role**.

### ✅ Endpoints that **require** JWT Token

| Endpoint                      | Method | Role Required | Description                               |
|-------------------------------|--------|----------------|------------------------------------------|
| `/auth/logout`                | POST   | Any Logged-in  | Invalidate a session                     |
| `/auth/refresh-jwt-token`     | POST   | Any Logged-in  | Refresh JWT token                        |
| `/auth/change-password`       | POST   | Any Logged-in  | Change password                          |
| `/auth/deactivate-user`       | POST   | Admin only     | Deactivate user account                  |
| `/auth/reactivate-user`       | POST   | Admin only     | Reactivate a user account                |
| `/auth/update-user`           | PUT    | Admin or User  | Update user info                         |
| `/auth/delete-user`           | DELETE | Admin only     | Permanently delete a user                |

## API Endpoints

Base_path: /api/v1/auth

| Method  | Endpoint                      | Description                                                                                       |
|---------|-------------------------------|-------------------------------------------------------------------------------------------------|
| GET     | Base_path/health            | Check if the authentication service is up and running.                                          |
| GET     | Base_path/last-user         | Retrieve the most recently registered user in the system.                                       |
| GET     | Base_path/get-user-by-mail  | Fetch user details using the user's email address.                                              |
| GET     | Base_path/list-users        | List all users registered in the system.                                                        |
| PUT     | Base_path/update-user       | Update existing user information (requires authenticated user).                                 |
| POST    | Base_path/register-user     | Register a new user. Rate limited to prevent abuse.                                             |
| POST    | Base_path/login             | Authenticate a user and return a JWT token. Rate limited.                                       |
| POST    | Base_path/send-mail-reset-code | Send a password reset code to the user's email. Rate limited.                                |
| POST    | Base_path/reset-password    | Reset user's password using the reset code. Rate limited.                                       |
| POST    | Base_path/generate-auth-code| Generate an authentication code for email verification or 2FA. Rate limited.                   |
| POST    | Base_path/verify-auth-code  | Verify the submitted authentication code. Rate limited.                                         |
| POST    | Base_path/logout            | Log out the current user by invalidating their session or token.                                |
| POST    | Base_path/refresh-jwt-token | Refresh the JWT token to extend session validity.                                              |
| POST    | Base_path/change-password   | Change the password for the authenticated user.                                                 |
| POST    | Base_path/deactivate-user   | Deactivate a user account (admin-only action).                                                  |
| POST    | Base_path/reactivate-user   | Reactivate a previously deactivated user account (admin-only action).                           |
| POST    | Base_path/check-mail-exists | Check if an email address is already registered in the system.                                  |
| POST    | Base_path/verify-mail-reset-code | Verify the password reset code sent to the user’s email.                                      |
| DELETE  | Base_path/delete-user       | Permanently delete a user account (admin-only action).                                          |

#### Notes:

Replace http://localhost with your actual server URL.

For endpoints requiring authorization, add the Authorization header accordingly.

Body data is JSON formatted and wrapped in single quotes '...' for shell compatibility.

| **Endpoint**                                                      | **Example cURL Command**                                                                                                                                                                                |
| ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET /api/v1/auth/health`                                         | `curl -X GET "http://localhost/api/v1/auth/health"`                                                                                                                                                     |
| `GET /api/v1/auth/last-user`                                      | `curl -X GET "http://localhost/api/v1/auth/last-user"`                                                                                                                                                  |
| `GET /api/v1/auth/get-user-by-mail?mail_address=user@example.com` | `curl -X GET "http://localhost/api/v1/auth/get-user-by-mail?mail_address=user@example.com"`                                                                                                             |
| `GET /api/v1/auth/list-users`                                     | `curl -X GET "http://localhost/api/v1/auth/list-users"`                                                                                                                                                 |
| `PUT /api/v1/auth/update-user`                                    | `curl -X PUT "http://localhost/api/v1/auth/update-user" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com","username":"newname","role":"Admin","activated":true}'`              |
| `POST /api/v1/auth/register-user`                                 | `curl -X POST "http://localhost/api/v1/auth/register-user" -H "Content-Type: application/json" -d '{"username":"newuser","mail_address":"newuser@example.com","password":"password123","role":"User"}'` |
| `POST /api/v1/auth/login`                                         | `curl -X POST "http://localhost/api/v1/auth/login" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com","password":"password123"}'`                                               |
| `POST /api/v1/auth/send-mail-reset-code`                          | `curl -X POST "http://localhost/api/v1/auth/send-mail-reset-code" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                         |
| `POST /api/v1/auth/reset-password`                                | `curl -X POST "http://localhost/api/v1/auth/reset-password" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com","reset_code":"123456","new_password":"newpass123"}'`             |
| `POST /api/v1/auth/generate-auth-code`                            | `curl -X POST "http://localhost/api/v1/auth/generate-auth-code" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                           |
| `POST /api/v1/auth/verify-auth-code`                              | `curl -X POST "http://localhost/api/v1/auth/verify-auth-code" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com","auth_code":"654321"}'`                                        |
| `POST /api/v1/auth/logout`                                        | `curl -X POST "http://localhost/api/v1/auth/logout"`                                                                                                                                                    |
| `POST /api/v1/auth/refresh-jwt-token`                             | `curl -X POST "http://localhost/api/v1/auth/refresh-jwt-token" -H "Authorization: Bearer <old_token>"`                                                                                                  |
| `POST /api/v1/auth/change-password`                               | `curl -X POST "http://localhost/api/v1/auth/change-password" -H "Content-Type: application/json" -d '{"old_password":"oldpass123","new_password":"newpass123"}'`                                        |
| `POST /api/v1/auth/deactivate-user`                               | `curl -X POST "http://localhost/api/v1/auth/deactivate-user" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                              |
| `POST /api/v1/auth/reactivate-user`                               | `curl -X POST "http://localhost/api/v1/auth/reactivate-user" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                              |
| `POST /api/v1/auth/check-mail-exists`                             | `curl -X POST "http://localhost/api/v1/auth/check-mail-exists" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                            |
| `POST /api/v1/auth/verify-mail-reset-code`                        | `curl -X POST "http://localhost/api/v1/auth/verify-mail-reset-code" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com","reset_code":"123456"}'`                                 |
| `DELETE /api/v1/auth/delete-user`                                 | `curl -X DELETE "http://localhost/api/v1/auth/delete-user" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com"}'`                                                                |



### `GET /health`

**Health check endpoint**

This endpoint is used to verify if the Authentication Service is running and healthy. It performs a basic internal check and returns a simple message.

#### 🔁 Request

- No parameters
- No authentication required
- This request does not require authentication or any parameters in the body or headers.

#### Response

- `200 OK`  
  Plain text: `"Authentication Service is up!"`

- `500 Internal Server Error`  
  Returned if something goes wrong internally


### `GET /last-user`

**Fetch the most recently created user**

This endpoint retrieves the latest user record from the database based on the `created_at` timestamp in descending order. It returns basic user information without sensitive data like passwords.

#### 🔁 Request

- No parameters
- No authentication required

#### Response

- `200 OK`  
  JSON object with the last user's details:

  ```json
  {
    "id": <int>,
    "username": "<string>",
    "mail_address": "<string>",
    "role": "<string>",
    "activated": <bool>,
    "created_at": "<timestamp>"
  }

- `500 Internal Server Error`
- If there is a failure querying the database or encoding the response.

#### 🪵 Logs
- On success, logs the time taken to fetch the user.
- On failure, logs the error message.


### `GET /get-user-by-mail`

**Fetch user details by email address**

This endpoint retrieves user information based on the `mail_address` provided as a query parameter. It returns basic user details without sensitive information like passwords.

#### 🔁 Request

- Query parameter:  
  `mail_address` (string) — **required**

- No authentication required

Example:  
GET /api/v1/auth/get-user-by-mail?mail_address=user@example.com


#### Response

- `200 OK`  
  JSON object with the user's details:

  ```json
  {
    "id": <int>,
    "username": "<string>",
    "mail_address": "<string>",
    "role": "<string>",
    "activated": <bool>,
    "created_at": "<timestamp>"
  }

- 400 Bad Request
- If the mail_address query parameter is missing.

- 404 Not Found
- If no user is found with the given mail address.

#### 🪵 Logs
- On success, logs the time taken to fetch the user by email.
- On failure, logs the error message.


### `GET /list-users`

**List all users (Admin only)**

This endpoint returns a list of all users in the system. Access is restricted to users with the `Admin` role.

#### 🔐 Authentication & Authorization

- Requires a valid JWT token in the request (usually in the `Authorization` header).
- Only users with the role `Admin` can access this endpoint.
- Returns `401 Unauthorized` if token is missing or invalid.
- Returns `403 Forbidden` if the user role is not `Admin`.

#### 🔁 Request

- No URL parameters or body required.
- JWT token required for authorization.

Example Header:  
Authorization: Bearer <jwt-token>

#### Response

- `200 OK`  
  JSON array of user objects:

  ```json
  [
    {
      "id": <int>,
      "username": "<string>",
      "mail_address": "<string>",
      "role": "<string>",
      "activated": <bool>,
      "login_status": <bool>,
      "created_at": "<timestamp>",
      "updated_at": "<timestamp>"
    },
    ...
  ]

- 401 Unauthorized
- If JWT token is missing or invalid.

- 403 Forbidden
- If the user is not an Admin.

- 500 Internal Server Error
- If database query or data scanning fails.

#### 🪵 Logs
- On success, logs the time taken for the admin to fetch all users.
- On unauthorized or forbidden access, logs warnings or errors accordingly.
- On failure, logs detailed error messages.

### `PUT /update-user`

**Update user information**

This endpoint allows updating user details such as username, role, and activation status.

#### 🔐 Authentication & Authorization

- Requires a valid JWT token.
- Admin users can update any user's data.
- Non-admin users can only update their own data, verified by matching the `mail_address` in the request with their own.

#### 🔁 Request

- Content-Type: `application/json`
- JSON body example:

```json
{
  "username": "newUsername",
  "role": "Admin",           // Optional, only admins should use this
  "activated": true,         // Optional
  "mail_address": "user@example.com"
}
```
mail_address is required to identify the user to update.

Other fields (username, role, activated) are optional and will be updated if provided.


Response

200 OK
Plain text: "User updated successfully"

400 Bad Request
If JSON body is invalid or missing required fields.

401 Unauthorized
If JWT is missing or invalid.

403 Forbidden
If a non-admin user tries to update another user's data.

404 Not Found
If no user matches the provided mail_address.

500 Internal Server Error
If database update fails or other server errors occur.


Authorization Logic
- Admin can update any user.

- Non-admin user can update only their own data (must match mail_address).

#### 🪵 Logs
- Logs errors for failed authentication, invalid payload, unauthorized attempts, and database errors.
- Logs successful update and time taken to process the request.


### `POST /register-user`

**Register a new user**

This endpoint registers a new user by inserting their data into the database after validating and hashing the password.

#### 🔁 Request

- Content-Type: `application/json`
- JSON body example:

```json
{
  "username": "johndoe",
  "mail_address": "john@example.com",
  "password": "securePassword123"
}
```

username, mail_address, and password are required.

#### 🔐 Security Measures
Email is normalized (lowercased + trimmed) before checking or inserting.

Passwords are hashed using bcrypt before storage.

Rate limiting is applied per IP using httprate, configured by:

r.Config.RateLimitRegister

r.Config.RateLimitWindowMinutes

✅ Success Response
201 Created
Plain text: "User registered successfully"

#### Error Responses
400 Bad Request
If the request body is missing or invalid.

409 Conflict
If the username or email is already registered.

500 Internal Server Error
If a database query, insert operation, or password hashing fails.

Logs
Logs an error if:

JSON parsing fails

User existence check fails

Password hashing fails

Insert operation fails

Logs an info message if the user is successfully registered, including the time taken for the operation.


### `POST /login`

**Authenticate a user and return a JWT**

This endpoint validates user credentials and returns a signed JWT token along with the user's role if authentication is successful.

---

#### 🔁 Request

- Content-Type: `application/json`
- JSON body example:

```json
{
  "mail_address": "john@example.com",
  "password": "securePassword123"
}
```

Both mail_address and password are required.

#### 🔐 Security Measures
Password is securely validated using bcrypt.CompareHashAndPassword.

Returns a JWT signed with the server’s secret key.

JWT contains: user_id, role, exp (expiry).

Rate limiting is applied per IP using httprate, configured by:

r.Config.RateLimitLogin

r.Config.RateLimitWindowMinutes

✅ Success Response
200 OK

Content-Type: application/json

Body:
```

{
  "token": "<JWT token>",
  "role": "Admin" // or "Sales Representative"
}
```


#### Error Responses
400 Bad Request
If the request body is missing or invalid.

401 Unauthorized
If the password is incorrect.

404 Not Found
If the user does not exist (depending on DB error returned).

500 Internal Server Error
If a database or JWT generation error occurs.

#### 🪵 Logs
Logs error if:

Request body decoding fails

User fetch from DB fails

Password validation fails

JWT signing fails

Logs info if the login is successful with the time taken.

#### 📌 Notes
The JWT token must be used in the Authorization header for protected endpoints:
Authorization: Bearer <token>

### `POST /send-mail-reset-code`

**Initiate password reset by sending a reset code to the user's email address.**

---

#### 🔁 Request

- Content-Type: `application/json`
- JSON body:

```json
{
  "mail_address": "user@example.com"
}
```
mail_address is required.
#### 🔐 Security Measures
Email is normalized (trimmed + lowercased) before processing.

No user existence is leaked in response (uniform response).

Email format is validated (basic @ check).

Rate-limited per IP using httprate:

r.Config.RateLimitResetCode

r.Config.RateLimitWindowMinutes

✅ Success Response
200 OK

Content-Type: text/plain

Body:

If the email exists, a reset link/code has been sent.

#### Error Responses
400 Bad Request
If the request body is malformed or email format is invalid.

500 Internal Server Error
If database access, reset code generation, or update fails.

#### 🛠️ Backend Flow
Parse and validate email.

Check user existence in the users table (by email).

If user found:

Generate reset code

Save reset code to DB via UpdateResetCode()

Send reset email via mailer.SendPasswordResetMail(...) (in background goroutine).

Always respond with a generic success message.

#### 🪵 Logging
Logs all errors (decoding, DB, mailing).

Logs when email is sent or skipped (for non-existing users).

Logs duration of the process.


### `POST /reset-password`

**Resets the user's password after verifying the reset code.**

---

#### 🔁 Request

- Content-Type: `application/json`
- JSON body:

```json
{
  "mail_address": "user@example.com",
  "new_password": "NewSecurePassword123!"
}
```
mail_address is required (will be normalized).

new_password is required and will be securely hashed.

#### 🔐 Security Measures
Email is trimmed + lowercased.

Password is hashed using bcrypt.

Password reset only allowed if reset_verified = true in the database.

After successful reset:

reset_verified is set to false.

updated_at is updated.

Rate-limited per IP via httprate:

r.Config.RateLimitResetPassword

r.Config.RateLimitWindowMinutes

✅ Success Response
200 OK

Content-Type: text/plain

Body:
```
Password reset successfully! User can use new password.
```

### `POST /generate-auth-code`

**Generates and sends an authentication code to the provided email address.**

---

#### 🔁 Request

- Content-Type: `application/json`
- JSON body:

```json
{
  "mail_address": "user@example.com"
}
```
mail_address is required, will be lowercased and trimmed.

Must contain an @.
#### 🔐 Security & Behavior
Rate-limited per IP:

Configured via: r.Config.RateLimitResetCode and r.Config.RateLimitWindowMinutes

If the email is new:

A new user is inserted with activated = false, and created_at, updated_at set to NOW().

If email already exists:

No change in user creation (ON CONFLICT DO NOTHING).

A 6-digit authentication code is generated and stored via UpdateAuthenticationCode.

Email is sent asynchronously using SendAuthenticationCode.

✅ Success Response
200 OK

Content-Type: text/plain

Body:
```
If the email exists, a verification code has been sent.
```

✅ Note: This response is generic to prevent user enumeration.
❌ Error Responses
400 Bad Request
If the payload is invalid or the email format is incorrect.

500 Internal Server Error
If:

Database insert fails

Auth code generation fails

Updating the DB with the auth code fails
#### 🛠️ Backend Flow
Parse and validate email.

Insert user if not exists (with activated=false).

Generate a 6-digit auth code.

Update the user's record with the new code.

Send verification code to email asynchronously.

Return a generic success message.
#### 🪵 Logging
Logs:

Errors for bad payloads, DB failures, email issues.

Info when auth code email is sent or skipped due to non-existing email.

Duration of request processing.
Email delivery is asynchronous (via goroutine), meaning the response doesn't block for sending.

### `POST /verify-auth-code`

**Verifies the provided authentication code and activates the user account.**

---

#### 🔁 Request

- Content-Type: `application/json`
- JSON body:

```json
{
  "mail_address": "user@example.com",
  "authentication_code": "123456"
}
```

mail_address: required, case-insensitive, trimmed

authentication_code: required, trimmed

🔐 Security & Behavior
Rate-limited per IP:

Configured via: r.Config.RateLimitResetCode and r.Config.RateLimitWindowMinutes

Checks if a user exists with the given mail_address.

Compares stored authentication_code with the submitted one.

If matched:

Sets activated = true for the user.
✅ Success Response
200 OK

Content-Type: text/plain

Body:
```
Authentication code verified and user activated.
```

#### Error Responses
400 Bad Request
If the request is malformed or fields are missing.

404 Not Found
If the mail_address doesn't exist in the database.

401 Unauthorized
If the authentication_code doesn't match the stored value.

500 Internal Server Error
If there is a database query or update error.

🛠️ Backend Flow
Parse and validate mail_address and authentication_code.

Query users table for authentication_code using mail_address.

If code matches:

Update activated = true.

Respond with success or appropriate error.

#### 🪵 Logging
Logs:

Errors for decoding, DB failures, or mismatches.

Warnings if the email doesn't exist.

Info log for successful activation with elapsed time.


### `POST /logout`

**Logs out the authenticated user by updating their login status.**

---

#### 🔁 Request

- Method: `POST`
- URL: `/logout`
- Headers:
  - `Authorization: Bearer <JWT_TOKEN>`

---

#### 🔐 Authentication

- Requires a valid JWT token in the `Authorization` header.
- Token is validated using `GetValidatedUserIDRole`.
- Extracted `userID` is used to update the login status.

---

#### ✅ Success Response

- `200 OK`
- Content-Type: `application/json`
- Body:

```json
{
  "message": "Logout successful"
}
```

#### Error Responses
401 Unauthorized
If the JWT token is missing, invalid, or expired.

500 Internal Server Error
If the user's login_status could not be updated in the database.

#### 🛠️ Backend Flow
Extract and validate JWT token.

Get userID from token claims.

Call updateLoginStatus(userID, false) to mark user as logged out.

Return success response or appropriate error.

#### 🪵 Logging
Logs:

Errors related to token validation or database update.

Info log on successful logout with elapsed time.


### `POST /refresh-jwt-token`

**Refreshes the JWT token for an authenticated user.**

---

#### 🔁 Request

- Method: `POST`
- URL: `/refresh-jwt-token`
- Headers:
  - `Authorization: Bearer <JWT_TOKEN>`

---

#### 🔐 Authentication

- Requires a valid JWT token in the `Authorization` header.
- Token is validated using `GetValidatedUserIDRole`.

---

#### ✅ Success Response

- `200 OK`
- Content-Type: `application/json`
- Body:

```json
{
  "token": "<NEW_JWT_TOKEN>"
}

```
❌ Error Responses
401 Unauthorized
If the JWT token is missing, invalid, or expired.

500 Internal Server Error

If the user's role cannot be fetched from the database.

If the JWT token generation fails.

#### 🛠️ Backend Flow
Extract and validate the JWT token.

Retrieve userID and role from the database.

Generate a new JWT token using the existing user ID and role.

Return the new token in the response.

🪵 Logging
Logs:

JWT validation failures.

Errors fetching the user role from DB.

Token generation failures.

Success log with elapsed time for token refresh.

#### 📌 Notes
This endpoint helps extend session duration without re-authentication.

The new token includes the same userID and role as the previous token.

Ideal for token rotation strategies to maintain security.
🔄 Token Details
JWT claims include:

- user_id

- role

- exp (expiration timestamp)

Token expiration is based on h.App.JWTExpiration configuration.

### `POST /change-password`

**Allows an authenticated user to change their password securely.**

---

#### 🔁 Request

- Method: `POST`
- URL: `/change-password`
- Headers:
  - `Authorization: Bearer <JWT_TOKEN>`
- Body (`application/json`):
```json
{
  "mail_address": "user@example.com",
  "old_password": "currentPassword123",
  "new_password": "newSecurePassword456"
}
```
🔐 Authentication
Requires a valid JWT token.

Token is parsed to extract the user ID.

The user ID is validated and used to ensure that:

The mail_address matches the authenticated user.

The old_password is correct.

✅ Success Response
200 OK

Body:
```
Password changed successfully
```
####  Error Responses
400 Bad Request – Malformed JSON or missing fields.

401 Unauthorized

Missing or invalid token.

Mismatched mail_address.

Incorrect old_password.

404 Not Found – User not found.

500 Internal Server Error – Database or hashing error.

#### 🛠️ Backend Flow
Extract and validate JWT token from Authorization header.

Parse user_id and validate it exists in the database.

Decode JSON body into ChangePasswordRequest struct.

Match the provided mail_address with the one in the DB.

Compare old password with stored hashed password.

Hash new password and update it in the database.

Respond with success.

#### 🪵 Logging
Logs:

Authorization issues (missing or invalid token).

Token parsing errors.

Email mismatches.

Incorrect old password attempts.

DB or hashing failures.

Password change success with timing.

#### 📌 Notes
The endpoint ensures the user is changing their own password.

Passwords are securely hashed using bcrypt.

updated_at field in the database is updated to track the change.

### `POST /deactivate-user`

**Deactivates a user account. Only accessible to users with the `Admin` role.**

---

#### 🔁 Request

- **Method**: `POST`
- **URL**: `/deactivate-user`
- **Headers**:
  - `Authorization: Bearer <JWT_TOKEN>`
- **Body** (`application/json`):
```json
{
  "mail_address": "user@example.com"
}
```
#### 🔐 Authentication & Authorization
Requires a valid JWT token in the Authorization header.

Role must be "Admin" to perform deactivation.

JWT is parsed and validated using GetValidatedUserRoleOnly (without DB lookup).

✅ Success Response
Status: 200 OK

Body:
```
User deactivated successfully
```

#### Error Responses
400 Bad Request – Invalid or missing mail_address in request body.

401 Unauthorized – Invalid or missing JWT token.

403 Forbidden – Authenticated user is not an Admin.

404 Not Found – No user found with the provided mail_address.

500 Internal Server Error – Database error during update.

#### 🛠️ Backend Flow
Parse and validate JWT using the secret key.

Extract userID and role from token.

Ensure the user has an "Admin" role.

Decode and validate the JSON body.

Update activated = false and updated_at = CURRENT_TIMESTAMP for the user.

Respond with success or appropriate error.

#### 🪵 Logging
Logs:

Authorization and role validation errors.

JSON parsing issues.

Deactivation success and failure with execution time.

Database update errors.

### `POST /reactivate-user`

**Reactivates a deactivated user account. Only accessible to users with the `Admin` role.**

---

#### 🔁 Request

- **Method**: `POST`
- **URL**: `/reactivate-user`
- **Headers**:
  - `Authorization: Bearer <JWT_TOKEN>`
- **Body** (`application/json`):
```json
{
  "mail_address": "user@example.com"
}
```
#### 🔐 Authentication & Authorization
Requires a valid JWT token in the Authorization header.

Role must be "Admin" to perform reactivation.

JWT is parsed and validated using GetValidatedUserRoleOnly without querying the database.

✅ Success Response
Status: 200 OK

Body:
```
User reactivated successfully!
```

#### Error Responses
400 Bad Request – Invalid or missing mail_address in request body.

401 Unauthorized – Invalid or missing JWT token.

403 Forbidden – Authenticated user is not an Admin.

404 Not Found – No user found with the provided mail_address.

500 Internal Server Error – Database error during update.

#### 🛠️ Backend Flow
Extract role from JWT without database lookup.

Verify the role is Admin.

Decode the JSON request body and validate mail_address.

Update the activated field to true and updated_at timestamp for the specified user.

Return success or error responses accordingly.

#### 🪵 Logging
Logs authorization errors.

Logs JSON decoding errors.

Logs database errors and no user found conditions.

Logs successful reactivation including execution time.

#### 📌 Notes
This endpoint performs a soft reactivation by setting activated = true.

updated_at is updated to the current timestamp.

No password or other fields are affected.

### `POST /check-mail-exists`

**Checks if an email address exists in the users database.**

---

#### Request

- **Method:** `POST`
- **URL:** `/check-mail-exists`
- **Headers:**
  - `Content-Type: application/json`
- **Body:**  
```json
{
  "mail_address": "user@example.com"
}
```

#### Response
Status: 200 OK

Body:
```
"Email exists in DB." — if the email is found

"Email does not exist in DB." — if the email is not found
```

#### Error Responses
400 Bad Request — invalid JSON or missing mail_address field

500 Internal Server Error — database query error

#### 🛠️ Backend Flow
Parses JSON request body to get the mail_address.

Queries the database to check if a user with the given email exists.

Returns "Email exists in DB." if found, otherwise "Email does not exist in DB.".

Logs each request and errors with elapsed time.

### `POST /verify-mail-reset-code`

**Verifies a password reset code for a given email address.**

---

#### Request

- **Method:** `POST`
- **URL:** `/verify-mail-reset-code`
- **Headers:**
  - `Content-Type: application/json`
- **Body:**
```json
{
  "mail_address": "user@example.com",
  "reset_code": "123456"
}
```
Response
Status: 200 OK

Body:
```
Reset code verified successfully
```

#### Error Responses
400 Bad Request — invalid JSON payload or missing required fields (mail_address or reset_code)

401 Unauthorized — invalid email or reset code combination

500 Internal Server Error — database query or update error

#### 🛠️ Backend Flow
Parses the request JSON to extract mail_address and reset_code.

Normalizes email and reset code strings (trim and lowercase).

Checks if a user exists with the matching email and reset code.

If valid, sets reset_verified flag to true in the database.

Returns success message on valid verification.

Logs requests, errors, and elapsed time.

#### Notes
This endpoint expects the reset code to be stored in the users table.

On successful verification, the user’s reset_verified status is set to true.

No authentication is required to call this endpoint.



### `DELETE /delete-user`
Deletes a user by their email address. Only accessible by Admin users.

Request
Method: DELETE

URL: /delete-user

Headers:

Content-Type: application/json

Authorization: Bearer <JWT token>

Body:
```json
{
  "mail_address": "user@example.com"
}
```

####  Response
Status: 200 OK
Body:
```json
User deleted successfully
```
Status: 400 Bad Request
Missing or invalid request payload.

Status: 401 Unauthorized
Missing or invalid JWT token.

Status: 403 Forbidden
User role is not Admin.

Status: 404 Not Found
No user found with the specified email address.

Status: 500 Internal Server Error
Database or internal error.

## 🚀 Installation & Usage
```
# Clone the repository
git clone https://github.com/MuratTunc/authentication-service.git
cd authentication-service

# Create your .env file
cp .env.example .env

# Start services
docker-compose up --build
```


## 📝 License

This project is licensed under the [MIT License](LICENSE) – see the LICENSE file for details.


## ⚠️ Disclaimer
```
This software is provided "as is", without any warranty.
The author is not liable for any damage, loss, or issues resulting from the use of this project.
```


## 💬 Contributing
Pull requests are welcome!
For major changes, please open an issue to discuss what you'd like to improve or add.

## 📫 Contact & Profiles
Location: Istanbul, Turkey – Kadıköy
Email: murat.tunc8558@gmail.com
Phone: +90 (531) 731-58-54

🔗 GitHub

🔗 LinkedIn

🔗 Dev.to

🔗 Medium

