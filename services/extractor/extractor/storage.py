"""MinIO/S3 client for the Extractor service.

Reads uploaded documents and stores extracted images.
"""

import io
import logging

import boto3
from botocore.config import Config as BotoConfig
from botocore.exceptions import ClientError

logger = logging.getLogger(__name__)


class StorageClient:
    def __init__(self, endpoint: str, bucket: str, access_key: str, secret_key: str):
        self.bucket = bucket
        self._client = boto3.client(
            "s3",
            endpoint_url=endpoint,
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            region_name="us-east-1",
            config=BotoConfig(signature_version="s3v4"),
        )

    def get(self, key: str) -> bytes:
        """Download a file from S3/MinIO and return its bytes."""
        resp = self._client.get_object(Bucket=self.bucket, Key=key)
        return resp["Body"].read()

    def put(self, key: str, data: bytes, content_type: str = "application/octet-stream") -> None:
        """Upload bytes to S3/MinIO."""
        self._client.put_object(
            Bucket=self.bucket,
            Key=key,
            Body=io.BytesIO(data),
            ContentType=content_type,
        )

    def exists(self, key: str) -> bool:
        """Check if a key exists in S3/MinIO."""
        try:
            self._client.head_object(Bucket=self.bucket, Key=key)
            return True
        except ClientError as e:
            if e.response["Error"]["Code"] == "404":
                return False
            raise
