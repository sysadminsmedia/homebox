#!/bin/bash

# update-docker-testing-branch.sh
# Script to update the docker/testing-image branch with selected feature branches

set -e  # Exit on error

# Configuration
DOCKER_BRANCH="docker/testing-image"
MAIN_BRANCH="main"

# Make sure we're starting from a clean state
git fetch origin
git checkout $MAIN_BRANCH
git pull origin $MAIN_BRANCH

# Checkout or create the docker testing branch
if git show-ref --verify --quiet refs/heads/$DOCKER_BRANCH; then
    git checkout $DOCKER_BRANCH
    git reset --hard origin/$MAIN_BRANCH  # Reset to match main
else
    git checkout -b $DOCKER_BRANCH origin/$MAIN_BRANCH
fi

# Function to merge a feature branch
merge_feature() {
    local branch=$1
    echo "Merging $branch into $DOCKER_BRANCH..."
    
    # Try to merge, but don't fail the script if there are conflicts
    if git merge --no-ff $branch -m "Merge $branch into $DOCKER_BRANCH for testing"; then
        echo "✅ Successfully merged $branch"
    else
        echo "⚠️ Conflicts detected when merging $branch"
        echo "Aborting this merge. Please resolve conflicts manually if needed."
        git merge --abort
    fi
}

# List of feature branches to merge
# Add or remove branches as needed
FEATURE_BRANCHES=(
    "feature/favorite-items"
    # Add more feature branches here
    # "feature/another-feature"
)

# Merge each feature branch
for branch in "${FEATURE_BRANCHES[@]}"; do
    # Check if the branch exists
    if git show-ref --verify --quiet refs/heads/$branch; then
        merge_feature $branch
    else
        echo "⚠️ Branch $branch does not exist locally. Skipping."
    fi
done

echo "Docker testing branch updated!"
echo "You can now build and test your Docker image from this branch."
