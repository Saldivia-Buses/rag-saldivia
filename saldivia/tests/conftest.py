# saldivia/tests/conftest.py
"""Pytest configuration: set env vars before any gateway module is imported."""
import os

# Allow gateway to import without JWT_SECRET set.
# Individual tests that need auth enabled mock the relevant variables.
os.environ.setdefault("BYPASS_AUTH", "true")
os.environ.setdefault("ENVIRONMENT", "development")
