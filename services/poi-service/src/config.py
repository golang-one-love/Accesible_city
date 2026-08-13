from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PORT: int = Field(default=8002, alias="PORT")
    HOST: str = Field(default="0.0.0.0", alias="HOST")

    POSTGRES_HOST: str = Field(default="localhost", alias="POSTGRES_HOST")
    POSTGRES_PORT: int = Field(default=5432, alias="POSTGRES_PORT")
    POSTGRES_USER: str = Field(default="poi_user", alias="POSTGRES_USER")
    POSTGRES_PASSWORD: str = Field(default="poi_pass", alias="POSTGRES_PASSWORD")
    POSTGRES_DB: str = Field(default="poi_db", alias="POSTGRES_DB")
    POSTGRES_SSL_MODE: str = Field(default="disable", alias="POSTGRES_SSL_MODE")

    DB_ECHO: bool = Field(default=False, alias="DB_ECHO")
    ENV: str = Field(default="development", alias="ENV")

    JWT_PUBLIC_KEY_PATH: str = Field(default="/keys/public.pem", alias="JWT_PUBLIC_KEY_PATH")

    @property
    def database_url(self) -> str:
        url = f"postgresql+asyncpg://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"
        if self.POSTGRES_SSL_MODE not in ("", "disable"):
            url += f"?ssl={self.POSTGRES_SSL_MODE}"
        return url


settings = Settings()