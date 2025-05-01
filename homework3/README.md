Homework3
======

В дз я сделал систему, которая обрабатывает информацию о рекламных баннерах (топик с кликами, топик с показами), джойнит их через kafka streams (по advertising_id) и выдает CTR (отношение числа кликов к числу показов) в другой топик с неким окном подсчета

## запуск кафки
```bash
docker compose up -d
```

## создание топиков

```bash
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
    --create \
    --topic clicks_v1 \
    --partitions 5 \
    --replication-factor 3 \
    --bootstrap-server broker1:9092
  
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
    --create \
    --topic views_v1 \
    --partitions 5 \
    --replication-factor 3 \
    --bootstrap-server broker1:9092
    
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
    --create \
    --topic ctr_v1 \
    --partitions 5 \
    --replication-factor 3 \
    --bootstrap-server broker1:9092
```

## запуск продюсеров
```bash
cd producers/cmd/producer
go run main.go
```

## запуск процессора
```bash
./gradlew run
```

## пример вывода:
```
Sending to ctr_v1 topic: Key=4, Value=CTRResult{advertisingId='4', clicks=5, views=19, ctr=0.2631578947368421, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=3, Value=CTRResult{advertisingId='3', clicks=2, views=18, ctr=0.1111111111111111, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=2, Value=CTRResult{advertisingId='2', clicks=1, views=25, ctr=0.04, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=6, Value=CTRResult{advertisingId='6', clicks=1, views=21, ctr=0.047619047619047616, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=9, Value=CTRResult{advertisingId='9', clicks=2, views=22, ctr=0.09090909090909091, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=1, Value=CTRResult{advertisingId='1', clicks=1, views=17, ctr=0.058823529411764705, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=1, Value=CTRResult{advertisingId='1', clicks=1, views=17, ctr=0.058823529411764705, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=7, Value=CTRResult{advertisingId='7', clicks=1, views=17, ctr=0.058823529411764705, windowStart=1746079020000, windowEnd=1746079040000}
Sending to ctr_v1 topic: Key=8, Value=CTRResult{advertisingId='8', clicks=4, views=12, ctr=0.3333333333333333, windowStart=1746079020000, windowEnd=1746079040000}

```




