#!/usr/bin/env python3
"""
Omni Build System - Standalone Entry Point

Pintasan:
- Jalankan tanpa argumen untuk membuka menu interaktif.
- Masih bisa jalan seperti biasa dengan subcommand.

Contoh langsung:
    python build.py quickfix           # Quick Fix (fix service issues, no rebuild) - FIRST OPTION
    python build.py smart              # Smart build (RECOMMENDED for code changes)
    python build.py full               # Full rebuild
    python build.py validate           # Validate environment only
    python build.py clean              # Cleanup Docker
    python build.py status             # Show container status
    python build.py smart --spec=lowspec
    python build.py smart --skip-frontend
    python build.py smart --dry-run    # Validation mode (no actual build)

Quick Fix vs Smart Build:
- Quick Fix: Untuk memperbaiki service yang bermasalah TANPA rebuild.
             Jika gagal, otomatis fallback ke Smart Build.
- Smart Build: Untuk perubahan kode yang butuh rebuild image.
"""
import sys
import os
import subprocess
from pathlib import Path

# Add python-build to path
project_root = Path(__file__).parent
python_build_dir = project_root / "scripts" / "python-build"
sys.path.insert(0, str(python_build_dir))

# ----- Auto-bootstrap venv ----------------------------------------------------
def _ensure_venv():
    """Auto-create venv and install deps if not already running inside one."""
    venv_dir = project_root / ".venv"
    if os.name == "nt":
        venv_python = venv_dir / "Scripts" / "python.exe"
    else:
        venv_python = venv_dir / "bin" / "python"

    # Already running inside the venv? Nothing to do.
    if sys.prefix != sys.base_prefix:
        return

    requirements = python_build_dir / "requirements.txt"

    # Create venv if missing
    if not venv_python.exists():
        print("[AUTO] Creating virtual environment (.venv)...")
        subprocess.check_call([sys.executable, "-m", "venv", str(venv_dir)])
        print("[OK] Virtual environment created.")

    # Check if deps already installed (marker file to skip slow pip check)
    marker = venv_dir / ".deps_installed"
    req_hash = ""
    if requirements.exists():
        import hashlib
        req_hash = hashlib.md5(requirements.read_bytes()).hexdigest()

    needs_install = True
    if marker.exists():
        try:
            if marker.read_text().strip() == req_hash:
                needs_install = False
        except Exception:
            pass

    if needs_install:
        print("[AUTO] Installing dependencies...")
        pip_cmd = [str(venv_python), "-m", "pip", "install", "-q", "--disable-pip-version-check", "-r", str(requirements)]
        subprocess.check_call(pip_cmd)
        marker.write_text(req_hash)
        print("[OK] Dependencies installed.")

    # Re-exec this script with the venv python, preserving PATH
    print("[AUTO] Re-launching with venv python...")
    print()
    relaunch_args = [str(venv_python), __file__] + sys.argv[1:]
    if os.name == "nt":
        # Windows: os.execv() doesn't replace the process, it spawns a child.
        # Use subprocess + sys.exit to avoid double-execution.
        result = subprocess.run(relaunch_args)
        sys.exit(result.returncode)
    else:
        os.execv(str(venv_python), relaunch_args)


def _ensure_path():
    """
    Ensure npm/node AND the container runtime (docker/podman) plus the
    project venv's Scripts folder are discoverable in PATH.

    Cross-platform: handles Windows (nvm, default install, Docker Desktop,
    Podman for Windows) and Linux (~/.local/bin, nvm, fnm, volta).
    """
    import shutil

    # (1) Always put the project venv on PATH so subprocess.run() can find
    #     tools installed inside it (podman-compose, docker-compose, etc.).
    #     Skipping this made `get_compose_command()` invisible to child
    #     processes even after `pip install podman-compose` in the venv.
    venv_scripts = project_root / ".venv" / ("Scripts" if os.name == "nt" else "bin")
    if venv_scripts.exists():
        path_sep = ";" if os.name == "nt" else ":"
        current = os.environ.get("PATH", "")
        if str(venv_scripts) not in current:
            os.environ["PATH"] = f"{venv_scripts}{path_sep}{current}"

    # (2) Container runtime search: docker/podman may live outside the
    #     default PATH the subprocess inherits (Windows Start-Process
    #     strips a lot of the interactive PATH). Look for well-known
    #     install locations and prepend them if the CLI is not visible.
    if os.name == "nt":
        runtime_candidates = [
            Path(r"C:\Program Files\Docker\Docker\resources\bin"),
            Path(r"C:\Program Files\RedHat\Podman"),
            Path(r"C:\Program Files\Podman\bin"),
        ]
        need_docker = shutil.which("docker") is None
        need_podman = shutil.which("podman") is None
        if need_docker or need_podman:
            path_sep = ";"
            current = os.environ.get("PATH", "")
            added_runtime: list[str] = []
            for cand in runtime_candidates:
                if not cand.exists():
                    continue
                cand_str = str(cand)
                if cand_str in current:
                    continue
                has_docker = (cand / "docker.exe").exists()
                has_podman = (cand / "podman.exe").exists()
                if (need_docker and has_docker) or (need_podman and has_podman):
                    added_runtime.append(cand_str)
            if added_runtime:
                os.environ["PATH"] = path_sep.join(added_runtime) + path_sep + os.environ.get("PATH", "")
                print(f"[AUTO] Added container runtime to PATH: {', '.join(added_runtime)}")

    # Quick check: npm already in PATH?
    npm_cmd = "npm.cmd" if os.name == "nt" else "npm"
    if shutil.which(npm_cmd):
        return  # Already good

    # Common locations to search
    try:
        import pwd as _pwd
        home = Path(_pwd.getpwuid(os.getuid()).pw_dir)
    except (ImportError, KeyError):
        home = Path.home()
    candidates = []

    if os.name == "nt":
        # Windows: nvm, default nodejs, AppData/Roaming/npm
        appdata = os.environ.get("APPDATA", "")
        localappdata = os.environ.get("LOCALAPPDATA", "")
        programfiles = os.environ.get("ProgramFiles", r"C:\Program Files")

        candidates = [
            Path(appdata) / "npm",
            Path(appdata) / "nvm" / "current",
            Path(programfiles) / "nodejs",
            Path(localappdata) / "fnm_multishells",
            home / ".volta" / "bin",
        ]
        # nvm-windows: find latest version
        nvm_root = Path(appdata) / "nvm"
        if nvm_root.exists():
            versions = sorted(nvm_root.glob("v*"), reverse=True)
            candidates.extend(versions)
    else:
        # Linux/Mac: ~/.local/bin, nvm, fnm, volta, system
        candidates = [
            home / ".local" / "bin",
            home / ".nvm" / "current" / "bin",
            home / ".volta" / "bin",
            home / ".fnm" / "current" / "bin",
            Path("/usr/local/bin"),
        ]
        # nvm: find default version
        nvm_dir = home / ".nvm" / "versions" / "node"
        if nvm_dir.exists():
            versions = sorted(nvm_dir.glob("v*"), reverse=True)
            for v in versions:
                candidates.append(v / "bin")

    # Add found paths to PATH
    path_sep = ";" if os.name == "nt" else ":"
    current_path = os.environ.get("PATH", "")
    added = []

    for candidate in candidates:
        if not candidate.exists():
            continue
        candidate_str = str(candidate)
        if candidate_str in current_path:
            continue
        # Check if npm exists here
        npm_check = candidate / ("npm.cmd" if os.name == "nt" else "npm")
        node_check = candidate / ("node.exe" if os.name == "nt" else "node")
        if npm_check.exists() or node_check.exists():
            added.append(candidate_str)

    if added:
        os.environ["PATH"] = path_sep.join(added) + path_sep + current_path
        print(f"[AUTO] Added to PATH: {', '.join(added)}")
    else:
        # Last resort: check if node_modules/.bin has npx
        print("[WARN] npm not found in PATH. Frontend build may fail.")
        print("       Install Node.js: https://nodejs.org/ or use nvm/fnm/volta")


_ensure_venv()
_ensure_path()

# ----- Interactive menu helper ------------------------------------------------
def _detect_available_runtimes() -> list[str]:
    """Return list of runtimes that are installed and responsive."""
    available = []
    import shutil as _shutil
    import subprocess as _sp
    for rt in ("docker", "podman"):
        if _shutil.which(rt):
            try:
                result = _sp.run([rt, "info"], capture_output=True, timeout=10)
                if result.returncode == 0:
                    available.append(rt)
            except Exception:
                pass
    return available


def _prompt_runtime_choice(available: list[str]) -> str | None:
    """If both runtimes available, ask user. Returns choice or None (auto)."""
    if len(available) < 2:
        return None

    print("-" * 60)
    print("  Container Runtime:")
    for idx, rt in enumerate(available, start=1):
        tag = "Docker Desktop" if rt == "docker" else "Podman"
        print(f"    {idx}. {tag}")
    print(f"    {len(available) + 1}. Auto-detect (env CONTAINER_RUNTIME)")
    print()

    while True:
        try:
            c = input(f"  Pilih runtime [1-{len(available) + 1}]: ").strip()
        except (EOFError, KeyboardInterrupt):
            return None
        if not c.isdigit():
            continue
        n = int(c)
        if 1 <= n <= len(available):
            return available[n - 1]
        if n == len(available) + 1:
            return None
        print("  Di luar range.")


def interactive_menu():
    """Prompt user to choose a build action; returns list of CLI args."""
    options = [
        ("Quick Fix (fix service issues, no rebuild)", ["quickfix"]),
        ("Smart build (recommended)", ["smart"]),
        ("Full rebuild (no cache)", ["full"]),
        ("Full rebuild + DB restore (for new PC)", ["full", "--restore"]),
        ("Smart build + DB restore", ["smart", "--restore"]),
        ("Backup database (NDJSON)", ["backup"]),
        ("Restore database (from NDJSON)", ["sync-import"]),
        ("Validate only (no build)", ["validate"]),
        ("Clean Docker resources", ["clean"]),
        ("Show container status", ["status"]),
        ("Keluar / Batal", None),
    ]

    available = _detect_available_runtimes()

    # Clear screen
    os.system('cls' if os.name == 'nt' else 'clear')

    print("=" * 60)
    print(" Omni Build Menu")
    print("=" * 60)
    rt_display = " + ".join(r.upper() for r in available) if available else "NONE"
    print(f"  Runtime terdeteksi: {rt_display}")
    print("-" * 60)

    # Runtime selection (only if both available)
    chosen = _prompt_runtime_choice(available)
    if chosen:
        os.environ["CONTAINER_RUNTIME"] = chosen
        print(f"  >> Runtime: {chosen.upper()}\n")
    elif available:
        # Show which will be used (env or auto-detect priority)
        env_rt = os.environ.get("CONTAINER_RUNTIME", "").strip().lower()
        if env_rt in ("docker", "podman"):
            print(f"  >> Runtime: {env_rt.upper()} (dari env)")
        else:
            print(f"  >> Runtime: {available[0].upper()} (auto-detect)")
    print()

    for idx, (label, _) in enumerate(options, start=1):
        print(f"  {idx}. {label}")
    print()

    while True:
        try:
            choice = input(f"Pilih menu [1-{len(options)}]: ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nBatal.")
            sys.exit(0)

        if not choice.isdigit():
            print(f"Input tidak valid, masukkan angka 1-{len(options)}.")
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
            
            # Remove script name and build.py-only flags from argv, so click
            # does not reject them as unknown options.
            filtered_argv = [
                arg for arg in sys.argv[1:]
                if not arg.startswith('--dry-run')
                and not arg.startswith('--validate-only')
                and not arg.startswith('--compact-vhdx')
            ]
            sys.argv = ['build.py'] + filtered_argv
            
            # Detect whether we're about to run a build that produces new
            # image layers so we can prune dangling images afterwards.
            _prune_after = filtered_argv and filtered_argv[0] in ("smart", "full", "quickfix")

            # Run CLI - catch SystemExit to allow returning to menu
            try:
                cli()
                exit_code = 0
            except SystemExit as e:
                # CLI called sys.exit(), capture the exit code
                exit_code = e.code if e.code is not None else 0

            # Post-build podman retention: prevent the 20-30 GB dangling
            # rubbish pile-up observed 2026-09-15. Keeps :latest of omni
            # repos + up to 3 recent dangling per bucket, and (since the
            # 2026-09-21 audit) also removes leaked testcontainers
            # containers and orphaned anonymous volumes, which previously
            # pinned images as "in use" and defeated this whole step.
            # Best-effort; a prune failure never masks a successful build.
            if _prune_after and exit_code == 0:
                try:
                    retention = project_root / "scripts" / "podman-retention.py"
                    if retention.exists():
                        print()
                        print("[RETENTION] Running podman retention (keep=3)...")
                        subprocess.run(
                            [sys.executable, str(retention), "--keep", "3"],
                            check=False,
                        )
                except Exception as retention_err:
                    print(f"[WARN] Retention step skipped: {retention_err}")

                # Pruning inside the VM cannot shrink ext4.vhdx (it is not
                # sparse), so dead space silently accumulates on C:. Report it
                # and point at the elevated fix. --compact-vhdx opts into
                # running that fix when this shell is already elevated.
                try:
                    scripts_dir = project_root / "scripts"
                    if str(scripts_dir) not in sys.path:
                        sys.path.insert(0, str(scripts_dir))
                    import vhdx_health

                    vhdx_health.report_and_maybe_compact(
                        project_root,
                        compact_requested="--compact-vhdx" in sys.argv,
                    )
                except Exception as vhdx_err:
                    print(f"[WARN] VHDX health check skipped: {vhdx_err}")

            return exit_code

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
        # Direct command mode - run once but still wait before exit
        exit_code = run_build()
        print()
        print("=" * 60)
        try:
            choice = input("Press Enter to exit, or 'm' to go to menu: ").strip().lower()
            if choice == 'm':
                # User wants to go to menu
                run_with_menu_loop()
            else:
                sys.exit(exit_code)
        except (EOFError, KeyboardInterrupt):
            print("\nGoodbye!")
            sys.exit(exit_code)
