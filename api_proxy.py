#!/usr/bin/env python3
import subprocess
import json
import urllib.request
import urllib.error
import random
import time
from http.server import HTTPServer, BaseHTTPRequestHandler

RPC_URL = "http://127.0.0.1:26657"

# =====================================================
# FREE QUOTA — 10 requests per IP per day
# =====================================================
FREE_LIMIT = 5
quota = {}  # { ip: { 'count': int, 'reset_at': timestamp } }

def check_quota(ip: str) -> tuple[bool, int]:
    """Returns (allowed, remaining)"""
    now = time.time()
    day = 86400  # 24 hours in seconds
    if ip not in quota or now > quota[ip]['reset_at']:
        quota[ip] = {'count': 0, 'reset_at': now + day}
    entry = quota[ip]
    if entry['count'] >= FREE_LIMIT:
        return False, 0
    return True, FREE_LIMIT - entry['count']

def increment_quota(ip: str):
    now = time.time()
    day = 86400
    if ip not in quota or now > quota[ip]['reset_at']:
        quota[ip] = {'count': 0, 'reset_at': now + day}
    quota[ip]['count'] += 1

def is_agent_alive(endpoint: str) -> bool:
    try:
        req = urllib.request.urlopen(f"{endpoint}/status", timeout=3)
        return req.status == 200
    except Exception:
        return False

def get_active_models():
    result = subprocess.run(
        ['portalchaind', 'query', 'model-registry', 'list-active', '--output', 'json'],
        capture_output=True, text=True
    )
    if result.returncode == 0:
        try:
            agents = json.loads(result.stdout)
            if not isinstance(agents, list):
                agents = agents.get('models', [])
            return agents
        except:
            pass
    return []

def select_agent(models, task_type):
    alive = [m for m in models if is_agent_alive(m.get('endpoint', ''))]
    if not alive:
        return None
    base_weight = 0.1
    rep_key = f"rep_{task_type}" if task_type in ['text', 'code', 'analysis'] else 'rep_text'
    weights = [float(m.get(rep_key, '0') or '0') + base_weight for m in alive]
    total = sum(weights)
    r = random.uniform(0, total)
    cumulative = 0
    for model, weight in zip(alive, weights):
        cumulative += weight
        if r <= cumulative:
            return model
    return alive[-1]

class Handler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def get_client_ip(self):
        forwarded = self.headers.get('X-Forwarded-For')
        if forwarded:
            return forwarded.split(',')[0].strip()
        return self.client_address[0]

    def send_cors_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_cors_headers()
        self.end_headers()

    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_cors_headers()
        self.end_headers()

        if self.path == '/agents':
            models = get_active_models()
            for agent in models:
                agent['online'] = is_agent_alive(agent.get('endpoint', ''))
            self.wfile.write(json.dumps(models).encode())

        elif self.path == '/reputations':
            agents = get_active_models()
            reps = []
            for agent in agents:
                addr = agent.get('operator', '')
                if addr:
                    rep_result = subprocess.run(
                        ['portalchaind', 'query', 'poi', 'reputation', addr, '--output', 'json'],
                        capture_output=True, text=True
                    )
                    tasks_result = subprocess.run(
                        ['portalchaind', 'query', 'poi', 'reports', addr, '--output', 'json'],
                        capture_output=True, text=True
                    )
                    total_tasks = 0
                    if tasks_result.returncode == 0:
                        try:
                            reports_data = json.loads(tasks_result.stdout)
                            reports = reports_data if isinstance(reports_data, list) else reports_data.get('reports', [])
                            total_tasks = sum(int(r.get('tasks_processed', 0)) for r in reports)
                        except:
                            pass
                    if rep_result.returncode == 0:
                        try:
                            rep_data = json.loads(rep_result.stdout)
                            reps.append({
                                'validator': addr,
                                'value': rep_data.get('reputation', {}).get('value', '0'),
                                'total_tasks': total_tasks
                            })
                        except:
                            reps.append({'validator': addr, 'value': '0', 'total_tasks': 0})
            self.wfile.write(json.dumps(reps).encode())

        elif self.path == '/status':
            result = subprocess.run(
                ['portalchaind', 'status'],
                capture_output=True, text=True
            )
            height = '0'
            if result.returncode == 0:
                try:
                    data = json.loads(result.stdout)
                    height = data['SyncInfo']['latest_block_height']
                except:
                    pass
            self.wfile.write(json.dumps({'block_height': height}).encode())

        else:
            self.wfile.write(b'{}')

    def do_POST(self):
        if self.path == '/ask':
            content_length = int(self.headers.get('Content-Length', 0))
            body = self.rfile.read(content_length)
            try:
                data = json.loads(body)
                query = data.get('query', '')
                task_type = data.get('task_type', 'text')
                paid = data.get('paid', False)

                # Check free quota (skip for paid requests)
                ip = self.get_client_ip()
                allowed, remaining = check_quota(ip)
                if not allowed and not paid:
                    self.send_response(429)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({
                        'error': 'free_limit_exceeded',
                        'message': 'Daily free limit reached. Switch to PAY mode to continue.',
                        'remaining': 0
                    }).encode())
                    return

                models = get_active_models()
                agent = select_agent(models, task_type)

                if not agent:
                    self.send_response(503)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': 'No agents available'}).encode())
                    return

                endpoint = agent.get('endpoint', '')
                req = urllib.request.Request(
                    f"{endpoint}/task",
                    data=json.dumps({'prompt': query, 'max_tokens': 500}).encode(),
                    headers={'Content-Type': 'application/json'},
                    method='POST'
                )
                with urllib.request.urlopen(req, timeout=120) as resp:
                    result = json.loads(resp.read())

                # Increment quota only on success
                increment_quota(ip)

                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.send_cors_headers()
                self.end_headers()
                self.wfile.write(json.dumps({
                    'result': result.get('result', ''),
                    'agent': endpoint,
                    'latency_ms': result.get('latency_ms', 0),
                    'task_type': task_type,
                    'remaining': remaining - 1,
                }).encode())

            except Exception as e:
                self.send_response(500)
                self.send_header('Content-Type', 'application/json')
                self.send_cors_headers()
                self.end_headers()
                self.wfile.write(json.dumps({'error': str(e)}).encode())

        elif self.path == '/broadcast-amino':
            content_length = int(self.headers.get('Content-Length', 0))
            body = self.rfile.read(content_length)
            try:
                data = json.loads(body)
                signed_doc = data.get('signed_doc', {})
                signature = data.get('signature', {})
                sender = data.get('sender', '')
                query = data.get('query', '')
                task_type = data.get('task_type', 'text')

                if not signed_doc or not signature or not sender:
                    self.send_response(400)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': 'Missing required fields'}).encode())
                    return

                msgs = signed_doc.get('msgs', [])
                if not msgs:
                    self.send_response(400)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': 'No messages in signed doc'}).encode())
                    return

                msg_value = msgs[0].get('value', {})
                query_hash = msg_value.get('query_hash', '')
                query_url = msg_value.get('query_url', '')

                result = subprocess.run([
                    'portalchaind', 'tx', 'tasks', 'create-task',
                    '--query', query_hash,
                    '--task-type', task_type,
                    '--from', 'validator',
                    '--keyring-backend', 'test',
                    '--chain-id', 'portalchain',
                    '--fees', '1000udaai',
                    '--yes',
                    '--output', 'json'
                ], capture_output=True, text=True, timeout=30)

                if result.returncode != 0:
                    self.send_response(400)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': result.stderr or result.stdout}).encode())
                    return

                tx_data = json.loads(result.stdout)
                code = tx_data.get('code', 0)
                if code != 0:
                    self.send_response(400)
                    self.send_header('Content-Type', 'application/json')
                    self.send_cors_headers()
                    self.end_headers()
                    self.wfile.write(json.dumps({
                        'error': tx_data.get('raw_log', 'Transaction failed'),
                        'code': code
                    }).encode())
                    return

                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.send_cors_headers()
                self.end_headers()
                self.wfile.write(json.dumps({
                    'txhash': tx_data.get('txhash', ''),
                    'code': 0,
                    'log': tx_data.get('raw_log', '')
                }).encode())

            except Exception as e:
                self.send_response(500)
                self.send_header('Content-Type', 'application/json')
                self.send_cors_headers()
                self.end_headers()
                self.wfile.write(json.dumps({'error': str(e)}).encode())

        else:
            self.send_response(404)
            self.end_headers()

HTTPServer(('0.0.0.0', 8090), Handler).serve_forever()
