#!/bin/bash

# Script to create test data in HomeBox for upgrade testing
# This script creates users, items, attachments, notifiers, locations, and tags

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
    
    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            curl -s -X "$method" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$API_URL$endpoint"
        else
            curl -s -X "$method" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                "$API_URL$endpoint"
        fi
    else
        if [ -n "$data" ]; then
            curl -s -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$API_URL$endpoint"
        else
            curl -s -X "$method" \
                -H "Content-Type: application/json" \
                "$API_URL$endpoint"
        fi
    fi
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
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$API_URL/users/register")
    
    echo "$response"
}

# Function to login and get token
login_user() {
    local email=$1
    local password=$2
    
    echo "Logging in user: $email" >&2
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$email\",\"password\":\"$password\"}" \
        "$API_URL/users/login")
    
    echo "$response" | jq -r '.token // empty'
}

# Function to create an item
create_item() {
    local token=$1
    local name=$2
    local description=$3
    local location_id=$4
    
    echo "Creating item: $name" >&2
    
    local payload="{\"name\":\"$name\",\"description\":\"$description\""
    
    if [ -n "$location_id" ]; then
        payload="$payload,\"locationId\":\"$location_id\""
    fi
    
    payload="$payload}"
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$API_URL/items")
    
    echo "$response"
}

# Function to create a location
create_location() {
    local token=$1
    local name=$2
    local description=$3
    
    echo "Creating location: $name" >&2
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"description\":\"$description\"}" \
        "$API_URL/locations")
    
    echo "$response"
}

# Function to create a tag
create_tag() {
    local token=$1
    local name=$2
    local description=$3
    
    echo "Creating tag: $name" >&2
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"description\":\"$description\"}" \
        "$API_URL/tags")
    
    echo "$response"
}

# Function to create a notifier
create_notifier() {
    local token=$1
    local name=$2
    local url=$3
    
    echo "Creating notifier: $name" >&2
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"url\":\"$url\",\"isActive\":true}" \
        "$API_URL/groups/notifiers")
    
    echo "$response"
}

# Function to attach a file to an item (creates a dummy attachment)
attach_file_to_item() {
    local token=$1
    local item_id=$2
    local filename=$3
    
    echo "Creating attachment for item: $item_id" >&2
    
    # Create a temporary file with some content
    local temp_file=$(mktemp)
    echo "This is a test attachment for $filename" > "$temp_file"
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $token" \
        -F "file=@$temp_file" \
        -F "type=attachment" \
        -F "name=$filename" \
        "$API_URL/items/$item_id/attachments")
    
    rm -f "$temp_file"
    
    echo "$response"
}

# Initialize test data storage
echo "{\"users\":[]}" > "$TEST_DATA_FILE"

echo "=== Step 1: Create first group with 5 users ==="

# Register first user (creates a new group)
user1_response=$(register_user "user1@homebox.test" "User One" "TestPassword123!")
user1_token=$(echo "$user1_response" | jq -r '.token // empty')
group_token=$(echo "$user1_response" | jq -r '.group.inviteToken // empty')

if [ -z "$user1_token" ]; then
    echo "Failed to register first user"
    echo "Response: $user1_response"
    exit 1
fi

echo "First user registered with token. Group token: $group_token"

# Store user1 data
jq --arg email "user1@homebox.test" \
   --arg password "TestPassword123!" \
   --arg token "$user1_token" \
   --arg group "1" \
   '.users += [{"email":$email,"password":$password,"token":$token,"group":$group}]' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

# Register 4 more users in the same group
for i in {2..5}; do
    echo "Registering user$i in group 1..."
    user_response=$(register_user "user${i}@homebox.test" "User $i" "TestPassword123!" "$group_token")
    user_token=$(echo "$user_response" | jq -r '.token // empty')
    
    if [ -z "$user_token" ]; then
        echo "Failed to register user$i"
        echo "Response: $user_response"
    else
        echo "user$i registered successfully"
        # Store user data
        jq --arg email "user${i}@homebox.test" \
           --arg password "TestPassword123!" \
           --arg token "$user_token" \
           --arg group "1" \
           '.users += [{"email":$email,"password":$password,"token":$token,"group":$group}]' \
           "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"
    fi
done

echo "=== Step 2: Create second group with 2 users ==="

# Register first user of second group
user6_response=$(register_user "user6@homebox.test" "User Six" "TestPassword123!")
user6_token=$(echo "$user6_response" | jq -r '.token // empty')
group2_token=$(echo "$user6_response" | jq -r '.group.inviteToken // empty')

if [ -z "$user6_token" ]; then
    echo "Failed to register user6"
    echo "Response: $user6_response"
    exit 1
fi

echo "user6 registered with token. Group 2 token: $group2_token"

# Store user6 data
jq --arg email "user6@homebox.test" \
   --arg password "TestPassword123!" \
   --arg token "$user6_token" \
   --arg group "2" \
   '.users += [{"email":$email,"password":$password,"token":$token,"group":$group}]' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

# Register second user in group 2
user7_response=$(register_user "user7@homebox.test" "User Seven" "TestPassword123!" "$group2_token")
user7_token=$(echo "$user7_response" | jq -r '.token // empty')

if [ -z "$user7_token" ]; then
    echo "Failed to register user7"
    echo "Response: $user7_response"
else
    echo "user7 registered successfully"
    # Store user7 data
    jq --arg email "user7@homebox.test" \
       --arg password "TestPassword123!" \
       --arg token "$user7_token" \
       --arg group "2" \
       '.users += [{"email":$email,"password":$password,"token":$token,"group":$group}]' \
       "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"
fi

echo "=== Step 3: Create locations for each group ==="

# Create locations for group 1 (using user1's token)
location1=$(create_location "$user1_token" "Living Room" "Main living area")
location1_id=$(echo "$location1" | jq -r '.id // empty')
echo "Created location: Living Room (ID: $location1_id)"

location2=$(create_location "$user1_token" "Garage" "Storage and tools")
location2_id=$(echo "$location2" | jq -r '.id // empty')
echo "Created location: Garage (ID: $location2_id)"

# Create location for group 2 (using user6's token)
location3=$(create_location "$user6_token" "Home Office" "Work from home space")
location3_id=$(echo "$location3" | jq -r '.id // empty')
echo "Created location: Home Office (ID: $location3_id)"

# Store locations
jq --arg loc1 "$location1_id" \
   --arg loc2 "$location2_id" \
   --arg loc3 "$location3_id" \
   '.locations = {"group1":[$loc1,$loc2],"group2":[$loc3]}' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

echo "=== Step 4: Create tags for each group ==="

# Create tags for group 1
tag1=$(create_tag "$user1_token" "Electronics" "Electronic devices")
tag1_id=$(echo "$tag1" | jq -r '.id // empty')
echo "Created tag: Electronics (ID: $tag1_id)"

tag2=$(create_tag "$user1_token" "Important" "High priority items")
tag2_id=$(echo "$tag2" | jq -r '.id // empty')
echo "Created tag: Important (ID: $tag2_id)"

# Create tag for group 2
tag3=$(create_tag "$user6_token" "Work Equipment" "Items for work")
tag3_id=$(echo "$tag3" | jq -r '.id // empty')
echo "Created tag: Work Equipment (ID: $tag3_id)"

# Store tags
jq --arg tag1 "$tag1_id" \
   --arg tag2 "$tag2_id" \
   --arg tag3 "$tag3_id" \
   '.tags = {"group1":[$tag1,$tag2],"group2":[$tag3]}' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

echo "=== Step 5: Create test notifier ==="

# Create notifier for group 1
notifier1=$(create_notifier "$user1_token" "TESTING" "https://example.com/webhook")
notifier1_id=$(echo "$notifier1" | jq -r '.id // empty')
echo "Created notifier: TESTING (ID: $notifier1_id)"

# Store notifier
jq --arg not1 "$notifier1_id" \
   '.notifiers = {"group1":[$not1]}' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

echo "=== Step 6: Create items for all users ==="

# Create items for users in group 1
declare -A user_tokens
user_tokens[1]=$user1_token
user_tokens[2]=$(echo "$user1_token") # Users in same group share data, but we'll use user1 token
user_tokens[3]=$(echo "$user1_token")
user_tokens[4]=$(echo "$user1_token")
user_tokens[5]=$(echo "$user1_token")

# Items for group 1 users
echo "Creating items for group 1..."
item1=$(create_item "$user1_token" "Laptop Computer" "Dell XPS 15 for work" "$location1_id")
item1_id=$(echo "$item1" | jq -r '.id // empty')
echo "Created item: Laptop Computer (ID: $item1_id)"

item2=$(create_item "$user1_token" "Power Drill" "DeWalt 20V cordless drill" "$location2_id")
item2_id=$(echo "$item2" | jq -r '.id // empty')
echo "Created item: Power Drill (ID: $item2_id)"

item3=$(create_item "$user1_token" "TV Remote" "Samsung TV remote control" "$location1_id")
item3_id=$(echo "$item3" | jq -r '.id // empty')
echo "Created item: TV Remote (ID: $item3_id)"

item4=$(create_item "$user1_token" "Tool Box" "Red metal tool box with tools" "$location2_id")
item4_id=$(echo "$item4" | jq -r '.id // empty')
echo "Created item: Tool Box (ID: $item4_id)"

item5=$(create_item "$user1_token" "Coffee Maker" "Breville espresso machine" "$location1_id")
item5_id=$(echo "$item5" | jq -r '.id // empty')
echo "Created item: Coffee Maker (ID: $item5_id)"

# Items for group 2 users
echo "Creating items for group 2..."
item6=$(create_item "$user6_token" "Monitor" "27 inch 4K monitor" "$location3_id")
item6_id=$(echo "$item6" | jq -r '.id // empty')
echo "Created item: Monitor (ID: $item6_id)"

item7=$(create_item "$user6_token" "Keyboard" "Mechanical keyboard" "$location3_id")
item7_id=$(echo "$item7" | jq -r '.id // empty')
echo "Created item: Keyboard (ID: $item7_id)"

# Store items
jq --argjson group1_items "[\"$item1_id\",\"$item2_id\",\"$item3_id\",\"$item4_id\",\"$item5_id\"]" \
   --argjson group2_items "[\"$item6_id\",\"$item7_id\"]" \
   '.items = {"group1":$group1_items,"group2":$group2_items}' \
   "$TEST_DATA_FILE" > "$TEST_DATA_FILE.tmp" && mv "$TEST_DATA_FILE.tmp" "$TEST_DATA_FILE"

echo "=== Step 7: Add attachments to items ==="

# Add attachments for group 1 items
echo "Adding attachments to group 1 items..."
attach_file_to_item "$user1_token" "$item1_id" "laptop-receipt.pdf"
attach_file_to_item "$user1_token" "$item1_id" "laptop-warranty.pdf"
attach_file_to_item "$user1_token" "$item2_id" "drill-manual.pdf"
attach_file_to_item "$user1_token" "$item3_id" "remote-guide.pdf"
attach_file_to_item "$user1_token" "$item4_id" "toolbox-inventory.txt"

# Add attachments for group 2 items
echo "Adding attachments to group 2 items..."
attach_file_to_item "$user6_token" "$item6_id" "monitor-receipt.pdf"
attach_file_to_item "$user6_token" "$item7_id" "keyboard-manual.pdf"

echo "=== Test Data Creation Complete ==="
echo "Test data file saved to: $TEST_DATA_FILE"
echo "Summary:"
echo "  - Users created: 7 (5 in group 1, 2 in group 2)"
echo "  - Locations created: 3"
echo "  - Tags created: 3"
echo "  - Notifiers created: 1"
echo "  - Items created: 7"
echo "  - Attachments created: 7"

# Display the test data file for verification
echo ""
echo "Test data:"
cat "$TEST_DATA_FILE" | jq '.'

exit 0
