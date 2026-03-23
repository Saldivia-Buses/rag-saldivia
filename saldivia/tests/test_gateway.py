# saldivia/tests/test_gateway.py
"""Tests for gateway auth and collection filtering."""
import pytest
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient
from fastapi import HTTPException

from saldivia.auth import User, Role, Permission
from saldivia.gateway import app, filter_collections


@pytest.fixture
def regular_user():
    return User(
        id=2,
        email="user@test.com",
        name="User",
        area_id=2,
        role=Role.USER,
        api_key_hash="user_hash"
    )


# Test filter_collections logic
def test_filter_collections_admin(admin_user):
    """Admin users should receive all collections they request."""
    requested = ["col1", "col2", "col3"]
    with patch("saldivia.gateway.db") as mock_db:
        result = filter_collections(admin_user, requested)
        assert result == requested
        # Admin should not trigger DB calls
        mock_db.get_user_collections.assert_not_called()


def test_filter_collections_user_restricted(regular_user):
    """Regular users should only receive allowed collections."""
    requested = ["col1", "col2", "col3"]
    allowed = ["col1", "col3"]

    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_collections.return_value = allowed
        result = filter_collections(regular_user, requested)
        assert result == ["col1", "col3"]
        mock_db.get_user_collections.assert_called_once_with(regular_user)


def test_filter_collections_user_no_access(regular_user):
    """Users with no access to any requested collection should get 403."""
    requested = ["col_other"]
    allowed = ["col_mine"]

    with patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_collections.return_value = allowed
        with pytest.raises(HTTPException) as exc_info:
            filter_collections(regular_user, requested)

        assert exc_info.value.status_code == 403
        assert "No access to requested collections" in exc_info.value.detail


def test_filter_collections_bypass_auth():
    """With user=None (BYPASS_AUTH mode), return all collections."""
    requested = ["col1", "col2"]
    with patch("saldivia.gateway.db") as mock_db:
        result = filter_collections(None, requested)
        assert result == requested
        mock_db.get_user_collections.assert_not_called()


# Test endpoints
def test_health_endpoint():
    """GET /health should return status=ok."""
    client = TestClient(app)
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert "auth_enabled" in data
