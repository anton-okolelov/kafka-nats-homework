Запустить кафку (KRAFT, 3 брокера)
```bash
docker compose up -d
```

 Создать топки clicks_v1
 
```bash
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
    --create \
    --topic clicks_v1 \
    --partitions 5 \
    --replication-factor 3 \
    --bootstrap-server broker1:9092
```

Посмотреть, где лежат реплики и т.д.
```bash
docker compose exec broker1 /opt/kafka/bin/kafka-topics.sh \
--describe \
--topic clicks_v1 \
--bootstrap-server broker1:9092
```

```
Topic: clicks_v1        TopicId: GrE_Bxj-R1qGr4RiS8S6Eg PartitionCount: 5       ReplicationFactor: 3    Configs: 
        Topic: clicks_v1        Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3,1,2      Elr:    LastKnownElr: 
        Topic: clicks_v1        Partition: 1    Leader: 1       Replicas: 1,2,3 Isr: 1,2,3      Elr:    LastKnownElr: 
        Topic: clicks_v1        Partition: 2    Leader: 2       Replicas: 2,3,1 Isr: 2,3,1      Elr:    LastKnownElr: 
        Topic: clicks_v1        Partition: 3    Leader: 2       Replicas: 2,3,1 Isr: 2,3,1      Elr:    LastKnownElr: 
        Topic: clicks_v1        Partition: 4    Leader: 3       Replicas: 3,1,2 Isr: 3,1,2      Elr:    LastKnownElr: 
```

Запустить продюсер
```bash
go run cmd/producer/main.go
```

```
➜  homework2 git:(master) ✗ go run cmd/producer/main.go
time=2025-04-27T19:55:58.588+02:00 level=INFO msg="starting producer"
time=2025-04-27T19:55:58.589+02:00 level=INFO msg="producing message" message="{SiteId:8 PriceCents:600 ClickedAt:2025-04-27T19:55:58+02:00}"
time=2025-04-27T19:55:59.624+02:00 level=INFO msg="producing message" message="{SiteId:6 PriceCents:310 ClickedAt:2025-04-27T19:55:59+02:00}"
time=2025-04-27T19:56:00.630+02:00 level=INFO msg="producing message" message="{SiteId:4 PriceCents:914 ClickedAt:2025-04-27T19:56:00+02:00}"
time=2025-04-27T19:56:01.640+02:00 level=INFO msg="producing message" message="{SiteId:7 PriceCents:43 ClickedAt:2025-04-27T19:56:01+02:00}"

```

Запустить консюмер
```bash
go run cmd/consumer/main.go
```

```
➜  homework2 git:(master) ✗ go run cmd/consumer/main.go
time=2025-04-27T19:56:08.520+02:00 level=INFO msg="Recieved message" object="{SiteId:3 PriceCents:474 ClickedAt:2025-04-27T19:55:25+02:00}"
time=2025-04-27T19:56:08.523+02:00 level=INFO msg="Recieved message" object="{SiteId:2 PriceCents:611 ClickedAt:2025-04-27T19:55:26+02:00}"
time=2025-04-27T19:56:08.523+02:00 level=INFO msg="Recieved message" object="{SiteId:2 PriceCents:504 ClickedAt:2025-04-27T19:55:27+02:00}"
time=2025-04-27T19:56:08.523+02:00 level=INFO msg="Recieved message" object="{SiteId:6 PriceCents:41 ClickedAt:2025-04-27T19:55:28+02:00}"
```

Прибить один из брокеров
```bash
docker compose stop broker2
```

Прибить еще один брокер (majority)
```bash
docker compose stop broker1
```

