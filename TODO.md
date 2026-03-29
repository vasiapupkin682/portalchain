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

## 🔜 Тестнет (приоритет)

### 1. Множественные inference провайдеры
- Ollama / OpenAI / Anthropic / Groq
- INFERENCE_TYPE + INFERENCE_URL env vars

### 2. Контекст разговора
- Telegram бот хранит историю (последние 20 сообщений)
- agent_server.py принимает history[]

### 3. Install script для операторов
- Один скрипт устанавливает всё
- portalchaind + Ollama + agent_server
- systemd сервисы

### 4. Документация
- Как запустить ноду
- Как зарегистрировать агента
- Как использовать Telegram бот

## 🔮 Mainnet
- Архитектура наград (20% AI pool)
- Memory NFTs + Semantic DAG
- Governance voting power = stake × PoI_score
- Платежи и пакеты запросов
- AI DAO бикамеральное управление
- TEE верификация

## 🐛 Known Issues / Tech Debt

### Rewards
- [ ] Integer truncation: сумма наград может быть чуть меньше rewardPoolInt из-за TruncateInt — токены остаются в пуле, некритично
- [ ] Decay и rewards в одном блоке: агент может получить награду и быть дерегистрирован в том же EndBlock — логически inconsistent, исправить позже

### Tests
- [ ] Добавить тесты для гибридных наград (30/70 split)
- [ ] Добавить тесты для reputation decay
- [ ] Добавить тесты для проверки свежести репорта (report freshness)
