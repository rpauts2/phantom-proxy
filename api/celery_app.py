"""
Celery - Task Queue for AI generation, reports, notifications
"""
import os
from celery import Celery

redis_url = os.getenv("REDIS_URL", "redis://localhost:6379/0")
app = Celery(
    "phantomproxy",
    broker=redis_url,
    backend=redis_url,
    include=["app.tasks"],
)

app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="UTC",
    enable_utc=True,
    task_track_started=True,
)
