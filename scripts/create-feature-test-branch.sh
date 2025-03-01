#!/bin/bash

# create-feature-test-branch.sh
# Script to create a test branch for a specific feature based on docker/custom-image

set -e  # Exit on error

# Check if a feature branch was provided
if [ $# -lt 1 ]; then
    echo "Usage: $0 <feature-branch-name> [test-branch-suffix]"
    echo "Example: $0 feature/homepage-accordions test"
    echo "This will create docker/testing-homepage-accordions-test branch"
    exit 1
fi

FEATURE_BRANCH=$1
TEST_SUFFIX=${2:-test}  # Default suffix is "test" if not provided

# Extract the feature name from the branch name
if [[ $FEATURE_BRANCH == feature/* ]]; then
    FEATURE_NAME=${FEATURE_BRANCH#feature/}
else
    FEATURE_NAME=$FEATURE_BRANCH
fi

# Create a properly named test branch
TEST_BRANCH="docker/testing-${FEATURE_NAME}-${TEST_SUFFIX}"

# Base branch for testing
BASE_BRANCH="docker/custom-image"

echo "Creating test branch $TEST_BRANCH based on $BASE_BRANCH with feature $FEATURE_BRANCH"

# Make sure we're starting from a clean state
git fetch origin
git checkout $BASE_BRANCH
git pull origin $BASE_BRANCH || true  # Continue even if there's no remote tracking

# Create the test branch
if git show-ref --verify --quiet refs/heads/$TEST_BRANCH; then
    echo "Branch $TEST_BRANCH already exists. Resetting it to match $BASE_BRANCH."
    git checkout $TEST_BRANCH
    git reset --hard $BASE_BRANCH
else
    echo "Creating new branch $TEST_BRANCH from $BASE_BRANCH."
    git checkout -b $TEST_BRANCH $BASE_BRANCH
fi

# Check if the feature branch exists
if ! git show-ref --verify --quiet refs/heads/$FEATURE_BRANCH; then
    echo "Error: Feature branch $FEATURE_BRANCH does not exist."
    exit 1
fi

# Try to cherry-pick or merge the feature branch
echo "Attempting to cherry-pick commits from $FEATURE_BRANCH..."

# Get the commit hash of the latest commit in the feature branch
FEATURE_COMMIT=$(git rev-parse $FEATURE_BRANCH)

# Try cherry-picking
if git cherry-pick $FEATURE_COMMIT; then
    echo "✅ Successfully cherry-picked $FEATURE_BRANCH"
else
    echo "⚠️ Cherry-pick failed. Trying merge instead..."
    git cherry-pick --abort
    
    if git merge --no-ff $FEATURE_BRANCH -m "Merge $FEATURE_BRANCH into $TEST_BRANCH for testing"; then
        echo "✅ Successfully merged $FEATURE_BRANCH"
    else
        echo "⚠️ Merge failed as well. Manual intervention required."
        echo "Please resolve conflicts and commit the changes."
        exit 1
    fi
fi

echo "Test branch $TEST_BRANCH created and updated with $FEATURE_BRANCH!"
echo "You can now build and test your Docker image from this branch:"
echo "docker build -t homebox-testing-${FEATURE_NAME} ."
