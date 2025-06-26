# Authentication-service

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Welcome to the Golang Auth Service — a lightweight, high-performance authentication microservice built using Go’s native net/http package with the Chi router for efficient routing. Designed with simplicity and scalability in mind, this service powers secure user authentication and management in any modern backend system.

### Features

✅ Built with Go for speed and concurrency

🧩 Uses Chi for idiomatic, lightweight HTTP routing

🗄️ PostgreSQL for robust and reliable user data persistence

🔒 JWT-based authentication with refresh tokens

📩 Email-based reset code & verification support

📦 Containerized with Docker for easy deployment

⚙️ Configurable via Docker Compose

### 🛠️ Tech Stack

| Technology       | Purpose                                      |
|------------------|----------------------------------------------|
| **Go (net/http)**| Core HTTP server logic                       |
| **Chi**          | Lightweight and idiomatic HTTP routing       |
| **PostgreSQL**   | Reliable user data persistence               |
| **JWT**          | Secure token-based authentication            |
| **Docker**       | Containerization and binary generation       |
| **Docker Compose**| Service orchestration and configuration     |


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


| #  | Method | Endpoint                              | Description                           | JWT-Protected | Example cURL Command                                                                                                                                                                              |
| -- | ------ | ------------------------------------- | ------------------------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | GET    | `/api/v1/auth/health`                 | Health check for the auth service     | No            | `curl -X GET http://localhost:8080/api/v1/auth/health`                                                                                                                                            |
| 2  | GET    | `/api/v1/auth/last-user`              | Get the most recently registered user | No            | `curl -X GET http://localhost:8080/api/v1/auth/last-user`                                                                                                                                         |
| 3  | GET    | `/api/v1/auth/get-user-by-mail`       | Get user by email address             | No            | `curl -G http://localhost:8080/api/v1/auth/get-user-by-mail --data-urlencode "mail_address=user@example.com"`                                                                                     |
| 4  | GET    | `/api/v1/auth/list-users`             | List all users                        | No            | `curl -X GET http://localhost:8080/api/v1/auth/list-users`                                                                                                                                        |
| 5  | PUT    | `/api/v1/auth/update-user`            | Update user details                   | Yes           | `curl -X PUT http://localhost:8080/api/v1/auth/update-user -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"mail_address":"user@example.com", "username":"NewName"}'` |
| 6  | POST   | `/api/v1/auth/register-user`          | Register a new user                   | No            | `curl -X POST http://localhost:8080/api/v1/auth/register-user -H "Content-Type: application/json" -d '{"username":"John", "mail_address":"john@example.com", "password":"password"}'`             |
| 7  | POST   | `/api/v1/auth/login`                  | Log in an existing user               | No            | `curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"mail_address":"john@example.com", "password":"password"}'`                                        |
| 8  | POST   | `/api/v1/auth/send-mail-reset-code`   | Send reset code via email             | No            | `curl -X POST http://localhost:8080/api/v1/auth/send-mail-reset-code -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`                                                |
| 9  | POST   | `/api/v1/auth/reset-password`         | Reset the user's password             | No            | `curl -X POST http://localhost:8080/api/v1/auth/reset-password -H "Content-Type: application/json" -d '{"mail_address":"john@example.com", "new_password":"newpassword"}'`                        |
| 10 | POST   | `/api/v1/auth/generate-auth-code`     | Generate authentication code          | No            | `curl -X POST http://localhost:8080/api/v1/auth/generate-auth-code -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`                                                  |
| 11 | POST   | `/api/v1/auth/verify-auth-code`       | Verify the sent authentication code   | No            | `curl -X POST http://localhost:8080/api/v1/auth/verify-auth-code -H "Content-Type: application/json" -d '{"mail_address":"john@example.com", "auth_code":"123456"}'`                              |
| 12 | POST   | `/api/v1/auth/logout`                 | Log out a user                        | Yes           | `curl -X POST http://localhost:8080/api/v1/auth/logout -H "Authorization: Bearer <token>"`                                                                                                        |
| 13 | POST   | `/api/v1/auth/refresh-jwt-token`      | Refresh JWT token                     | Yes           | `curl -X POST http://localhost:8080/api/v1/auth/refresh-jwt-token -H "Authorization: Bearer <refresh_token>"`                                                                                     |
| 14 | POST   | `/api/v1/auth/change-password`        | Change user password                  | Yes           | `curl -X POST http://localhost:8080/api/v1/auth/change-password -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"old_password":"old", "new_password":"new"}'`         |
| 15 | POST   | `/api/v1/auth/deactivate-user`        | Deactivate a user account             | Yes           | `curl -X POST http://localhost:8080/api/v1/auth/deactivate-user -H "Authorization: Bearer <admin_token>" -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`            |
| 16 | POST   | `/api/v1/auth/reactivate-user`        | Reactivate a user account             | Yes           | `curl -X POST http://localhost:8080/api/v1/auth/reactivate-user -H "Authorization: Bearer <admin_token>" -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`            |
| 17 | POST   | `/api/v1/auth/check-mail-exists`      | Check if email exists in DB           | No            | `curl -X POST http://localhost:8080/api/v1/auth/check-mail-exists -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`                                                   |
| 18 | POST   | `/api/v1/auth/verify-mail-reset-code` | Verify password reset code            | No            | `curl -X POST http://localhost:8080/api/v1/auth/verify-mail-reset-code -H "Content-Type: application/json" -d '{"mail_address":"john@example.com", "reset_code":"123456"}'`                       |
| 19 | DELETE | `/api/v1/auth/delete-user`            | Delete a user account                 | Yes           | `curl -X DELETE http://localhost:8080/api/v1/auth/delete-user -H "Authorization: Bearer <admin_token>" -H "Content-Type: application/json" -d '{"mail_address":"john@example.com"}'`              |


#### Notes:

Replace http://localhost with your actual server URL.

For endpoints requiring authorization, add the Authorization header accordingly.

Body data is JSON formatted and wrapped in single quotes '...' for shell compatibility.

## 🚀 Installation & Usage
You can either use the prebuilt binary directly or run the service using Docker Compose.

### ✅ Option 1: Use Prebuilt Binary (Cloud/VM/Kubernetes)
Download the binary from your cloud path (e.g. from Kubernetes or remote VM):
```
back-end/authentication-service/bin/authentication-serviceBinary
```

Make the binary executable:
```
chmod +x authentication-serviceBinary
```
Run the service:
```
./authentication-serviceBinary
```

Make sure to set environment variables using the .env file if needed.


### ✅ Option 2: Build and Run Using Docker Compose
Clone the repository:
```
git clone https://github.com/MuratTunc/authentication-service.git
cd authentication-service/back-end/build-tools
```

Build and start all services:
```
./build-services.sh build-all
```

Environment Configuration:
```
The default .env file is located at:
back-end/build-tools/.env
```
### ⚠️ Important: Don’t forget to update the secrets (e.g. JWT keys, database credentials) before deploying to production.



