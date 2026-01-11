# Milestone: M3_Beta (Multi-Machine Orchestration)

**Objective**: Extend Vendatta beyond single-machine isolation to support multi-machine coordination and advanced virtualization providers. This phase introduces QEMU-based full machine emulation for complex development scenarios requiring different OS/architectures or complete system isolation.

**Acceptance Criteria** (per user stories):
- Developers can test cross-platform applications (Linux → macOS, x86 → ARM)
- Teams can run distributed systems (web server + database + cache on separate machines)
- CI/CD can replicate production multi-machine topologies locally
- Performance-critical apps can be developed with accurate hardware simulation

## 🎯 Success Criteria
- [ ] QEMU provider supports full machine emulation with custom OS images.
- [ ] Multi-machine sessions allow coordination between isolated machines.
- [ ] Coordination protocols enable service discovery across machines.
- [ ] CLI supports multi-machine workspace management (`create-multi`, `connect`, `disconnect`).
- [ ] Performance benchmarks: Machine startup <60s, cross-machine latency <100ms.
- [ ] Deterministic multi-machine recreation via enhanced lockfile.

## 🛠 Implementation Tasks

| ID | Title | Priority | Status | Test Plan |
| :--- | :--- | :--- | :--- | :--- |
| **QEM-01** | QEMU Provider Implementation | 🔥 High | [🚧 Pending] | [TP-QEM-01](#test-plan-qem-01) |
| **MUL-01** | Multi-Machine Session Management | 🔥 High | [🚧 Pending] | [TP-MUL-01](#test-plan-mul-01) |
| **COR-01** | Cross-Machine Coordination | ⚡ Med | [🚧 Pending] | [TP-COR-01](#test-plan-cor-01) |
| **CLI-05** | Multi-Machine CLI Commands | ⚡ Med | [🚧 Pending] | [TP-CLI-05](#test-plan-cli-05) |
| **LCK-02** | Enhanced Lockfile for Multi-Machine | ⚡ Med | [🚧 Pending] | [TP-LCK-02](#test-plan-lck-02) |

## 🔗 Task Dependencies
- **QEM-01** depends on INF infrastructure from M1/M2
- **MUL-01** depends on QEM-01 and COR-01
- **CLI-05** depends on MUL-01
- **LCK-02** depends on LCK-01 from M2
- **COR-01** depends on QEM-01 and networking from INF

## 📋 Detailed Test Plans

### **TP-QEM-01: QEMU Provider Implementation**
**Objective**: Enable full machine emulation for complex development scenarios.

**Requirements:**
- Support popular OS images: Ubuntu, Alpine, Debian, CentOS, Fedora
- Automatic image download and caching from trusted sources
- Custom image support via local files or URLs
- Image versioning and updates

**Unit Tests:**
- ✅ **QEMU Launch**: Successfully start QEMU VMs with custom images
- ✅ **Image Management**: Download and cache OS images with integrity verification
- ✅ **Network Configuration**: Bridge networking for cross-machine communication
- ✅ **Resource Limits**: CPU/memory allocation and monitoring

**Integration Tests:**
- ✅ **VM Lifecycle**: Create/start/stop/destroy QEMU machines
- ✅ **SSH Access**: Automatic SSH key setup and connection
- ✅ **Service Installation**: Package managers work inside VMs

**E2E Scenarios:**
```bash
# Test 1: QEMU Machine Creation
# 1. Run 'vendatta workspace create-multi test --machines 2 --provider qemu'
# 2. Verify 2 QEMU VMs start with unique IPs
# 3. SSH into each machine successfully
# 4. Install packages and run services
# Expected: VMs running, network accessible, startup <60s
```

---

### **TP-MUL-01: Multi-Machine Session Management**
**Objective**: Manage coordinated multi-machine environments.

**Unit Tests:**
- ✅ **Session Creation**: Create sessions with multiple machines
- ✅ **Machine Roles**: Assign roles (master, worker, database, etc.)
- ✅ **Dependency Management**: Start machines in correct order
- ✅ **Health Monitoring**: Track machine status and auto-recovery

**Integration Tests:**
- ✅ **Session Persistence**: Save/restore multi-machine sessions
- ✅ **Concurrent Operations**: Parallel machine management
- ✅ **Failure Handling**: Graceful degradation when machines fail

---

### **TP-COR-01: Cross-Machine Coordination**
**Objective**: Enable service discovery and communication across machines.

**Unit Tests:**
- ✅ **Service Registry**: Register services across machines
- ✅ **DNS Resolution**: Automatic DNS for cross-machine services
- ✅ **Load Balancing**: Distribute requests across machine instances
- ✅ **Security**: Encrypted communication channels

**Integration Tests:**
- ✅ **Multi-Tier Apps**: Web server on one machine, database on another
- ✅ **Microservices**: Service mesh across multiple VMs
- ✅ **Data Replication**: Database clusters spanning machines

**E2E Scenarios:**
```bash
# Test 1: Multi-Machine Web App
# 1. Create session with web VM and db VM
# 2. Deploy app to web VM, database to db VM
# 3. Configure cross-machine networking
# 4. Verify app connects to database
# Expected: Full app functional, latency <100ms
```

---

### **TP-CLI-05: Multi-Machine CLI Commands**
**Objective**: Extend CLI for multi-machine operations.

**Requirements:**
- ✅ **`create-multi`**: Create workspaces with multiple machines
- ✅ **`connect`**: SSH into specific machines
- ✅ **`disconnect`**: Graceful machine shutdown
- ✅ **`status`**: Show multi-machine session status
- ✅ **`scale`**: Add/remove machines from running sessions

**Test Plan:**
```bash
# Test CLI commands work correctly
vendatta workspace create-multi test --machines 3
vendatta workspace connect test machine-1
vendatta workspace scale test --add 1
vendatta workspace status test  # Shows 4 machines
```

---

### **TP-LCK-02: Enhanced Lockfile for Multi-Machine**
**Objective**: Ensure deterministic recreation of complex multi-machine environments.

**Unit Tests:**
- ✅ **Multi-Machine Locking**: Capture all machine configs and images
- ✅ **Dependency Resolution**: Lock machine startup order and networking
- ✅ **Version Pinning**: Pin OS images and package versions
- ✅ **Integrity Verification**: Detect tampering with multi-machine configs

**Integration Tests:**
- ✅ **Deterministic Recreation**: Identical environments from lockfile
- ✅ **Offline Mode**: Work without internet if all cached
- ✅ **Upgrade Path**: Safe migration between lockfile versions

---

## 🏗 Infrastructure Requirements (for handover)

### **CI Integration**
- **Performance Testing**: Benchmark multi-machine startup and coordination
- **Resource Management**: CI runners with sufficient CPU/memory for QEMU testing
- **Image Caching**: Pre-cache common OS images in CI

### **Handover Guidelines**
- **QEMU Expertise**: Team members should understand QEMU networking and image management
- **Multi-Machine Testing**: Use integration tests for cross-machine scenarios
- **Resource Planning**: Multi-machine environments require significant resources (4-8 CPUs, 8-16GB RAM per session)
- **Security**: Implement proper isolation between machines and host system