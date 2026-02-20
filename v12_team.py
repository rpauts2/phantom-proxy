#!/usr/bin/env python3
"""
PhantomProxy v12.3 PRO++ — Team Management Module
Управление командой, ролями и задачами

© 2026 PhantomSec Labs. All rights reserved.
"""

import sqlite3
import hashlib
import secrets
from datetime import datetime
from pathlib import Path

DB_PATH = Path(__file__).parent / 'phantom.db'

# === ROLES & PERMISSIONS ===

ROLES = {
    'admin': {
        'description': 'Full access to all features',
        'permissions': [
            'users.manage', 'users.view',
            'campaigns.manage', 'campaigns.view',
            'clients.manage', 'clients.view',
            'billing.manage', 'billing.view',
            'reports.generate', 'reports.view',
            'compliance.view', 'compliance.manage',
            'settings.manage'
        ]
    },
    'operator': {
        'description': 'Can run campaigns and view reports',
        'permissions': [
            'campaigns.manage', 'campaigns.view',
            'clients.view',
            'reports.generate', 'reports.view',
            'billing.view'
        ]
    },
    'viewer': {
        'description': 'Read-only access',
        'permissions': [
            'campaigns.view',
            'clients.view',
            'reports.view',
            'billing.view'
        ]
    },
    'client': {
        'description': 'Client portal access only',
        'permissions': [
            'portal.view',
            'reports.download'
        ]
    }
}

class TeamManager:
    """Управление командой и ролями"""
    
    def __init__(self, db_path=DB_PATH):
        self.db_path = db_path
        self.init_tables()
    
    def init_tables(self):
        """Инициализация таблиц"""
        conn = sqlite3.connect(self.db_path)
        c = conn.cursor()
        
        # Задачи
        c.execute('''CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY,
            title TEXT, description TEXT,
            assigned_to INTEGER, created_by INTEGER,
            campaign_id INTEGER, status TEXT,
            priority TEXT, due_date TEXT,
            created_at TEXT, completed_at TEXT
        )''')
        
        # Сессии пользователей (для audit)
        c.execute('''CREATE TABLE IF NOT EXISTS user_sessions (
            id INTEGER PRIMARY KEY,
            user_id INTEGER, token TEXT,
            ip_address TEXT, user_agent TEXT,
            created_at TEXT, expires_at TEXT,
            last_active TEXT
        )''')
        
        # Activity log
        c.execute('''CREATE TABLE IF NOT EXISTS activity_log (
            id INTEGER PRIMARY KEY,
            user_id INTEGER, action TEXT,
            resource_type TEXT, resource_id INTEGER,
            details TEXT, ip_address TEXT,
            created_at TEXT
        )''')
        
        conn.commit()
        conn.close()
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def create_user(self, username, password, role='operator', email=''):
        """Создание пользователя"""
        conn = self.get_db()
        c = conn.cursor()
        
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        api_key = secrets.token_urlsafe(32)
        
        try:
            c.execute('''INSERT INTO users 
                (username, password_hash, role, api_key, email, created_at)
                VALUES (?, ?, ?, ?, ?, ?)''',
                (username, password_hash, role, api_key, email, datetime.now().isoformat()))
            
            user_id = c.lastrowid
            conn.commit()
            
            self.log_activity(user_id, 'user_created', 'user', user_id, {'role': role})
            
            return {'success': True, 'user_id': user_id, 'api_key': api_key}
        except sqlite3.IntegrityError:
            return {'success': False, 'error': 'Username already exists'}
        finally:
            conn.close()
    
    def authenticate(self, username, password):
        """Аутентификация пользователя"""
        conn = self.get_db()
        c = conn.cursor()
        
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        
        c.execute('SELECT * FROM users WHERE username=? AND password_hash=?',
                 (username, password_hash))
        user = c.fetchone()
        
        if user:
            user_dict = dict(user)
            del user_dict['password_hash']  # Не возвращаем hash
            
            # Создаём сессию
            token = secrets.token_urlsafe(64)
            expires_at = datetime.now() + timedelta(hours=24)
            
            c.execute('''INSERT INTO user_sessions 
                (user_id, token, expires_at, created_at)
                VALUES (?, ?, ?, ?)''',
                (user['id'], token, expires_at.isoformat(), datetime.now().isoformat()))
            conn.commit()
            
            user_dict['token'] = token
            user_dict['expires_at'] = expires_at.isoformat()
            
            self.log_activity(user['id'], 'login', 'session', None, {})
            
            return {'success': True, 'user': user_dict}
        else:
            return {'success': False, 'error': 'Invalid credentials'}
        
        conn.close()
    
    def check_permission(self, user_id, permission):
        """Проверка прав доступа"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT role FROM users WHERE id=?', (user_id,))
        user = c.fetchone()
        
        if not user:
            return False
        
        role = user['role']
        role_permissions = ROLES.get(role, {}).get('permissions', [])
        
        conn.close()
        
        return permission in role_permissions
    
    def assign_task(self, title, assigned_to, created_by, campaign_id=None, 
                    description='', priority='medium', due_date=None):
        """Назначение задачи"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO tasks 
            (title, description, assigned_to, created_by, campaign_id, 
             status, priority, due_date, created_at)
            VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?)''',
            (title, description, assigned_to, created_by, campaign_id, 
             priority, due_date, datetime.now().isoformat()))
        
        task_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self.log_activity(created_by, 'task_created', 'task', task_id, 
                         {'assigned_to': assigned_to})
        
        return {'success': True, 'task_id': task_id}
    
    def get_user_tasks(self, user_id, status=None):
        """Получение задач пользователя"""
        conn = self.get_db()
        c = conn.cursor()
        
        if status:
            c.execute('''SELECT * FROM tasks WHERE assigned_to=? AND status=? 
                        ORDER BY created_at DESC''', (user_id, status))
        else:
            c.execute('''SELECT * FROM tasks WHERE assigned_to=? 
                        ORDER BY created_at DESC''', (user_id,))
        
        tasks = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return tasks
    
    def update_task_status(self, task_id, status, user_id):
        """Обновление статуса задачи"""
        conn = self.get_db()
        c = conn.cursor()
        
        completed_at = datetime.now().isoformat() if status == 'completed' else None
        
        c.execute('UPDATE tasks SET status=?, completed_at=? WHERE id=?',
                 (status, completed_at, task_id))
        conn.commit()
        conn.close()
        
        self.log_activity(user_id, 'task_updated', 'task', task_id, {'status': status})
        
        return {'success': True}
    
    def log_activity(self, user_id, action, resource_type, resource_id, details):
        """Логирование активности"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO activity_log 
            (user_id, action, resource_type, resource_id, details, created_at)
            VALUES (?, ?, ?, ?, ?, ?)''',
            (user_id, action, resource_type, resource_id, json.dumps(details), 
             datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
    
    def get_activity_log(self, user_id=None, limit=100):
        """Получение лога активности"""
        conn = self.get_db()
        c = conn.cursor()
        
        if user_id:
            c.execute('''SELECT * FROM activity_log WHERE user_id=? 
                        ORDER BY created_at DESC LIMIT ?''', (user_id, limit))
        else:
            c.execute('''SELECT * FROM activity_log ORDER BY created_at DESC LIMIT ?''', 
                     (limit,))
        
        activities = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return activities
    
    def get_team_stats(self):
        """Статистика команды"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Количество пользователей по ролям
        c.execute('SELECT role, COUNT(*) as count FROM users GROUP BY role')
        users_by_role = {row['role']: row['count'] for row in c.fetchall()}
        
        # Активные задачи
        c.execute("SELECT status, COUNT(*) as count FROM tasks WHERE status != 'completed' GROUP BY status")
        tasks_by_status = {row['status']: row['count'] for row in c.fetchall()}
        
        # Всего задач
        c.execute('SELECT COUNT(*) FROM tasks')
        total_tasks = c.fetchone()[0]
        
        # Завершено задач
        c.execute("SELECT COUNT(*) FROM tasks WHERE status = 'completed'")
        completed_tasks = c.fetchone()[0]
        
        conn.close()
        
        return {
            'users_by_role': users_by_role,
            'tasks_by_status': tasks_by_status,
            'total_tasks': total_tasks,
            'completed_tasks': completed_tasks,
            'completion_rate': round(completed_tasks / max(1, total_tasks) * 100, 1)
        }

# Import for JSON
import json
from datetime import timedelta

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.3 PRO++ — Team Management Module")
    print("="*60)
    
    manager = TeamManager()
    print("✅ Team Manager initialized")
    
    # Test create user
    print("\n👥 Creating test users...")
    result = manager.create_user('operator1', 'password123', role='operator', email='op1@test.com')
    if result['success']:
        print(f"   ✅ User created: ID {result['user_id']}")
        print(f"   🔑 API Key: {result['api_key']}")
    else:
        print(f"   ⚠️  {result['error']}")
    
    # Test authenticate
    print("\n🔐 Testing authentication...")
    result = manager.authenticate('operator1', 'password123')
    if result['success']:
        print(f"   ✅ Authenticated: {result['user']['username']}")
        print(f"   🎫 Token: {result['user']['token'][:20]}...")
    else:
        print(f"   ❌ {result['error']}")
    
    # Test permissions
    print("\n🔑 Testing permissions...")
    permissions = ['campaigns.view', 'users.manage', 'billing.view']
    for perm in permissions:
        has_perm = manager.check_permission(2, perm)  # ID 2 = operator1
        icon = '✅' if has_perm else '❌'
        print(f"   {icon} {perm}")
    
    # Test task assignment
    print("\n📋 Creating test task...")
    result = manager.assign_task(
        title='Setup Q1 Campaign',
        assigned_to=2,
        created_by=1,
        description='Configure phishing campaign for Q1',
        priority='high',
        due_date='2026-03-01'
    )
    if result['success']:
        print(f"   ✅ Task created: ID {result['task_id']}")
    
    # Test team stats
    print("\n📊 Team Statistics:")
    stats = manager.get_team_stats()
    print(f"   Users by role: {stats['users_by_role']}")
    print(f"   Tasks: {stats['total_tasks']} total, {stats['completed_tasks']} completed")
    print(f"   Completion rate: {stats['completion_rate']}%")
    
    print("\n✅ All team management features ready!")
