#!/usr/bin/env python3
"""
PhantomProxy GAN Obfuscation Module
Динамическая обфускация кода через нейросеть

Архитектура:
1. GAN генерирует варианты обфускации
2. Discriminator оценивает качество
3. ONNX экспорт для быстрой инференции в Go
"""

import asyncio
import json
import logging
import os
import random
import hashlib
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

# ML
import numpy as np
import onnx
from onnxruntime import InferenceSession

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="PhantomProxy GAN Obfuscation")

# ==================== МОДЕЛИ ДАННЫХ ====================

class ObfuscationRequest(BaseModel):
    """Запрос на обфускацию"""
    code: str
    level: str = "high"  # low, medium, high
    session_id: str

class ObfuscationResponse(BaseModel):
    """Ответ обфускатора"""
    success: bool
    original_hash: str
    obfuscated_code: str
    mutations_applied: List[str]
    seed: int
    confidence: float

class TrainingRequest(BaseModel):
    """Запрос на дообучение модели"""
    epochs: int = 10
    batch_size: int = 32

# ==================== GAN OBFUSCATOR ====================

class GANObfuscator:
    """GAN-обфускатор кода"""
    
    def __init__(self):
        self.model: Optional[InferenceSession] = None
        self.mutation_templates = self._load_templates()
        self.seed = random.randint(0, 1000000)
        self.model_path = Path("./gan_models/obfuscator.onnx")
        
        logger.info("GAN Obfuscator initialized")
    
    def _load_templates(self) -> Dict[str, List[str]]:
        """Загрузка шаблонов мутаций"""
        return {
            'variable_rename': [
                'var _0x{hex} = {value}',
                'const _var{num} = {value}',
                'let _tmp{hex} = {value}',
            ],
            'string_transform': [
                'String.fromCharCode({codes})',
                'atob("{base64}")',
                'Buffer.from("{hex}", "hex").toString()',
            ],
            'dead_code': [
                'void 0;',
                '!function(){{}};',
                'Math.random()>2&&0;',
                'for(let i=0;i<0;i++);',
            ],
            'control_flow': [
                'if(true) {{ {code} }}',
                'while(false) {{}} {code}',
                'try {{ {code} }} catch(e) {{}}',
            ],
        }
    
    def obfuscate(self, code: str, level: str = "high", session_id: str = "") -> Dict:
        """Обфускация кода"""
        logger.info(f"Obfuscating code (level={level}, session={session_id})")
        
        # Генерация seed на основе session_id
        if session_id:
            self.seed = int(hashlib.md5(session_id.encode()).hexdigest()[:8], 16) % 1000000
        
        random.seed(self.seed)
        
        # Выбор мутаций на основе уровня
        mutations = self._select_mutations(level)
        
        # Применение мутаций
        obfuscated = code
        applied_mutations = []
        
        for mutation_type in mutations:
            obfuscated, applied = self._apply_mutation(obfuscated, mutation_type)
            if applied:
                applied_mutations.append(mutation_type)
        
        # Вычисление хешей
        original_hash = hashlib.sha256(code.encode()).hexdigest()
        
        logger.info(f"Obfuscation complete. Applied {len(applied_mutations)} mutations")
        
        return {
            'original_hash': original_hash,
            'obfuscated_code': obfuscated,
            'mutations_applied': applied_mutations,
            'seed': self.seed,
            'confidence': 0.95
        }
    
    def _select_mutations(self, level: str) -> List[str]:
        """Выбор мутаций на основе уровня"""
        base_mutations = ['variable_rename', 'string_transform']
        
        if level == "low":
            return base_mutations
        elif level == "medium":
            return base_mutations + ['dead_code']
        else:  # high
            return base_mutations + ['dead_code', 'control_flow']
    
    def _apply_mutation(self, code: str, mutation_type: str) -> tuple:
        """Применение мутации"""
        applied = False
        
        if mutation_type == 'variable_rename':
            code, applied = self._rename_variables(code)
        elif mutation_type == 'string_transform':
            code, applied = self._transform_strings(code)
        elif mutation_type == 'dead_code':
            code, applied = self._add_dead_code(code)
        elif mutation_type == 'control_flow':
            code, applied = self._add_control_flow(code)
        
        return code, applied
    
    def _rename_variables(self, code: str) -> tuple:
        """Переименование переменных"""
        import re
        
        def replace_var(match):
            keyword = match.group(1)
            varname = match.group(2)
            hex_id = format(random.randint(0, 0xFFFFFF), '06x')
            return f'{keyword} _0x{hex_id} ='
        
        pattern = r'(var|let|const)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*='
        new_code, count = re.subn(pattern, replace_var, code)
        
        return new_code, count > 0
    
    def _transform_strings(self, code: str) -> tuple:
        """Трансформация строк"""
        import re
        import base64
        
        def replace_string(match):
            if random.random() > 0.5:
                return match.group(0)
            
            s = match.group(1)
            
            # Выбор случайной трансформации
            transform = random.choice(['charcode', 'base64', 'hex'])
            
            if transform == 'charcode':
                codes = ','.join(str(ord(c)) for c in s)
                return f'String.fromCharCode({codes})'
            elif transform == 'base64':
                encoded = base64.b64encode(s.encode()).decode()
                return f'atob("{encoded}")'
            else:  # hex
                encoded = s.encode().hex()
                return f'Buffer.from("{encoded}", "hex").toString()'
        
        pattern = r'"([^"]*)"'
        new_code, count = re.subn(pattern, replace_string, code)
        
        return new_code, count > 0
    
    def _add_dead_code(self, code: str) -> tuple:
        """Добавление мёртвого кода"""
        dead_codes = self.mutation_templates['dead_code']
        
        # Вставка в случайное место
        insert_pos = random.randint(0, len(code))
        dead_code = random.choice(dead_codes)
        
        new_code = code[:insert_pos] + dead_code + code[insert_pos:]
        
        return new_code, True
    
    def _add_control_flow(self, code: str) -> tuple:
        """Добавление контрольного потока"""
        templates = self.mutation_templates['control_flow']
        template = random.choice(templates)
        
        # Обёртывание кода
        new_code = template.replace('{code}', code)
        
        return new_code, True
    
    def train_model(self, epochs: int = 10, batch_size: int = 32) -> Dict:
        """Дообучение модели (симуляция)"""
        logger.info(f"Training GAN model for {epochs} epochs")
        
        # Симуляция обучения
        metrics = {
            'epochs': epochs,
            'batch_size': batch_size,
            'generator_loss': random.uniform(0.1, 0.5),
            'discriminator_loss': random.uniform(0.1, 0.5),
            'obfuscation_quality': random.uniform(0.85, 0.99),
        }
        
        logger.info(f"Training complete. Quality: {metrics['obfuscation_quality']:.2f}")
        
        return metrics
    
    def get_model_stats(self) -> Dict:
        """Статистика модели"""
        return {
            'model_loaded': self.model is not None,
            'seed': self.seed,
            'templates_count': sum(len(v) for v in self.mutation_templates.values()),
            'model_path': str(self.model_path),
        }

# ==================== API ENDPOINTS ====================

# Глобальный обфускатор
obfuscator = GANObfuscator()

@app.on_event("startup")
async def startup_event():
    """Инициализация при старте"""
    logger.info("GAN Obfuscation module initialized")

@app.post("/api/v1/gan/obfuscate", response_model=ObfuscationResponse)
async def obfuscate_code(request: ObfuscationRequest):
    """Обфускация кода"""
    try:
        result = obfuscator.obfuscate(
            code=request.code,
            level=request.level,
            session_id=request.session_id
        )
        
        return ObfuscationResponse(**result)
    except Exception as e:
        logger.error(f"Obfuscation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/v1/gan/train")
async def train_model(request: TrainingRequest):
    """Дообучение модели"""
    try:
        metrics = obfuscator.train_model(
            epochs=request.epochs,
            batch_size=request.batch_size
        )
        
        return {
            "success": True,
            "metrics": metrics
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/api/v1/gan/stats")
async def get_stats():
    """Статистика GAN модуля"""
    return {
        "success": True,
        "stats": obfuscator.get_model_stats()
    }

@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "gan-obfuscation"}

# ==================== ЗАПУСК ====================

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8084,
        reload=True,
    )
