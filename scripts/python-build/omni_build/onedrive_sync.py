#!/usr/bin/env python3
"""
OneDrive Sync for OMNI Database Backups

Mekanisme sync yang menggabungkan kebaikan Extensions (git-tracked manifest)
dan kepraktisan cloud storage (OneDrive auto-sync untuk binary data).

Workflow:
  1. Backup → backups/smart/ (lokal)
  2. Sync to OneDrive → OneDrive/Data/omni-backup/
  3. Git push → manifest.json + schema (kecil, diffable)
  4. Di PC lain: Git pull → OneDrive sync → Restore

Keuntungan:
  - Git repo tetap kecil (~50MB setelah cleanup)
  - Binary data auto-sync via OneDrive (zero effort)
  - Multi-PC seamless (semua PC dengan OneDrive login)
  - Versioning built-in (OneDrive file history)
"""
import json
import shutil
import subprocess
import sys
from datetime import datetime
from pathlib import Path
from typing import List, Tuple


class OneDriveSync:
    """Handle sync between local backups/smart/ and OneDrive."""

    def __init__(self, project_root: Path = None):
        self.project_root = project_root or Path.cwd()
        self.local_backup_dir = self.project_root / "backups" / "smart"
        self.onedrive_dir = self._find_onedrive() / "Data" / "omni-backup"

    def _find_onedrive(self) -> Path:
        """Find OneDrive directory."""
        # Common OneDrive paths
        candidates = [
            Path.home() / "OneDrive",
            Path.home() / "OneDrive - Personal",
            Path("C:") / "Users" / Path.home().name / "OneDrive",
        ]
        for candidate in candidates:
            if candidate.exists():
                return candidate
        raise RuntimeError("OneDrive not found. Please ensure OneDrive is installed and synced.")

    def sync_to_cloud(self) -> Tuple[bool, str]:
        """Sync local backup to OneDrive."""
        if not self.local_backup_dir.exists():
            return False, f"Local backup not found: {self.local_backup_dir}"

        print(f"[SYNC] Local: {self.local_backup_dir}")
        print(f"[SYNC] OneDrive: {self.onedrive_dir}")

        # Ensure OneDrive directory exists
        self.onedrive_dir.mkdir(parents=True, exist_ok=True)

        # Create sync manifest
        sync_manifest = {
            "synced_at": datetime.now().isoformat(),
            "machine": subprocess.run(["hostname"], capture_output=True, text=True).stdout.strip(),
            "source": str(self.local_backup_dir),
            "destination": str(self.onedrive_dir),
        }

        # Write sync manifest
        manifest_path = self.onedrive_dir / "_sync_manifest.json"
        with open(manifest_path, 'w') as f:
            json.dump(sync_manifest, f, indent=2)

        # Sync using robocopy (Windows) or rsync (Linux/Mac)
        if sys.platform == "win32":
            # Robocopy: mirror, exclude manifest.json (tracked by git)
            cmd = [
                "robocopy",
                str(self.local_backup_dir),
                str(self.onedrive_dir),
                "/MIR",  # Mirror
                "/R:3",  # 3 retries
                "/W:5",  # 5 sec wait between retries
                "/MT:8", # Multi-threaded
                "/XD", "", # No exclude dirs
                "/XF", "", # No exclude files (copy everything)
                "/NDL",  # No directory list
                "/NFL",  # No file list
            ]
        else:
            # rsync for Linux/Mac
            cmd = [
                "rsync",
                "-avz",      # Archive, verbose, compress
                "--delete",   # Delete extraneous files
                f"{self.local_backup_dir}/",
                f"{self.onedrive_dir}/",
            ]

        print(f"[SYNC] Running: {' '.join(cmd[:5])}...")
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode in [0, 1, 2, 3]:  # Robocopy success codes
            # Count files synced
            files_copied = len(list(self.onedrive_dir.rglob("*")))
            return True, f"Synced {files_copied} files to OneDrive"
        else:
            return False, f"Sync failed: {result.stderr[:200]}"

    def sync_from_cloud(self) -> Tuple[bool, str]:
        """Sync from OneDrive to local."""
        if not self.onedrive_dir.exists():
            return False, f"OneDrive backup not found: {self.onedrive_dir}"

        print(f"[SYNC] OneDrive: {self.onedrive_dir}")
        print(f"[SYNC] Local: {self.local_backup_dir}")

        # Ensure local directory exists
        self.local_backup_dir.mkdir(parents=True, exist_ok=True)

        # Check sync manifest
        manifest_path = self.onedrive_dir / "_sync_manifest.json"
        if manifest_path.exists():
            with open(manifest_path) as f:
                sync_info = json.load(f)
            print(f"[SYNC] Last synced: {sync_info.get('synced_at', 'unknown')}")
            print(f"[SYNC] From machine: {sync_info.get('machine', 'unknown')}")

        # Sync from OneDrive to local
        if sys.platform == "win32":
            cmd = [
                "robocopy",
                str(self.onedrive_dir),
                str(self.local_backup_dir),
                "/MIR",
                "/R:3",
                "/W:5",
                "/MT:8",
                "/NDL",
                "/NFL",
            ]
        else:
            cmd = [
                "rsync",
                "-avz",
                "--delete",
                f"{self.onedrive_dir}/",
                f"{self.local_backup_dir}/",
            ]

        print(f"[SYNC] Running: {' '.join(cmd[:5])}...")
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode in [0, 1, 2, 3]:
            files_copied = len(list(self.local_backup_dir.rglob("*")))
            return True, f"Synced {files_copied} files from OneDrive"
        else:
            return False, f"Sync failed: {result.stderr[:200]}"

    def status(self) -> dict:
        """Show sync status."""
        status = {
            "local_backup_exists": self.local_backup_dir.exists(),
            "onedrive_backup_exists": self.onedrive_dir.exists(),
            "local_size_mb": 0,
            "onedrive_size_mb": 0,
            "local_files": 0,
            "onedrive_files": 0,
        }

        if self.local_backup_dir.exists():
            files = list(self.local_backup_dir.rglob("*"))
            status["local_files"] = len(files)
            status["local_size_mb"] = sum(f.stat().st_size for f in files if f.is_file()) / (1024*1024)

        if self.onedrive_dir.exists():
            files = list(self.onedrive_dir.rglob("*"))
            status["onedrive_files"] = len(files)
            status["onedrive_size_mb"] = sum(f.stat().st_size for f in files if f.is_file()) / (1024*1024)

        return status


def main():
    """CLI entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="OMNI OneDrive Sync")
    parser.add_argument("action", choices=["to-cloud", "from-cloud", "status"],
                       help="Sync direction or status check")
    parser.add_argument("--project-root", type=Path, default=Path.cwd(),
                       help="Project root directory")

    args = parser.parse_args()

    sync = OneDriveSync(args.project_root)

    if args.action == "to-cloud":
        success, message = sync.sync_to_cloud()
        print(f"[{'OK' if success else 'FAIL'}] {message}")
        sys.exit(0 if success else 1)

    elif args.action == "from-cloud":
        success, message = sync.sync_from_cloud()
        print(f"[{'OK' if success else 'FAIL'}] {message}")
        sys.exit(0 if success else 1)

    elif args.action == "status":
        status = sync.status()
        print("=== SYNC STATUS ===")
        print(f"Local backup: {'EXISTS' if status['local_backup_exists'] else 'NOT FOUND'}")
        if status['local_backup_exists']:
            print(f"  Files: {status['local_files']}")
            print(f"  Size: {status['local_size_mb']:.1f} MB")
        print()
        print(f"OneDrive backup: {'EXISTS' if status['onedrive_backup_exists'] else 'NOT FOUND'}")
        if status['onedrive_backup_exists']:
            print(f"  Files: {status['onedrive_files']}")
            print(f"  Size: {status['onedrive_size_mb']:.1f} MB")


if __name__ == "__main__":
    main()
