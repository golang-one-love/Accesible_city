
import boto3
from botocore.config import Config

from src.config import settings
from src.domain.repositories import PhotoStorage


class S3Storage(PhotoStorage):
    def __init__(self):
        endpoint = f"http://{settings.MINIO_ENDPOINT}" if not settings.MINIO_USE_SSL else f"https://{settings.MINIO_ENDPOINT}"
        self.client = boto3.client(
            "s3",
            endpoint_url=endpoint,
            aws_access_key_id=settings.MINIO_ACCESS_KEY,
            aws_secret_access_key=settings.MINIO_SECRET_KEY,
            config=Config(signature_version="s3v4", s3={"addressing_style": settings.S3_ADDRESSING_STYLE}),
            region_name=settings.S3_REGION,
        )
        self.bucket = settings.MINIO_BUCKET
        public_endpoint = settings.MINIO_PUBLIC_URL
        self.public_client = None
        if public_endpoint and public_endpoint != endpoint:
            self.public_client = boto3.client(
                "s3",
                endpoint_url=public_endpoint,
                aws_access_key_id=settings.MINIO_ACCESS_KEY,
                aws_secret_access_key=settings.MINIO_SECRET_KEY,
                config=Config(signature_version="s3v4", s3={"addressing_style": settings.S3_ADDRESSING_STYLE}),
                region_name=settings.S3_REGION,
            )
        self._ensure_bucket()

    def _ensure_bucket(self) -> None:
        try:
            self.client.head_bucket(Bucket=self.bucket)
        except self.client.exceptions.ClientError:
            try:
                self.client.create_bucket(Bucket=self.bucket)
            except Exception as e:
                print(f"[s3] bucket '{self.bucket}' not found and create_bucket failed: {e}. "
                      "Pre-create it in the storage panel if uploads are expected.")
        except Exception as e:
            print(f"[s3] storage endpoint unreachable at startup ({e}). "
                  "Uploads will fail until storage is available, service keeps running.")

    async def upload(self, key: str, data: bytes, content_type: str) -> str:
        self.client.put_object(
            Bucket=self.bucket,
            Key=key,
            Body=data,
            ContentType=content_type,
        )
        return key

    async def download(self, key: str) -> bytes:
        response = self.client.get_object(Bucket=self.bucket, Key=key)
        return response["Body"].read()

    async def delete(self, key: str) -> bool:
        try:
            self.client.delete_object(Bucket=self.bucket, Key=key)
            return True
        except self.client.exceptions.ClientError:
            return False

    async def generate_presigned_url(self, key: str, expires_in: int = 3600) -> str:
        client = self.public_client or self.client
        return client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self.bucket, "Key": key},
            ExpiresIn=expires_in,
        )