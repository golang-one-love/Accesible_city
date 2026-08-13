from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PORT: int = Field(default=8000, alias="PORT")
    HOST: str = Field(default="0.0.0.0", alias="HOST")

    POSTGRES_HOST: str = Field(default="localhost", alias="POSTGRES_HOST")
    POSTGRES_PORT: int = Field(default=5432, alias="POSTGRES_PORT")
    POSTGRES_USER: str = Field(default="barrier_user", alias="POSTGRES_USER")
    POSTGRES_PASSWORD: str = Field(default="barrier_pass", alias="POSTGRES_PASSWORD")
    POSTGRES_DB: str = Field(default="barrier_db", alias="POSTGRES_DB")
    POSTGRES_SSL_MODE: str = Field(default="disable", alias="POSTGRES_SSL_MODE")

    REDIS_HOST: str = Field(default="localhost", alias="REDIS_HOST")
    REDIS_PORT: int = Field(default=6379, alias="REDIS_PORT")
    REDIS_PASSWORD: str = Field(default="", alias="REDIS_PASSWORD")
    REDIS_DB: int = Field(default=0, alias="REDIS_DB")

    MINIO_ENDPOINT: str = Field(default="localhost:9000", alias="MINIO_ENDPOINT")
    MINIO_PUBLIC_URL: str = Field(default="http://localhost:9000", alias="MINIO_PUBLIC_URL")
    MINIO_ACCESS_KEY: str = Field(default="minioadmin", alias="MINIO_ACCESS_KEY")
    MINIO_SECRET_KEY: str = Field(default="minioadmin", alias="MINIO_SECRET_KEY")
    MINIO_BUCKET: str = Field(default="barrier-photos", alias="MINIO_BUCKET")
    MINIO_USE_SSL: bool = Field(default=False, alias="MINIO_USE_SSL")
    S3_REGION: str = Field(default="us-east-1", alias="S3_REGION")
    S3_ADDRESSING_STYLE: str = Field(default="path", alias="S3_ADDRESSING_STYLE")

    STREAM_BARRIER_CREATED: str = Field(default="barrier.created", alias="STREAM_BARRIER_CREATED")

    JWT_PUBLIC_KEY_PATH: str = Field(default="/keys/public.pem", alias="JWT_PUBLIC_KEY_PATH")

    DB_ECHO: bool = Field(default=False, alias="DB_ECHO")
    ENV: str = Field(default="development", alias="ENV")

    @property
    def database_url(self) -> str:
        url = f"postgresql+asyncpg://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"
        if self.POSTGRES_SSL_MODE not in ("", "disable"):
            url += f"?ssl={self.POSTGRES_SSL_MODE}"
        return url

    @property
    def redis_url(self) -> str:
        if self.REDIS_PASSWORD:
            return f"redis://:{self.REDIS_PASSWORD}@{self.REDIS_HOST}:{self.REDIS_PORT}/{self.REDIS_DB}"
        return f"redis://{self.REDIS_HOST}:{self.REDIS_PORT}/{self.REDIS_DB}"


settings = Settings()