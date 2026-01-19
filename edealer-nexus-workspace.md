# eDealer Nexus Workspace Setup

This document describes the nexus workspace configuration created for eDealer on linuxbox host.

## Files Created

The following files have been created in `examples/epson-eshop/edealer/` to enable nexus workspace support:

### 1. Nexus Configuration (`.nexus/config.yaml`)
Complete workspace configuration for eDealer including:
- Provider: LXC (provider-agnostic remote support on linuxbox)
- Services: Web (Puma), Worker (Sidekiq), Cron Worker, PostgreSQL, Redis
- Port mappings for local access:
  - Web: `linuxbox:23100`
  - PostgreSQL: `linuxbox:23101`
  - Redis: `linuxbox:23102`
- Resource allocation: 4 CPUs, 4GB RAM, 20GB disk
- Database initialization with seed data
- SSH access configuration

### 2. Setup Guide (`.nexus/SETUP_GUIDE.md`)
Comprehensive guide for:
- Creating and connecting to workspaces
- Accessing the application
- Running tests (RSpec and Playwright)
- Database management
- Debugging and troubleshooting

### 3. Playwright Configuration (`playwright.config.ts`)
E2E testing setup with:
- Multiple browsers: Chromium, Firefox, WebKit
- Mobile testing: Chrome and Safari
- Base URL: `http://localhost:5000`
- HTML report generation
- Auto server startup

### 4. Test Suite (`tests/auth.spec.ts`)
Example Playwright tests covering:
- Admin login with seed credentials
- User portal navigation
- Error handling
- Page performance
- API health checks

### 5. Package Configuration (`package.json`)
NPM package definition with:
- Playwright test dependency (`@playwright/test@^1.48.0`)
- Custom test scripts:
  - `npm run playwright:test` - Headless tests
  - `npm run playwright:headed` - Interactive tests
  - `npm run playwright:debug` - Debug mode
  - `npm run playwright:report` - View HTML report

## Quick Start

### Create and Connect to Workspace

```bash
# Create the workspace
nexus workspace create --config examples/epson-eshop/edealer/.nexus/config.yaml edealer

# Connect to workspace
nexus workspace connect edealer
```

### Access the Application

Once workspace is running:

- **User Portal**: http://linuxbox:23100
- **Admin Portal**: http://linuxbox:23100/admins
- **Admin Credentials**: admin@example.com / 1234Qwer!

### Run Tests

```bash
# Install dependencies
nexus workspace shell edealer -c "cd /workspace && npm install"

# Run Playwright tests (headless)
nexus workspace shell edealer -c "cd /workspace && npm run playwright:test"

# Run Playwright tests (interactive)
nexus workspace shell edealer -c "cd /workspace && npm run playwright:headed"

# View test report
nexus workspace shell edealer -c "cd /workspace && npm run playwright:report"
```

### Database Management

```bash
# Reset database with fresh seed data
nexus workspace shell edealer -c "cd /workspace && bundle exec rails db:reset"

# View database logs
nexus workspace shell edealer -c "tail -f /workspace/log/development.log"
```

## Seed Data

The eDealer application includes seed data for testing:

### Admin Account
- Email: `admin@example.com`
- Password: `1234Qwer!`
- Access: http://linuxbox:23100/admins

### User Accounts
Defined in `db/seeds/development/user.seeds.csv`:
- See individual seed files for complete account list
- Seed files location: `db/seeds/development/*.seeds.rb`

## Testing Strategy

### Unit & Integration Tests
```bash
# Run RSpec tests
nexus workspace shell edealer -c "cd /workspace && bundle exec rspec"

# With coverage
nexus workspace shell edealer -c "cd /workspace && make test report-coverage"
```

### End-to-End Tests (Playwright)
The Playwright test suite (`tests/auth.spec.ts`) validates:
1. Authentication flows (admin & user logins)
2. Error handling for invalid credentials
3. Session management
4. Navigation between portals
5. Page load performance
6. API health endpoint

## Performance Metrics

- Web service port: 23100 (5000 internal)
- Database port: 23101 (5432 internal)
- Cache port: 23102 (6379 internal)
- Page load time goal: < 5 seconds
- Application startup: ~30 seconds
- Service health check interval: 10 seconds

## Troubleshooting

### Database Connection Failed
```bash
# Verify PostgreSQL service
nexus workspace shell edealer -c "pg_isready -h localhost -p 5432"

# Check database configuration
nexus workspace shell edealer -c "cd /workspace && bundle exec rails db:status"
```

### Redis Connection Failed
```bash
# Test Redis connectivity
nexus workspace shell edealer -c "redis-cli ping"

# View Redis info
nexus workspace shell edealer -c "redis-cli info"
```

### Playwright Tests Failing
```bash
# Install Chromium browser dependencies
nexus workspace shell edealer -c "npx playwright install"

# Run with verbose logging
nexus workspace shell edealer -c "cd /workspace && npm run playwright:test -- --verbose"
```

## Integration with CI/CD

For CI/CD pipelines, use headless mode:

```bash
nexus workspace shell edealer -c "cd /workspace && npm run playwright:test"
```

Tests will:
- Run in headless chromium
- Generate HTML reports in `playwright-report/`
- Retry failures up to 2 times
- Use single worker for reliability
- Generate trace files for debugging

## Next Steps

1. **Create Workspace**: `nexus workspace create --config examples/epson-eshop/edealer/.nexus/config.yaml edealer`
2. **Connect**: `nexus workspace connect edealer`
3. **Run Tests**: `npm run playwright:test`
4. **Review Reports**: `npm run playwright:report`

See `examples/epson-eshop/edealer/.nexus/SETUP_GUIDE.md` for detailed commands and advanced usage.
