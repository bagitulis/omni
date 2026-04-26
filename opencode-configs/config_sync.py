"""
config_sync.py - Sync opencode-configs between project locations

Two sync modes:
  1. Hub sync: D:\\Project\\ai (hub) → auto/extensions/omni (one-way, silent)
  2. Interactive sync: any project → any other project (with diffs + prompts)

Hub sync is called automatically on AI.py startup (unless --no-sync).
Interactive sync is called from AI.py menu [X] or standalone.

Called from AI.py menu [X] or standalone:
  python config_sync.py                    # Interactive
  python config_sync.py extensions->omni   # Direct
  python config_sync.py --hub              # Hub sync to all targets
  python config_sync.py --hub --dry-run    # Preview hub sync
"""

import difflib
import json
import shutil
import sys
from pathlib import Path

# Known project locations
LOCATIONS = {
    "extensions": Path("D:/Project/extensions"),
    "omni": Path("D:/Project/omni"),
    "auto": Path("D:/Project/auto"),
}

# Files to sync (relative to project root) — full file copy
SYNC_FILES = [
    "AI.py",
    "opencode-configs/ai_profiles.py",
    "opencode-configs/ai_sync.py",
    "opencode-configs/config_sync.py",
    "opencode-configs/opencode-enowx.json",
    "opencode-configs/opencode-plugin.json",
]

# Special: profiles JSON syncs only "profiles" section (not "shared" rules)
PROFILES_FILE = "opencode-configs/opencode-profiles.json"

# ANSI colors
_R = "\033[91m"
_G = "\033[92m"
_Y = "\033[93m"
_C = "\033[96m"
_B = "\033[1m"
_D = "\033[2m"
_0 = "\033[0m"


def _read(path: Path) -> str | None:
    if not path.exists():
        return None
    return path.read_text(encoding="utf-8")


def _diff_files(src: Path, dst: Path, label: str) -> dict | None:
    """Compare two files. Returns diff info dict or None if identical."""
    s, d = _read(src), _read(dst)
    if s is None and d is None:
        return None
    if s == d:
        return None

    diff = list(difflib.unified_diff(
        (d or "").splitlines(keepends=True),
        (s or "").splitlines(keepends=True),
        fromfile=f"[target] {label}",
        tofile=f"[source] {label}",
        lineterm="",
    ))
    if not diff:
        return None

    return {
        "label": label, "src": src, "dst": dst, "diff": diff,
        "src_ok": s is not None, "dst_ok": d is not None,
    }


def _diff_profiles(src: Path, dst: Path) -> dict | None:
    """Compare only 'profiles' section (skip 'shared' rules)."""
    s, d = _read(src), _read(dst)
    if s is None and d is None:
        return None

    try:
        sd = json.loads(s) if s else {}
        dd = json.loads(d) if d else {}
    except json.JSONDecodeError as e:
        return {
            "label": PROFILES_FILE, "src": src, "dst": dst,
            "diff": [f"JSON parse error: {e}"],
            "src_ok": s is not None, "dst_ok": d is not None, "error": True,
        }

    sp = sd.get("profiles", {})
    dp = dd.get("profiles", {})
    sm = {k: v for k, v in sd.items() if k not in ("shared", "profiles")}
    dm = {k: v for k, v in dd.items() if k not in ("shared", "profiles")}

    if sp == dp and sm == dm:
        return None

    st = json.dumps({"_meta": sm, "profiles": sp}, indent=2)
    dt = json.dumps({"_meta": dm, "profiles": dp}, indent=2)

    diff = list(difflib.unified_diff(
        dt.splitlines(keepends=True), st.splitlines(keepends=True),
        fromfile=f"[target] {PROFILES_FILE} (profiles only)",
        tofile=f"[source] {PROFILES_FILE} (profiles only)",
        lineterm="",
    ))
    if not diff:
        return None

    return {
        "label": f"{PROFILES_FILE} (profiles section)", "src": src, "dst": dst,
        "diff": diff, "src_ok": True, "dst_ok": True,
        "profiles_only": True, "src_profiles": sp, "src_meta": sm,
    }


def _apply_one(d: dict):
    """Apply a single diff entry."""
    if d.get("profiles_only"):
        dst_data = json.loads(d["dst"].read_text(encoding="utf-8"))
        dst_data["profiles"] = d["src_profiles"]
        dst_data.update(d["src_meta"])
        d["dst"].write_text(
            json.dumps(dst_data, indent=2, ensure_ascii=False) + "\n",
            encoding="utf-8",
        )
    else:
        d["dst"].parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(d["src"], d["dst"])


def _color_diff(lines: list[str]) -> str:
    out = []
    for ln in lines:
        if ln.startswith("+++") or ln.startswith("---"):
            out.append(f"{_B}{ln}{_0}")
        elif ln.startswith("@@"):
            out.append(f"{_C}{ln}{_0}")
        elif ln.startswith("+"):
            out.append(f"{_G}{ln}{_0}")
        elif ln.startswith("-"):
            out.append(f"{_R}{ln}{_0}")
        else:
            out.append(ln)
    return "\n".join(out)


def _detect_self(caller_dir: Path | None = None) -> str | None:
    """Detect which location the caller is in."""
    if caller_dir is None:
        return None
    caller_dir = caller_dir.resolve()
    for name, loc in LOCATIONS.items():
        if caller_dir == loc.resolve() or str(caller_dir).startswith(str(loc.resolve())):
            return name
    return None


# ── Hub Sync (one-way: hub → targets) ───────────────────────────────────────


def _copy_if_newer(src: Path, dst: Path) -> str:
    """Copy src to dst if src is newer. Returns status: 'copied', 'skipped', 'new'."""
    if not src.exists():
        return "missing"
    if not dst.exists():
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)
        return "new"
    if src.stat().st_mtime > dst.stat().st_mtime:
        shutil.copy2(src, dst)
        return "copied"
    return "skipped"


def sync_from_hub(
    hub_path: Path, target_path: Path, dry_run: bool = False
) -> tuple[int, int, int, list[str], list[str]]:
    """Sync files from hub to a single target project.

    Returns:
        Tuple of (copied, new, skipped, details, errors).
    """
    from ai_constants import HUB_SYNC_FILES, HUB_SYNC_EXCLUDE

    copied = 0
    new = 0
    skipped = 0
    details: list[str] = []
    errors: list[str] = []
    exclude_set = set(HUB_SYNC_EXCLUDE)

    for rel in HUB_SYNC_FILES:
        if rel in exclude_set:
            continue
        src = hub_path / rel
        dst = target_path / rel
        if not src.exists():
            continue
        try:
            if dry_run:
                s, d = _read(src), _read(dst)
                if s != d:
                    label = "would_copy" if dst.exists() else "would_create"
                    details.append(f"  {label}: {rel}")
                    if dst.exists():
                        copied += 1
                    else:
                        new += 1
                else:
                    skipped += 1
            else:
                status = _copy_if_newer(src, dst)
                if status == "copied":
                    copied += 1
                    details.append(f"  updated: {rel}")
                elif status == "new":
                    new += 1
                    details.append(f"  created: {rel}")
                else:
                    skipped += 1
        except OSError as e:
            errors.append(f"  {rel}: {e}")
    return copied, new, skipped, details, errors


def hub_sync_to_all(
    hub_path: Path | None = None, dry_run: bool = False, quiet: bool = False
) -> bool:
    """Sync hub configs to all target projects.

    Args:
        hub_path: Override hub path (defaults to AI_HUB_PATH from constants).
        dry_run: Preview without copying.
        quiet: Suppress output (for auto-sync on startup).

    Returns:
        True if all syncs succeeded, False if any errors.
    """
    from ai_constants import AI_HUB_PATH, HUB_SYNC_TARGETS

    hub = hub_path or AI_HUB_PATH
    if not hub.exists():
        if not quiet:
            print(f"   {_R}Hub not found: {hub}{_0}")
        return False

    success = True
    total_changed = 0
    for name, target in HUB_SYNC_TARGETS.items():
        if not target.exists():
            if not quiet:
                print(f"   {_Y}Skipping {name}: {target} not found{_0}")
            continue
        copied, new, _skipped, details, errors = sync_from_hub(hub, target, dry_run=dry_run)
        total_changed += copied + new
        if errors:
            success = False
            if not quiet:
                print(f"   {_R}{name}: {len(errors)} error(s){_0}")
                for e in errors:
                    print(f"   {_R}{e}{_0}")
        elif not quiet and (copied or new):
            tag = "[DRY RUN] " if dry_run else ""
            print(f"   {_G}{tag}{name}: {copied} updated, {new} new{_0}")
            for d in details:
                print(f"   {_D}{d}{_0}")

    if not quiet and total_changed == 0 and success:
        print(f"   {_G}All targets in sync with hub.{_0}")
    return success


# ── Interactive Sync (project ↔ project) ────────────────────────────────────


def run_interactive(caller_dir: Path | None = None):
    """Interactive config sync — pick direction, review diffs, apply.

    Args:
        caller_dir: The project root that called this (to auto-detect location).
    """
    self_name = _detect_self(caller_dir)

    print()
    print(f"   {_B}--- Config Sync (models + scripts, excludes rules) ---{_0}")
    print()

    for name, path in LOCATIONS.items():
        tag = " <-- you are here" if name == self_name else ""
        ok = "ok" if path.exists() else "missing"
        print(f"   {name:15s} {_D}{path}{_0}  [{ok}]{tag}")

    print()
    names = list(LOCATIONS.keys())
    print(f"   {_B}Pick source (syncs to ALL others):{_0}")
    for i, name in enumerate(names, 1):
        tag = " *" if name == self_name else ""
        print(f"   [{i}] {name}{tag}")
    print(f"   [Q] Back")
    print()

    valid_choices = [str(i) for i in range(1, len(names) + 1)]
    choice = input(f"   Select [{', '.join(valid_choices)}, Q]: ").strip().lower()
    if choice in valid_choices:
        src = names[int(choice) - 1]
        targets = [n for n in names if n != src]
        for dst in targets:
            _run(src, dst)
            print()
    else:
        print("   Cancelled.")


def run_cli(direction: str):
    """CLI config sync — e.g. 'extensions->omni'."""
    parts = direction.lower().replace(" ", "").split("->")
    if len(parts) != 2 or parts[0] not in LOCATIONS or parts[1] not in LOCATIONS:
        valid = ", ".join(f"{a}->{b}" for a in LOCATIONS for b in LOCATIONS if a != b)
        print(f"   [ERROR] Invalid direction. Use: {valid}")
        return
    _run(parts[0], parts[1])


def _run(src_name: str, dst_name: str):
    """Core sync: compare, show diffs, prompt, apply."""
    src_root = LOCATIONS[src_name]
    dst_root = LOCATIONS[dst_name]

    print()
    print(f"   {_B}Comparing: {src_name} --> {dst_name}{_0}")
    print(f"   {_D}Source: {src_root}{_0}")
    print(f"   {_D}Target: {dst_root}{_0}")
    print()

    diffs = []
    for rel in SYNC_FILES:
        r = _diff_files(src_root / rel, dst_root / rel, rel)
        if r:
            diffs.append(r)

    r = _diff_profiles(src_root / PROFILES_FILE, dst_root / PROFILES_FILE)
    if r:
        diffs.append(r)

    if not diffs:
        print(f"   {_G}All files are in sync!{_0}")
        return

    print(f"   {_Y}Found {len(diffs)} file(s) with differences:{_0}")
    print()
    for i, d in enumerate(diffs, 1):
        st = "NEW" if not d.get("dst_ok") else "MODIFIED"
        print(f"   [{i}] {d['label']}  ({st})")

    print()
    print(f"   {_B}Options:{_0}")
    print(f"   [D] Show diffs for each file, then decide per file")
    print(f"   [A] Apply all changes (overwrite target)")
    print(f"   [Q] Quit without changes")
    print()

    action = input("   Select [D, A, Q]: ").strip().lower()

    if action == "q":
        print("   Cancelled.")
        return

    if action == "a":
        applied = 0
        for d in diffs:
            if d.get("error"):
                print(f"   {_R}Skipping {d['label']} (JSON error){_0}")
                continue
            _apply_one(d)
            applied += 1
            print(f"   {_G}Synced: {d['label']}{_0}")
        print(f"\n   {_G}{_B}Done! {applied}/{len(diffs)} applied ({src_name} --> {dst_name}){_0}")
        return

    # Per-file review
    applied = skipped = 0
    for d in diffs:
        print()
        print(f"   {'='*56}")
        print(f"   {_B}{d['label']}{_0}")
        print(f"   {'='*56}")

        if d.get("error"):
            print(f"   {_R}{d['diff'][0]}{_0}")
            skipped += 1
            continue

        print(_color_diff(d["diff"]))
        print()

        ch = input("   Apply this change? [Y/n/q]: ").strip().lower()
        if ch == "q":
            print("   Stopped.")
            break
        if ch in ("", "y", "yes"):
            _apply_one(d)
            applied += 1
            print(f"   {_G}Applied{_0}")
        else:
            skipped += 1
            print(f"   {_D}Skipped{_0}")

    print(f"\n   {_B}Summary: {applied} applied, {skipped} skipped{_0}")


# ── Standalone entry point ──────────────────────────────────────────────────

if __name__ == "__main__":
    import io as _io
    if sys.stdout.encoding != "utf-8":
        sys.stdout = _io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")
        sys.stderr = _io.TextIOWrapper(sys.stderr.buffer, encoding="utf-8", errors="replace")

    args = [a.lower() for a in sys.argv[1:]]
    if "--hub" in args:
        dry = "--dry-run" in args
        hub_sync_to_all(dry_run=dry)
    elif len(sys.argv) > 1 and "->" in sys.argv[1]:
        run_cli(sys.argv[1])
    else:
        run_interactive()
