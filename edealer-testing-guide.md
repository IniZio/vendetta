# Testing eDealer on Fresh Remote Machine - Complete Guide

This guide walks you through testing the eDealer Nexus workspace setup from a completely fresh remote machine.

## Prerequisites

Before starting, ensure your remote machine has:

- **OS**: Ubuntu 22.04 or compatible Linux
- **CPU**: 4+ cores (for LXC containers)
- **RAM**: 8GB+ (for running workspace + services)
- **Disk**: 50GB+ free space
- **Network**: SSH access to nexus host

### Required Software

You need to install on the remote machine:

```bash
# Install Go (required for nexus binary)
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Git
sudo apt-get update
sudo apt-get install -y git

# Install Docker (optional, for workspace provider flexibility)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install LXC/LXD (recommended for this setup)
sudo snap install lxd --classic
sudo lxd init --auto
```

## Step 1: Clone the Repository

```bash
# Create a working directory
mkdir -p ~/projects
cd ~/projects

# Clone the nexus repository
git clone https://github.com/IniZio/nexus.git
cd nexus

# Verify you're on main branch
git branch -a
git log --oneline -3
```

**Expected Output:**
```
* main
  origin/main
1fb5ae3 feat(edealer): add nexus workspace configuration for playwright e2e testing on linuxbox
7eb0340 feat(lxc): add port forwarding support with LXC proxy devices
8a91527 feat(ssh): auto-configure SSH access with GitHub public keys
```

## Step 2: Verify eDealer Configuration Files

```bash
# Check the edealer workspace configuration exists
ls -la examples/epson-eshop/edealer/.nexus/

# View the main configuration
cat examples/epson-eshop/edealer/.nexus/config.yaml | head -50

# View the setup guide
cat examples/epson-eshop/edealer/.nexus/SETUP_GUIDE.md | head -30

# Check playwright configuration
cat examples/epson-eshop/edealer/playwright.config.ts

# Check test files
ls -la examples/epson-eshop/edealer/tests/

# Check package.json
cat examples/epson-eshop/edealer/package.json
```

**Expected Files:**
```
examples/epson-eshop/edealer/.nexus/
├── config.yaml          (221 lines - main workspace config)
└── SETUP_GUIDE.md       (Comprehensive setup documentation)

examples/epson-eshop/edealer/
├── playwright.config.ts (54 lines - playwright setup)
├── package.json         (17 lines - npm dependencies)
└── tests/
    └── auth.spec.ts     (Test suite with auth flows)
```

## Step 3: Build Nexus Binary

```bash
# Build the nexus binary
make build

# Verify build succeeded
./bin/nexus --version

# Check binary works
./bin/nexus coordination start --help
```

**Expected Output:**
```
nexus version v0.X.X
Usage: nexus coordination start [options]
```

## Step 4: Setup Coordination Server

```bash
# Start the coordination server in background
./bin/nexus coordination start &

# Wait for it to start (should take ~5 seconds)
sleep 5

# Verify server is running
curl -s http://localhost:3001/health | jq .

# Check it's listening
netstat -tuln | grep 3001
```

**Expected Output:**
```
LISTEN  ... 0.0.0.0:3001 ...
```

## Step 5: Create eDealer Workspace

This is the main test - creating a workspace with the edealer configuration.

```bash
# Create the workspace
./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  edealer

# Wait for workspace to be created (this takes 2-3 minutes)
# You should see output showing:
# - Container/VM creation
# - Service startup
# - Database initialization
# - Health checks passing

# Verify workspace was created
./bin/nexus workspace list

# Check workspace details
./bin/nexus workspace show edealer
```

**Expected Output:**
```
✓ Workspace created: edealer
  Provider: lxc
  Status: running
  Services: web, worker, postgres, redis, cron-worker
```

## Step 6: Connect to Workspace (SSH)

```bash
# Get SSH connection details
./bin/nexus workspace show edealer --ssh

# Copy the SSH key path (usually ~/.ssh/nexus_edealer_key)
# SSH into the workspace
ssh -i ~/.ssh/nexus_edealer_key dev@<workspace-host>

# Once connected, verify services are running
docker ps  # if using docker
lxc list   # if using lxc

# Check database connection
psql -h db -U postgres -d edealer_development -c "SELECT version();"

# Check Redis connection
redis-cli -h redis-db ping

# Exit the SSH session
exit
```

## Step 7: Access Web Application

```bash
# Get workspace network info
./bin/nexus workspace show edealer

# Access the application
# User Portal: http://<your-machine-ip>:23100
# Admin Portal: http://<your-machine-ip>:23100/admins

# For local testing, use:
curl -s http://localhost:23100 | head -20

# Verify health endpoint
curl -s http://localhost:23100/up | jq .
```

**Expected Output:**
```
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

## Step 8: Setup Playwright Tests

```bash
# SSH into workspace
./bin/nexus workspace shell edealer

# Navigate to app directory
cd /workspace

# Install Playwright dependencies
npm install

# Verify installation
npm list @playwright/test
```

**Expected Output:**
```
edealer@1.0.0 /workspace
└── @playwright/test@1.48.0
```

## Step 9: Run Playwright Tests

```bash
# From workspace shell, run tests in headless mode
npm run playwright:test

# OR run specific test file
npx playwright test tests/auth.spec.ts

# Watch the output for test results
# Expected: All tests should PASS with seed data
```

**Expected Test Output:**
```
✓ eDealer Authentication Flows
  ✓ Admin Login › should login with valid admin credentials
  ✓ Admin Login › should show error with invalid credentials
  ✓ Admin Login › should logout successfully
  ✓ User Login › should access user portal
  ✓ Navigation › should navigate between portals
  ✓ Page Load Performance › should load user portal within acceptable time
  ✓ API Health › should respond to health check endpoint

7 passed (2.3s)
```

## Step 10: Generate Test Report

```bash
# From workspace shell, generate HTML report
npm run playwright:report

# Exit workspace shell
exit

# Copy report to your local machine (if needed)
./bin/nexus workspace shell edealer -c "cd /workspace && tar czf playwright-report.tar.gz playwright-report/"
./bin/nexus workspace copy edealer /workspace/playwright-report.tar.gz ./
tar xzf playwright-report.tar.gz
# Open playwright-report/index.html in your browser
```

## Step 11: Test Admin Account Login (Manual)

```bash
# Open browser to http://localhost:23100/admins

# Use seed credentials:
# Email: admin@example.com
# Password: 1234Qwer!

# Verify you can:
# 1. Login successfully
# 2. Access dashboard
# 3. Navigate to different sections
# 4. Logout

# Check the logs
./bin/nexus workspace shell edealer -c "tail -f /workspace/log/development.log"
```

## Step 12: Verify Database and Services

```bash
# Check all services are healthy
./bin/nexus workspace show edealer --services

# Expected output:
# web:      RUNNING  http://localhost:23100
# worker:   RUNNING  (background)
# postgres: RUNNING  :23101
# redis:    RUNNING  :23102

# Verify database has seed data
./bin/nexus workspace shell edealer -c "cd /workspace && bundle exec rails db:seed:dump"

# Check admin exists
./bin/nexus workspace shell edealer -c "cd /workspace && bundle exec rails runner 'puts Admin.pluck(:email).inspect'"

# Expected output: ["admin@example.com"]
```

## Step 13: Run Additional Tests (RSpec)

```bash
# From workspace shell
cd /workspace

# Run RSpec tests (if you want to test backend too)
bundle exec rspec

# With coverage report
COVERAGE=true bundle exec rspec

# View coverage
cat coverage/index.html
```

## Troubleshooting

### Issue: Workspace creation fails

```bash
# Check coordination server is running
curl http://localhost:3001/health

# Check LXC/Docker availability
lxc list  # for LXC
docker ps # for Docker

# View detailed workspace logs
./bin/nexus workspace show edealer --logs

# Restart coordination server
killall nexus
./bin/nexus coordination start &
```

### Issue: Cannot connect to database

```bash
# Check database service is running
./bin/nexus workspace shell edealer -c "pg_isready -h db -p 5432"

# Check database URL is correct
./bin/nexus workspace shell edealer -c "echo $DATABASE_URL"

# Try connecting directly
./bin/nexus workspace shell edealer -c "psql $DATABASE_URL -c 'SELECT version();'"
```

### Issue: Redis connection failed

```bash
# Verify Redis is running
./bin/nexus workspace shell edealer -c "redis-cli -h redis-db ping"

# Check Redis configuration
./bin/nexus workspace shell edealer -c "redis-cli -h redis-db config get maxmemory"
```

### Issue: Playwright tests fail

```bash
# Check Chromium is installed
./bin/nexus workspace shell edealer -c "npx playwright install"

# Run tests with verbose output
./bin/nexus workspace shell edealer -c "cd /workspace && npx playwright test --verbose"

# Check browser compatibility
./bin/nexus workspace shell edealer -c "cd /workspace && npx playwright test --list"
```

### Issue: Port conflicts (23100, 23101, 23102)

If ports are already in use, either:

1. **Stop existing workspace:**
   ```bash
   ./bin/nexus workspace delete edealer
   ```

2. **Modify port mappings** in `examples/epson-eshop/edealer/.nexus/config.yaml`:
   ```yaml
   ports:
     web: 24100      # Changed from 23100
     postgres: 24101 # Changed from 23101
     redis: 24102    # Changed from 23102
   ```

3. **Recreate workspace** with new ports

## Success Checklist

- [ ] Repository cloned successfully
- [ ] nexus binary built and runs
- [ ] Coordination server started
- [ ] eDealer workspace created and running
- [ ] Can SSH into workspace
- [ ] Can access http://localhost:23100 (User Portal)
- [ ] Can access http://localhost:23100/admins (Admin Portal)
- [ ] Can login with admin@example.com / 1234Qwer!
- [ ] Playwright tests pass (7/7)
- [ ] Test report generated successfully
- [ ] Database seeds loaded correctly
- [ ] All 5 services healthy (web, worker, postgres, redis, cron-worker)

## Performance Expectations

| Operation | Expected Time | Status |
|-----------|--------------|--------|
| Build binary | 2-5 minutes | ⏱️ |
| Coordination server start | ~5 seconds | ⏱️ |
| Workspace creation | 2-3 minutes | ⏱️ |
| Web service startup | ~30 seconds | ⏱️ |
| Database initialization | ~30 seconds | ⏱️ |
| Page load time | < 5 seconds | ⏱️ |
| Playwright test suite | ~2-3 minutes | ⏱️ |
| Report generation | ~1 minute | ⏱️ |
| **Total time (first run)** | **~15 minutes** | ⏱️ |

## Next Steps After Testing

1. **If tests pass:** Great! The workspace is fully functional
   - Proceed with production deployment
   - Integrate tests into CI/CD pipeline
   - Add more test cases as needed

2. **If tests fail:** 
   - Check troubleshooting section above
   - Review workspace logs: `./bin/nexus workspace show edealer --logs`
   - Check individual service logs
   - Report issue with full error output

3. **To run tests repeatedly:**
   ```bash
   # Quick test run
   ./bin/nexus workspace shell edealer -c "cd /workspace && npm run playwright:test"
   
   # With browser visualization
   ./bin/nexus workspace shell edealer -c "cd /workspace && npm run playwright:headed"
   
   # Single test file
   ./bin/nexus workspace shell edealer -c "cd /workspace && npx playwright test tests/auth.spec.ts"
   ```

4. **To cleanup:**
   ```bash
   # Delete workspace
   ./bin/nexus workspace delete edealer
   
   # Stop coordination server
   pkill nexus
   ```

## Documentation References

- Detailed Setup: `examples/epson-eshop/edealer/.nexus/SETUP_GUIDE.md`
- eDealer App README: `examples/epson-eshop/edealer/README.md`
- Nexus Main README: `README.md`
- Playwright Docs: https://playwright.dev/
- eDealer Repository: https://github.com/oursky/epson-shop

---

**Ready to test? Start at Step 1 above and follow through to Step 13!**
