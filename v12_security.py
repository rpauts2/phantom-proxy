#!/usr/bin/env python3
"""
PhantomProxy v12.4 PRO+++ — Security Module
2FA TOTP, Session Management, Advanced Security

© 2026 PhantomSec Labs. All rights reserved.
"""

import sqlite3
import hashlib
import secrets
import base64
import hmac
import struct
import time
import json
from datetime import datetime, timedelta
from pathlib import Path

DB_PATH = Path(__file__).parent / 'phantom.db'

class TOTP:
    """Time-based One-Time Password (TOTP) implementation"""
    
    @staticmethod
    def generate_secret():
        """Генерация секрета"""
        return base64.b32encode(secrets.token_bytes(20)).decode('utf-8')
    
    @staticmethod
    def generate_totp(secret, time_step=30, digits=6):
        """Генерация TOTP кода"""
        try:
            # Текущий time step
            counter = int(time.time() // time_step)
            
            # Pack counter as 8-byte big-endian
            counter_bytes = struct.pack('>Q', counter)
            
            # HMAC-SHA1
            key = base64.b32decode(secret.upper() + '=' * (-len(secret) % 8))
            hmac_hash = hmac.new(key, counter_bytes, hashlib.sha1).digest()
            
            # Dynamic truncation
            offset = hmac_hash[-1] & 0x0F
            code = struct.unpack('>I', hmac_hash[offset:offset+4])[0]
            code &= 0x7FFFFFFF
            code %= 10 ** digits
            
            return str(code).zfill(digits)
        except Exception as e:
            return None
    
    @staticmethod
    def verify_totp(secret, user_code, window=1):
        """Проверка TOTP кода с окном"""
        current_time_step = int(time.time() // 30)
        
        for offset in range(-window, window + 1):
            expected_code = TOTP.generate_totp(secret, time_step=30)
            # Recalculate with offset
            adjusted_time = current_time_step + offset
            counter_bytes = struct.pack('>Q', adjusted_time)
            
            key = base64.b32decode(secret.upper() + '=' * (-len(secret) % 8))
            hmac_hash = hmac.new(key, counter_bytes, hashlib.sha1).digest()
            offset_byte = hmac_hash[-1] & 0x0F
            code = struct.unpack('>I', hmac_hash[offset_byte:offset_byte+4])[0]
            code &= 0x7FFFFFFF
            code %= 10 ** 6
            
            if str(code).zfill(6) == user_code:
                return True
        
        return False

class SecurityManager:
    """Менеджер безопасности"""
    
    def __init__(self, db_path=DB_PATH):
        self.db_path = db_path
        self.init_tables()
    
    def init_tables(self):
        """Инициализация таблиц"""
        conn = sqlite3.connect(self.db_path)
        c = conn.cursor()
        
        # 2FA секреты
        c.execute('''CREATE TABLE IF NOT EXISTS user_2fa (
            user_id INTEGER PRIMARY KEY,
            secret TEXT, enabled INTEGER DEFAULT 0,
            backup_codes TEXT, created_at TEXT
        )''')
        
        # Активные сессии
        c.execute('''CREATE TABLE IF NOT EXISTS active_sessions (
            id INTEGER PRIMARY KEY,
            user_id INTEGER, token TEXT UNIQUE,
            ip_address TEXT, user_agent TEXT,
            created_at TEXT, expires_at TEXT,
            last_activity TEXT, is_valid INTEGER DEFAULT 1
        )''')
        
        # Login attempts
        c.execute('''CREATE TABLE IF NOT EXISTS login_attempts (
            id INTEGER PRIMARY KEY,
            username TEXT, ip_address TEXT,
            success INTEGER, timestamp TEXT
        )''')
        
        conn.commit()
        conn.close()
    
    def get_db(self):
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn
    
    def enable_2fa(self, user_id):
        """Включение 2FA"""
        secret = TOTP.generate_secret()
        backup_codes = [secrets.token_urlsafe(8) for _ in range(10)]
        
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT OR REPLACE INTO user_2fa 
            (user_id, secret, enabled, backup_codes, created_at)
            VALUES (?, ?, 1, ?, ?)''',
            (user_id, secret, json.dumps(backup_codes), datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
        
        return {
            'success': True,
            'secret': secret,
            'backup_codes': backup_codes,
            'qr_provisioning_uri': f"otpauth://totp/PhantomSec:Labs?secret={secret}&issuer=PhantomSec%20Labs"
        }
    
    def disable_2fa(self, user_id):
        """Отключение 2FA"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE user_2fa SET enabled=0 WHERE user_id=?', (user_id,))
        conn.commit()
        conn.close()
        
        return {'success': True}
    
    def verify_2fa(self, user_id, code):
        """Проверка 2FA кода"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT secret, enabled FROM user_2fa WHERE user_id=?', (user_id,))
        row = c.fetchone()
        
        if not row or not row['enabled']:
            conn.close()
            return {'success': False, 'error': '2FA not enabled'}
        
        is_valid = TOTP.verify_totp(row['secret'], code)
        
        conn.close()
        
        if is_valid:
            return {'success': True}
        else:
            return {'success': False, 'error': 'Invalid code'}
    
    def verify_backup_code(self, user_id, code):
        """Проверка backup кода"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('SELECT backup_codes FROM user_2fa WHERE user_id=?', (user_id,))
        row = c.fetchone()
        
        if not row:
            conn.close()
            return {'success': False, 'error': '2FA not configured'}
        
        backup_codes = json.loads(row['backup_codes'])
        
        if code in backup_codes:
            # Remove used code
            backup_codes.remove(code)
            c.execute('UPDATE user_2fa SET backup_codes=? WHERE user_id=?',
                     (json.dumps(backup_codes), user_id))
            conn.commit()
            conn.close()
            return {'success': True}
        
        conn.close()
        return {'success': False, 'error': 'Invalid backup code'}
    
    def create_session(self, user_id, ip_address, user_agent, expires_hours=24):
        """Создание сессии"""
        token = secrets.token_urlsafe(64)
        expires_at = datetime.now() + timedelta(hours=expires_hours)
        
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO active_sessions 
            (user_id, token, ip_address, user_agent, created_at, expires_at, last_activity)
            VALUES (?, ?, ?, ?, ?, ?, ?)''',
            (user_id, token, ip_address, user_agent, datetime.now().isoformat(),
             expires_at.isoformat(), datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
        
        return {'success': True, 'token': token, 'expires_at': expires_at.isoformat()}
    
    def validate_session(self, token):
        """Проверка сессии"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT * FROM active_sessions 
            WHERE token=? AND is_valid=1 AND expires_at > ?''',
            (token, datetime.now().isoformat()))
        
        session = c.fetchone()
        
        if session:
            # Update last activity
            c.execute('UPDATE active_sessions SET last_activity=? WHERE token=?',
                     (datetime.now().isoformat(), token))
            conn.commit()
        
        conn.close()
        
        if session:
            return {'success': True, 'session': dict(session)}
        else:
            return {'success': False, 'error': 'Invalid or expired session'}
    
    def invalidate_session(self, token):
        """Завершение сессии"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE active_sessions SET is_valid=0 WHERE token=?', (token,))
        conn.commit()
        conn.close()
        
        return {'success': True}
    
    def invalidate_all_sessions(self, user_id):
        """Завершение всех сессий пользователя"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('UPDATE active_sessions SET is_valid=0 WHERE user_id=?', (user_id,))
        conn.commit()
        conn.close()
        
        return {'success': True}
    
    def log_login_attempt(self, username, ip_address, success):
        """Логирование попытки входа"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''INSERT INTO login_attempts 
            (username, ip_address, success, timestamp)
            VALUES (?, ?, ?, ?)''',
            (username, ip_address, 1 if success else 0, datetime.now().isoformat()))
        
        conn.commit()
        conn.close()
    
    def check_brute_force(self, username, ip_address, max_attempts=5, window_minutes=15):
        """Проверка на brute force"""
        conn = self.get_db()
        c = conn.cursor()
        
        window_start = (datetime.now() - timedelta(minutes=window_minutes)).isoformat()
        
        c.execute('''SELECT COUNT(*) FROM login_attempts 
            WHERE (username=? OR ip_address=?) AND success=0 AND timestamp > ?''',
            (username, ip_address, window_start))
        
        failed_attempts = c.fetchone()[0]
        conn.close()
        
        if failed_attempts >= max_attempts:
            return {'blocked': True, 'attempts': failed_attempts}
        else:
            return {'blocked': False, 'attempts': failed_attempts}
    
    def get_user_sessions(self, user_id):
        """Получение активных сессий пользователя"""
        conn = self.get_db()
        c = conn.cursor()
        
        c.execute('''SELECT * FROM active_sessions 
            WHERE user_id=? AND is_valid=1 AND expires_at > ?
            ORDER BY created_at DESC''',
            (user_id, datetime.now().isoformat()))
        
        sessions = [dict(row) for row in c.fetchall()]
        conn.close()
        
        return sessions
    
    def get_security_stats(self):
        """Статистика безопасности"""
        conn = self.get_db()
        c = conn.cursor()
        
        # Active sessions
        c.execute('SELECT COUNT(*) FROM active_sessions WHERE is_valid=1')
        active_sessions = c.fetchone()[0]
        
        # 2FA enabled users
        c.execute('SELECT COUNT(*) FROM user_2fa WHERE enabled=1')
        twofa_enabled = c.fetchone()[0]
        
        # Failed logins (last 24h)
        c.execute('''SELECT COUNT(*) FROM login_attempts 
            WHERE success=0 AND timestamp > datetime("now", "-1 day")''')
        failed_logins = c.fetchone()[0]
        
        conn.close()
        
        return {
            'active_sessions': active_sessions,
            'twofa_enabled_users': twofa_enabled,
            'failed_logins_24h': failed_logins
        }

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.4 PRO+++ — Security Module")
    print("="*60)
    
    security = SecurityManager()
    print("✅ Security Manager initialized")
    
    # Test TOTP
    print("\n🔐 TOTP Test:")
    secret = TOTP.generate_secret()
    print(f"   Generated Secret: {secret}")
    
    code = TOTP.generate_totp(secret)
    print(f"   Current TOTP Code: {code}")
    
    is_valid = TOTP.verify_totp(secret, code)
    print(f"   Verification: {'✅ Valid' if is_valid else '❌ Invalid'}")
    
    # Test 2FA setup
    print("\n📱 2FA Setup Test:")
    result = security.enable_2fa(user_id=1)
    if result['success']:
        print(f"   ✅ 2FA enabled for user 1")
        print(f"   Secret: {result['secret']}")
        print(f"   Backup Codes: {len(result['backup_codes'])} codes")
    
    # Test session management
    print("\n🎫 Session Management Test:")
    result = security.create_session(
        user_id=1,
        ip_address='192.168.1.100',
        user_agent='Mozilla/5.0',
        expires_hours=24
    )
    if result['success']:
        print(f"   ✅ Session created")
        print(f"   Token: {result['token'][:30]}...")
        
        # Validate
        validation = security.validate_session(result['token'])
        print(f"   Validation: {'✅ Valid' if validation['success'] else '❌ Invalid'}")
    
    # Test security stats
    print("\n📊 Security Statistics:")
    stats = security.get_security_stats()
    print(f"   Active Sessions: {stats['active_sessions']}")
    print(f"   2FA Enabled: {stats['twofa_enabled_users']}")
    print(f"   Failed Logins (24h): {stats['failed_logins_24h']}")
    
    print("\n✅ All security features ready!")
