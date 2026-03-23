# saldivia/tests/conftest.py
"""Pytest configuration: set env vars before any gateway module is imported."""
import os

# Allow gateway to import without JWT_SECRET set.
# Individual tests that need auth enabled mock the relevant variables.
os.environ.setdefault("BYPASS_AUTH", "true")
os.environ.setdefault("ENVIRONMENT", "development")

import pytest
from fastapi.testclient import TestClient
from unittest.mock import MagicMock
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key, hash_password


@pytest.fixture
def admin_user():
    key, hash_val = generate_api_key()
    return User(
        id=1, email="admin@test.com", name="Admin", area_id=1,
        role=Role.ADMIN, api_key_hash=hash_val,
        password_hash=hash_password("admin123"),
        active=True,
    )


@pytest.fixture
def client():
    return TestClient(app, raise_server_exceptions=False)
