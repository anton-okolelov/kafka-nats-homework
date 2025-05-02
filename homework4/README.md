Homework4
====

Поднять контейнеры
```bash
docker compose up --build -d
```


Создать топик transactions

```bash
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
    --create \
    --topic transactions \
    --partitions 5 \
    --replication-factor 3 \
    --bootstrap-server broker1:9092
```


Создать транзакцию (с плюсом или минусом)
```
curl -X POST \
  http://localhost:8080/transactions \
    -H 'Content-Type: application/json' \
    -d '{
        "account_id": 1,
        "transaction_id": "f3f60dae-ac2a-4bb1-a67b-792b16b5f596",
        "amount": 100.0,
        "transaction_date": "2025-04-27T19:55:58+02:00"
    }'
   
```

Сравнить балансы (watch)

первая база
```bash
echo "select * from balances; \n\watch 1" | PGPASSWORD=pass psql -h 127.0.0.1 -p 5432 -U user db 
```

вторая база
```bash
echo "select * from balances; \n\watch 1" | PGPASSWORD=pass psql -h 127.0.0.1 -p 5433 -U user db 
```


```
 account_id | balance 
------------+---------
          1 |     600
          2 |     200
     100500 |     100
     100501 |    -200
(4 rows)

```