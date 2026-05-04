"""
API Client for OMNI Backend.
Handles authentication, token management, and API calls.
"""

import sys
import time
from typing import Optional

try:
    import requests
except ImportError:
    print("Installing requests...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "requests", "-q"])
    import requests

from config import BASE_URL, LOGIN_ENDPOINT, USERNAME, PASSWORD, REQUEST_TIMEOUT


class OmniAPIClient:
    """Client for OMNI backend API with automatic auth."""

    def __init__(self, base_url: str = BASE_URL, username: str = USERNAME, password: str = PASSWORD):
        self.base_url = base_url.rstrip("/")
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.token: Optional[str] = None
        self.tenant_id: Optional[str] = None
        self.token_expires_at: float = 0

    def login(self) -> bool:
        """Authenticate and store token."""
        url = f"{self.base_url}{LOGIN_ENDPOINT}"
        payload = {"username": self.username, "password": self.password}

        print(f"[AUTH] Logging in as '{self.username}'...")
        resp = self.session.post(url, json=payload, timeout=REQUEST_TIMEOUT)

        if resp.status_code != 200:
            print(f"[AUTH] Login failed: HTTP {resp.status_code}")
            print(f"       Response: {resp.text[:200]}")
            return False

        data = resp.json()
        if not data.get("success"):
            print(f"[AUTH] Login failed: {data.get('error', 'Unknown error')}")
            return False

        # Extract token from response
        login_data = data.get("data", {})
        self.token = login_data.get("access_token") or login_data.get("token")
        self.tenant_id = login_data.get("tenant_id")

        if not self.token:
            print("[AUTH] Login succeeded but no token in response")
            return False

        # Set token expiry (default 30 min, refresh at 25 min)
        expires_in = login_data.get("expires_in", 1800)
        self.token_expires_at = time.time() + expires_in - 300

        # Set auth header for all subsequent requests
        self.session.headers["Authorization"] = f"Bearer {self.token}"

        print(f"[AUTH] Login successful. Tenant: {self.tenant_id}")
        return True

    def ensure_auth(self) -> bool:
        """Ensure we have a valid token, re-login if expired."""
        if self.token and time.time() < self.token_expires_at:
            return True
        return self.login()

    def get(self, endpoint: str, params: Optional[dict] = None) -> dict:
        """Make authenticated GET request. Returns JSON response."""
        if not self.ensure_auth():
            raise RuntimeError("Authentication failed")

        url = f"{self.base_url}{endpoint}"
        resp = self.session.get(url, params=params, timeout=REQUEST_TIMEOUT)

        if resp.status_code != 200:
            return {"success": False, "error": f"HTTP {resp.status_code}: {resp.text[:200]}"}

        return resp.json()

    def post(self, endpoint: str, json_data: Optional[dict] = None) -> dict:
        """Make authenticated POST request. Returns JSON response."""
        if not self.ensure_auth():
            raise RuntimeError("Authentication failed")

        url = f"{self.base_url}{endpoint}"
        resp = self.session.post(url, json=json_data, timeout=REQUEST_TIMEOUT)

        if resp.status_code != 200:
            # Try to parse error from response body
            try:
                error_body = resp.json()
                return {"success": False, "error": error_body.get("error", f"HTTP {resp.status_code}")}
            except Exception:
                return {"success": False, "error": f"HTTP {resp.status_code}: {resp.text[:200]}"}

        return resp.json()
