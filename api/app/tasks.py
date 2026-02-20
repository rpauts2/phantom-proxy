"""Celery tasks - AI generation, PDF reports, notifications"""
try:
    from celery_app import app
except ImportError:
    from api.celery_app import app


@app.task
def generate_phishing_email(target_data: dict, template: str):
    """AI-generated personalized email (LangGraph/Llama)"""
    return {"status": "queued", "task": "generate_email"}


@app.task
def generate_pdf_report(campaign_id: str):
    """Async PDF report generation"""
    return {"status": "queued", "campaign": campaign_id}
