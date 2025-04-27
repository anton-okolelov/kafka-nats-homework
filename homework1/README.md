----
Запускаем nats, билдим go

```bash
docker compose up --build -d
```

Запускаем продюсер
```bash
docker compose run app /app/producer
```

Запускаем консюмер
```bash
docker compose run  app /app/consumer
```
