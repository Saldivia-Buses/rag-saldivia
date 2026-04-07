"""Tests for S3/MinIO storage client — mocked boto3."""

from unittest.mock import MagicMock, patch
from io import BytesIO

import pytest
from botocore.exceptions import ClientError

from extractor.storage import StorageClient


@patch("extractor.storage.boto3")
def test_get_returns_bytes(mock_boto3):
    """Storage.get downloads and returns file bytes."""
    mock_s3 = MagicMock()
    mock_boto3.client.return_value = mock_s3
    mock_s3.get_object.return_value = {"Body": BytesIO(b"pdf content")}

    client = StorageClient("http://minio:9000", "docs", "key", "secret")
    result = client.get("tenant/doc-1/original.pdf")

    assert result == b"pdf content"
    mock_s3.get_object.assert_called_once_with(
        Bucket="docs", Key="tenant/doc-1/original.pdf"
    )


@patch("extractor.storage.boto3")
def test_put_uploads_bytes(mock_boto3):
    """Storage.put uploads bytes with content type."""
    mock_s3 = MagicMock()
    mock_boto3.client.return_value = mock_s3

    client = StorageClient("http://minio:9000", "docs", "key", "secret")
    client.put("tenant/doc-1/images/p1.png", b"png data", "image/png")

    mock_s3.put_object.assert_called_once()
    call_kwargs = mock_s3.put_object.call_args[1]
    assert call_kwargs["Bucket"] == "docs"
    assert call_kwargs["Key"] == "tenant/doc-1/images/p1.png"
    assert call_kwargs["ContentType"] == "image/png"


@patch("extractor.storage.boto3")
def test_exists_returns_true(mock_boto3):
    """Storage.exists returns True when object exists."""
    mock_s3 = MagicMock()
    mock_boto3.client.return_value = mock_s3

    client = StorageClient("http://minio:9000", "docs", "key", "secret")
    assert client.exists("tenant/doc-1/original.pdf") is True
    mock_s3.head_object.assert_called_once()


@patch("extractor.storage.boto3")
def test_exists_returns_false_on_404(mock_boto3):
    """Storage.exists returns False when object doesn't exist."""
    mock_s3 = MagicMock()
    mock_boto3.client.return_value = mock_s3
    mock_s3.head_object.side_effect = ClientError(
        {"Error": {"Code": "404", "Message": "Not Found"}}, "HeadObject"
    )

    client = StorageClient("http://minio:9000", "docs", "key", "secret")
    assert client.exists("nonexistent") is False


@patch("extractor.storage.boto3")
def test_exists_raises_on_other_errors(mock_boto3):
    """Storage.exists raises on non-404 errors."""
    mock_s3 = MagicMock()
    mock_boto3.client.return_value = mock_s3
    mock_s3.head_object.side_effect = ClientError(
        {"Error": {"Code": "403", "Message": "Forbidden"}}, "HeadObject"
    )

    client = StorageClient("http://minio:9000", "docs", "key", "secret")
    with pytest.raises(ClientError):
        client.exists("forbidden-key")
