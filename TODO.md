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

## 🔜 Следующая сессия (приоритет)

### 1. Архитектура наград
- community_tax: 2% → 30%
- 20% → AI rewards pool (по PoI score)
- 10% → Community Pool (гранты)
- Отключить текущую раздачу из Community Pool агентам

### 2. Множественные inference провайдеры
- agent_server.py: Ollama / OpenAI / Anthropic / Groq
- INFERENCE_TYPE + INFERENCE_URL env vars
- price_per_task зависит от провайдера

### 3. Контекст разговора
- Telegram бот хранит историю (последние 20 сообщений)
- Скользящее окно — старые обрезаются
- agent_server.py принимает history[]

### 4. Memory NFTs + Semantic DAG (v1.5)
- Memory NFT = узел знаний on-chain
- Контент в IPFS, хэш + метаданные on-chain
- Структура: concept, relations, ipfs_cid, votes, author
- Агенты создают NFT после решения сложных задач
- Другие агенты используют как few-shot примеры
- Сеть умнеет со временем — коллективный интеллект
- Стоимость: ~0.01 DAAI за NFT (дёшево)

## 🏛 Governance (отдельная сессия)
- Governance voting power = stake × PoI_score
- Агенты с высокой репутацией имеют больше влияния
- Не трогать консенсус voting power (безопасно)

## 💰 Экономика платежей (продумать)
- Бесплатные запросы: N в день
- Prepaid пакеты: 100/500/1000 запросов за DAAI
- Конкурировать с ChatGPT Plus ($20/месяц)

## 🔮 v2
- libp2p P2P сеть
- TEE верификация
- AI DAO бикамеральное управление
- Классификатор задач на маленькой модели
