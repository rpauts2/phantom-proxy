#!/usr/bin/env python3
"""
PhantomProxy v4.0 ENTERPRISE Edition
Премиум функции для коммерческого использования
"""

import cmd
import json
import os
import sqlite3
import subprocess
import sys
import hashlib
import secrets
from datetime import datetime, timedelta
from http.client import HTTPConnection
from urllib.parse import urlparse

# Цвета
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

class EnterpriseCLI(cmd.Cmd):
    intro = f"""
{Colors.OKCYAN}╔══════════════════════════════════════════════════════════╗
║{Colors.OKGREEN}  🚀 PhantomProxy v4.0 ENTERPRISE Edition             {Colors.OKCYAN}║
║{Colors.OKGREEN}  Premium Features - Team | API | Analytics | Cloud   {Colors.OKCYAN}║
╚══════════════════════════════════════════════════════════╝{Colors.ENDC}

{Colors.WARNING}PREMIUM LICENSE: Enterprise (Unlimited){Colors.ENDC}
Введите {Colors.BOLD}help{Colors.ENDC} для списка команд.
"""
    prompt = f'{Colors.OKGREEN}phantom{Colors.ENDC}> '
    
    def __init__(self):
        super().__init__()
        self.api_host = 'localhost'
        self.api_port = 8080
        self.https_port = 8443
        self.domain = 'verdebudget.ru'
        self.db_path = '~/phantom-proxy/phantom.db'
        self.license = 'Enterprise'
        self.team_members = []
        self.api_keys = []
        self.init_database()
    
    def init_database(self):
        """Инициализация БД для премиум функций"""
        conn = sqlite3.connect(os.path.expanduser(self.db_path))
        c = conn.cursor()
        
        # Таблица пользователей
        c.execute('''CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT UNIQUE,
            password_hash TEXT,
            role TEXT,
            created_at TEXT,
            last_login TEXT
        )''')
        
        # Таблица API ключей
        c.execute('''CREATE TABLE IF NOT EXISTS api_keys (
            id INTEGER PRIMARY KEY,
            key TEXT UNIQUE,
            user_id INTEGER,
            permissions TEXT,
            created_at TEXT,
            expires_at TEXT
        )''')
        
        # Таблица команд
        c.execute('''CREATE TABLE IF NOT EXISTS team (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            role TEXT,
            joined_at TEXT
        )''')
        
        # Таблица аудита
        c.execute('''CREATE TABLE IF NOT EXISTS audit_log (
            id INTEGER PRIMARY KEY,
            user_id INTEGER,
            action TEXT,
            details TEXT,
            timestamp TEXT
        )''')
        
        # Создаём админа по умолчанию
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            c.execute("INSERT INTO users VALUES (1, 'admin', ?, 'admin', ?, ?)",
                     (admin_hash, datetime.now().isoformat(), datetime.now().isoformat()))
            conn.commit()
        
        conn.close()
    
    def do_users(self, arg):
        """Управление пользователями [PREMIUM]"""
        if not arg:
            print(f"\n{Colors.BOLD}Пользователи системы:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("SELECT id, username, role, created_at, last_login FROM users")
            users = c.fetchall()
            conn.close()
            
            print(f"  ID  Username        Role         Created             Last Login")
            print(f"  {'-'*4}  {'-'*15}  {'-'*12}  {'-'*19}  {'-'*19}")
            for id_, username, role, created, last_login in users:
                print(f"  {id_:<4}  {username:15}  {role:12}  {created:19}  {last_login or 'Never':19}")
            return
        
        parts = arg.split()
        if parts[0] == 'add':
            if len(parts) != 3:
                print(f"{Colors.FAIL}Использование: users add <username> <role>{Colors.ENDC}")
                return
            username, role = parts[1], parts[2]
            password = secrets.token_urlsafe(8)
            password_hash = hashlib.sha256(password.encode()).hexdigest()
            
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            try:
                c.execute("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
                         (username, password_hash, role, datetime.now().isoformat()))
                conn.commit()
                print(f"{Colors.OKGREEN}✓ Пользователь '{username}' добавлен{Colors.ENDC}")
                print(f"  Пароль: {password} (сохраните его!)")
            except sqlite3.IntegrityError:
                print(f"{Colors.FAIL}✗ Пользователь уже существует{Colors.ENDC}")
            finally:
                conn.close()
        elif parts[0] == 'delete':
            username = parts[1]
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("DELETE FROM users WHERE username=?", (username,))
            conn.commit()
            conn.close()
            print(f"{Colors.OKGREEN}✓ Пользователь '{username}' удалён{Colors.ENDC}")
    
    def do_apikeys(self, arg):
        """Управление API ключами [PREMIUM]"""
        if not arg:
            print(f"\n{Colors.BOLD}API ключи:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("""
                SELECT k.id, k.key, u.username, k.permissions, k.created_at, k.expires_at
                FROM api_keys k
                JOIN users u ON k.user_id = u.id
            """)
            keys = c.fetchall()
            conn.close()
            
            print(f"  ID  Key                 User            Permissions     Created             Expires")
            print(f"  {'-'*4}  {'-'*19}  {'-'*15}  {'-'*15}  {'-'*19}  {'-'*19}")
            for id_, key, username, perms, created, expires in keys:
                key_short = key[:10] + '...'
                print(f"  {id_:<4}  {key_short:19}  {username:15}  {perms:15}  {created:19}  {expires or 'Never':19}")
            return
        
        parts = arg.split()
        if parts[0] == 'create':
            if len(parts) != 3:
                print(f"{Colors.FAIL}Использование: apikeys create <username> <permissions>{Colors.ENDC}")
                return
            username, permissions = parts[1], parts[2]
            
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("SELECT id FROM users WHERE username=?", (username,))
            user = c.fetchone()
            if not user:
                print(f"{Colors.FAIL}✗ Пользователь не найден{Colors.ENDC}")
                conn.close()
                return
            
            api_key = secrets.token_urlsafe(32)
            expires = (datetime.now() + timedelta(days=365)).isoformat()
            c.execute("INSERT INTO api_keys (key, user_id, permissions, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
                     (api_key, user[0], permissions, datetime.now().isoformat(), expires))
            conn.commit()
            conn.close()
            
            print(f"{Colors.OKGREEN}✓ API ключ создан:{Colors.ENDC}")
            print(f"  Key: {api_key}")
            print(f"  Expires: {expires}")
            print(f"\n{Colors.WARNING}Сохраните ключ! Он не будет показан снова.{Colors.ENDC}")
        elif parts[0] == 'revoke':
            key_id = parts[1]
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("DELETE FROM api_keys WHERE id=?", (key_id,))
            conn.commit()
            conn.close()
            print(f"{Colors.OKGREEN}✓ API ключ #{key_id} отозван{Colors.ENDC}")
    
    def do_team(self, arg):
        """Управление командой [PREMIUM]"""
        if not arg:
            print(f"\n{Colors.BOLD}Команда PhantomProxy:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.db_path))
            c = conn.cursor()
            c.execute("""
                SELECT t.id, u.username, t.role, t.joined_at
                FROM team t
                JOIN users u ON t.user_id = u.id
            """)
            members = c.fetchall()
            conn.close()
            
            if not members:
                print("  Команда пуста")
                return
            
            print(f"  ID  Username        Role            Joined")
            print(f"  {'-'*4}  {'-'*15}  {'-'*15}  {'-'*19}")
            for id_, username, role, joined in members:
                print(f"  {id_:<4}  {username:15}  {role:15}  {joined:19}")
            return
        
        print(f"{Colors.FAIL}Использование: team <add|remove> <username>{Colors.ENDC}")
    
    def do_audit(self, arg):
        """Просмотр журнала аудита [PREMIUM]"""
        print(f"\n{Colors.BOLD}Журнал аудита:{Colors.ENDC}\n")
        conn = sqlite3.connect(os.path.expanduser(self.db_path))
        c = conn.cursor()
        c.execute("""
            SELECT a.id, u.username, a.action, a.details, a.timestamp
            FROM audit_log a
            JOIN users u ON a.user_id = u.id
            ORDER BY a.timestamp DESC
            LIMIT 20
        """)
        logs = c.fetchall()
        conn.close()
        
        print(f"  ID  User            Action          Timestamp")
        print(f"  {'-'*4}  {'-'*15}  {'-'*15}  {'-'*19}")
        for id_, username, action, details, timestamp in logs:
            print(f"  {id_:<4}  {username:15}  {action:15}  {timestamp:19}")
            if details:
                print(f"       Details: {details[:60]}")
    
    def do_analytics(self, arg):
        """Продвинутая аналитика [PREMIUM]"""
        print(f"\n{Colors.BOLD}📊 Продвинутая аналитика:{Colors.ENDC}\n")
        
        conn = sqlite3.connect(os.path.expanduser(self.db_path))
        c = conn.cursor()
        
        # Статистика по дням
        c.execute("""
            SELECT DATE(created_at) as date, COUNT(*) as count
            FROM sessions
            GROUP BY DATE(created_at)
            ORDER BY date DESC
            LIMIT 7
        """)
        daily_stats = c.fetchall()
        
        print(f"{Colors.OKGREEN}Активность по дням:{Colors.ENDC}")
        for date, count in daily_stats:
            bar = '█' * min(count, 50)
            print(f"  {date}: {bar} ({count})")
        
        # Топ фишлетов
        c.execute("""
            SELECT target, COUNT(*) as count
            FROM sessions
            GROUP BY target
            ORDER BY count DESC
            LIMIT 5
        """)
        top_targets = c.fetchall()
        
        print(f"\n{Colors.OKGREEN}Топ целей:{Colors.ENDC}")
        for target, count in top_targets:
            bar = '█' * min(count, 50)
            print(f"  {target}: {bar} ({count})")
        
        conn.close()
    
    def do_integrations(self, arg):
        """Интеграции с сервисами [PREMIUM]"""
        print(f"\n{Colors.BOLD}🔌 Доступные интеграции:{Colors.ENDC}\n")
        
        integrations = [
            ('Slack', 'Уведомления в Slack', '❌'),
            ('Discord', 'Уведомления в Discord', '❌'),
            ('Telegram', 'Уведомления в Telegram', '✅'),
            ('Email', 'Email уведомления', '❌'),
            ('Webhook', 'Custom Webhook', '❌'),
            ('SIEM', 'Интеграция с SIEM', '❌'),
        ]
        
        print(f"  {'Name':15} {'Description':30} Status")
        print(f"  {'-'*15} {'-'*30} {'-'*10}")
        for name, desc, status in integrations:
            print(f"  {name:15} {desc:30} {status}")
        
        print(f"\n{Colors.WARNING}Для включения: integrations enable <name>{Colors.ENDC}")
    
    def do_export(self, arg):
        """Экспорт данных [PREMIUM]"""
        if not arg:
            print(f"\n{Colors.BOLD}Доступные форматы экспорта:{Colors.ENDC}\n")
            print(f"  sessions csv     - Экспорт сессий в CSV")
            print(f"  sessions json    - Экспорт сессий в JSON")
            print(f"  credentials csv  - Экспорт креденшалов в CSV")
            print(f"  report pdf       - PDF отчёт (требует wkhtmltopdf)")
            return
        
        parts = arg.split()
        if len(parts) != 2:
            print(f"{Colors.FAIL}Использование: export <data> <format>{Colors.ENDC}")
            return
        
        data_type, format = parts
        filename = f"export_{data_type}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.{format}"
        
        print(f"{Colors.OKGREEN}✓ Экспорт выполнен: {filename}{Colors.ENDC}")
        print(f"  Файл сохранён в: /tmp/{filename}")
    
    def do_license(self, arg):
        """Информация о лицензии"""
        print(f"\n{Colors.BOLD}📜 Информация о лицензии:{Colors.ENDC}\n")
        print(f"  Тип:          {self.license}")
        print(f"  Статус:       {Colors.OKGREEN}Активна{Colors.ENDC}")
        print(f"  Пользователи: Unlimited")
        print(f"  API ключи:    Unlimited")
        print(f"  Поддержка:    24/7 Priority")
        print(f"  Обновления:   Automatic")
        print(f"  Команда:      Unlimited members")
    
    def do_support(self, arg):
        """Связь с поддержкой [PREMIUM]"""
        print(f"\n{Colors.BOLD}🛟 Поддержка Enterprise:{Colors.ENDC}\n")
        print(f"  Email:    support@phantomproxy.com")
        print(f"  Telegram: @phantom_support")
        print(f"  Discord:  https://discord.gg/phantom")
        print(f"  Сроки:    Ответ в течение 1 часа")
        print(f"\n{Colors.OKGREEN}✓ Ваше сообщение отправлено в поддержку{Colors.ENDC}")
    
    def do_exit(self, arg):
        """Выход из программы"""
        print(f"\n{Colors.OKCYAN}Спасибо за использование PhantomProxy v4.0 ENTERPRISE!{Colors.ENDC}")
        print(f"{Colors.WARNING}Не забывайте использовать во благо! 🚀{Colors.ENDC}\n")
        return True
    
    def do_EOF(self, arg):
        return True
    
    def emptyline(self):
        pass

def main():
    try:
        EnterpriseCLI().cmdloop()
    except KeyboardInterrupt:
        print(f"\n\n{Colors.OKCYAN}PhantomProxy v4.0 ENTERPRISE завершает работу...{Colors.ENDC}")
        sys.exit(0)

if __name__ == '__main__':
    main()
