#!/usr/bin/env python3
"""
Omni Build System - Standalone Entry Point

Pintasan:
- Jalankan tanpa argumen untuk membuka menu interaktif.
- Masih bisa jalan seperti biasa dengan subcommand (smart, quick, full, validate, clean, status).

Contoh langsung:
    python build.py smart              # Smart build (RECOMMENDED)
    python build.py quick              # Quick restart
    python build.py full               # Full rebuild
    python build.py validate           # Validate environment only
    python build.py clean              # Cleanup Docker
    python build.py status             # Show container status
    python build.py smart --spec=lowspec
    python build.py smart --skip-frontend
    python build.py smart --dry-run    # Validation mode (no actual build)
"""
import sys
import os
from pathlib import Path

# Add python-build to path
project_root = Path(__file__).parent
python_build_dir = project_root / "scripts" / "python-build"
sys.path.insert(0, str(python_build_dir))

# ----- Interactive menu helper ------------------------------------------------
def interactive_menu():
    """Prompt user to choose a build action; returns list of CLI args."""
    options = [
        ("Smart build (recommended)", ["smart"]),
        ("Quick restart (no rebuild)", ["quick"]),
        ("Full rebuild (no cache)", ["full"]),
        ("Validate only (no build)", ["validate"]),
        ("Clean Docker resources", ["clean"]),
        ("Show container status", ["status"]),
        ("Smart build - DRY RUN", ["smart", "--dry-run"]),
        ("Keluar / Batal", None),
    ]

    # Clear screen
    os.system('cls' if os.name == 'nt' else 'clear')
    
    print("=" * 60)
    print(" Omni Build Menu")
    print("=" * 60)
    for idx, (label, _) in enumerate(options, start=1):
        print(f"  {idx}. {label}")
    print()

    while True:
        try:
            choice = input("Pilih menu [1-8]: ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nBatal.")
            sys.exit(0)

        if not choice.isdigit():
            print("Input tidak valid, masukkan angka 1-8.")
            continue

        idx = int(choice)
        if 1 <= idx <= len(options):
            args = options[idx - 1][1]
            if args is None:
                print("Batal.")
                sys.exit(0)
            return args

        print("Input di luar range, coba lagi.")


def run_build():
    """Execute a single build command. Returns exit code."""
    # Check if running in dry-run mode
    DRY_RUN = "--dry-run" in sys.argv or "--validate-only" in sys.argv
    
    if DRY_RUN:
        print("=" * 60)
        print("  DRY-RUN MODE - Validation Only")
        print("=" * 60)
        print()
        print("This mode will:")
        print("  - Verify all Python modules can be imported")
        print("  - Check configuration is valid")
        print("  - Validate file structure")
        print("  - Simulate build flow WITHOUT actual execution")
        print()
        print("=" * 60)
        print()

    # Import and run CLI
    try:
        # Test imports first
        from omni_build.config import Config
        from omni_build.models import BuildMode, SpecLevel
        from omni_build.cli import BuildOrchestrator
        from omni_build.logger import log_info, log_success, log_error
        
        print("[OK] All Python modules imported successfully")
        print()
        
        if DRY_RUN:
            # Dry-run validation mode
            print("[DRY-RUN] Loading configuration...")
            try:
                config = Config.from_env()
                print(f"[OK] Configuration loaded successfully")
                print(f"   - Project root: {config.project_root}")
                print(f"   - Frontend dir: {config.frontend_dir}")
                print(f"   - Docker ready check: configured")
                print()
            except Exception as e:
                print(f"[FAIL] Configuration error: {e}")
                return 1
            
            print("[DRY-RUN] Validating file structure...")
            required_files = [
                config.project_root / "docker-compose.tunnel.yml",
                config.frontend_dir / "package.json",
                python_build_dir / "omni_build" / "cli.py",
                python_build_dir / "omni_build" / "docker_manager.py",
                python_build_dir / "omni_build" / "error_handler.py",
                python_build_dir / "omni_build" / "health_checker.py",
            ]
            
            all_ok = True
            for file in required_files:
                if file.exists():
                    print(f"   [OK] {file.relative_to(project_root)}")
                else:
                    print(f"   [FAIL] MISSING: {file.relative_to(project_root)}")
                    all_ok = False
            print()
            
            if not all_ok:
                print("[FAIL] Some required files are missing!")
                return 1
            
            print("[DRY-RUN] Testing BuildOrchestrator instantiation...")
            try:
                orchestrator = BuildOrchestrator(config)
                print("[OK] BuildOrchestrator created successfully")
                print()
            except Exception as e:
                print(f"[FAIL] BuildOrchestrator error: {e}")
                return 1
            
            print("[DRY-RUN] Testing error patterns...")
            try:
                from omni_build.error_handler import ErrorHandler
                error_handler = ErrorHandler()
                pattern_count = len(error_handler.error_patterns)
                print(f"[OK] Error handler loaded: {pattern_count} patterns configured")
                
                # Check for critical patterns
                critical_patterns = [
                    "DockerDNSPostgresError",
                    "BackendHealthCheckTimeout",
                    "PostgresDataCorruption",
                ]
                
                pattern_names = [p.name for p in error_handler.error_patterns]
                for pattern in critical_patterns:
                    if pattern in pattern_names:
                        print(f"   [OK] {pattern}")
                    else:
                        print(f"   [FAIL] MISSING: {pattern}")
                        all_ok = False
                print()
            except Exception as e:
                print(f"[FAIL] Error handler test failed: {e}")
                return 1
            
            if all_ok:
                print("=" * 60)
                print("  [OK] DRY-RUN VALIDATION PASSED")
                print("=" * 60)
                print()
                print("All checks passed! Build system is ready.")
                print()
                print("To run actual build:")
                print("  python build.py smart")
                print("  python build.py quick")
                print()
                return 0
            else:
                print("=" * 60)
                print("  [FAIL] DRY-RUN VALIDATION FAILED")
                print("=" * 60)
                print()
                print("Some checks failed. Please fix the issues above.")
                print()
                return 1
        
        else:
            # Normal build mode - import and run CLI
            from omni_build.cli import cli
            
            # Remove script name and dry-run flags from argv
            filtered_argv = [arg for arg in sys.argv[1:] if not arg.startswith('--dry-run') and not arg.startswith('--validate-only')]
            sys.argv = ['build.py'] + filtered_argv
            
            # Run CLI
            cli()
            return 0

    except ImportError as e:
        print()
        print("=" * 60)
        print("  [FAIL] IMPORT ERROR")
        print("=" * 60)
        print()
        print(f"Failed to import required module: {e}")
        print()
        print("Please ensure:")
        print("  1. Virtual environment is set up:")
        print("     python scripts/python-build/setup.py")
        print()
        print("  2. You're in the virtual environment:")
        print("     .venv\\Scripts\\activate  (Windows)")
        print("     source .venv/bin/activate  (Linux/Mac)")
        print()
        print("  3. Dependencies are installed:")
        print("     pip install -r scripts/python-build/requirements.txt")
        print()
        return 1

    except Exception as e:
        print()
        print("=" * 60)
        print("  [FAIL] UNEXPECTED ERROR")
        print("=" * 60)
        print()
        print(f"Error: {e}")
        print()
        import traceback
        traceback.print_exc()
        print()
        return 1


def run_with_menu_loop():
    """Run build in a loop with menu."""
    while True:
        # Show menu and get user choice
        selected_args = interactive_menu()
        
        # Set up argv for the build command
        sys.argv = ['build.py'] + selected_args
        
        # Execute the build
        exit_code = run_build()
        
        # After build completes, ask to return to menu
        print()
        print("=" * 60)
        try:
            input("Press Enter to return to menu...")
        except (EOFError, KeyboardInterrupt):
            print("\nGoodbye!")
            sys.exit(0)


# Main execution
if __name__ == "__main__":
    # Check if we should run in interactive menu mode
    # Always show menu when no args are provided, even if stdin isn't a TTY
    # (e.g., when launched from VS Code task/debugger).
    INTERACTIVE_MODE = len(sys.argv) == 1

    if INTERACTIVE_MODE:
        # Run with menu loop
        run_with_menu_loop()
    else:
        # Direct command mode - run once
        exit_code = run_build()
        sys.exit(exit_code)
