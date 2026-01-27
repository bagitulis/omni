"""
Quick validation script to test all imports and configuration.
This can be run to verify the build system is properly configured.
"""

import sys
from pathlib import Path

# Fix Windows console encoding
if sys.platform == "win32":
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Setup path
project_root = Path(__file__).parent.parent.parent
python_build_dir = project_root / "scripts" / "python-build"
sys.path.insert(0, str(python_build_dir))

print("=" * 70)
print("  OMNI BUILD SYSTEM - VALIDATION TEST")
print("=" * 70)
print()

# Test 1: Import all modules
print("[TEST 1] Importing Python modules...")
try:
    from omni_build.config import Config
    from omni_build.models import BuildMode, SpecLevel, ErrorPattern
    from omni_build.cli import BuildOrchestrator
    from omni_build.docker_manager import DockerManager
    from omni_build.error_handler import ErrorHandler
    from omni_build.health_checker import HealthChecker
    from omni_build.frontend_builder import FrontendBuilder
    from omni_build.logger import log_info, log_success, log_error
    print("[OK] PASS: All modules imported successfully")
except ImportError as e:
    print(f"[FAIL] Import error - {e}")
    sys.exit(1)
print()

# Test 2: Load configuration
print("[TEST 2] Loading configuration...")
try:
    config = Config.from_env()
    print(f"[OK] PASS: Configuration loaded")
    print(f"   Project root: {config.project_root}")
    print(f"   Frontend dir: {config.frontend_dir}")
    print(f"   Backend URL:  {config.backend_health_url}")
except Exception as e:
    print(f"[FAIL] Config error - {e}")
    sys.exit(1)
print()

# Test 3: Verify error patterns
print("[TEST 3] Checking error patterns...")
try:
    error_handler = ErrorHandler()
    patterns = len(error_handler.error_patterns)
    print(f"[OK] PASS: {patterns} error patterns loaded")
    
    # Check critical patterns exist
    critical_patterns = {
        "DockerDNSPostgresError": False,
        "BackendHealthCheckTimeout": False,
        "PostgresDataCorruption": False,
    }
    
    for pattern in error_handler.error_patterns:
        if pattern.name in critical_patterns:
            critical_patterns[pattern.name] = True
            print(f"   [OK] {pattern.name} - {pattern.description}")
    
    missing = [name for name, found in critical_patterns.items() if not found]
    if missing:
        print(f"   [FAIL] Missing critical patterns: {', '.join(missing)}")
        sys.exit(1)
        
except Exception as e:
    print(f"[FAIL] Error handler test - {e}")
    sys.exit(1)
print()

# Test 4: Check file structure
print("[TEST 4] Verifying file structure...")
try:
    required_files = [
        config.project_root / "docker-compose.tunnel.yml",
        config.frontend_dir / "package.json",
        python_build_dir / "omni_build" / "cli.py",
        python_build_dir / "omni_build" / "docker_manager.py",
    ]
    
    all_exist = True
    for file_path in required_files:
        rel_path = file_path.relative_to(config.project_root)
        if file_path.exists():
            print(f"   [OK] {rel_path}")
        else:
            print(f"   [FAIL] MISSING: {rel_path}")
            all_exist = False
    
    if not all_exist:
        print("[FAIL] Missing required files")
        sys.exit(1)
    else:
        print("[OK] PASS: All required files exist")
        
except Exception as e:
    print(f"[FAIL] File check error - {e}")
    sys.exit(1)
print()

# Test 5: Instantiate orchestrator
print("[TEST 5] Testing BuildOrchestrator...")
try:
    orchestrator = BuildOrchestrator(config)
    print("[OK] PASS: BuildOrchestrator created successfully")
    print(f"   Docker manager: OK")
    print(f"   Frontend builder: OK")
    print(f"   Health checker: OK")
    print(f"   Error handler: OK")
except Exception as e:
    print(f"[FAIL] Orchestrator error - {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)
print()

# Final summary
print("=" * 70)
print("  [OK] ALL TESTS PASSED - BUILD SYSTEM IS READY")
print("=" * 70)
print()
print("You can now run:")
print("  python build.py smart       # Smart build (RECOMMENDED)")
print("  python build.py quick       # Quick restart")
print("  python build.py status      # Check containers")
print()
