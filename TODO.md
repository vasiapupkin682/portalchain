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
- Токеномика — агенты получают награды из Community Pool
- Faucet — /faucet <address> в Telegram боте
- Слэшинг — 10% стейка при 3+ sampling failures

## 🔜 Следующая сессия

### Архитектура наград (важно)
- Изменить community_tax: 2% → 30%
- 20% → AI rewards pool (по PoI score)
- 10% → Community Pool (гранты)
- Отключить текущую раздачу из Community Pool агентам

### Множественные inference провайдеры
- agent_server.py: поддержка Ollama / OpenAI / Anthropic / Groq
- INFERENCE_TYPE env var
- Оператор выбирает провайдера

### Governance voting power
- Безопасный вариант: governance voting power = stake × PoI_score
- Не трогать консенсус voting power
- Агенты с высокой репутацией имеют больше влияния в DAO

## 💰 Экономика платежей (продумать)
- Бесплатные запросы: N в день для новых пользователей
- Prepaid пакеты: 100/500/1000 запросов за DAAI
- MsgBuyRequestPackage on-chain
- Конкурировать с ChatGPT Plus ($20/месяц)

## 🔮 v2
- libp2p P2P сеть
- Memory NFTs
- TEE верификация
- AI DAO бикамеральное управление
- Классификатор задач на маленькой модели
