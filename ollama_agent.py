#!/usr/bin/env python3
"""
🤖 AI-агент для PortalChain с интеграцией Ollama (Llama 3.2)
Агент советуется с Llama 3.2 перед каждым решением
"""

import subprocess
import json
import random
import time
import sys
import os
import requests
from datetime import datetime
import signal
import re

# ========== КОНФИГУРАЦИЯ ==========
VALIDATOR_NAME = "alice"
CHAIN_ID = "portalchain"
INITIAL_EPOCH = 6000
SLEEP_BETWEEN_TX = 5
OLLAMA_URL = "http://localhost:11434/api/generate"
OLLAMA_MODEL = "llama3.2"
REQUEST_TIMEOUT = 60  # секунд на размышления

class LlamaAgent:
    def __init__(self, validator_name):
        self.name = f"Agent-{validator_name}"
        self.validator = validator_name
        self.epoch = INITIAL_EPOCH
        self.memory = self.load_memory()
        self.stats = {
            'total_reports': 0,
            'successful_txs': 0,
            'failed_txs': 0
        }
        
        # Проверяем доступность Ollama
        self.check_ollama()
        
        print(f"\n🤖 Агент {self.name} инициализирован")
        print(f"🧠 Модель: {OLLAMA_MODEL}")
        print(f"📊 Всего отчётов в памяти: {len(self.memory)}")
    
    def check_ollama(self):
        """Проверяет, доступна ли Ollama"""
        try:
            response = requests.get("http://localhost:11434/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get('models', [])
                model_names = [m['name'] for m in models]
                print(f"✅ Ollama доступна. Модели: {', '.join(model_names)}")
                
                # Проверяем, есть ли нужная модель
                if not any(OLLAMA_MODEL in m for m in model_names):
                    print(f"⚠️ Модель {OLLAMA_MODEL} не найдена!")
                    print(f"   Скачайте её: ollama pull {OLLAMA_MODEL}")
            else:
                print("⚠️ Ollama отвечает, но что-то не так")
        except requests.exceptions.ConnectionError:
            print("❌ Ollama не доступна! Запустите: ollama serve")
            print("   Агент будет использовать запасную логику")
        except requests.exceptions.Timeout:
            print("⏰ Таймаут при проверке Ollama")
    
    def load_memory(self):
        """Загружает историю из файла"""
        mem_file = f"agent_{self.validator}_memory.json"
        if os.path.exists(mem_file):
            try:
                with open(mem_file, 'r') as f:
                    return json.load(f)
            except:
                return []
        return []
    
    def save_memory(self):
        """Сохраняет историю"""
        mem_file = f"agent_{self.validator}_memory.json"
        with open(mem_file, 'w') as f:
            json.dump(self.memory, f, indent=2)
    
    def consult_llama(self, context):
        """Советуется с Llama 3.2 для принятия решения с увеличенным timeout"""
        
        # Формируем промпт, понятный для Llama 3.2
        prompt = f"""You are an AI agent managing a validator in the PortalChain blockchain network.

Current situation:
- Epoch number: {context['epoch']}
- Blockchain height: {context['height']}
- Your current reputation score: {context['reputation']:.4f}
- Your last task count: {context['last_tasks']}
- Your last reliability: {context['last_reliability']:.3f}

Your goal: Maintain a high reputation while processing tasks efficiently.

Based on this context, decide:
1. How many tasks to process this epoch (choose between 80-150)
2. What reliability target to aim for (choose between 0.85-0.99)

Think step by step, then respond with ONLY a valid JSON object in this exact format:
{{"tasks": 120, "reliability": 0.95, "reasoning": "Brief explanation of your decision"}}

No other text, just the JSON object.
"""
        
        try:
            # УВЕЛИЧИВАЕМ TIMEOUT ДО 60 СЕКУНД
            response = requests.post(OLLAMA_URL, json={
                "model": OLLAMA_MODEL,
                "prompt": prompt,
                "stream": False,
                "temperature": 0.7,
                "max_tokens": 150,
                "options": {
                    "num_predict": 150,
                    "temperature": 0.7
                }
            }, timeout=REQUEST_TIMEOUT)
            
            if response.status_code == 200:
                result = response.json()
                response_text = result.get('response', '').strip()
                
                # Пытаемся извлечь JSON из ответа
                try:
                    # Ищем JSON в ответе (на случай, если модель добавила лишний текст)
                    json_match = re.search(r'\{.*\}', response_text, re.DOTALL)
                    if json_match:
                        decision = json.loads(json_match.group())
                        print(f"🤔 Llama: {decision.get('reasoning', '')[:80]}...")
                        return decision
                    else:
                        print(f"⚠️ Не удалось найти JSON в ответе: {response_text[:100]}")
                        return self.fallback_decision(context)
                except json.JSONDecodeError as e:
                    print(f"⚠️ Ошибка парсинга JSON: {e}")
                    print(f"Ответ: {response_text[:200]}")
                    return self.fallback_decision(context)
            else:
                print(f"⚠️ Ollama ошибка: {response.status_code}")
                return self.fallback_decision(context)
                
        except requests.exceptions.Timeout:
            print(f"⏰ Таймаут! Llama думает дольше {REQUEST_TIMEOUT} секунд. Увеличьте REQUEST_TIMEOUT в коде.")
            return self.fallback_decision(context)
        except Exception as e:
            print(f"⚠️ Ошибка связи с Ollama: {e}")
            return self.fallback_decision(context)
    
    def fallback_decision(self, context=None):
        """Умная запасная логика, учитывающая контекст"""
        print("🤖 Использую запасную логику...")
        
        # Берём последнюю репутацию из памяти
        last_reports = self.memory[-5:] if self.memory else []
        
        if last_reports:
            # Если в последнее время репутация падала, работаем осторожнее
            avg_rel = sum(r['reliability'] for r in last_reports) / len(last_reports)
            if avg_rel < 0.9:
                tasks = random.randint(60, 80)  # меньше задач
                reliability = round(random.uniform(0.92, 0.96), 2)  # выше надёжность
                reasoning = "fallback: cautious mode due to recent low reliability"
            else:
                tasks = random.randint(90, 110)
                reliability = round(random.uniform(0.88, 0.94), 2)
                reasoning = "fallback: normal mode"
        else:
            tasks = random.randint(80, 120)
            reliability = round(random.uniform(0.85, 0.95), 2)
            reasoning = "fallback: no history available"
        
        return {
            "tasks": tasks,
            "reliability": reliability,
            "reasoning": reasoning
        }
    
    def get_current_reputation(self):
        """Получает текущую репутацию валидатора из блокчейна"""
        try:
            result = subprocess.run(
                ["portalchaind", "q", "poi", "reputation", self.validator],
                capture_output=True, text=True, check=True
            )
            # Парсим JSON ответ
            data = json.loads(result.stdout)
            return float(data['reputation']['value'])
        except:
            return 0.0
    
    def get_network_height(self):
        """Получает текущую высоту блока"""
        try:
            result = subprocess.run(
                ["portalchaind", "status"],
                capture_output=True, text=True, check=True
            )
            data = json.loads(result.stdout)
            return int(data['SyncInfo']['latest_block_height'])
        except:
            return 0
    
    def get_last_report(self):
        """Получает последний отчёт валидатора"""
        if self.memory:
            return self.memory[-1]
        return {'tasks': 100, 'reliability': 0.95}
    
    def make_decision(self):
        """Принимает решение с помощью Llama"""
        # Собираем контекст
        context = {
            'epoch': self.epoch,
            'height': self.get_network_height(),
            'reputation': self.get_current_reputation(),
            'last_reliability': self.get_last_report().get('reliability', 0.95),
            'last_tasks': self.get_last_report().get('tasks', 100)
        }
        
        # Советуемся с Llama
        decision = self.consult_llama(context)
        
        # Генерируем остальные метрики на основе решения
        tasks = decision.get('tasks', 100)
        weighted_sum = tasks * random.randint(45, 55)
        latency = random.randint(100, 200)
        failures = random.choices([0, 1], weights=[0.95, 0.05])[0]
        
        return {
            'epoch': self.epoch,
            'tasks': tasks,
            'weighted_sum': weighted_sum,
            'latency': latency,
            'reliability': decision.get('reliability', 0.95),
            'failures': failures,
            'reasoning': decision.get('reasoning', 'No reasoning provided'),
            'timestamp': int(time.time())
        }
    
    def submit_report(self, decision):
        """Отправляет отчёт в блокчейн и проверяет на sampling selection"""
        cmd = [
            "portalchaind", "tx", "poi", "submit-report",
            "--epoch", str(decision['epoch']),
            "--tasks-processed", str(decision['tasks']),
            "--weighted-task-sum", str(decision['weighted_sum']),
            "--avg-latency", str(decision['latency']),
            "--reliability", str(decision['reliability']),
            "--sampling-failures", str(decision['failures']),
            "--from", self.validator,
            "--chain-id", CHAIN_ID,
            "--yes",
            "--output", "json"
        ]
        
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            tx_data = json.loads(result.stdout)
            
            report = {
                'epoch': decision['epoch'],
                'txhash': tx_data['txhash'],
                'reliability': decision['reliability'],
                'tasks': decision['tasks'],
                'reasoning': decision.get('reasoning', ''),
                'timestamp': decision['timestamp']
            }

            sampled = self._check_sampling_selected(tx_data)
            if sampled:
                report['sampling_status'] = 'pending_verification'
                report['sampling_deadline'] = decision['epoch'] + 100
                print(f"   SAMPLED - report needs verification (deadline: epoch {report['sampling_deadline']})")

            self.memory.append(report)
            self.stats['total_reports'] += 1
            self.stats['successful_txs'] += 1
            
            print(f"Epoch {decision['epoch']}: rel={decision['reliability']:.3f} tasks={decision['tasks']}")
            print(f"   Llama: {decision.get('reasoning', '')[:100]}")
            print(f"   tx: {tx_data['txhash'][:16]}...")
            
            return tx_data
            
        except subprocess.CalledProcessError as e:
            print(f"TX error: {e.stderr}")
            self.stats['failed_txs'] += 1
            return None
        except json.JSONDecodeError:
            print(f"JSON parse error in tx response")
            self.stats['failed_txs'] += 1
            return None

    def _check_sampling_selected(self, tx_data):
        """Returns True if the tx events contain a sampling_selected event."""
        events = tx_data.get('events', [])
        if isinstance(events, list):
            for ev in events:
                ev_type = ev.get('type', '')
                if ev_type == 'sampling_selected':
                    return True
        logs = tx_data.get('logs', [])
        if isinstance(logs, list):
            for log_entry in logs:
                for ev in log_entry.get('events', []):
                    if ev.get('type') == 'sampling_selected':
                        return True
        return False

    # ------------------------------------------------------------------
    # Verification of other agents' reports
    # ------------------------------------------------------------------

    def get_pending_samplings(self):
        """Queries the chain for pending sampling records."""
        try:
            result = subprocess.run(
                ["portalchaind", "q", "poi", "list-sampling-records",
                 "--status", "pending", "--output", "json"],
                capture_output=True, text=True, check=True
            )
            data = json.loads(result.stdout)
            return data.get('records', data.get('sampling_records', []))
        except (subprocess.CalledProcessError, json.JSONDecodeError) as e:
            print(f"Could not query pending samplings: {e}")
            return []

    def consult_llama_verify(self, report_to_verify):
        """Asks Llama whether a report from another validator looks legitimate."""
        prompt = (
            "You are verifying an epoch report from another validator.\n"
            "Report metrics:\n"
            f"- Tasks processed: {report_to_verify.get('tasks_processed', 'N/A')}\n"
            f"- Reliability: {report_to_verify.get('reliability', 'N/A')}\n"
            f"- Avg latency: {report_to_verify.get('avg_latency', 'N/A')}ms\n"
            f"- Sampling failures: {report_to_verify.get('sampling_failures', 'N/A')}\n\n"
            "Is this report legitimate?\n"
            "Respond with ONLY a valid JSON object:\n"
            '{"passed": true, "reasoning": "brief explanation"}'
        )

        try:
            response = requests.post(OLLAMA_URL, json={
                "model": OLLAMA_MODEL,
                "prompt": prompt,
                "stream": False,
                "options": {
                    "num_predict": 150,
                    "temperature": 0.3
                }
            }, timeout=REQUEST_TIMEOUT)

            if response.status_code == 200:
                text = response.json().get('response', '').strip()
                json_match = re.search(r'\{.*\}', text, re.DOTALL)
                if json_match:
                    return json.loads(json_match.group())
        except Exception as e:
            print(f"Llama verify error: {e}")

        return {"passed": True, "reasoning": "fallback: approved by default"}

    def _fetch_epoch_report(self, epoch, validator):
        """Fetches a stored epoch report from the chain."""
        try:
            result = subprocess.run(
                ["portalchaind", "q", "poi", "epoch-report",
                 str(epoch), validator, "--output", "json"],
                capture_output=True, text=True, check=True
            )
            return json.loads(result.stdout).get('report', json.loads(result.stdout))
        except (subprocess.CalledProcessError, json.JSONDecodeError):
            return None

    def check_pending_verifications(self):
        """Finds pending sampling records and verifies those belonging to other validators."""
        records = self.get_pending_samplings()
        if not records:
            return

        for record in records:
            validator = record.get('validator', '')
            epoch = record.get('epoch', 0)

            if validator == self.validator:
                continue

            report = self._fetch_epoch_report(epoch, validator)
            if report is None:
                print(f"Could not fetch report for epoch {epoch} / {validator}, skipping")
                continue

            verdict = self.consult_llama_verify(report)
            passed = verdict.get('passed', True)
            reasoning = verdict.get('reasoning', '')

            print(f"Verifying epoch {epoch} validator {validator}: "
                  f"passed={passed} ({reasoning[:60]})")

            self._send_verify_sampling(epoch, validator, passed)

    def _send_verify_sampling(self, epoch, validator, passed):
        """Sends MsgVerifySampling to the chain."""
        cmd = [
            "portalchaind", "tx", "poi", "verify-sampling",
            "--epoch", str(epoch),
            "--validator", validator,
            "--passed", str(passed).lower(),
            "--from", self.validator,
            "--chain-id", CHAIN_ID,
            "--yes",
            "--output", "json"
        ]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            tx_data = json.loads(result.stdout)
            print(f"   Verification tx: {tx_data.get('txhash', 'unknown')[:16]}...")
        except subprocess.CalledProcessError as e:
            print(f"   Verification tx failed: {e.stderr}")
        except json.JSONDecodeError:
            print(f"   Verification tx: response not JSON")
    
    def run(self):
        """Main agent loop: verify others, submit own report, repeat."""
        print(f"\nStarting agent {self.name}")
        print(f"Interval: {SLEEP_BETWEEN_TX}s")
        print("=" * 60)
        
        try:
            while True:
                self.check_pending_verifications()

                decision = self.make_decision()
                self.submit_report(decision)

                self.save_memory()

                self.epoch += 1
                time.sleep(SLEEP_BETWEEN_TX)
                
        except KeyboardInterrupt:
            self.print_stats()
            self.save_memory()
            sys.exit(0)
    
    def print_stats(self):
        """Показывает статистику"""
        print("\n" + "="*60)
        print(f"📊 СТАТИСТИКА АГЕНТА {self.name}")
        print("="*60)
        print(f"Всего отчётов: {self.stats['total_reports']}")
        print(f"Успешных: {self.stats['successful_txs']}")
        print(f"Ошибок: {self.stats['failed_txs']}")
        if self.memory:
            avg_rel = sum(r['reliability'] for r in self.memory) / len(self.memory)
            print(f"Средняя надёжность: {avg_rel:.3f}")
        print("="*60)

if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description='Запуск AI-агента для PortalChain')
    parser.add_argument('--from', dest='validator', default='alice',
                       help='Имя валидатора (alice/bob)')
    
    args = parser.parse_args()
    
    agent = LlamaAgent(args.validator)
    agent.run()