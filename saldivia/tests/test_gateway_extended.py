"""Tests for SDA gateway extensions."""
import pytest
from saldivia.auth.models import hash_password, verify_password

def test_password_hashing():
    pw = "supersecret123"
    hashed = hash_password(pw)
    assert verify_password(pw, hashed)
    assert not verify_password("wrong", hashed)
    assert hashed != pw
