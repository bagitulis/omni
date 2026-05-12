"""
Frontend Builder module for Omni Build System.

SRP: This module ONLY handles frontend build operations (npm install, build).
No Docker operations, no health checks.

Cross-platform: Works on Windows (cmd/powershell) and Linux/Mac (bash/zsh).
Auto-detects npm/node location even when not in PATH.
"""
import hashlib
import os
import shutil
import subprocess
import time
from pathlib import Path

from omni_build.config import Config
from omni_build.logger import log_error, log_info, log_success, log_warning


def _find_npm() -> str:
    """
    Find npm executable path. Cross-platform.
    Returns the npm command string (may be full path if not in PATH).
    """
    # Check PATH first
    npm_name = "npm.cmd" if os.name == "nt" else "npm"
    found = shutil.which(npm_name)
    if found:
        return found

    # Also try plain "npm" on Windows (git bash, WSL)
    if os.name == "nt":
        found = shutil.which("npm")
        if found:
            return found

    # Search common locations
    home = Path.home()
    candidates = []

    if os.name == "nt":
        appdata = os.environ.get("APPDATA", "")
        programfiles = os.environ.get("ProgramFiles", r"C:\Program Files")
        candidates = [
            Path(appdata) / "npm" / "npm.cmd",
            Path(programfiles) / "nodejs" / "npm.cmd",
        ]
        # nvm-windows
        nvm_root = Path(appdata) / "nvm"
        if nvm_root.exists():
            versions = sorted(nvm_root.glob("v*"), reverse=True)
            for v in versions:
                candidates.append(v / "npm.cmd")
    else:
        candidates = [
            home / ".local" / "bin" / "npm",
            Path("/usr/local/bin/npm"),
            Path("/usr/bin/npm"),
        ]
        # nvm
        nvm_dir = home / ".nvm" / "versions" / "node"
        if nvm_dir.exists():
            versions = sorted(nvm_dir.glob("v*"), reverse=True)
            for v in versions:
                candidates.append(v / "bin" / "npm")
        # volta
        volta_bin = home / ".volta" / "bin" / "npm"
        candidates.append(volta_bin)
        # fnm
        fnm_dir = home / ".fnm" / "node-versions"
        if fnm_dir.exists():
            versions = sorted(fnm_dir.glob("v*"), reverse=True)
            for v in versions:
                candidates.append(v / "installation" / "bin" / "npm")

    for candidate in candidates:
        if candidate.exists():
            return str(candidate)

    # Not found anywhere
    return "npm"  # Fall back, will fail with clear error


def _parse_npm_command(npm_path: str, command_str: str) -> list[str]:
    """
    Parse an npm command string into a list suitable for subprocess.
    Replaces the leading 'npm' with the resolved npm_path.

    Args:
        npm_path: Resolved path to npm executable (may contain spaces)
        command_str: Command string like 'npm ci --legacy-peer-deps'

    Returns:
        List of command parts, e.g. ['C:\\Program Files\\nodejs\\npm.cmd', 'ci', '--legacy-peer-deps']
    """
    import shlex
    parts = command_str.split()
    if parts and parts[0] in ('npm', 'npm.cmd'):
        parts[0] = npm_path
    return parts

class FrontendBuilder:
    """
    Manages frontend build operations with caching and retry logic.
    
    Responsibilities:
    - Check if dependencies need reinstallation (hash-based)
    - Install npm dependencies
    - Build frontend (npm run build)
    - Retry on failure
    - Auto-detect npm/node (cross-platform)
    """
    
    def __init__(self, config: Config) -> None:
        """
        Initialize frontend builder.
        
        Args:
            config: Configuration instance
        """
        self.config = config
        self.frontend_dir = config.frontend_dir
        self.package_json = self.frontend_dir / "package.json"
        self.package_lock = self.frontend_dir / "package-lock.json"
        self.node_modules = self.frontend_dir / "node_modules"
        self.hash_file = self.frontend_dir / ".build_cache" / "package-lock.hash"
        self._npm_path = _find_npm()
        self._env = self._build_env()
    
    def _build_env(self) -> dict:
        """
        Build environment dict for subprocess calls.
        Ensures npm/node parent directory is in PATH.
        """
        env = os.environ.copy()
        npm_dir = str(Path(self._npm_path).parent)
        path_sep = ";" if os.name == "nt" else ":"
        current_path = env.get("PATH", "")
        if npm_dir not in current_path:
            env["PATH"] = npm_dir + path_sep + current_path
        return env
    
    def _get_file_hash(self, path: Path) -> str:
        """
        Get SHA256 hash of a file.
        
        Returns:
            SHA256 hash string, or empty string if file doesn't exist
        """
        if not path.exists():
            return ""
        with open(path, "rb") as f:
            return hashlib.sha256(f.read()).hexdigest()
    
    def _get_combined_hash(self) -> str:
        """
        Get combined hash of package.json + package-lock.json.

        Hashing both files catches the case where a developer adds a package
        to package.json but forgets to run npm install (so package-lock.json
        is still unchanged). Either file changing triggers a reinstall.

        Returns:
            SHA256 of the concatenated hashes
        """
        combined = self._get_file_hash(self.package_json) + self._get_file_hash(self.package_lock)
        return hashlib.sha256(combined.encode()).hexdigest()
    
    def _get_cached_hash(self) -> str:
        """
        Get previously cached hash.
        
        Returns:
            Cached hash or empty string if not found
        """
        if not self.hash_file.exists():
            return ""
        
        try:
            return self.hash_file.read_text().strip()
        except Exception:
            return ""
    
    def _save_hash(self, hash_value: str) -> None:
        """
        Save hash to cache file.
        
        Args:
            hash_value: Hash to save
        """
        self.hash_file.parent.mkdir(parents=True, exist_ok=True)
        self.hash_file.write_text(hash_value)
    
    def check_needs_rebuild(self) -> bool:
        """
        Check if dependencies need reinstallation.

        Compares combined hash of package.json + package-lock.json against
        the last saved hash. This detects both:
        - package-lock.json changes (normal npm install flow)
        - package.json-only changes (dev forgot to run npm install)
        
        Returns:
            True if reinstall needed, False if cache is valid
        """
        if not self.config.frontend_config.cache_enabled:
            log_info("Dependency caching disabled - will reinstall")
            return True
        
        if not self.node_modules.exists():
            log_info("node_modules not found - install required")
            return True
        
        current_hash = self._get_combined_hash()
        cached_hash = self._get_cached_hash()
        
        if current_hash != cached_hash:
            log_info("package.json / package-lock.json changed - reinstall required")
            return True
        
        log_success("Dependencies unchanged - using cache")
        return False
    
    def install_dependencies(self) -> bool:
        """
        Install npm dependencies with retry logic.
        
        Returns:
            True if install succeeded, False otherwise
        """
        max_retries = self.config.frontend_config.max_retries
        install_cmd = self.config.frontend_config.install_command
        
        # Build command as list (safe for paths with spaces)
        cmd_parts = _parse_npm_command(self._npm_path, install_cmd)
        
        log_info(f"Installing dependencies: {install_cmd}")
        log_info(f"Using npm: {self._npm_path}")
        
        for attempt in range(1, max_retries + 1):
            log_info(f"Install attempt {attempt}/{max_retries}...")
            
            try:
                result = subprocess.run(
                    cmd_parts,
                    shell=False,
                    cwd=str(self.frontend_dir),
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=self.config.npm_install_timeout,
                    env=self._env,
                )
                
                if result.returncode == 0:
                    log_success(f"Dependencies installed (attempt {attempt})")
                    # Save hash for caching
                    self._save_hash(self._get_combined_hash())
                    return True
                
                # Install failed
                log_error(f"Install failed (attempt {attempt}): {result.returncode}")
                
                # Check for integrity errors
                if "EINTEGRITY" in result.stderr or "integrity" in result.stderr.lower():
                    log_warning("NPM integrity error detected")
                    if attempt < max_retries:
                        log_info("Cleaning cache and retrying...")
                        self._clean_npm_cache()
                        time.sleep(5)  # Wait after cache clean
                        continue
                
                # Check for network errors
                if "ENOTFOUND" in result.stderr or "fetch failed" in result.stderr.lower():
                    log_warning("Network error detected")
                    if attempt < max_retries:
                        log_info("Retrying in 10 seconds...")
                        time.sleep(10)  # Longer wait for network issues
                        continue
                
                # Generic error
                if attempt >= max_retries:
                    log_error("Max retries reached - install failed")
                    log_error(result.stderr[:1000])
                    return False
                
            except subprocess.TimeoutExpired:
                log_error(f"Install timed out after {self.config.npm_install_timeout}s")
                if attempt >= max_retries:
                    return False
                log_info("Retrying with fresh cache...")
                self._clean_npm_cache()
                time.sleep(5)
            
            except Exception as e:
                log_error(f"Install error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def build(self, force_install: bool = False) -> bool:
        """
        Build frontend with dependency check and retry logic.
        
        Args:
            force_install: Force dependency reinstallation
            
        Returns:
            True if build succeeded, False otherwise
        """
        log_info("Starting frontend build...")
        
        # Check if dependencies need reinstallation
        if force_install or self.check_needs_rebuild():
            if not self.install_dependencies():
                log_error("Failed to install dependencies")
                return False
        
        # Build frontend
        max_retries = self.config.frontend_config.max_retries
        build_cmd = self.config.frontend_config.build_command
        
        # Build command as list (safe for paths with spaces)
        build_cmd_parts = _parse_npm_command(self._npm_path, build_cmd)
        
        log_info(f"Building frontend: {build_cmd}")
        
        for attempt in range(1, max_retries + 1):
            log_info(f"Build attempt {attempt}/{max_retries}...")
            
            try:
                result = subprocess.run(
                    build_cmd_parts,
                    shell=False,
                    cwd=str(self.frontend_dir),
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=self.config.npm_build_timeout,
                    env=self._env,
                )
                
                if result.returncode == 0:
                    log_success(f"Frontend build completed (attempt {attempt})")
                    
                    # Verify output directory exists
                    output_dir = self.frontend_dir / self.config.frontend_config.output_directory
                    if output_dir.exists():
                        log_success(f"Build output verified: {output_dir}")
                        return True
                    else:
                        log_error(f"Build succeeded but output not found: {output_dir}")
                        return False  # No output = build failed
                
                # Build failed - DETAILED ERROR LOGGING
                log_error(f"Build failed (attempt {attempt}): {result.returncode}")
                
                # Combine all output for analysis
                error_output = result.stderr + "\n" + result.stdout
                
                # Show last 30 lines of output for debugging
                output_lines = error_output.strip().split('\n')
                last_lines = output_lines[-30:] if len(output_lines) > 30 else output_lines
                
                log_error("=== BUILD OUTPUT (last 30 lines) ===")
                for line in last_lines:
                    if line.strip():
                        print(f"  {line}")
                log_error("=== END BUILD OUTPUT ===")
                
                # Check for TypeScript errors (cannot auto-fix)
                if "TS" in error_output and "error" in error_output:
                    log_error("TypeScript compilation errors detected - requires manual fix")
                    return False
                
                # Check for syntax errors
                if "SyntaxError" in error_output:
                    log_error("JavaScript/TypeScript syntax errors - requires manual fix")
                    return False
                
                # Check for memory issues
                if "heap out of memory" in error_output.lower() or "fatal error" in error_output.lower():
                    log_error("Node.js memory error - try increasing heap size")
                    log_info("Set NODE_OPTIONS=--max-old-space-size=4096 in .env")
                    return False
                
                # Check for missing modules
                if "Cannot find module" in error_output or "Module not found" in error_output:
                    log_warning("Missing module detected - forcing dependency reinstall")
                    if attempt < max_retries:
                        log_info("Reinstalling dependencies...")
                        self._clean_npm_cache()
                        if not self.install_dependencies():
                            return False
                        continue
                
                # Generic retry
                if attempt >= max_retries:
                    log_error("Max retries reached - build failed")
                    return False
                
                # Wait before retry
                time.sleep(3)
                
            except subprocess.TimeoutExpired:
                log_error(f"Build timed out after {self.config.npm_build_timeout}s")
                if attempt >= max_retries:
                    return False
                time.sleep(5)
            
            except Exception as e:
                log_error(f"Build error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def _clean_npm_cache(self) -> None:
        """Clean npm cache."""
        log_info("Cleaning npm cache...")
        try:
            subprocess.run(
                [self._npm_path, "cache", "clean", "--force"],
                shell=False,
                cwd=str(self.frontend_dir),
                capture_output=True,
                encoding='utf-8',
                errors='replace',
                timeout=60,
                env=self._env,
            )
            log_success("NPM cache cleaned")
        except Exception as e:
            log_warning(f"Could not clean npm cache: {e}")
