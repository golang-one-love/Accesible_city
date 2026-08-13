from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PORT: int = Field(default=8001, alias="PORT")
    HOST: str = Field(default="0.0.0.0", alias="HOST")

    POSTGRES_HOST: str = Field(default="localhost", alias="POSTGRES_HOST")
    POSTGRES_PORT: int = Field(default=5432, alias="POSTGRES_PORT")
    POSTGRES_USER: str = Field(default="moderation_user", alias="POSTGRES_USER")
    POSTGRES_PASSWORD: str = Field(default="moderation_pass", alias="POSTGRES_PASSWORD")
    POSTGRES_DB: str = Field(default="moderation_db", alias="POSTGRES_DB")
    POSTGRES_SSL_MODE: str = Field(default="disable", alias="POSTGRES_SSL_MODE")

    REDIS_HOST: str = Field(default="localhost", alias="REDIS_HOST")
    REDIS_PORT: int = Field(default=6379, alias="REDIS_PORT")
    REDIS_PASSWORD: str = Field(default="", alias="REDIS_PASSWORD")
    REDIS_DB: int = Field(default=0, alias="REDIS_DB")

    STREAM_BARRIER_CREATED: str = Field(default="barrier.created", alias="STREAM_BARRIER_CREATED")
    STREAM_BARRIER_APPROVED: str = Field(default="barrier.approved", alias="STREAM_BARRIER_APPROVED")
    STREAM_BARRIER_REJECTED: str = Field(default="barrier.rejected", alias="STREAM_BARRIER_REJECTED")
    CONSUMER_GROUP_MODERATION: str = Field(default="moderation-group", alias="CONSUMER_GROUP_MODERATION")

    JWT_PUBLIC_KEY_PATH: str = Field(default="/keys/public.pem", alias="JWT_PUBLIC_KEY_PATH")

    BARRIER_SERVICE_URL: str = Field(default="http://barrier-service:8000", alias="BARRIER_SERVICE_URL")

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