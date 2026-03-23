# saldivia/tests/test_tier.py
import pytest
from saldivia.tier import classify_tier, TierConfig, TIERS


def test_classify_tier_by_pages_tiny():
    assert classify_tier(page_count=5) == "tiny"

def test_classify_tier_by_pages_small():
    assert classify_tier(page_count=50) == "small"

def test_classify_tier_by_pages_medium():
    assert classify_tier(page_count=150) == "medium"

def test_classify_tier_by_pages_large():
    assert classify_tier(page_count=500) == "large"

def test_classify_tier_boundary_tiny_small():
    assert classify_tier(page_count=20) == "tiny"
    assert classify_tier(page_count=21) == "small"

def test_classify_tier_boundary_small_medium():
    assert classify_tier(page_count=80) == "small"
    assert classify_tier(page_count=81) == "medium"

def test_classify_tier_boundary_medium_large():
    assert classify_tier(page_count=300) == "medium"
    assert classify_tier(page_count=301) == "large"

def test_classify_tier_by_file_size_when_no_pages():
    assert classify_tier(page_count=None, file_size=50_000) == "tiny"
    assert classify_tier(page_count=None, file_size=500_000) == "small"
    assert classify_tier(page_count=None, file_size=5_000_000) == "medium"
    assert classify_tier(page_count=None, file_size=50_000_000) == "large"

def test_classify_tier_pages_take_priority_over_size():
    # Si hay pages, ignorar file_size
    assert classify_tier(page_count=5, file_size=999_999_999) == "tiny"

def test_classify_tier_no_args_defaults_to_tiny():
    # Sin args, debe devolver tier válido sin crash
    result = classify_tier()
    assert result in TIERS

def test_tiers_have_required_fields():
    for name, tier in TIERS.items():
        assert isinstance(tier, TierConfig)
        assert hasattr(tier, "poll_interval")
        assert hasattr(tier, "restart_after")
        assert hasattr(tier, "timeout")
        assert tier.poll_interval > 0
        assert tier.restart_after > 0
        assert tier.timeout > 0

def test_classify_tier_returns_valid_tier_name():
    for pages in [5, 50, 150, 500]:
        result = classify_tier(page_count=pages)
        assert result in TIERS, f"classify_tier({pages}) returned invalid tier: {result}"

def test_tiers_ordered_by_complexity():
    # Los tiers deben tener timeouts crecientes (large > medium > small > tiny)
    assert TIERS["large"].timeout > TIERS["medium"].timeout
    assert TIERS["medium"].timeout > TIERS["small"].timeout
    assert TIERS["small"].timeout > TIERS["tiny"].timeout
