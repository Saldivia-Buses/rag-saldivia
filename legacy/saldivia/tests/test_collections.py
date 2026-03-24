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
    with patch('saldivia.collections.connections'), \
         patch('saldivia.collections.utility') as mock_util, \
         patch('saldivia.collections.Collection') as mock_collection_cls:

        mock_util.list_collections.return_value = ["tecpia"]

        mock_sparse_field = MagicMock()
        mock_sparse_field.name = "sparse"
        mock_index = MagicMock()
        mock_index.params = {"index_type": "HNSW"}
        mock_coll = MagicMock()
        mock_coll.num_entities = 42
        mock_coll.schema.fields = [mock_sparse_field]
        mock_coll.indexes = [mock_index]
        mock_collection_cls.return_value = mock_coll

        manager = CollectionManager()
        result = manager.stats("tecpia")

        assert result is not None
        assert result.name == "tecpia"
        assert result.entity_count == 42
        assert result.has_sparse is True
        assert result.index_type == "HNSW"
