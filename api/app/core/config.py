"""Configuration - Pydantic Settings"""
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    service_name: str = "phantomproxy-api"
    database_url: str = "postgresql+asyncpg://phantom:phantom@localhost:5432/phantom"
    redis_url: str = "redis://localhost:6379/0"
    cors_origins: list[str] = ["http://localhost:3000", "http://localhost:3001"]
    otlp_endpoint: str | None = None  # OpenTelemetry collector
    environment: str = "development"

    model_config = {"env_file": ".env", "extra": "ignore"}


settings = Settings()
