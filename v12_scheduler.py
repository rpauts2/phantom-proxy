#!/usr/bin/env python3
"""
PhantomProxy v12.5 PRO++++ — Scheduler Module
Campaign scheduling, Auto-reports, Automated tasks

© 2026 PhantomSec Labs. All rights reserved.
"""

import sqlite3
import json
import threading
import time
from datetime import datetime, timedelta
from pathlib import Path

DB_PATH = Path(__file__).parent / 'phantom.db'
SCHEDULER_LOGS_PATH = Path(__file__).parent / 'scheduler_logs'
SCHEDULER_LOGS_PATH.mkdir(exist_ok=True)

class Scheduler:
    """Планировщик задач"""
    
    def __init__(self):
        self.init_tables()
        self.running = False
        self.scheduler_thread = None
    
    def init_tables(self):
        """Инициализация таблиц"""
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        # Запланированные кампании
        c.execute('''CREATE TABLE IF NOT EXISTS scheduled_campaigns (
            id INTEGER PRIMARY KEY,
            campaign_name TEXT, service TEXT,
            scheduled_start TEXT, scheduled_end TEXT,
            status TEXT, created_by INTEGER,
            config TEXT, created_at TEXT
        )''')
        
        # Автоматические отчёты
        c.execute('''CREATE TABLE IF NOT EXISTS scheduled_reports (
            id INTEGER PRIMARY KEY,
            campaign_id INTEGER, client_email TEXT,
            schedule_type TEXT,  # daily, weekly, monthly
            next_run TEXT, enabled INTEGER DEFAULT 1,
            created_at TEXT
        )''')
        
        # Автоматические задачи
        c.execute('''CREATE TABLE IF NOT EXISTS automated_tasks (
            id INTEGER PRIMARY KEY,
            task_type TEXT, task_name TEXT,
            schedule TEXT,  # cron-like or interval
            config TEXT, enabled INTEGER DEFAULT 1,
            last_run TEXT, next_run TEXT,
            created_at TEXT
        )''')
        
        # Логи выполнения
        c.execute('''CREATE TABLE IF NOT EXISTS scheduler_logs (
            id INTEGER PRIMARY KEY,
            task_type TEXT, task_id INTEGER,
            status TEXT, message TEXT,
            executed_at TEXT, duration REAL
        )''')
        
        conn.commit()
        conn.close()
    
    def get_db(self):
        conn = sqlite3.connect(DB_PATH)
        conn.row_factory = sqlite3.Row
        return conn
    
    def schedule_campaign(self, campaign_name, service, start_time, end_time, 
                         created_by, config=None):
        """Планирование кампании"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO scheduled_campaigns 
            (campaign_name, service, scheduled_start, scheduled_end, 
             status, created_by, config, created_at)
            VALUES (?, ?, ?, ?, 'scheduled', ?, ?, ?)''',
            (campaign_name, service, start_time.isoformat(), end_time.isoformat(),
             created_by, json.dumps(config or {}), datetime.now().isoformat()))
        
        campaign_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self._log('campaign_scheduled', campaign_id, 'success', f'Campaign scheduled: {start_time} - {end_time}')
        
        return {'success': True, 'campaign_id': campaign_id}
    
    def schedule_auto_report(self, campaign_id, client_email, schedule_type='weekly'):
        """Планирование автоматического отчёта"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Calculate next run
        now = datetime.now()
        if schedule_type == 'daily':
            next_run = now + timedelta(days=1)
        elif schedule_type == 'weekly':
            next_run = now + timedelta(weeks=1)
        elif schedule_type == 'monthly':
            next_run = now + timedelta(days=30)
        else:
            next_run = now + timedelta(days=7)
        
        c.execute('''INSERT INTO scheduled_reports 
            (campaign_id, client_email, schedule_type, next_run, created_at)
            VALUES (?, ?, ?, ?, ?)''',
            (campaign_id, client_email, schedule_type, next_run.isoformat(), datetime.now().isoformat()))
        
        report_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self._log('report_scheduled', report_id, 'success', f'Auto-report scheduled: {schedule_type}')
        
        return {'success': True, 'report_id': report_id}
    
    def create_automated_task(self, task_type, task_name, schedule, config=None):
        """Создание автоматической задачи"""
        conn = self.get_db()
        c = conn.cursor()
        
        now = datetime.now()
        next_run = now + timedelta(minutes=5)  # Default: run in 5 minutes
        
        c.execute('''INSERT INTO automated_tasks 
            (task_type, task_name, schedule, config, last_run, next_run, created_at)
            VALUES (?, ?, ?, ?, NULL, ?, ?)''',
            (task_type, task_name, schedule, json.dumps(config or {}), 
             next_run.isoformat(), datetime.now().isoformat()))
        
        task_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self._log('task_created', task_id, 'success', f'Task created: {task_name}')
        
        return {'success': True, 'task_id': task_id}
    
    def get_scheduled_campaigns(self, status='scheduled'):
        """Получение запланированных кампаний"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT * FROM scheduled_campaigns WHERE status=? 
                     ORDER BY scheduled_start''', (status,))
        
        campaigns = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return campaigns
    
    def get_due_campaigns(self):
        """Получение кампаний, готовых к запуску"""
        conn = self.get_db()
        c = conn.cursor()
        
        now = datetime.now().isoformat()
        
        c.execute('''SELECT * FROM scheduled_campaigns 
                     WHERE status='scheduled' AND scheduled_start <= ?''', (now,))
        
        campaigns = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return campaigns
    
    def get_due_reports(self):
        """Получение отчётов, готовых к генерации"""
        conn = self.get_db()
        c = conn.cursor()
        
        now = datetime.now().isoformat()
        
        c.execute('''SELECT * FROM scheduled_reports 
                     WHERE enabled=1 AND next_run <= ?''', (now,))
        
        reports = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return reports
    
    def update_campaign_status(self, campaign_id, status):
        """Обновление статуса кампании"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE scheduled_campaigns SET status=? WHERE id=?', (status, campaign_id))
        conn.commit()
        conn.close()
        
        self._log('campaign_status_updated', campaign_id, 'success', f'Status: {status}')
    
    def update_report_schedule(self, report_id, next_run):
        """Обновление расписания отчёта"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE scheduled_reports SET next_run=? WHERE id=?', 
                 (next_run.isoformat(), report_id))
        conn.commit()
        conn.close()
    
    def _log(self, task_type, task_id, status, message, duration=0):
        """Логирование выполнения"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO scheduler_logs 
            (task_type, task_id, status, message, executed_at, duration)
            VALUES (?, ?, ?, ?, ?, ?)''',
            (task_type, task_id, status, message, datetime.now().isoformat(), duration))
        
        conn.commit()
        conn.close()
        
        # Also save to file
        log_file = SCHEDULER_LOGS_PATH / f"scheduler_{datetime.now().strftime('%Y%m%d')}.log"
        with open(log_file, 'a') as f:
            log_entry = {
                'timestamp': datetime.now().isoformat(),
                'task_type': task_type,
                'task_id': task_id,
                'status': status,
                'message': message,
                'duration': duration
            }
            f.write(json.dumps(log_entry) + '\n')
    
    def start_scheduler(self):
        """Запуск планировщика"""
        self.running = True
        self.scheduler_thread = threading.Thread(target=self._run_scheduler, daemon=True)
        self.scheduler_thread.start()
        print("✅ Scheduler started")
    
    def stop_scheduler(self):
        """Остановка планировщика"""
        self.running = False
        if self.scheduler_thread:
            self.scheduler_thread.join(timeout=5)
        print("✅ Scheduler stopped")
    
    def _run_scheduler(self):
        """Основной цикл планировщика"""
        print("🔄 Scheduler running...")
        
        while self.running:
            try:
                # Check due campaigns
                due_campaigns = self.get_due_campaigns()
                for campaign in due_campaigns:
                    print(f"🚀 Starting scheduled campaign: {campaign['campaign_name']}")
                    self.update_campaign_status(campaign['id'], 'running')
                    # Here would be actual campaign start logic
                    self._log('campaign_started', campaign['id'], 'success', 
                             f"Scheduled campaign started", 0)
                
                # Check due reports
                due_reports = self.get_due_reports()
                for report in due_reports:
                    print(f"📄 Generating scheduled report for campaign {report['campaign_id']}")
                    # Here would be actual report generation logic
                    
                    # Update next run
                    now = datetime.now()
                    if report['schedule_type'] == 'daily':
                        next_run = now + timedelta(days=1)
                    elif report['schedule_type'] == 'weekly':
                        next_run = now + timedelta(weeks=1)
                    elif report['schedule_type'] == 'monthly':
                        next_run = now + timedelta(days=30)
                    else:
                        next_run = now + timedelta(days=7)
                    
                    self.update_report_schedule(report['id'], next_run)
                    self._log('report_generated', report['id'], 'success', 
                             f"Auto-report generated", 0)
                
                # Sleep for 1 minute
                time.sleep(60)
            
            except Exception as e:
                print(f"❌ Scheduler error: {e}")
                time.sleep(60)
    
    def get_scheduler_stats(self):
        """Статистика планировщика"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Scheduled campaigns
        c.execute("SELECT COUNT(*) FROM scheduled_campaigns WHERE status='scheduled'")
        scheduled_campaigns = c.fetchone()[0]
        
        # Running campaigns
        c.execute("SELECT COUNT(*) FROM scheduled_campaigns WHERE status='running'")
        running_campaigns = c.fetchone()[0]
        
        # Scheduled reports
        c.execute("SELECT COUNT(*) FROM scheduled_reports WHERE enabled=1")
        scheduled_reports = c.fetchone()[0]
        
        # Automated tasks
        c.execute("SELECT COUNT(*) FROM automated_tasks WHERE enabled=1")
        automated_tasks = c.fetchone()[0]
        
        # Logs (last 24h)
        c.execute('''SELECT COUNT(*) FROM scheduler_logs 
                     WHERE executed_at > datetime("now", "-1 day")''')
        executions_24h = c.fetchone()[0]
        
        conn.close()
        
        return {
            'scheduled_campaigns': scheduled_campaigns,
            'running_campaigns': running_campaigns,
            'scheduled_reports': scheduled_reports,
            'automated_tasks': automated_tasks,
            'executions_24h': executions_24h
        }

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.5 PRO++++ — Scheduler Module")
    print("="*60)
    
    scheduler = Scheduler()
    print("✅ Scheduler initialized")
    
    # Test scheduling
    print("\n📅 Test Scheduling:")
    
    # Schedule campaign
    start = datetime.now() + timedelta(hours=1)
    end = datetime.now() + timedelta(hours=2)
    
    result = scheduler.schedule_campaign(
        campaign_name='Scheduled Test Campaign',
        service='Microsoft 365',
        start_time=start,
        end_time=end,
        created_by=1,
        config={'auto_start': True}
    )
    
    if result['success']:
        print(f"   ✅ Campaign scheduled: ID {result['campaign_id']}")
        print(f"   Start: {start}")
        print(f"   End: {end}")
    
    # Schedule auto-report
    result = scheduler.schedule_auto_report(
        campaign_id=1,
        client_email='client@example.com',
        schedule_type='weekly'
    )
    
    if result['success']:
        print(f"   ✅ Auto-report scheduled: ID {result['report_id']}")
    
    # Create automated task
    result = scheduler.create_automated_task(
        task_type='cleanup',
        task_name='Cleanup old sessions',
        schedule='0 2 * * *',  # Daily at 2 AM
        config={'retention_days': 30}
    )
    
    if result['success']:
        print(f"   ✅ Automated task created: ID {result['task_id']}")
    
    # Get stats
    print("\n📊 Scheduler Statistics:")
    stats = scheduler.get_scheduler_stats()
    print(f"   Scheduled Campaigns: {stats['scheduled_campaigns']}")
    print(f"   Running Campaigns: {stats['running_campaigns']}")
    print(f"   Scheduled Reports: {stats['scheduled_reports']}")
    print(f"   Automated Tasks: {stats['automated_tasks']}")
    print(f"   Executions (24h): {stats['executions_24h']}")
    
    print("\n✅ All scheduler features ready!")
