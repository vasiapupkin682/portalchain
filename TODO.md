# PortalChain TODO

## ✅ Выполнено
- x/poi — отчёты, репутация, random sampling
- x/constitution — S1-S4 священные принципы
- agent_server.py — реальный инференс через Llama
- Telegram бот DAAI
- x/model-registry — реестр агентов со стейком
- Репутация по категориям (text/code/analysis)
- Smart routing — взвешенный случайный выбор агента
- DAAI токен + scripts/init.sh

## 🔜 Следующая сессия

### Токеномика
- AI Treasury аккаунт в genesis
- x/poi EndBlocker: каждые 100 блоков распределять из Community Pool
- Пропорционально PoI score агента
- Автоматически без клейма

### Faucet
- Команда /faucet <address> в Telegram боте
- Лимит: 1 раз в 24 часа на адрес
- Выдавать 1000 DAAI

### Слэшинг агентов
- SamplingFailures > порога → слэшинг части стейка
- Защита от некачественных агентов

## 🔮 v2
- Классификатор задач на маленькой модели
- libp2p P2P сеть
- Memory NFTs
- TEE верификация
- AI DAO бикамеральное управление
