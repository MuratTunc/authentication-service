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

Base path: /api/v1/auth


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

Logs
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

Logs
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

Logs
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
- mail_address is required to identify the user to update.

- Other fields (username, role, activated) are optional and will be updated if provided.


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

 Logs
- Logs errors for failed authentication, invalid payload, unauthorized attempts, and database errors.
- Logs successful update and time taken to process the request.





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

