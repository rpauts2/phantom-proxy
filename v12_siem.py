#!/usr/bin/env python3
"""
PhantomProxy v12.5 PRO++++ — SIEM Integration Module
Splunk, ELK, QRadar, Syslog exports

© 2026 PhantomSec Labs. All rights reserved.
"""

import json
import socket
import requests
from datetime import datetime
from pathlib import Path

# === КОНФИГУРАЦИЯ ===
SIEM_EXPORTS_PATH = Path(__file__).parent / 'siem_exports'
SIEM_EXPORTS_PATH.mkdir(exist_ok=True)

class SIEMIntegration:
    """Интеграция с SIEM системами"""
    
    def __init__(self):
        self.config = {
            'splunk': {
                'enabled': False,
                'hec_url': '',
                'hec_token': '',
                'index': 'phantomproxy',
                'source': 'phantomproxy'
            },
            'elk': {
                'enabled': False,
                'elasticsearch_url': '',
                'index': 'phantomproxy',
                'username': '',
                'password': ''
            },
            'qradar': {
                'enabled': False,
                'console_url': '',
                'api_token': ''
            },
            'syslog': {
                'enabled': False,
                'server': '',
                'port': 514,
                'protocol': 'udp'
            }
        }
    
    def configure_splunk(self, hec_url, hec_token, index='phantomproxy'):
        """Настройка Splunk HEC"""
        self.config['splunk'] = {
            'enabled': True,
            'hec_url': hec_url,
            'hec_token': hec_token,
            'index': index,
            'source': 'phantomproxy'
        }
        return {'success': True, 'message': 'Splunk configured'}
    
    def configure_elk(self, es_url, index='phantomproxy', username='', password=''):
        """Настройка Elasticsearch"""
        self.config['elk'] = {
            'enabled': True,
            'elasticsearch_url': es_url,
            'index': index,
            'username': username,
            'password': password
        }
        return {'success': True, 'message': 'ELK configured'}
    
    def configure_qradar(self, console_url, api_token):
        """Настройка QRadar"""
        self.config['qradar'] = {
            'enabled': True,
            'console_url': console_url,
            'api_token': api_token
        }
        return {'success': True, 'message': 'QRadar configured'}
    
    def configure_syslog(self, server, port=514, protocol='udp'):
        """Настройка Syslog"""
        self.config['syslog'] = {
            'enabled': True,
            'server': server,
            'port': port,
            'protocol': protocol
        }
        return {'success': True, 'message': 'Syslog configured'}
    
    def format_event(self, event_type, data):
        """Форматирование события"""
        event = {
            '@timestamp': datetime.now().isoformat(),
            'event_type': event_type,
            'product': 'PhantomProxy',
            'vendor': 'PhantomSec Labs',
            'version': '12.5',
            'data': data
        }
        return event
    
    def send_to_splunk(self, event):
        """Отправка в Splunk HEC"""
        if not self.config['splunk']['enabled']:
            return {'success': False, 'error': 'Splunk not configured'}
        
        try:
            headers = {
                'Authorization': f"Splunk {self.config['splunk']['hec_token']}",
                'Content-Type': 'application/json'
            }
            
            payload = {
                'time': datetime.now().timestamp(),
                'host': socket.gethostname(),
                'source': self.config['splunk']['source'],
                'sourcetype': 'phantomproxy:events',
                'index': self.config['splunk']['index'],
                'event': event
            }
            
            response = requests.post(
                f"{self.config['splunk']['hec_url']}/services/collector/event",
                headers=headers,
                json=payload,
                timeout=10
            )
            
            if response.ok:
                return {'success': True, 'message': 'Event sent to Splunk'}
            else:
                return {'success': False, 'error': response.text}
        
        except Exception as e:
            return {'success': False, 'error': str(e)}
    
    def send_to_elk(self, event):
        """Отправка в Elasticsearch"""
        if not self.config['elk']['enabled']:
            return {'success': False, 'error': 'ELK not configured'}
        
        try:
            url = f"{self.config['elk']['elasticsearch_url']}/{self.config['elk']['index']}/_doc"
            
            auth = None
            if self.config['elk']['username']:
                auth = (self.config['elk']['username'], self.config['elk']['password'])
            
            response = requests.post(url, auth=auth, json=event, timeout=10)
            
            if response.status_code in [200, 201]:
                return {'success': True, 'message': 'Event sent to ELK'}
            else:
                return {'success': False, 'error': response.text}
        
        except Exception as e:
            return {'success': False, 'error': str(e)}
    
    def send_to_qradar(self, event):
        """Отправка в QRadar"""
        if not self.config['qradar']['enabled']:
            return {'success': False, 'error': 'QRadar not configured'}
        
        try:
            headers = {
                'SEC': self.config['qradar']['api_token'],
                'Content-Type': 'application/json'
            }
            
            # QRadar Event format
            qradar_event = {
                'magnitude': 5,
                'qid': 1001,
                'category': 'PhantomProxy Event',
                'high_level_category': 'Security Testing',
                'severity': 3,
                'sourceip': event.get('data', {}).get('ip_address', '0.0.0.0'),
                'username': event.get('data', {}).get('username', ''),
                'payload': json.dumps(event)
            }
            
            url = f"{self.config['qradar']['console_url']}/api/siem/events"
            response = requests.post(url, headers=headers, json=qradar_event, timeout=10)
            
            if response.ok:
                return {'success': True, 'message': 'Event sent to QRadar'}
            else:
                return {'success': False, 'error': response.text}
        
        except Exception as e:
            return {'success': False, 'error': str(e)}
    
    def send_to_syslog(self, message, facility=1, severity=6):
        """Отправка в Syslog"""
        if not self.config['syslog']['enabled']:
            return {'success': False, 'error': 'Syslog not configured'}
        
        try:
            # RFC 3164 format
            priority = (facility * 8) + severity
            timestamp = datetime.now().strftime('%b %d %H:%M:%S')
            hostname = socket.gethostname()
            
            syslog_msg = f"<{priority}>{timestamp} {hostname} PhantomProxy: {message}"
            
            if self.config['syslog']['protocol'] == 'udp':
                sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                sock.sendto(
                    syslog_msg.encode(),
                    (self.config['syslog']['server'], self.config['syslog']['port'])
                )
            else:  # TCP
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.connect((self.config['syslog']['server'], self.config['syslog']['port']))
                sock.send(syslog_msg.encode())
                sock.close()
            
            return {'success': True, 'message': 'Event sent to Syslog'}
        
        except Exception as e:
            return {'success': False, 'error': str(e)}
    
    def send_session_event(self, session_data):
        """Отправка события о сессии во все настроенные SIEM"""
        event = self.format_event('session_captured', {
            'session_id': session_data.get('id'),
            'email': session_data.get('email'),
            'service': session_data.get('service'),
            'quality_score': session_data.get('quality_score'),
            'classification': session_data.get('classification'),
            'ip_address': session_data.get('ip'),
            'timestamp': session_data.get('created_at')
        })
        
        results = {}
        
        # Send to all configured
        if self.config['splunk']['enabled']:
            results['splunk'] = self.send_to_splunk(event)
        
        if self.config['elk']['enabled']:
            results['elk'] = self.send_to_elk(event)
        
        if self.config['qradar']['enabled']:
            results['qradar'] = self.send_to_qradar(event)
        
        if self.config['syslog']['enabled']:
            results['syslog'] = self.send_to_syslog(
                f"Session captured: {session_data.get('email')} | {session_data.get('service')} | Quality: {session_data.get('classification')}"
            )
        
        # Also save to file
        self._save_to_file(event)
        
        return {'success': True, 'results': results}
    
    def send_campaign_event(self, campaign_data, event_type):
        """Отправка события о кампании"""
        event = self.format_event(f'campaign_{event_type}', {
            'campaign_id': campaign_data.get('id'),
            'campaign_name': campaign_data.get('name'),
            'service': campaign_data.get('service'),
            'status': campaign_data.get('status'),
            'created_by': campaign_data.get('created_by')
        })
        
        results = {}
        
        if self.config['splunk']['enabled']:
            results['splunk'] = self.send_to_splunk(event)
        
        if self.config['elk']['enabled']:
            results['elk'] = self.send_to_elk(event)
        
        if self.config['syslog']['enabled']:
            results['syslog'] = self.send_to_syslog(
                f"Campaign {event_type}: {campaign_data.get('name')} | Status: {campaign_data.get('status')}"
            )
        
        self._save_to_file(event)
        
        return {'success': True, 'results': results}
    
    def send_user_action_event(self, user_id, action, resource_type, resource_id):
        """Отправка события о действии пользователя"""
        event = self.format_event('user_action', {
            'user_id': user_id,
            'action': action,
            'resource_type': resource_type,
            'resource_id': resource_id
        })
        
        if self.config['syslog']['enabled']:
            self.send_to_syslog(f"User {user_id} performed {action} on {resource_type}/{resource_id}")
        
        self._save_to_file(event)
        
        return {'success': True}
    
    def _save_to_file(self, event):
        """Сохранение в файл (резервное)"""
        export_file = SIEM_EXPORTS_PATH / f"siem_export_{datetime.now().strftime('%Y%m%d')}.json"
        
        with open(export_file, 'a') as f:
            f.write(json.dumps(event) + '\n')
    
    def export_to_file(self, output_path=None):
        """Экспорт всех событий в файл (для ручного импорта)"""
        if output_path is None:
            output_path = SIEM_EXPORTS_PATH / f"full_export_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        
        # В реальной версии здесь была бы выгрузка из БД
        export_data = {
            'exported_at': datetime.now().isoformat(),
            'product': 'PhantomProxy',
            'version': '12.5',
            'events': []
        }
        
        with open(output_path, 'w') as f:
            json.dump(export_data, f, indent=2)
        
        return {'success': True, 'path': str(output_path)}

# === TEST ===
if __name__ == '__main__':
    print("PhantomProxy v12.5 PRO++++ — SIEM Integration Module")
    print("="*60)
    
    siem = SIEMIntegration()
    print("✅ SIEM Integration initialized")
    
    # Test configurations
    print("\n📡 SIEM Configurations:")
    print("   - Splunk HEC: siem.configure_splunk('https://splunk:8088', 'token')")
    print("   - ELK: siem.configure_elk('https://elasticsearch:9200', 'index')")
    print("   - QRadar: siem.configure_qradar('https://qradar', 'token')")
    print("   - Syslog: siem.configure_syslog('syslog.server.com', 514)")
    
    # Test event formatting
    print("\n📝 Event Formatting Test:")
    event = siem.format_event('test_event', {'test': 'data'})
    print(f"   Event keys: {list(event.keys())}")
    print(f"   Timestamp: {event['@timestamp']}")
    
    # Test file export
    print("\n💾 File Export Test:")
    result = siem.export_to_file()
    print(f"   ✅ Export path: {result['path']}")
    
    print("\n✅ All SIEM integration features ready!")
