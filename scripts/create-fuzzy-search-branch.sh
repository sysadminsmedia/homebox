#!/bin/bash

# Script to create a feature branch for implementing fuzzy search logic
# This will set up the initial branch structure and files needed for fuzzy search

set -e

# Configuration
FEATURE_BRANCH="feature/fuzzy-search"
MAIN_BRANCH="main"

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -p, --push              Push the new branch to remote after creation"
    echo
    echo "This script creates a new feature branch for implementing fuzzy search logic."
    exit 0
}

# Parse command line arguments
PUSH_BRANCH=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            ;;
        -p|--push)
            PUSH_BRANCH=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Save the current branch to return to it later
CURRENT_BRANCH=$(git branch --show-current)

# Function to clean up and return to the original branch
cleanup() {
    if [ $? -ne 0 ]; then
        echo "An error occurred. Cleaning up..."
    fi
    echo "Returning to branch: $CURRENT_BRANCH"
    git checkout "$CURRENT_BRANCH"
}

# Set up trap to ensure we return to the original branch on exit
trap cleanup EXIT

# Check if the feature branch already exists
if git show-ref --verify --quiet "refs/heads/$FEATURE_BRANCH"; then
    echo "Branch $FEATURE_BRANCH already exists"
    read -p "Do you want to delete it and create a new one? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git branch -D "$FEATURE_BRANCH"
    else
        echo "Aborting."
        exit 1
    fi
fi

# Create the feature branch from main
echo "Creating branch $FEATURE_BRANCH from $MAIN_BRANCH"
git checkout "$MAIN_BRANCH"
git pull origin "$MAIN_BRANCH"
git checkout -b "$FEATURE_BRANCH"

# Create initial files for fuzzy search implementation
echo "Setting up initial files for fuzzy search implementation"

# Create a README file for the feature
cat > FUZZY_SEARCH_README.md << 'EOF'
# Fuzzy Search Feature

This feature adds fuzzy search capabilities to the Homebox application, allowing users to find items even when they don't remember the exact spelling or complete name.

## Implementation Details

The fuzzy search implementation uses:

1. Frontend fuzzy matching for immediate results
2. Backend fuzzy search for more comprehensive results

## Files Modified

- `frontend/pages/items.vue`: Added fuzzy search toggle and UI elements
- `frontend/composables/useFuzzySearch.ts`: Added fuzzy search logic
- `backend/app/api/handlers/v1/v1_ctrl_items.go`: Modified search endpoint to support fuzzy matching

## Usage

To use fuzzy search:

1. Navigate to the Items page
2. Toggle on "Fuzzy Search" in the search options
3. Enter your search query - results will include close matches

EOF

# Push the branch if requested
if [ "$PUSH_BRANCH" = true ]; then
    echo "Pushing branch to remote"
    git add FUZZY_SEARCH_README.md
    git commit -m "feat: initial setup for fuzzy search feature"
    git push -u origin "$FEATURE_BRANCH"
fi

echo "Done! Feature branch $FEATURE_BRANCH has been created"
echo "You can now start implementing the fuzzy search feature"
