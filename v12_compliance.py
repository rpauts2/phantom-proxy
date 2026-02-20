#!/usr/bin/env python3
"""
PhantomProxy v12.0 - RED TEAM PROFESSIONAL EDITION
Модуль: Scope Enforcement + Compliance Logging

Для легального использования в рамках Red Team engagements
Только для аккредитованных организаций с письменными разрешениями
"""

import os
import json
import sqlite3
import hashlib
import secrets
from datetime import datetime, timedelta
from pathlib import Path
from cryptography.fernet import Fernet

# === КОНФИГУРАЦИЯ ===
LOGS_PATH = Path(__file__).parent / 'compliance_logs'
DB_PATH = Path(__file__).parent / 'phantom.db'
SCOPE_CONFIG_PATH = Path(__file__).parent / 'scope_config.json'

# Создаём директорию
LOGS_PATH.mkdir(exist_ok=True)

# === ENCRYPTED LOGGING ===
class ComplianceLogger:
    """Зашифрованное логирование всех действий для аудита"""
    
    def __init__(self):
        self.key_file = LOGS_PATH / '.encryption_key'
        self.key = self._get_or_create_key()
        self.cipher = Fernet(self.key)
    
    def _get_or_create_key(self):
        """Получение или создание ключа шифрования"""
        if self.key_file.exists():
            with open(self.key_file, 'rb') as f:
                return f.read()
        else:
            key = Fernet.generate_key()
            with open(self.key_file, 'wb') as f:
                f.write(key)
            return key
    
    def log_action(self, user_id, action, details, ip_address=''):
        """Логирование действия"""
        log_entry = {
            'timestamp': datetime.now().isoformat(),
            'user_id': user_id,
            'action': action,
            'details': details,
            'ip_address': ip_address,
            'hash': ''
        }
        
        # Создаём hash для верификации
        log_entry['hash'] = hashlib.sha256(
            json.dumps(log_entry, sort_keys=True).encode()
        ).hexdigest()
        
        # Шифруем
        encrypted = self.cipher.encrypt(json.dumps(log_entry).encode())
        
        # Сохраняем
        log_file = LOGS_PATH / f"audit_{datetime.now().strftime('%Y%m%d')}.log"
        with open(log_file, 'ab') as f:
            f.write(encrypted + b'\n')
        
        # Также в БД
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute('''INSERT INTO audit_log 
            (user_id, action, details, ip_address, created_at) 
            VALUES (?, ?, ?, ?, ?)''',
            (user_id, action, json.dumps(details), ip_address, datetime.now().isoformat()))
        conn.commit()
        conn.close()
        
        return log_entry['hash']
    
    def read_logs(self, date=None):
        """Чтение логов (для аудита)"""
        if date is None:
            date = datetime.now().strftime('%Y%m%d')
        
        log_file = LOGS_PATH / f"audit_{date}.log"
        
        if not log_file.exists():
            return []
        
        entries = []
        with open(log_file, 'rb') as f:
            for line in f:
                try:
                    decrypted = self.cipher.decrypt(line.strip())
                    entry = json.loads(decrypted)
                    entries.append(entry)
                except:
                    continue
        
        return entries
    
    def verify_integrity(self, date=None):
        """Проверка целостности логов"""
        entries = self.read_logs(date)
        
        for entry in entries:
            stored_hash = entry.pop('hash', '')
            calculated_hash = hashlib.sha256(
                json.dumps(entry, sort_keys=True).encode()
            ).hexdigest()
            
            if stored_hash != calculated_hash:
                return False, f"Integrity violation at {entry.get('timestamp', 'unknown')}"
        
        return True, "All logs verified"

# === SCOPE ENFORCEMENT ===
class ScopeEnforcement:
    """Контроль соблюдения границ кампании (RoE)"""
    
    def __init__(self):
        self.scope_config = self._load_scope()
        self.logger = ComplianceLogger()
    
    def _load_scope(self):
        """Загрузка конфигурации Scope"""
        if SCOPE_CONFIG_PATH.exists():
            with open(SCOPE_CONFIG_PATH, 'r') as f:
                return json.load(f)
        else:
            # Default scope
            return {
                'allowed_domains': [],
                'blacklisted_domains': ['gov', 'mil', 'edu'],  # Пример
                'max_emails_per_campaign': 1000,
                'max_concurrent_campaigns': 5,
                'allowed_hours': {'start': 9, 'end': 18},
                'kill_switch_enabled': True,
                'auto_stop_on_real_data': True
            }
    
    def save_scope(self, scope_config):
        """Сохранение конфигурации Scope"""
        with open(SCOPE_CONFIG_PATH, 'w') as f:
            json.dump(scope_config, f, indent=2)
        self.scope_config = scope_config
    
    def check_domain_allowed(self, domain):
        """Проверка разрешён ли домен"""
        # Blacklist проверка
        for blacklisted in self.scope_config.get('blacklisted_domains', []):
            if blacklisted in domain.lower():
                return False, f"Domain {domain} is blacklisted"
        
        # Whitelist проверка (если указана)
        allowed = self.scope_config.get('allowed_domains', [])
        if allowed:
            if not any(allowed_domain in domain.lower() for allowed_domain in allowed):
                return False, f"Domain {domain} is not in allowed list"
        
        return True, "Domain allowed"
    
    def check_campaign_limit(self, campaign_id):
        """Проверка лимитов кампании"""
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        # Количество писем в кампании
        c.execute('SELECT COUNT(*) FROM sessions WHERE campaign_id=?', (campaign_id,))
        count = c.fetchone()[0]
        
        max_emails = self.scope_config.get('max_emails_per_campaign', 1000)
        
        conn.close()
        
        if count >= max_emails:
            return False, f"Campaign limit reached: {count}/{max_emails}"
        
        return True, f"Campaign within limits: {count}/{max_emails}"
    
    def check_time_allowed(self):
        """Проверка разрешённого времени"""
        allowed_hours = self.scope_config.get('allowed_hours', {'start': 9, 'end': 18})
        current_hour = datetime.now().hour
        
        if allowed_hours['start'] <= current_hour < allowed_hours['end']:
            return True, "Within allowed hours"
        else:
            return False, f"Outside allowed hours ({allowed_hours['start']}:00-{allowed_hours['end']}:00)"
    
    def detect_real_data_exfil(self, session_data):
        """Детекция попытки эксфильтрации реальных данных"""
        # Простые эвристики
        red_flags = []
        
        # Проверка на реальные пароли (не тестовые)
        password = session_data.get('password', '')
        if password and not any(test in password.lower() for test in ['test', 'demo', 'example']):
            # Пароль похож на реальный
            if len(password) >= 8 and any(c.isdigit() for c in password):
                red_flags.append('Real-looking password detected')
        
        # Проверка на корпоративные email
        email = session_data.get('email', '')
        if email and any(corp in email.lower() for corp in ['company', 'corp', 'enterprise']):
            red_flags.append('Corporate email detected')
        
        if self.scope_config.get('auto_stop_on_real_data', True) and red_flags:
            return False, red_flags
        
        return True, []
    
    def enforce(self, action, data):
        """Принудительная проверка перед действием"""
        violations = []
        
        # Проверка домена
        if 'domain' in data:
            allowed, msg = self.check_domain_allowed(data['domain'])
            if not allowed:
                violations.append(msg)
        
        # Проверка времени
        allowed, msg = self.check_time_allowed()
        if not allowed:
            violations.append(msg)
        
        # Проверка лимитов
        if 'campaign_id' in data:
            allowed, msg = self.check_campaign_limit(data['campaign_id'])
            if not allowed:
                violations.append(msg)
        
        # Проверка на реальные данные
        if action == 'capture_credentials':
            allowed, red_flags = self.detect_real_data_exfil(data)
            if not allowed:
                violations.extend(red_flags)
        
        # Логирование
        self.logger.log_action('system', f'scope_check_{action}', {
            'data': data,
            'violations': violations,
            'blocked': len(violations) > 0
        })
        
        if violations:
            return False, violations
        return True, []

# === KILL SWITCH ===
class KillSwitch:
    """Аварийная остановка кампании"""
    
    def __init__(self):
        self.active = False
        self.reason = None
        self.triggered_at = None
    
    def trigger(self, reason):
        """Активация kill switch"""
        self.active = True
        self.reason = reason
        self.triggered_at = datetime.now().isoformat()
        
        # Логирование
        logger = ComplianceLogger()
        logger.log_action('system', 'kill_switch_triggered', {
            'reason': reason,
            'triggered_at': self.triggered_at
        })
        
        print(f"🚨 KILL SWITCH ACTIVATED: {reason}")
        
        # Остановка кампаний
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        c.execute("UPDATE campaigns SET status='stopped_by_killswitch' WHERE status='active'")
        conn.commit()
        conn.close()
    
    def reset(self):
        """Сброс kill switch"""
        self.active = False
        self.reason = None
        self.triggered_at = None
        
        logger = ComplianceLogger()
        logger.log_action('system', 'kill_switch_reset', {})
        
        print("✅ Kill switch reset")
    
    def is_active(self):
        """Проверка статуса"""
        return self.active

# === PROJECT/CAMPAIGN MANAGER ===
class CampaignManager:
    """Управление кампаниями и проектами"""
    
    def __init__(self):
        self.logger = ComplianceLogger()
        self.scope = ScopeEnforcement()
    
    def create_project(self, client_name, roe_hash, start_date, end_date, responsible):
        """Создание проекта"""
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute('''INSERT INTO projects 
            (client_name, roe_hash, start_date, end_date, responsible, status, created_at)
            VALUES (?, ?, ?, ?, ?, 'active', ?)''',
            (client_name, roe_hash, start_date, end_date, responsible, datetime.now().isoformat()))
        
        project_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self.logger.log_action(responsible, 'project_created', {
            'project_id': project_id,
            'client_name': client_name,
            'roe_hash': roe_hash
        })
        
        return project_id
    
    def create_campaign(self, project_id, name, service, subdomains, responsible):
        """Создание кампании"""
        # Проверка scope
        allowed, violations = self.scope.enforce('create_campaign', {
            'project_id': project_id
        })
        
        if not allowed:
            return None, violations
        
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute('''INSERT INTO campaigns 
            (project_id, name, service, subdomains, status, created_by, created_at)
            VALUES (?, ?, ?, ?, 'active', ?, ?)''',
            (project_id, name, service, json.dumps(subdomains), responsible, datetime.now().isoformat()))
        
        campaign_id = c.lastrowid
        conn.commit()
        conn.close()
        
        self.logger.log_action(responsible, 'campaign_created', {
            'campaign_id': campaign_id,
            'project_id': project_id,
            'name': name
        })
        
        return campaign_id, []
    
    def get_campaign_status(self, campaign_id):
        """Статус кампании"""
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute('SELECT * FROM campaigns WHERE id=?', (campaign_id,))
        campaign = dict(c.fetchone())
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE campaign_id=?', (campaign_id,))
        sessions_count = c.fetchone()[0]
        
        conn.close()
        
        campaign['sessions_count'] = sessions_count
        
        return campaign
    
    def stop_campaign(self, campaign_id, reason, responsible):
        """Остановка кампании"""
        conn = sqlite3.connect(DB_PATH)
        c = conn.cursor()
        
        c.execute("UPDATE campaigns SET status=?, stopped_reason=?, stopped_at=? WHERE id=?",
                 ('stopped', reason, datetime.now().isoformat(), campaign_id))
        
        conn.commit()
        conn.close()
        
        self.logger.log_action(responsible, 'campaign_stopped', {
            'campaign_id': campaign_id,
            'reason': reason
        })

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.0 - Scope Enforcement + Compliance Logging")
    print("="*60)
    
    # Test logging
    logger = ComplianceLogger()
    print("✅ Compliance Logger initialized")
    
    hash = logger.log_action('admin', 'test_action', {'test': 'data'})
    print(f"📝 Test log entry created: {hash}")
    
    # Test scope
    scope = ScopeEnforcement()
    print("✅ Scope Enforcement initialized")
    
    allowed, msg = scope.check_domain_allowed('test.company.com')
    print(f"🔍 Domain check: {msg}")
    
    # Test kill switch
    kill = KillSwitch()
    print("✅ Kill Switch initialized")
    
    # Test campaign manager
    cm = CampaignManager()
    print("✅ Campaign Manager initialized")
    
    print("\n✅ All v12.0 compliance modules ready!")
