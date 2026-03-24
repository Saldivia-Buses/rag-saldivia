# saldivia/tests/test_mode_manager.py
import pytest
from unittest.mock import patch, MagicMock

def test_mode_manager_initial_state():
    from saldivia.mode_manager import ModeManager, Mode
    manager = ModeManager(gpu_memory_gb=98)
    assert manager.current_mode == Mode.QUERY

def test_mode_manager_can_switch_to_ingest():
    from saldivia.mode_manager import ModeManager, Mode
    manager = ModeManager(gpu_memory_gb=98)
    assert manager.can_switch_to(Mode.INGEST) == True

def test_mode_manager_memory_requirements():
    from saldivia.mode_manager import ModeManager, Mode, MEMORY_REQUIREMENTS
    assert MEMORY_REQUIREMENTS[Mode.QUERY] < 50  # NIMs only
    assert MEMORY_REQUIREMENTS[Mode.INGEST] < 95  # NIMs + VLM
