# saldivia/tests/test_collections.py
import pytest
from unittest.mock import patch, MagicMock

def test_collection_manager_list():
    from saldivia.collections import CollectionManager
    with patch('saldivia.collections.connections') as mock_conn, \
         patch('saldivia.collections.utility') as mock_util:
        mock_util.list_collections.return_value = ["tecpia", "docs"]
        manager = CollectionManager()
        collections = manager.list()
        assert "tecpia" in collections

def test_collection_manager_stats():
    from saldivia.collections import CollectionManager
    manager = CollectionManager()
    # Will test with mock
