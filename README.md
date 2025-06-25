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


## 📜 License
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

