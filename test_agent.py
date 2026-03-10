#!/usr/bin/env python3
"""
🤖 Тестовый AI-агент для PortalChain
Симулирует работу настоящего агента: принимает решения, учится на ошибках, 
адаптирует своё поведение
"""

import subprocess
import json
import random
import time
import sys
import os
from datetime import datetime
import signal

# ========== КОНФИГУРАЦИЯ ==========
VALIDATOR_NAME = "alice"
CHAIN_ID = "portalchain"
INITIAL_EPOCH = 5000  # начнём с чистой эпохи
SLEEP_BETWEEN_TX = 3  # секунд между отчётами

# Файл для сохранения истории (чтобы агент "помнил")
HISTORY_FILE = "agent_memory.json"

class PortalChainAIAgent:
    def __init__(self):
        self.name = "TestAgent-01"
        self.epoch = INITIAL_EPOCH
        self.memory = self.load_memory()
        print(f"🤖 Агент {self.name} инициализирован")
        print(f"🧠 Память: {len(self.memory.get('history', []))} прошлых отчётов")
    
    def load_memory(self):
        """Загружает историю из файла (долгосрочная память)"""
        if os.path.exists(HISTORY_FILE):
            try:
                with open(HISTORY_FILE, 'r') as f:
                    return json.load(f)
            except:
                return {"history": [], "stats": {"total_reports": 0, "avg_reliability": 0.95}}
        return {"history": [], "stats": {"total_reports": 0, "avg_reliability": 0.95}}
    
    def save_memory(self):
        """Сохраняет историю"""
        with open(HISTORY_FILE, 'w') as f:
            json.dump(self.memory, f, indent=2)
    
    def analyze_performance(self, last_reports):
        """Анализирует последние отчёты и решает, как улучшить работу"""
        if len(last_reports) < 5:
            return 1.0  # пока недостаточно данных
        
        # Считаем среднюю надёжность
        reliabilities = [r.get('reliability', 0.95) for r in last_reports[-5:]]
        avg_rel = sum(reliabilities) / len(reliabilities)
        
        # Если надёжность падает, агент "учится" и улучшает её
        if avg_rel < 0.9:
            print("📈 Агент учится на ошибках: повышаю качество работы")
            return min(avg_rel + 0.05, 0.99)
        else:
            # Случайные колебания
            return avg_rel + random.uniform(-0.02, 0.02)
    
    def make_decision(self):
        """Принимает решение: сколько задач взять, с какой надёжностью"""
        # Загружаем последние отчёты из памяти
        last_reports = self.memory.get('history', [])[-20:]
        
        # Анализируем прошлую производительность
        base_reliability = self.analyze_performance(last_reports)
        
        # Базовые метрики с интеллектуальными корректировками
        tasks = random.randint(80, 120)
        
        # Если агент "устал" (каждый 10-й раз), берёт меньше задач
        if self.epoch % 10 == 0:
            tasks = int(tasks * 0.7)
            print("😴 Агент немного устал, беру меньше задач")
        
        # Чем больше задач, тем выше может быть латентность
        weighted_sum = tasks * random.randint(40, 60)
        
        # Латентность зависит от количества задач
        base_latency = 100 + (tasks // 2)
        latency = int(random.gauss(base_latency, 15))
        latency = max(50, min(500, latency))  # ограничиваем
        
        # Надёжность с учётом обучения
        reliability = round(random.gauss(base_reliability, 0.02), 3)
        reliability = max(0.7, min(0.99, reliability))
        
        # Ошибки случаются редко, но метко
        failures = random.choices([0, 1, 2], weights=[0.92, 0.06, 0.02])[0]
        
        return {
            'epoch': self.epoch,
            'tasks': tasks,
            'weighted_sum': weighted_sum,
            'latency': latency,
            'reliability': reliability,
            'failures': failures,
            'timestamp': int(time.time())
        }
    
    def submit_report(self, decision):
        """Отправляет отчёт в блокчейн"""
        cmd = [
            "portalchaind", "tx", "poi", "submit-report",
            "--epoch", str(decision['epoch']),
            "--tasks-processed", str(decision['tasks']),
            "--weighted-task-sum", str(decision['weighted_sum']),
            "--avg-latency", str(decision['latency']),
            "--reliability", str(decision['reliability']),
            "--sampling-failures", str(decision['failures']),
            "--from", VALIDATOR_NAME,
            "--chain-id", CHAIN_ID,
            "--yes",
            "--output", "json"
        ]
        
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            tx_data = json.loads(result.stdout)
            
            # Сохраняем в память
            report_record = {
                'epoch': decision['epoch'],
                'txhash': tx_data['txhash'],
                'reliability': decision['reliability'],
                'tasks': decision['tasks'],
                'timestamp': decision['timestamp']
            }
            self.memory['history'].append(report_record)
            self.memory['stats']['total_reports'] = len(self.memory['history'])
            self.save_memory()
            
            print(f"✅ Эпоха {decision['epoch']}: tx={tx_data['txhash'][:8]}... "
                  f"rel={decision['reliability']} tasks={decision['tasks']} "
                  f"fail={decision['failures']}")
            return tx_data
            
        except subprocess.CalledProcessError as e:
            print(f"❌ Ошибка: {e.stderr}")
            return None
    
    def run(self):
        """Основной цикл агента"""
        print(f"\n🤖 Агент {self.name} запущен в {datetime.now().strftime('%H:%M:%S')}")
        print(f"📤 Отправка отчётов каждые {SLEEP_BETWEEN_TX} сек\n")
        
        try:
            while True:
                # Принимаем решение
                decision = self.make_decision()
                
                # Отправляем отчёт
                self.submit_report(decision)
                
                # Переходим к следующей эпохе
                self.epoch += 1
                
                # Случайная пауза (иногда агент думает дольше)
                time.sleep(SLEEP_BETWEEN_TX * random.uniform(0.8, 1.2))
                
        except KeyboardInterrupt:
            print(f"\n\n🛑 Агент остановлен. Всего отправлено: {len(self.memory['history'])} отчётов")
            self.save_memory()
            sys.exit(0)

def signal_handler(sig, frame):
    print("\n\n🛑 Получен сигнал остановки, сохраняю память...")
    sys.exit(0)

if __name__ == "__main__":
    signal.signal(signal.SIGINT, signal_handler)
    
    # Проверяем, доступен ли portalchaind
    try:
        subprocess.run(["portalchaind", "version"], capture_output=True, check=True)
    except:
        print("❌ portalchaind не найден! Убедись, что сеть установлена")
        sys.exit(1)
    
    agent = PortalChainAIAgent()
    agent.run()
