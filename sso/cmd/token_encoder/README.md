Успехов! 🚀

```
make run-encoder token="v2.public..."
make build-encoder
```bash

Теперь вы можете легко кодировать токены для передачи через gRPC метаданные.

## 🎉 Готово!

---

- [sso/pkg/utils/grpc_metadata.go](../pkg/utils/grpc_metadata.go) - Функции утилит
- [QUICK_START.md](../../QUICK_START.md) - Быстрый старт
- [GRPC_EXAMPLES.md](../../GRPC_EXAMPLES.md) - Примеры на разных языках
- [TOKEN_METADATA_FIX.md](../../TOKEN_METADATA_FIX.md) - Подробное объяснение

## 📚 Дополнительно

---

A: Да, работает с любыми строками.
**Q: Работает ли с другими типами токенов?**  

A: Не храните - кодируйте на лету перед каждым запросом.
**Q: Где хранить закодированный токен?**  

A: Нет, Base64 одинаково быстро везде.
**Q: Это медленнее, чем прямое кодирование?**  

A: Нет, используйте функции кодирования в вашем приложении (Base64).
**Q: Могу ли я использовать это в продакшене?**  

A: Утилита нужна для отладки и тестирования. В реальном коде используйте `utils.AddAuthTokenToContext()`.
**Q: Для чего нужна эта утилита, если есть utils.AddAuthTokenToContext()?**  

## 🆘 Часто задаваемые вопросы

---

✅ **Документация** - Полная справка через `-help`  
✅ **Интеграция** - Легко встраивается в скрипты  
✅ **Отладка** - Можно декодировать обратно  
✅ **Универсальность** - Работает с любыми токенами  
✅ **Простота** - Одна команда для кодирования  

## ✨ Особенности

---

```
resp, _ := client.AssignRole(ctx, request)
ctx := utils.AddAuthTokenToContext(ctx, token)
// Просто используйте утилиту из pkg/utils
```go
### Вариант 2: В Go коде (РЕКОМЕНДУЕТСЯ)

```
curl -H "authorization-bin: $ENCODED" http://api/protected
# 3. Используйте в запросе

ENCODED=$(./bin/token_encoder -token "$TOKEN" | tail -1 | awk '{print $NF}')
# 2. Закодируйте его

TOKEN=$(curl -X POST http://api/login -d '{"user":"x","pass":"y"}' | jq '.accessToken')
# 1. Получите токен при логине
```bash
### Вариант 1: Используя утилиту

## 🎯 Workflow

---

| **Постман/Insomnia** | Скопируйте закодированный токен в header |
| **Скрипты** | Кодируйте через pipe: `echo $TOKEN \| token_encoder` |
| **Отладка** | Декодируйте токен с флагом `-decode` |
| **Работа в коде** | Используйте `utils.AddAuthTokenToContext()` |
| **Локальное тестирование** | Используйте token_encoder утилиту |
|----------|---------|
| Сценарий | Решение |

## 💡 Для чего это нужно

---

```
  localhost:50051 auth.User/GetProfile
  -d '{"user_id": 123}' \
grpcurl -H "authorization-bin: $ENCODED" \
# Использовать в grpcurl

ENCODED=$(./bin/token_encoder -token "$TOKEN" | grep "authorization-bin:" | awk '{print $NF}')
TOKEN="v2.public..."
# Получить и закодировать токен
```bash
#### Bash + curl + grpcurl

```
var headers = new Metadata { { "authorization-bin", encoded } };
string encoded = Convert.ToBase64String(Encoding.UTF8.GetBytes(token));
```csharp
#### C#

```
headers.put(authKey, encoded);
Metadata headers = new Metadata();
String encoded = Base64.getEncoder().encodeToString(token.getBytes());
```java
#### Java

```
client.assignRole(request, metadata, callback);
metadata.add('authorization-bin', encoded);
const metadata = new grpc.Metadata();
const encoded = Buffer.from(token).toString('base64');
```javascript
#### JavaScript/Node.js

```
stub.AssignRole(request, metadata=metadata)
metadata = [('authorization-bin', encoded)]
encoded = base64.b64encode(token.encode()).decode()
import base64
```python
#### Python

```
ctx := metadata.NewOutgoingContext(context.Background(), md)
md := metadata.Pairs("authorization-bin", encoded)
encoded := base64.StdEncoding.EncodeToString([]byte(token))
```go
#### Go (вручную)

```
resp, _ := client.AssignRole(ctx, request)
ctx := utils.AddAuthTokenToContext(ctx, accessToken)

import "sso/sso/pkg/utils"
```go
#### Go (автоматически)

### Использование в проектах

```
└── README.md             ← Эта документация
├── Makefile              ← Для сборки
│   └── main.go           ← Утилита для кодирования
├── token_encoder/
sso/cmd/
```

### Структура файлов

## 🔗 Интеграция

---

```
cat encoded_token.txt
./bin/token_encoder -token "v2.public..." > encoded_token.txt
```bash
#### 4. Сохранить в файл

```
make run-encoder token="dj2public1..." decode=1
# Декодирование

make run-encoder token="v2.public..."
# Кодирование
```bash
#### 3. Через make

```
./bin/token_encoder -decode -token "dj2public1eyJhcHBfaWQi..."
```bash
#### 2. Декодировать закодированный токен

```
./bin/token_encoder -token "v2.public.eyJhcHBfaWQiOjE..."
```bash
#### 1. Кодировать токен (по умолчанию)

### Примеры

```
token_encoder -help
```bash

### Все флаги

## 📖 Полное использование

---

```
  authorization-bin: dj2public1eyJhcHBfaWQiOjEsImV4cCI6MTc2NzQ3MTg4NSwiaWF0IjoxNjM5NzE5ODg1LCJ1c2VyX2lkIjo1fUW5vIEsQ0FgYLzMVVX0N1oyfDFAQWkjkb6AI7htgwwaAxgU.bnVsbA==
📝 Используйте в метаданных:

dj2public1eyJhcHBfaWQiOjEsImV4cCI6MTc2NzQ3MTg4NSwiaWF0IjoxNjM5NzE5ODg1LCJ1c2VyX2lkIjo1fUW5vIEsQ0FgYLzMVVX0N1oyfDFAQWkjkb6AI7htgwwaAxgU.bnVsbA==
✅ Закодированный токен (для header 'authorization-bin'):
```
Вывод:

### 3. Получить закодированный токен

```
make run-encoder token="v2.public.eyJhcHBfaWQiOjE..."
# Или через make

./bin/token_encoder -token "v2.public.eyJhcHBfaWQiOjE..."
# Кодировать токен в Base64
```bash

### 2. Использовать утилиту

```
make build-all
# Или собрать все утилиты

make build-encoder
# Собрать только token_encoder

cd sso
```bash

### 1. Собрать утилиту

## 🚀 Быстрый старт

---

```
dj2public1eyJhcHBfaWQi...     (безопасно для gRPC метаданных)
              ↓ Base64 кодирование
v2.public.eyJhcHBfaWQiOjE...  (исходный токен со спецсимволами)
```

Кодируйте токен в Base64 перед отправкой:

## ✅ Решение

```
❌ Error: Metadata string value "v2.public.eyJhcHBfaWQiOjE..." contains illegal characters
```

gRPC метаданные требуют ASCII-совместимых символов, но PASETO токены содержат спецсимволы:

## ⚠️ Проблема

**Token Encoder** - это простая утилита для кодирования/декодирования PASETO токенов в Base64 для безопасной передачи через gRPC метаданные.

## 📝 Описание


