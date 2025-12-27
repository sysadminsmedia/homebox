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

    # Validate response is proper JSON
    if ! echo "$response" | jq '.' > /dev/null 2>&1; then
        echo "Invalid API response for $endpoint: $response" >&2
        exit 1
    fi

    echo "$response"
}

# Function to initialize the test data JSON file
initialize_test_data() {
    echo "Initializing test data JSON file: $TEST_DATA_FILE"
    if [ -f "$TEST_DATA_FILE" ]; then
        echo "Removing existing test data file..."
        rm -f "$TEST_DATA_FILE"
    fi
    echo "{\"users\":[],\"locations\":[],\"labels\":[],\"items\":[],\"attachments\":[],\"notifiers\":[]}" > "$TEST_DATA_FILE"
}

# Function to add content to JSON data file
add_to_test_data() {
    local key=$1
    local value=$2

    jq --argjson data "$value" ".${key} += [\$data]" "$TEST_DATA_FILE" > "${TEST_DATA_FILE}.tmp" && mv "${TEST_DATA_FILE}.tmp" "$TEST_DATA_FILE"
}

# Register a user and get their auth token
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

    api_call "POST" "/users/register" "$payload"
}

# Main logic for creating test data
initialize_test_data

# Group 1: Create 5 users
echo "=== Creating Group 1 Users ==="
group1_user1_response=$(register_user "user1@homebox.test" "User One" "password123")
group1_user1_token=$(echo "$group1_user1_response" | jq -r '.token // empty')
group1_invite_token=$(echo "$group1_user1_response" | jq -r '.group.inviteToken // empty')

if [ -z "$group1_user1_token" ]; then
    echo "Failed to register the first group user" >&2
    exit 1
fi
add_to_test_data "users" "{\"email\": \"user1@homebox.test\", \"token\": \"$group1_user1_token\", \"group\": 1}"

# Add 4 more users to the same group
for user in 2 3 4 5; do
    response=$(register_user "user$user@homebox.test" "User $user" "password123" "$group1_invite_token")
    token=$(echo "$response" | jq -r '.token // empty')
    add_to_test_data "users" "{\"email\": \"user$user@homebox.test\", \"token\": \"$token\", \"group\": 1}"
done

# Group 2: Create 2 users
echo "=== Creating Group 2 Users ==="
group2_user1_response=$(register_user "user6@homebox.test" "User Six" "password123")
group2_user1_token=$(echo "$group2_user1_response" | jq -r '.token // empty')
group2_invite_token=$(echo "$group2_user1_response" | jq -r '.group.inviteToken // empty')
add_to_test_data "users" "{\"email\": \"user6@homebox.test\", \"token\": \"$group2_user1_token\", \"group\": 2}"

response=$(register_user "user7@homebox.test" "User Seven" "password123" "$group2_invite_token")
group2_user2_token=$(echo "$response" | jq -r '.token // empty')
add_to_test_data "users" "{\"email\": \"user7@homebox.test\", \"token\": \"$group2_user2_token\", \"group\": 2}"

# Create Locations
echo "=== Creating Locations ==="
group1_locations=()
group1_locations+=("$(api_call "POST" "/locations" "{ \"name\": \"Living Room\", \"description\": \"Family area\" }" "$group1_user1_token")")
group1_locations+=("$(api_call "POST" "/locations" "{ \"name\": \"Garage\", \"description\": \"Storage area\" }" "$group1_user1_token")")
group2_locations=()
group2_locations+=("$(api_call "POST" "/locations" "{ \"name\": \"Office\", \"description\": \"Workspace\" }" "$group2_user1_token")")

# Add Locations to Test Data
for loc in "${group1_locations[@]}"; do
    loc_id=$(echo "$loc" | jq -r '.id // empty')
    add_to_test_data "locations" "{\"id\": \"$loc_id\", \"group\": 1}"
done

for loc in "${group2_locations[@]}"; do
    loc_id=$(echo "$loc" | jq -r '.id // empty')
    add_to_test_data "locations" "{\"id\": \"$loc_id\", \"group\": 2}"
done

# Create Labels
echo "=== Creating Labels ==="
label1=$(api_call "POST" "/labels" "{ \"name\": \"Electronics\", \"description\": \"Devices\" }" "$group1_user1_token")
add_to_test_data "labels" "$label1"

label2=$(api_call "POST" "/labels" "{ \"name\": \"Important\", \"description\": \"High Priority\" }" "$group1_user1_token")
add_to_test_data "labels" "$label2"

# Create Items and Attachments
echo "=== Creating Items and Attachments ==="
item1=$(api_call "POST" "/items" "{ \"name\": \"Laptop\", \"description\": \"Work laptop\", \"locationId\": \"$(echo ${group1_locations[0]} | jq -r '.id // empty')\" }" "$group1_user1_token")
item1_id=$(echo "$item1" | jq -r '.id // empty')
add_to_test_data "items" "{\"id\": \"$item1_id\", \"group\": 1}"

attachment1=$(api_call "POST" "/items/$item1_id/attachments" "" "$group1_user1_token")
add_to_test_data "attachments" "{\"id\": \"$(echo $attachment1 | jq -r '.id // empty')\", \"itemId\": \"$item1_id\"}"

# Create Test Notifier
echo "=== Creating Notifiers ==="
notifier=$(api_call "POST" "/notifiers" "{ \"name\": \"TESTING\", \"url\": \"https://example.com/webhook\", \"isActive\": true }" "$group1_user1_token")
add_to_test_data "notifiers" "$notifier"

echo "=== Test Data Creation Complete ==="
cat "$TEST_DATA_FILE" | jq