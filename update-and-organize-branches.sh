#!/bin/bash

# Script to update main branch from upstream and organize branches
# Created: $(date)

set -e  # Exit immediately if a command exits with a non-zero status
echo "Starting branch update and organization process..."

# Function to check if a branch exists
branch_exists() {
    git show-ref --verify --quiet refs/heads/$1
    return $?
}

# Function to check if a branch is clean (no uncommitted changes)
is_branch_clean() {
    # Ignore this script file itself when checking for uncommitted changes
    if [ -z "$(git status --porcelain | grep -v "update-and-organize-branches.sh")" ]; then
        return 0  # Clean
    else
        return 1  # Not clean
    fi
}

# Store current branch
CURRENT_BRANCH=$(git symbolic-ref --short HEAD)
echo "Current branch: $CURRENT_BRANCH"

# Check if current branch has uncommitted changes
if ! is_branch_clean; then
    echo "ERROR: You have uncommitted changes in your current branch."
    echo "Please commit or stash your changes before running this script."
    exit 1
fi

# Update main branch from upstream
echo -e "\n=== Updating main branch from upstream ==="
if branch_exists "main"; then
    git checkout main
    
    # First, ensure local main is up-to-date with origin/main
    echo "Updating local main from origin/main..."
    git pull origin main --ff-only || {
        echo "Cannot fast-forward local main to origin/main."
        echo "This could be due to local commits that aren't on origin."
        echo "Trying to reset local main to origin/main..."
        git fetch origin
        git reset --hard origin/main
    }
    
    # Now update from upstream
    echo "Updating from upstream/main..."
    git fetch upstream
    git merge upstream/main
    
    # Push to origin
    echo "Pushing updated main to origin..."
    git push origin main
    
    echo "Main branch updated successfully from upstream"
else
    echo "ERROR: Main branch does not exist locally"
    exit 1
fi

# Rename Docker-related branches
echo -e "\n=== Organizing Docker-related branches ==="
if branch_exists "custom-docker-image"; then
    git checkout custom-docker-image
    git branch -m docker/custom-image
    git push origin docker/custom-image
    git push origin --delete custom-docker-image || echo "Note: Could not delete remote branch custom-docker-image. It may not exist or you may not have permission."
    echo "Renamed 'custom-docker-image' to 'docker/custom-image'"
fi

# Rename feature branches
echo -e "\n=== Organizing feature branches ==="

# Array of feature branches to rename
FEATURE_BRANCHES=("asset-id-lookup" "clean-asset-id-toggle" "favorite-items" "fuzzy-search-logic")

for branch in "${FEATURE_BRANCHES[@]}"; do
    if branch_exists "$branch"; then
        git checkout "$branch"
        NEW_NAME="feature/$branch"
        git branch -m "$NEW_NAME"
        git push origin "$NEW_NAME"
        git push origin --delete "$branch" || echo "Note: Could not delete remote branch $branch. It may not exist or you may not have permission."
        echo "Renamed '$branch' to '$NEW_NAME'"
    else
        echo "Branch '$branch' does not exist locally, skipping"
    fi
done

# Return to original branch
echo -e "\n=== Returning to original branch ==="
if branch_exists "$CURRENT_BRANCH"; then
    git checkout "$CURRENT_BRANCH"
    echo "Returned to original branch: $CURRENT_BRANCH"
elif [[ "$CURRENT_BRANCH" == "custom-docker-image" ]]; then
    git checkout docker/custom-image
    echo "Original branch was renamed, now on: docker/custom-image"
elif [[ " ${FEATURE_BRANCHES[@]} " =~ " ${CURRENT_BRANCH} " ]]; then
    git checkout "feature/$CURRENT_BRANCH"
    echo "Original branch was renamed, now on: feature/$CURRENT_BRANCH"
else
    git checkout main
    echo "Original branch no longer exists, checked out main instead"
fi

echo -e "\n=== Branch organization complete ==="
echo "Your branches have been organized according to the new naming convention:"
echo "- Docker branches: docker/*"
echo "- Feature branches: feature/*"
echo "- Main branch: updated from upstream"

# List all branches for reference
echo -e "\nCurrent branches:"
git branch -a
