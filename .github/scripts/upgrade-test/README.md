# HomeBox Upgrade Testing Workflow

This document describes the automated upgrade testing workflow for HomeBox.

## Overview

The upgrade test workflow is designed to ensure data integrity and functionality when upgrading HomeBox from one version to another. It automatically:

1. Deploys a stable version of HomeBox
2. Creates test data (users, items, locations, tags, notifiers, attachments)
3. Upgrades to the latest version from the main branch
4. Verifies all data and functionality remain intact

## Workflow File

**Location**: `.github/workflows/upgrade-test.yaml`

## Trigger Conditions

The workflow runs:
- **Daily**: Automatically at 2 AM UTC (via cron schedule)
- **Manual**: Can be triggered manually via GitHub Actions UI
- **On Push**: When changes are made to the workflow files or test scripts

## Test Scenarios

### 1. Environment Setup
- Pulls the latest stable HomeBox Docker image from GHCR
- Starts the application with test configuration
- Ensures the service is healthy and ready

### 2. Data Creation

The workflow creates comprehensive test data using the `create-test-data.sh` script:

#### Users and Groups
- **Group 1**: 5 users (user1@homebox.test through user5@homebox.test)
- **Group 2**: 2 users (user6@homebox.test and user7@homebox.test)
- All users have password: `TestPassword123!`

#### Locations
- **Group 1**: Living Room, Garage
- **Group 2**: Home Office

#### Tags
- **Group 1**: Electronics, Important
- **Group 2**: Work Equipment

#### Items
- **Group 1**: 5 items (Laptop Computer, Power Drill, TV Remote, Tool Box, Coffee Maker)
- **Group 2**: 2 items (Monitor, Keyboard)

#### Attachments
- Multiple attachments added to various items (receipts, manuals, warranties)

#### Notifiers
- **Group 1**: Test notifier named "TESTING"

### 3. Upgrade Process

1. Stops the stable version container
2. Builds a fresh image from the current main branch
3. Copies the database to a new location
4. Starts the new version with the existing data

### 4. Verification Tests

The Playwright test suite (`upgrade-verification.spec.ts`) verifies:

- ✅ **User Authentication**: All 7 users can log in with their credentials
- ✅ **Data Persistence**: All items, locations, and tags are present
- ✅ **Attachments**: File attachments are correctly associated with items
- ✅ **Notifiers**: The "TESTING" notifier is still configured
- ✅ **UI Functionality**: Version display, theme switching work correctly
- ✅ **Data Isolation**: Groups can only see their own data

## Test Data File

The setup script generates a JSON file at `/tmp/test-users.json` containing:

```json
{
  "users": [
    {
      "email": "user1@homebox.test",
      "password": "TestPassword123!",
      "token": "...",
      "group": "1"
    },
    ...
  ],
  "locations": {
    "group1": ["location-id-1", "location-id-2"],
    "group2": ["location-id-3"]
  },
  "tags": {...},
  "items": {...},
  "notifiers": {...}
}
```

This file is used by the Playwright tests to verify data integrity.

## Scripts

### create-test-data.sh

**Location**: `.github/scripts/upgrade-test/create-test-data.sh`

**Purpose**: Creates all test data via the HomeBox REST API

**Environment Variables**:
- `HOMEBOX_URL`: Base URL of the HomeBox instance (default: http://localhost:7745)
- `TEST_DATA_FILE`: Path to output JSON file (default: /tmp/test-users.json)

**Requirements**:
- `curl`: For API calls
- `jq`: For JSON processing

**Usage**:
```bash
export HOMEBOX_URL=http://localhost:7745
./.github/scripts/upgrade-test/create-test-data.sh
```

## Running Tests Locally

To run the upgrade tests locally:

### Prerequisites
```bash
# Install dependencies
sudo apt-get install -y jq curl docker.io

# Install pnpm and Playwright
cd frontend
pnpm install
pnpm exec playwright install --with-deps chromium
```

### Run the test
```bash
# Start stable version
docker run -d \
  --name homebox-test \
  -p 7745:7745 \
  -e HBOX_OPTIONS_ALLOW_REGISTRATION=true \
  -v /tmp/homebox-data:/data \
  ghcr.io/sysadminsmedia/homebox:latest

# Wait for startup
sleep 10

# Create test data
export HOMEBOX_URL=http://localhost:7745
./.github/scripts/upgrade-test/create-test-data.sh

# Stop container
docker stop homebox-test
docker rm homebox-test

# Build new version
docker build -t homebox:test .

# Start new version with existing data
docker run -d \
  --name homebox-test \
  -p 7745:7745 \
  -e HBOX_OPTIONS_ALLOW_REGISTRATION=true \
  -v /tmp/homebox-data:/data \
  homebox:test

# Wait for startup
sleep 10

# Run verification tests
cd frontend
TEST_DATA_FILE=/tmp/test-users.json \
E2E_BASE_URL=http://localhost:7745 \
pnpm exec playwright test \
  --project=chromium \
  test/upgrade/upgrade-verification.spec.ts

# Cleanup
docker stop homebox-test
docker rm homebox-test
```

## Artifacts

The workflow produces several artifacts:

1. **playwright-report-upgrade-test**: HTML report of test results
2. **playwright-traces**: Detailed traces for debugging failures
3. **Docker logs**: Collected on failure for troubleshooting

## Failure Scenarios

The workflow will fail if:
- The stable version fails to start
- Test data creation fails
- The new version fails to start with existing data
- Any verification test fails
- Database migrations fail

## Troubleshooting

### Test Data Creation Fails

Check the Docker logs:
```bash
docker logs homebox-old
```

Verify the API is accessible:
```bash
curl http://localhost:7745/api/v1/status
```

### Verification Tests Fail

1. Download the Playwright report from GitHub Actions artifacts
2. Review the HTML report for detailed failure information
3. Check traces for visual debugging

### Database Issues

If migrations fail:
```bash
# Check database file
ls -lh /tmp/homebox-data-new/homebox.db

# Check Docker logs for migration errors
docker logs homebox-new
```

## Future Enhancements

Potential improvements:
- [ ] Test multiple upgrade paths (e.g., v0.10 → v0.11 → v0.12)
- [ ] Test with PostgreSQL backend in addition to SQLite
- [ ] Add performance benchmarks
- [ ] Test with larger datasets
- [ ] Add API-level verification in addition to UI tests
- [ ] Test backup and restore functionality

## Related Files

- `.github/workflows/upgrade-test.yaml` - Main workflow definition
- `.github/scripts/upgrade-test/create-test-data.sh` - Data generation script
- `frontend/test/upgrade/upgrade-verification.spec.ts` - Playwright verification tests
- `.github/workflows/e2e-partial.yaml` - Standard E2E test workflow (for reference)

## Support

For issues or questions about this workflow:
1. Check the GitHub Actions run logs
2. Review this documentation
3. Open an issue in the repository
