# PortalChain TODO

## ✅ Выполнено
- x/poi — отчёты, репутация, random sampling
- x/constitution — S1-S4 священные принципы
- agent_server.py — реальный инференс через Llama
- Telegram бот DAAI
- x/model-registry — реестр агентов

## 🔜 Следующая сессия (Месяц 5)

### Репутация по категориям задач
- Добавить rep_text, rep_code, rep_analysis, rep_general в ModelRecord
- Добавить task_type в MsgSubmitEpochReport
- Обновить UpdateReputation() — обновлять по категории
- Классификация задач в telegram_bot.py по ключевым словам

### Роутинг задач
- Telegram бот читает x/model-registry
- Относительный порог: rep > 20% от максимальной репутации в сети
- Взвешенный случайный выбор агента (weighted random)
- Два уровня: light задачи для слабых агентов, любые для сильных

### Токеномика (Месяц 5-6)
- AI Treasury аккаунт в genesis
- x/poi EndBlocker: каждые 100 блоков распределять из Community Pool
- Пропорционально PoI score агента
- Автоматически без клейма

### Faucet
- Команда /faucet <address> в Telegram боте
- Лимит: 1 раз в 24 часа на адрес
- Выдавать 100 PORTAL

## 🔮 v2 (после тестнета)
- Классификатор задач на маленькой модели (llama3.2:1b)
- libp2p P2P сеть
- Memory NFTs
- TEE верификация
- Semantic DAG
- Circuit Breaker
- AI DAO бикамеральное управление
