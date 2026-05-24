#!/usr/bin/env python3
"""
Cross-repo skill reference integrity validator.

Checks all 4 repos:
  - Registered skills (in opencode-profiles.json shared.agents / shared.categories)
    must exist as a directory in .opencode/skills/<name>/
  - Skills on disk that are unregistered are documented (not silently ignored)
  - Seeded-failure mode for testing

Classification:
  RESTORE   = skill referenced but missing from disk → needs creation
  REGISTER  = skill on disk but unregistered → needs profile entry
  OPTIONAL  = skill on disk but intentionally unmanaged (documented)
  RETIRE    = skill referenced but should be removed (owner decision)
  OK        = skill referenced AND exists on disk
"""

import json
import os
import sys
import argparse
from pathlib import Path

REPOS = {
    "extensions": "/home/ena/Project/extensions",
    "ads-analytics": "/home/ena/Project/ads-analytics",
    "auto": "/home/ena/Project/auto",
    "omni": "/home/ena/Project/omni",
}

PROFILES_REL = "opencode-configs/opencode-profiles.json"
SKILL_ROOT_REL = ".opencode/skills"


def extract_skill_refs(profiles_path):
    """Extract all skill references from shared.agents.*.skills and shared.categories.*.skills"""
    with open(profiles_path, "r") as f:
        data = json.load(f)

    shared = data.get("shared", {})
    refs = {}  # skill_name -> [locations]

    # From shared.agents
    agents = shared.get("agents", {})
    for agent_name, agent_config in agents.items():
        skills = agent_config.get("skills", [])
        for skill in skills:
            refs.setdefault(skill, []).append(f"agents.{agent_name}")

    # From shared.categories
    categories = shared.get("categories", {})
    for cat_name, cat_config in categories.items():
        skills = cat_config.get("skills", [])
        for skill in skills:
            refs.setdefault(skill, []).append(f"categories.{cat_name}")

    # Check prompt_append for inline skill references (e.g. /skill-name)
    # but these are load-time references, not registration. Skip for now.

    return refs


def list_on_disk_skills(repo_root):
    """List all skill directories in .opencode/skills/"""
    skill_root = os.path.join(repo_root, SKILL_ROOT_REL)
    if not os.path.isdir(skill_root):
        return []
    return sorted([
        d.name for d in Path(skill_root).iterdir()
        if d.is_dir() and not d.name.startswith(".")
    ])


def classify_finding(skill_name, referenced_in, on_disk):
    """Classify a finding based on reference and disk status."""
    if on_disk and not referenced_in:
        return "unregistered"
    if not on_disk and referenced_in:
        return "missing"
    return "ok"


def run_validation(repos, seed_failure=None):
    """Run validation across all repos. Returns (findings, exit_code)."""
    findings = []
    has_errors = False

    for repo_name, repo_root in repos.items():
        profiles_path = os.path.join(repo_root, PROFILES_REL)

        if not os.path.exists(profiles_path):
            findings.append({
                "repo": repo_name,
                "type": "error",
                "message": f"Profiles not found: {profiles_path}",
            })
            has_errors = True
            continue

        # Extract references
        refs = extract_skill_refs(profiles_path)

        # List on-disk skills
        disk_skills = set(list_on_disk_skills(repo_root))

        # Seeded failure: inject a fake reference
        if seed_failure and repo_name == seed_failure.get("repo"):
            fake_skill = seed_failure.get("skill", "this-skill-does-not-exist")
            refs[fake_skill] = ["seeded-test-failure"]
            findings.append({
                "repo": repo_name,
                "type": "info",
                "message": f"Seeded failure skill '{fake_skill}' injected for testing",
            })

        # Cross-reference: referenced skills on disk?
        for skill_name, locations in sorted(refs.items()):
            exists = skill_name in disk_skills
            if not exists:
                findings.append({
                    "repo": repo_name,
                    "type": "error",
                    "category": "RESTORE",
                    "skill": skill_name,
                    "locations": locations,
                    "message": f"Skill '{skill_name}' referenced in {', '.join(locations)} but NOT found on disk in .opencode/skills/",
                })
                has_errors = True
            else:
                findings.append({
                    "repo": repo_name,
                    "type": "ok",
                    "category": "OK",
                    "skill": skill_name,
                    "locations": locations,
                    "message": f"Skill '{skill_name}' OK — exists on disk",
                })

        # Cross-reference: skills on disk but unregistered?
        registered_skills = set(refs.keys())
        unregistered = disk_skills - registered_skills
        for skill_name in sorted(unregistered):
            # Check if it's a BMAD skill (only .opencode/skills is checked)
            findings.append({
                "repo": repo_name,
                "type": "warn",
                "category": "OPTIONAL/REGISTER",
                "skill": skill_name,
                "locations": [],
                "message": f"Skill '{skill_name}' exists on disk but is NOT registered in opencode-profiles.json",
            })

    return findings, 1 if has_errors else 0


def print_report(findings, exit_code):
    """Print a structured report."""
    print("=" * 72)
    print("CROSS-REPO SKILL REFERENCE INTEGRITY VALIDATOR REPORT")
    print("=" * 72)
    print()

    by_repo = {}
    for f in findings:
        by_repo.setdefault(f["repo"], []).append(f)

    total_errors = sum(1 for f in findings if f.get("type") == "error")
    total_warns = sum(1 for f in findings if f.get("type") == "warn")
    total_ok = sum(1 for f in findings if f.get("type") == "ok")
    total_info = sum(1 for f in findings if f.get("type") == "info")

    for repo_name in sorted(by_repo.keys()):
        repo_findings = by_repo[repo_name]
        errors = [f for f in repo_findings if f.get("type") == "error"]
        warns = [f for f in repo_findings if f.get("type") == "warn"]
        oks = [f for f in repo_findings if f.get("type") == "ok"]
        infos = [f for f in repo_findings if f.get("type") == "info"]

        print(f"─── {repo_name} ───")
        print(f"  Errors: {len(errors)} | Warnings: {len(warns)} | OK: {len(oks)}")

        for f in infos:
            print(f"  ℹ️  {f['message']}")

        for f in errors:
            print(f"  ❌ [{f['category']}] {f['message']}")

        for f in warns:
            print(f"  ⚠️  [{f['category']}] {f['message']}")

        for f in oks:
            print(f"  ✅ {f['message']}")

        print()

    print("─" * 72)
    print(f"TOTAL: {total_errors} errors, {total_warns} warnings, {total_ok} ok, {total_info} info")
    print(f"Exit code: {exit_code}")
    print("=" * 72)

    return total_errors, total_warns


def main():
    parser = argparse.ArgumentParser(description="Cross-repo skill reference validator")
    parser.add_argument("--seed-failure", metavar="REPO:SKILL",
                        help="Inject a fake skill reference for testing (format: repo:skill_name)")
    parser.add_argument("--exit-zero", action="store_true",
                        help="Always exit 0 (for CI when warnings are acceptable)")
    args = parser.parse_args()

    seed_failure = None
    if args.seed_failure:
        parts = args.seed_failure.split(":", 1)
        if len(parts) != 2:
            print("ERROR: --seed-failure must be in format REPO:SKILL_NAME", file=sys.stderr)
            sys.exit(2)
        seed_failure = {"repo": parts[0], "skill": parts[1]}
        if seed_failure["repo"] not in REPOS:
            print(f"ERROR: Unknown repo '{seed_failure['repo']}'. Valid: {', '.join(REPOS.keys())}", file=sys.stderr)
            sys.exit(2)

    findings, exit_code = run_validation(REPOS, seed_failure)
    total_errors, total_warns = print_report(findings, exit_code)

    if args.exit_zero:
        sys.exit(0)
    sys.exit(exit_code)


if __name__ == "__main__":
    main()
