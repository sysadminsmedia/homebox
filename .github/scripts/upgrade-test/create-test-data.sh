#!/bin/bash

# Script to create test data in HomeBox for upgrade testing

set -e

HOMEBOX_URL="${HOMEBOX_URL:-http://localhost:7745}"
API_URL="${HOMEBOX_URL}/api/v1"
TEST_DATA_FILE="${TEST_DATA_FILE:-/tmp/test-users.json}"

echo "Creating test data in HomeBox at $HOMEBOX_URL"

# Function to make API calls with error handling
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4

    local response
    if [ -n "$token" ]; then
        response=$(curl -s -X "$method" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint")
    else
        response=$(curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint")
    fi

    # Validate if the response is valid JSON
    if ! echo "$response" | jq '.' > /dev/null 2>&1; then
        echo "Invalid API response for $endpoint: $response" >&2
        exit 1
    fi

    echo "$response"
}

# Function to initialize test data storage
initialize_test_data() {
    echo "Initializing test data file: $TEST_DATA_FILE"
    if [ -f "$TEST_DATA_FILE" ]; then
        echo "Found existing test data file. Removing..."
        rm -f "$TEST_DATA_FILE"
    fi
    echo "{\"users\":[],\"locations\":[],\"labels\":[],\"notifiers\":[],\"items\":[]}" > "$TEST_DATA_FILE"
}

# Function to register a user and get token
register_user() {
    local email=$1
    local name=$2
    local password=$3
    local group_token=$4

    echo "Registering user: $email"

    local payload="{\"email\":\"$email\",\"name\":\"$name\",\"password\":\"$password\""
    if [ -n "$group_token" ]; then
        payload="$payload,\"groupToken\":\"$group_token\""
    fi
    payload="$payload}"

    local response
    response=$(api_call "POST" "/users/register" "$payload")
    echo "$response"
}

# Function to append user data to TEST_DATA_FILE
store_user() {
    local email=$1
    local password=$2
    local token=$3
    local group=$4

    jq --arg email "$email" \
       --arg password "$password" \
       --arg token "$token" \
       --arg group "$group" \
       '.users += [{"email":$email, "password":$password, "token":$token, "group":$group}]' \
       "$TEST_DATA_FILE" > "${TEST_DATA_FILE}.tmp" && mv "${TEST_DATA_FILE}.tmp" "$TEST_DATA_FILE"
}

# Initialize the test data file
initialize_test_data

# Step 1: Register the first user and create the first group
echo "=== Step 1: Create first group with 5 users ==="
user1_response=$(register_user "user1@homebox.test" "User One" "TestPassword123!")
user1_token=$(echo "$user1_response" | jq -r '.token // empty')
group_token=$(echo "$user1_response" | jq -r '.group.inviteToken // empty')

if [ -z "$user1_token" ]; then
    echo "Failed to register first user"
    echo "Response: $user1_response"
    exit 1
fi

# Store the first user
store_user "user1@homebox.test" "TestPassword123!" "$user1_token" "1"

# Register 4 more users in the same group
for i in {2..5}; do
    echo "Registering user$i in group 1..."
    user_response=$(register_user "user${i}@homebox.test" "User $i" "TestPassword123!" "$group_token")
    user_token=$(echo "$user_response" | jq -r '.token // empty')
    if [ -z "$user_token" ]; then
        echo "Failed to register user$i"
        echo "Response: $user_response"
    else
        store_user "user${i}@homebox.test" "TestPassword123!" "$user_token" "1"
    fi
done

# Step 2: Create second group with 2 users
echo "=== Step 2: Create second group with 2 users ==="
user6_response=$(register_user "user6@homebox.test" "User Six" "TestPassword123!")
user6_token=$(echo "$user6_response" | jq -r '.token // empty')
group2_token=$(echo "$user6_response" | jq -r '.group.inviteToken // empty')

if [ -z "$user6_token" ]; then
    echo "Failed to register user6"
    echo "Response: $user6_response"
    exit 1
fi

# Store user6
store_user "user6@homebox.test" "TestPassword123!" "$user6_token" "2"

user7_response=$(register_user "user7@homebox.test" "User Seven" "TestPassword123!" "$group2_token")
user7_token=$(echo "$user7_response" | jq -r '.token // empty')
if [ -z "$user7_token" ]; then
    echo "Failed to register user7"
    echo "Response: $user7_response"
else
    store_user "user7@homebox.test" "TestPassword123!" "$user7_token" "2"
fi

# Final Step: Log the created users for debugging
echo "=== Users Created ==="
cat "$TEST_DATA_FILE" | jq '.users'