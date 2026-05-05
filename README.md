# KafkaTSLGTester

CLI-инструмент для нагрузочного тестирования Kafka. Генерирует структурированные JSON-сообщения по шаблонам из `config.yaml` и отправляет их в топики с заданным параллелизмом.

## Возможности

- Генерация JSON-сообщений по декларативным шаблонам полей
- Несколько блоков (`blocks`) с независимыми настройками
- Параллельная отправка через `workers` горутин на блок
- Поддержка TLS (в т.ч. mTLS через PFX/P12-сертификат) и SASL/PLAIN
- Подключение к одному брокеру или кластеру
- Статистика по завершении: количество отправленных сообщений и скорость (msg/sec)
- Graceful shutdown по Ctrl+C / SIGTERM

## Сборка

Требуется Go 1.21+.

```bash
go build -o kafkatsgltest.exe .
```

## Запуск

```bash
# Обычный запуск
kafkatsgltest.exe -config config.yaml

# С подробным логом (уровень DEBUG)
kafkatsgltest.exe -config config.yaml -verbose

# Явное указание конфига
kafkatsgltest.exe -config /path/to/my-config.yaml
```

## Конфигурация

Все настройки находятся в `config.yaml`.

### Подключение к Kafka

```yaml
kafka:
  host: localhost      # хост одного брокера
  port: 9092

  # Кластер — перечислите несколько брокеров (приоритет над host+port)
  # brokers:
  #   - broker1.example.com:9092
  #   - broker2.example.com:9092

  topic: tslg_app_log  # топик по умолчанию для всех блоков

  # Аутентификация SASL/PLAIN (опционально — задайте оба поля)
  user: user
  password: password

  # TLS-транспорт
  secure: true              # true = включить TLS (и SASL, если заданы user/password)
  tls_skip_verify: false    # не проверять сертификат брокера (для dev/self-signed)

  # Truststore — CA-сертификаты для проверки сертификата брокера (PKCS12/P12)
  tls_truststore_file: server.truststore.p12
  tls_truststore_password: "changeit"

  # Keystore — клиентский сертификат + ключ для mTLS (PKCS12/P12, опционально)
  tls_keystore_file: client.keystore.p12
  tls_keystore_password: "changeit"
```

### Параметры по умолчанию

```yaml
defaults:
  batch_size: 5   # сколько сообщений отправить за одну итерацию воркера
  workers: 4      # число горутин на блок
```

### Словарь случайных слов

Используется паттернами `$(random_word_N)` и может использоваться в `$(random_text_N)`.

```yaml
random_words:
  - account
  - select
  - message
  - error
```

### Блоки сообщений

Каждый блок описывает один шаблон сообщений. Блоков может быть несколько.

```yaml
blocks:
  - count: 100        # сколько сообщений сгенерировать
    key: tslg_log     # ключ сообщения Kafka
    topic: my_topic   # переопределить глобальный топик (опционально)
    workers: 4        # переопределить defaults.workers (опционально)
    batch_size: 5     # переопределить defaults.batch_size (опционально)
    fields:
      # ... шаблоны полей ...
```

### Паттерны значений полей

| Паттерн | Пример результата | Описание |
|---|---|---|
| `"статическое значение"` | `"TSLG"` | Значение как есть |
| `$(CURRENT_TIMESTAMP)` | `2026-05-05T10:30:00Z` | Текущее время UTC, без долей секунды |
| `$(CURRENT_TIMESTAMP:3)` | `2026-05-05T10:30:00.123Z` | С миллисекундами |
| `$(CURRENT_TIMESTAMP:6)` | `2026-05-05T10:30:00.123456Z` | С микросекундами |
| `$(CURRENT_TIMESTAMP:9)` | `2026-05-05T10:30:00.123456789Z` | С наносекундами |
| `$(one_of:a,b,c)` | `"b"` | Случайный выбор из списка. Повторяйте варианты для взвешенного выбора |
| `$(num:0to500)` | `"317"` | Случайное целое число в диапазоне [min, max] |
| `$(uuid)` | `"550e8400-e29b-41d4-a716-..."` | UUID v4 |
| `$(random_hex_16)` | `"3f9a1b2c4d5e6f70"` | Случайная hex-строка длиной N |
| `$(random_ip)` | `"192.168.1.42"` | Случайный IPv4-адрес |
| `$(random_text_100)` | `"xKp2mRqZ..."` | N случайных символов (a-zA-Z0-9) |
| `$(random_word_20)` | `"app-loaderwebhookfail"` | Слова из `random_words` слитно, обрезанные до N символов |
| `"prefix-$(uuid)-$(num:1to9)"` | `"prefix-abc...-7"` | Шаблон с несколькими паттернами |

Вложенные YAML-объекты автоматически становятся вложенными JSON-объектами:

```yaml
fields:
  tslgMdc:
    requestId: "$(uuid)"
    userId:    "$(num:1000to9999)"
```

### Полный пример config.yaml

```yaml
kafka:
  host: localhost
  port: 9092
  topic: tslg_app_log
  user: user
  password: password
  secure: true
  tls_skip_verify: false
  tls_truststore_file: server.truststore.p12
  tls_truststore_password: "changeit"
  tls_keystore_file: client.keystore.p12
  tls_keystore_password: "changeit"

defaults:
  batch_size: 5
  workers: 4

random_words:
  - account
  - select
  - message
  - error
  - critical
  - kafka
  - TOPIC
  - ignore
  - summary
  - app-loader
  - loadbalancer
  - webhook
  - fail
  - healthy

blocks:
  - count: 100
    key: tslg_log
    fields:
      projectCode: "TSLG"
      localTime:    "$(CURRENT_TIMESTAMP:3)"
      risCode:      "131"
      appType:      "$(one_of:NET,JAVA,GO)"
      level:        "$(one_of:INFO,INFO,INFO,ERROR,WARNING,DEBUG)"
      appName:      "$(one_of:axdp-client-api,axdp-ldm-streamer,axdp-gateway)"
      levelInt:     "$(num:0to500)"
      text:         "$(random_text_100)"
      namespace:    "dk5-axdp01-axdp-dev1-1"
      source_type:  "$(one_of:k8s,docker,host)"
      hostAddress:  "$(random_ip)"
      hostName:     "$(random_text_20)"
      podName:      "$(random_word_20)"
      traceId:      "$(uuid)"
      eventId:      "$(uuid)"
      messageId:    "$(uuid)"
      callId:       "$(uuid)"
      spanId:       "$(random_hex_16)"
      requestId:    "$(uuid)"
      tslgMdc:
        requestId: "$(uuid)"
        userId:    "$(num:1000to9999)"
      tslgOtherFields:
        loggerName:   "com.example.axdp.Service"
        threadName:   "$(num:1to200)"
        callerMethod: "$(one_of:processRequest,handleEvent,sendResponse,readData,writeLog)"
        callerClass:  "$(one_of:AxdpController,EventHandler,GatewayService,LdmStreamer)"
```

## Пример выходного сообщения

```json
{
  "appName": "axdp-ldm-streamer",
  "appType": "GO",
  "callId": "a1b2c3d4-e5f6-4789-abcd-ef0123456789",
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "hostAddress": "192.168.45.12",
  "hostName": "xKp2mRqZ8nLvBtYw3jQs",
  "level": "INFO",
  "levelInt": "243",
  "localTime": "2026-05-05T10:30:00.123Z",
  "messageId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "namespace": "dk5-axdp01-axdp-dev1-1",
  "podName": "app-loaderwebhookfail",
  "projectCode": "TSLG",
  "requestId": "c0c0a000-dead-4bee-f00d-deadbeef1234",
  "risCode": "131",
  "source_type": "k8s",
  "spanId": "3f9a1b2c4d5e6f70",
  "text": "xKp2mRqZ8nLvBtYw3jQsNfAeHiOuDcPgMbXkWzTrVySqJl...",
  "traceId": "b8d6e4c2-a0f8-4163-9e5b-7d3c1a2b4f6e",
  "tslgMdc": {
    "requestId": "e3d5b7a9-c1e3-4f5a-b2d4-f6a8c0e2d4b6",
    "userId": "4721"
  },
  "tslgOtherFields": {
    "callerClass": "AxdpController",
    "callerMethod": "processRequest",
    "loggerName": "com.example.axdp.Service",
    "threadName": "87"
  }
}
```

## Архитектура

```
main.go
  └─ config.Load()          — разбор YAML
  └─ kafka.NewProducer()    — один AsyncProducer на всё приложение
  └─ engine.Run()
       ├─ block[0]
       │    ├─ worker-0  ─►  builder.Compile()  ─►  producer.Send()
       │    ├─ worker-1  ─►  builder.Compile()  ─►  producer.Send()
       │    └─ worker-N  ─►  builder.Compile()  ─►  producer.Send()
       └─ block[1]
            └─ ...
```

Каждый воркер компилирует поля независимо, получая собственные генераторы со своим источником случайности — гонок данных нет без дополнительных блокировок.

## Тесты

```bash
go test ./...
```
