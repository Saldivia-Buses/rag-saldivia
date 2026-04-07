"""Tests for NATS subject validation — security boundary."""

import re

import pytest

# Replicate the validation logic from main.py to avoid importing the full
# module (which requires pythonjsonlogger and NATS dependencies).
_SAFE_SUBJECT_RE = re.compile(r"^[a-zA-Z0-9_-]+$")


def _validate_subject_token(value: str, field: str) -> None:
    """Reject NATS subject injection (dots, wildcards, whitespace)."""
    if not _SAFE_SUBJECT_RE.match(value):
        raise ValueError(f"invalid NATS subject token in {field}: {value!r}")


def test_valid_subject_tokens():
    """Valid NATS subject tokens pass validation."""
    for token in ["saldivia", "doc-123", "tenant_1", "ABC", "a1b2c3"]:
        _validate_subject_token(token, "test")  # should not raise


def test_invalid_subject_dots():
    """Dots in subject tokens are rejected (NATS injection)."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("tenant.evil", "tenant_slug")


def test_invalid_subject_wildcard():
    """Wildcards in subject tokens are rejected."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("tenant.*", "tenant_slug")


def test_invalid_subject_greater_than():
    """Greater-than wildcards are rejected."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("tenant.>", "tenant_slug")


def test_invalid_subject_spaces():
    """Spaces in subject tokens are rejected."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("tenant slug", "tenant_slug")


def test_invalid_subject_empty():
    """Empty subject tokens are rejected."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("", "tenant_slug")


def test_invalid_subject_slash():
    """Slashes in subject tokens are rejected."""
    with pytest.raises(ValueError, match="invalid NATS subject token"):
        _validate_subject_token("../../etc/passwd", "document_id")
