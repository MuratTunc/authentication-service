#!/bin/bash

# Load environment variables from .env file
ENV_FILE="../build-tools/.env"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
else
  echo "⚠️ .env file not found at $ENV_FILE"
  exit 1
fi


# Install jq (JSON parsing utility) if not already installed
if ! command -v jq &> /dev/null
then
  echo "jq could not be found, installing..."
  sudo apt-get update
  sudo apt-get install -y jq
fi

test_start() {
    echo -e "********************************************************************"
    echo -e "$AUTHENTICATION_SERVICE_NAME API END POINT INTEGRATION TESTS STARTS..."
    echo -e "********************************************************************"
}

test_end() {
    echo -e "********************************************************************"
    echo -e "$AUTHENTICATION_SERVICE_NAME API END POINT INTEGRATION TESTS END..."
    echo -e "********************************************************************"
}

# Global test user parameters
TEST_USERNAME="testuser"
TEST_MAIL_ADDRESS="testuser@example.com"
TEST_PASSWORD="TestPassword123"
TEST_ROLE="Admin"
TEST_PHONE_NUMBER="+1234567890"
TEST_LANGUAGE="en"
NEW_PASSWORD="NewPass456"
MAIL_RESET_CODE_FOR_TEST="123456"

# Define API URLs
BASE_URL="http://localhost:$AUTHENTICATION_SERVICE_PORT/api/v1"
HEALTH_CHECK_URL="$BASE_URL/auth/health"
REGISTER_URL="$BASE_URL/auth/register-user"
LAST_USER_URL="$BASE_URL/auth/last-user"
DELETE_USER_URL="$BASE_URL/auth/delete-user"
LOGIN_URL="$BASE_URL/auth/login"
LOGOUT_URL="$BASE_URL/auth/logout"
GET_USER_URL="$BASE_URL/auth/get-user-by-mail"
REFRESH_TOKEN_URL="$BASE_URL/auth/refresh-jwt-token"
UPDATE_USER_URL="$BASE_URL/auth/update-user"
CHANGE_PASSWORD_URL="$BASE_URL/auth/change-password"
SEND_AUTH_CODE_URL="$BASE_URL/auth/send-mail-reset-code"
VERIFY_RESET_CODE_URL="$BASE_URL/auth/verify-mail-reset-code"
RESET_PASSWORD_URL="$BASE_URL/auth/reset-password"
LIST_USERS_URL="$BASE_URL/auth/list-users"
DEACTIVATE_USER_URL="$BASE_URL/auth/deactivate-user"
REACTIVATE_USER_URL="$BASE_URL/auth/reactivate-user"
CHECK_MAIL_EXIST_URL="$BASE_URL/auth/check-mail-exists"

TOKEN=""

health_check() {
  echo ""
  echo "===>TEST END POINT--->HEALTH CHECK"
  echo
  echo "REQUEST URL: $HEALTH_CHECK_URL"

  REQUEST_TYPE="GET"

  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "COMMAND: curl -X $REQUEST_TYPE \"$HEALTH_CHECK_URL\""

  HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X $REQUEST_TYPE "$HEALTH_CHECK_URL")

  HTTP_BODY=$(echo "$HEALTH_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$HEALTH_RESPONSE" | tail -n1)

  echo "Health Check Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "Service is healthy!"
  else
    echo "❌ Health check failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
    exit 1
  fi

  echo "✅ Health Check successfully"
  echo "----------------------------------------"
  echo
}

register_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> REGISTER NEW USER"
  echo

  local payload=$(cat <<EOF
{
  "username": "$TEST_USERNAME",
  "mail_address": "$TEST_MAIL_ADDRESS",
  "password": "$TEST_PASSWORD",
  "role": "$TEST_ROLE",
  "phone_number": "$TEST_PHONE_NUMBER",
  "language_preference": "$TEST_LANGUAGE"
}
EOF
)

  echo "REQUEST URL: $REGISTER_URL"
  echo "REQUEST TYPE: POST"
  echo "REQUEST PAYLOAD: $payload"

  REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REGISTER_URL" \
    -H "Content-Type: application/json" \
    -d "$payload")

  HTTP_BODY=$(echo "$REGISTER_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$REGISTER_RESPONSE" | tail -n1)

  echo "Register User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 201 ]; then
    echo "✅ User registered successfully"
  elif [ "$HTTP_STATUS" -eq 409 ]; then
    # Conflict — user exists, try deleting then retry registering
    echo "⚠️ User already exists. Trying to delete and retry registration..."

    if delete_user_test; then
      echo "🗑️ User deleted successfully. Retrying registration..."

      # Retry registration
      REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REGISTER_URL" \
        -H "Content-Type: application/json" \
        -d "$payload")

      HTTP_BODY=$(echo "$REGISTER_RESPONSE" | sed '$ d')
      HTTP_STATUS=$(echo "$REGISTER_RESPONSE" | tail -n1)

      echo "Retry Register User Response Body: $HTTP_BODY"
      echo "Retry HTTP Status Code: $HTTP_STATUS"

      if [ "$HTTP_STATUS" -eq 201 ]; then
        echo "✅ User registered successfully on retry"
      else
        echo "❌ Retry registration failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
        exit 1
      fi
    else
      echo "❌ Failed to delete existing user. Cannot retry registration."
      exit 1
    fi
  else
    echo "❌ User registration failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
    exit 1
  fi

  echo "----------------------------------------"
  echo
}




last_user_test() {
  echo ""
  echo "===>TEST END POINT--->GET LAST USER"
  echo

  echo "REQUEST URL: $LAST_USER_URL"
  echo "REQUEST TYPE: GET"

  LAST_USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$LAST_USER_URL" \
    -H "Content-Type: application/json")

  HTTP_BODY=$(echo "$LAST_USER_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$LAST_USER_RESPONSE" | tail -n1)

  echo "Last User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -ne 200 ]; then
    echo "❌ Failed to fetch last user with status code $HTTP_STATUS. Response: $HTTP_BODY"
  fi


  ACTUAL_USERNAME=$(echo "$HTTP_BODY" | jq -r '.username')
  ACTUAL_MAIL_ADDRESS=$(echo "$HTTP_BODY" | jq -r '.mail_address')
  ACTUAL_ROLE=$(echo "$HTTP_BODY" | jq -r '.role')

  if [[ "$ACTUAL_USERNAME" == "$TEST_USERNAME" && \
        "$ACTUAL_MAIL_ADDRESS" == "$TEST_MAIL_ADDRESS" && \
        "$ACTUAL_ROLE" == "$TEST_ROLE" ]]; then
    echo "✅ Last user data matches the registered user!"
  else
    echo "❌ Last user data does NOT match the registered user!"
    echo "Expected username: $TEST_USERNAME, got: $ACTUAL_USERNAME"
    echo "Expected mail_address: $TEST_MAIL_ADDRESS, got: $ACTUAL_MAIL_ADDRESS"
    echo "Expected role: $TEST_ROLE, got: $ACTUAL_ROLE"

  fi
  echo "----------------------------------------"
  echo
}
login_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> LOGIN USER"
  echo

  # Build JSON payload
  local payload
  payload=$(jq -n \
    --arg mail "$TEST_MAIL_ADDRESS" \
    --arg pwd "$TEST_PASSWORD" \
    '{mail_address: $mail, password: $pwd}')

  echo "REQUEST URL: $LOGIN_URL"
  echo "REQUEST TYPE: POST"
  echo "REQUEST PAYLOAD: $payload"

  # Make the POST request
  local response
  response=$(curl -s -w "\n%{http_code}" -X POST "$LOGIN_URL" \
    -H "Content-Type: application/json" \
    -d "$payload")

  # Separate body and status
  local http_body http_status
  http_body=$(echo "$response" | sed '$ d')
  http_status=$(echo "$response" | tail -n1)

  echo "Login Response Body: $http_body"
  echo "HTTP Status Code: $http_status"

  if [ "$http_status" -eq 200 ]; then
    # Extract token and role
    local token role
    token=$(echo "$http_body" | jq -r '.token // empty')
    role=$(echo "$http_body" | jq -r '.role // empty')

    if [[ -n "$token" ]]; then
      echo "✅ User logged in successfully"
      echo "🪪 JWT Token:"
      echo "----------------------------------------"
      echo "$token"
      echo "----------------------------------------"
      echo "🧑‍💼 User Role: $role"
    else
      echo "❌ Login succeeded but token is missing or invalid JSON"
      exit 1
    fi
  else
    echo "❌ Login failed with status code $http_status"
    echo "Response: $http_body"
    exit 1
  fi

  echo "----------------------------------------"
  echo

  # Optionally export the token if needed globally
  export TOKEN="$token"
}


logout_user_test() {
  echo ""
  echo "===> TEST END POINT ---> LOGOUT USER"
  echo

  if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    echo "❌ No JWT token found from login, cannot test logout"
    exit 1
  fi

  echo "REQUEST URL: $LOGOUT_URL"
  echo "REQUEST TYPE: POST"
  echo "Authorization: Bearer <token>"

  LOGOUT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$LOGOUT_URL" \
    -H "Authorization: Bearer $TOKEN")

  HTTP_BODY=$(echo "$LOGOUT_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$LOGOUT_RESPONSE" | tail -n1)

  echo "Logout Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ User logged out successfully"
  else
    echo "❌ Logout failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

refresh_jwt_token_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> REFRESH JWT TOKEN"
  echo

  # Use the token you got from login test as a bearer token here
  if [ -z "$TOKEN" ]; then
    echo "❌ No JWT token available to test refresh"
    exit 1
  fi

  echo "REQUEST URL: $REFRESH_TOKEN_URL"
  echo "REQUEST TYPE: POST"
  echo "Authorization: Bearer $TOKEN"

  REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REFRESH_TOKEN_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

  HTTP_BODY=$(echo "$REFRESH_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$REFRESH_RESPONSE" | tail -n1)

  echo "Refresh Token Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    NEW_TOKEN=$(echo "$HTTP_BODY" | jq -r '.token')
    if [[ "$NEW_TOKEN" != "null" && "$NEW_TOKEN" != "" ]]; then
      echo "✅ JWT token refreshed successfully"
      echo "🆕 New JWT Token:"
      echo "----------------------------------------"
      echo "$NEW_TOKEN"
    else
      echo "❌ Refresh succeeded but new token is missing"
      exit 1
    fi
  else
    echo "❌ Refresh failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

get_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> GET USER BY MAIL ADDRESS"
  echo

  local request_url="${GET_USER_URL}?mail_address=${TEST_MAIL_ADDRESS}"

  echo "REQUEST URL: $request_url"
  echo "REQUEST TYPE: GET"

  GET_USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$request_url" \
    -H "Content-Type: application/json")

  HTTP_BODY=$(echo "$GET_USER_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$GET_USER_RESPONSE" | tail -n1)

  echo "Get User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    USERNAME=$(echo "$HTTP_BODY" | jq -r '.username')
    if [[ "$USERNAME" != "null" && "$USERNAME" != "" ]]; then
      echo "✅ User fetched successfully: $USERNAME"
    else
      echo "❌ User data missing from response"
      exit 1
    fi
  else
    echo "❌ Get user failed with status code $HTTP_STATUS. Response: $HTTP_BODY"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

update_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> UPDATE USER"

  REQUEST_TYPE="PUT"
  AUTH_HEADER="Authorization: Bearer $JWT_TOKEN"

  # Prepare the update payload
  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS",
  "username": "updateduser",
  "role": "Admin",
  "activated": true
}
EOF
)

  echo "REQUEST URL: $UPDATE_USER_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "AUTH HEADER: $AUTH_HEADER"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X $REQUEST_TYPE $UPDATE_USER_URL \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$REQUEST_PAYLOAD")

  # Extract body and status
  RESPONSE_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  echo "Update User Response Body: $RESPONSE_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ User updated successfully"
  else
    echo "❌ Failed to update user"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

change_password_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> CHANGE PASSWORD"

  REQUEST_TYPE="POST"
  AUTH_HEADER="Authorization: Bearer $TOKEN"

  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS",
  "old_password": "$TEST_PASSWORD",
  "new_password": "$NEW_PASSWORD"
}
EOF
)

  echo "REQUEST URL: $CHANGE_PASSWORD_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "AUTH HEADER: $AUTH_HEADER"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X $REQUEST_TYPE $CHANGE_PASSWORD_URL \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$REQUEST_PAYLOAD")

  RESPONSE_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  echo "Change Password Response Body: $RESPONSE_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Password changed successfully"
  else
    echo "❌ Failed to change password"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

send_forgot_password_code_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> SEND FORGOT PASSWORD RESET CODE"

  REQUEST_TYPE="POST"

  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS"
}
EOF
)

  echo "REQUEST URL: $SEND_AUTH_CODE_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X $REQUEST_TYPE $SEND_AUTH_CODE_URL \
    -H "Content-Type: application/json" \
    -d "$REQUEST_PAYLOAD")

  RESPONSE_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  echo "Send Reset Code Response Body: $RESPONSE_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Reset code sent to email successfully"
  else
    echo "❌ Failed to send reset code"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}


verify_mail_reset_code_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> VERIFY RESET CODE"

  REQUEST_TYPE="POST"

  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS",
  "reset_code": "$MAIL_RESET_CODE_FOR_TEST"
}
EOF
)

  echo "REQUEST URL: $VERIFY_RESET_CODE_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X $REQUEST_TYPE $VERIFY_RESET_CODE_URL \
    -H "Content-Type: application/json" \
    -d "$REQUEST_PAYLOAD")

  RESPONSE_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  echo "Verify Reset Code Response Body: $RESPONSE_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Reset code verified successfully"
  else
    echo "❌ Failed to verify reset code"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}


reset_password_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> RESET PASSWORD"

  REQUEST_TYPE="POST"
  AUTH_HEADER="Authorization: Bearer $TOKEN"

  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS",
  "new_password": "$TEST_PASSWORD"
}
EOF
)

  echo "REQUEST URL: $RESET_PASSWORD_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "AUTH HEADER: $AUTH_HEADER"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" -X $REQUEST_TYPE $RESET_PASSWORD_URL \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$REQUEST_PAYLOAD")

  RESPONSE_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  echo "Reset Password Response Body: $RESPONSE_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Password reset successfully"
  else
    echo "❌ Failed to reset password"
    exit 1
  fi
  echo "----------------------------------------"
  echo
}


list_users_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> LIST USERS (ADMIN ONLY)"
  echo

  if [ -z "$LIST_USERS_URL" ]; then
    echo "❌ LIST_USERS_URL is not set. Please define it before running the test."
    exit 1
  fi

  if [ -z "$TOKEN" ]; then
    echo "❌ TOKEN is not set. Make sure login_user_test() was successful."
    exit 1
  fi

  echo "REQUEST URL: $LIST_USERS_URL"
  echo "REQUEST TYPE: GET"
  echo "Authorization: Bearer $TOKEN"

  LIST_USERS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$LIST_USERS_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

  HTTP_BODY=$(echo "$LIST_USERS_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$LIST_USERS_RESPONSE" | tail -n1)

  echo "List Users Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ -z "$HTTP_STATUS" ]; then
    echo "❌ No HTTP status received. Check the URL and the curl command."
    exit 1
  fi

  if [ "$HTTP_STATUS" -eq 200 ]; then
    # Extract role from the first user in the list using grep and cut
    ROLE=$(echo "$HTTP_BODY" | grep -o '"role":"[^"]*"' | head -n 1 | cut -d':' -f2 | tr -d '"')

    echo "✅ Successfully listed users."
    echo "----------------------------------------"
    echo "🧑‍💼 User Role: $ROLE"
  else
    echo "❌ Failed to list users. Check if token is valid and user is admin."
    exit 1
  fi
  echo "----------------------------------------"
  echo
}

deactivate_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> DEACTIVATE USER (ADMIN ONLY)"
  echo

  if [ -z "$DEACTIVATE_USER_URL" ]; then
    echo "❌ DEACTIVATE_USER_URL is not set. Please define it before running the test."
    exit 1
  fi

  if [ -z "$TOKEN" ]; then
    echo "❌ TOKEN is not set. Make sure login_user_test() was successful."
    exit 1
  fi

  if [ -z "$TEST_MAIL_ADDRESS" ]; then
    echo "❌ TEST_MAIL_ADDRESS is not set. Please define the target user's email."
    exit 1
  fi

  echo "REQUEST URL: $DEACTIVATE_USER_URL"
  echo "REQUEST TYPE: POST"
  echo "Authorization: Bearer $TOKEN"
  echo "Payload: { \"mail_address\": \"$TEST_MAIL_ADDRESS\" }"

  DEACTIVATE_USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$DEACTIVATE_USER_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"mail_address\": \"$TEST_MAIL_ADDRESS\"}")

  HTTP_BODY=$(echo "$DEACTIVATE_USER_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$DEACTIVATE_USER_RESPONSE" | tail -n1)

  echo "Deactivate User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ -z "$HTTP_STATUS" ]; then
    echo "❌ No HTTP status received. Check the URL and the curl command."
  fi

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Successfully deactivated user: $TEST_MAIL_ADDRESS"
  else
    echo "❌ Failed to deactivate user. Check if token is valid and user has admin rights."
    exit 1
  fi

  echo "----------------------------------------"
  echo
}

reactivate_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> REACTIVATE USER (ADMIN ONLY)"
  echo

  if [ -z "$REACTIVATE_USER_URL" ]; then
    echo "❌ REACTIVATE_USER_URL is not set. Please define it before running the test."
  fi

  if [ -z "$TOKEN" ]; then
    echo "❌ TOKEN is not set. Make sure login_user_test() was successful."
  fi

  if [ -z "$TEST_MAIL_ADDRESS" ]; then
    echo "❌ TEST_MAIL_ADDRESS is not set. Make sure last_user_test() was successful."
  fi

  echo "REQUEST URL: $REACTIVATE_USER_URL"
  echo "REQUEST TYPE: POST"
  echo "Authorization: Bearer $TOKEN"
  echo "Mail Address: $TEST_MAIL_ADDRESS"

  REACTIVATE_USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$REACTIVATE_USER_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"mail_address\": \"$TEST_MAIL_ADDRESS\"}")

  HTTP_BODY=$(echo "$REACTIVATE_USER_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$REACTIVATE_USER_RESPONSE" | tail -n1)

  echo "Reactivate User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ -z "$HTTP_STATUS" ]; then
    echo "❌ No HTTP status received. Check the URL and the curl command."
  fi

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ User reactivated successfully."
  else
    echo "❌ Failed to reactivate user. Status code: $HTTP_STATUS"
    exit 1
  fi

  echo "----------------------------------------"
  echo
}

stress_rate_limit_test() {
  echo ""
  echo "===> STRESS TEST ENDPOINT ---> LOGIN RATE LIMIT"
  echo

  local payload=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS",
  "password": "$TEST_PASSWORD"
}
EOF
)

  local TOTAL_REQUESTS=120
  local PARALLEL_REQUESTS=30

  echo "REQUEST URL: $LOGIN_URL"
  echo "REQUEST TYPE: POST"
  echo "REQUEST PAYLOAD: $payload"
  echo "TOTAL REQUESTS: $TOTAL_REQUESTS"
  echo "PARALLEL REQUESTS: $PARALLEL_REQUESTS"
  echo "Running stress test..."

  seq $TOTAL_REQUESTS | xargs -P$PARALLEL_REQUESTS -I{} \
    curl -s -o /dev/null -w "%{http_code}\n" -X POST "$LOGIN_URL" \
    -H "Content-Type: application/json" \
    -d "$payload" \
  | sort | uniq -c | while read count status; do
    case "$status" in
      200)
        echo "✅ $count OK (200) - Login success"
        ;;
      429)
        echo "⛔️ $count Too Many Requests (429) - Rate limit hit"
        ;;
      *)
        echo "❌ $count Unexpected status ($status)"
        ;;
    esac
  done

  echo "----------------------------------------"
  echo
}


check_mail_exist_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> CHECK MAIL EXISTS"
  echo

  if [ -z "$CHECK_MAIL_EXIST_URL" ]; then
    echo "❌ CHECK_MAIL_EXIST_URL is not set. Please define it before running the test."
  fi

  if [ -z "$TEST_MAIL_ADDRESS" ]; then
    echo "❌ TEST_MAIL_ADDRESS is not set."
  fi

  echo "REQUEST URL: $CHECK_MAIL_EXIST_URL"
  echo "REQUEST TYPE: POST"
  echo "Mail Address: $TEST_MAIL_ADDRESS"

  CHECK_MAIL_EXIST_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$CHECK_MAIL_EXIST_URL" \
    -H "Content-Type: application/json" \
    -d "{\"mail_address\": \"$TEST_MAIL_ADDRESS\"}")

  HTTP_BODY=$(echo "$CHECK_MAIL_EXIST_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$CHECK_MAIL_EXIST_RESPONSE" | tail -n1)

  echo "Check Mail Exists Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Mail existence check succeeded."
  else
    echo "❌ Failed to check mail existence. Status code: $HTTP_STATUS"
    exit 1
  fi

  echo "----------------------------------------"
  echo
}

verify_mail_address_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> VERIFY MAIL ADDRESS"
  echo

  if [ -z "$VERIFY_MAIL_URL" ]; then
    echo "❌ VERIFY_MAIL_URL is not set. Please define it before running the test."
  fi

  if [ -z "$TEST_MAIL_ADDRESS" ]; then
    echo "❌ TEST_MAIL_ADDRESS is not set."
  fi

  echo "REQUEST URL: $VERIFY_MAIL_URL"
  echo "REQUEST TYPE: POST"
  echo "Mail Address: $TEST_MAIL_ADDRESS"

  VERIFY_MAIL_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$VERIFY_MAIL_URL" \
    -H "Content-Type: application/json" \
    -d "{\"mail_address\": \"$TEST_MAIL_ADDRESS\"}")

  HTTP_BODY=$(echo "$VERIFY_MAIL_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$VERIFY_MAIL_RESPONSE" | tail -n1)

  echo "Verify Mail Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ Mail verification request succeeded."
  else
    echo "❌ Mail verification failed. Status code: $HTTP_STATUS"
    exit 1
  fi

  echo "----------------------------------------"
  echo
}



# With JWT
delete_user_test() {
  echo ""
  echo "===> TEST ENDPOINT ---> DELETE USER BY MAIL"

  if [ -z "$DELETE_USER_URL" ]; then
    echo "❌ DELETE_USER_URL is not set. Please define it before running the test."
    exit 1
  fi

  if [ -z "$TOKEN" ]; then
    echo "❌ TOKEN is not set. Make sure login_user_test() was successful."
    exit 1
  fi

  if [ -z "$TEST_MAIL_ADDRESS" ]; then
    echo "❌ TEST_MAIL_ADDRESS is not set. Make sure last_user_test() was successful."
    exit 1
  fi

  REQUEST_TYPE="DELETE"
  REQUEST_PAYLOAD=$(cat <<EOF
{
  "mail_address": "$TEST_MAIL_ADDRESS"
}
EOF
)

  echo "REQUEST URL: $DELETE_USER_URL"
  echo "REQUEST TYPE: $REQUEST_TYPE"
  echo "Authorization: Bearer $TOKEN"
  echo "REQUEST PAYLOAD: $REQUEST_PAYLOAD"

  HTTP_RESPONSE=$(curl -s -w "\n%{http_code}" -X $REQUEST_TYPE "$DELETE_USER_URL" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$REQUEST_PAYLOAD")

  HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed '$ d')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tail -n1)

  echo "Delete User Response Body: $HTTP_BODY"
  echo "HTTP Status Code: $HTTP_STATUS"

  if [ -z "$HTTP_STATUS" ]; then
    echo "❌ No HTTP status received. Check the URL and the curl command."
  fi

  if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ User deleted successfully"
  else
    echo "❌ Failed to delete user. Status code: $HTTP_STATUS"
    echo "Response: $HTTP_BODY"
  fi

  echo "----------------------------------------"
  echo
}



# Run tests
test_start                               # 🟢 Initialize the test suite (start timer or header)

health_check                             # ✅ Check if the service is running and responding

register_user_test                       # 🆕 Register a new user to use in subsequent tests

get_user_test                            # 🔍 Get the registered user's details by mail address

last_user_test                           # 📬 Save the last user's email address to a variable (used later)

login_user_test                          # 🔐 Log in using the registered credentials to obtain JWT token

check_mail_exist_test

list_users_test                          # 📋 List all users (admin-only endpoint to verify role/token)

deactivate_user_test                     # 🚫 Deactivate the user account (admin-only)

reactivate_user_test                     # ✅ Reactivate the user account (admin-only)

sleep 1                                  # ⏸️ Pause to ensure changes propagate (e.g. DB/state consistency)

change_password_test                     # 🔐 Change password while authenticated (authorized endpoint)

send_forgot_password_code_test           # 📧 Send a reset code to user’s email for password recovery

verify_mail_reset_code_test              # 🔑 Verify the mail reset code sent to the user

reset_password_test                      # 🔄 Reset the user’s password using verified code

refresh_jwt_token_test                   # ♻️ Refresh the JWT token to maintain session

update_user_test                         # 📝 Update user info like name or preferences

logout_user_test                         # 🚪 Log out the current user and invalidate the token

stress_rate_limit_test

sleep 10                                  # ⏸️ Pause to ensure stress test is finished.

delete_user_test                         # 🗑️ Delete the user from the database (clean up test user)

test_end                                 # 🔚 End the test suite (show summary or footer)

