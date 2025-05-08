----
Запускаем 

```bash
docker compose up --build -d
```

в контейнерах есть паб/саб, которые общаются между собой
```
 docker compose logs -f sub
sub-1  | Starting...
sub-1  | Слушаю топик pubsub
sub-1  | Консюмер получил сообщение: Producer Id: 41, message: 1
sub-1  | Консюмер получил сообщение: Producer Id: 41, message: 2
sub-1  | Консюмер получил сообщение: Producer Id: 41, message: 3
sub-1  | Консюмер получил сообщение: Producer Id: 41, message: 4
```


и request/reply
```
✗ docker compose logs -f request
request-1  | Starting...
request-1  | Producer Id: 21, request: 1
request-1  | Producer 21 received reply: Otvet
request-1  | Producer Id: 21, request: 2
request-1  | Producer 21 received reply: Otvet
request-1  | Producer Id: 21, request: 3
request-1  | Producer 21 received reply: Otvet
request-1  | Producer Id: 21, request: 4
request-1  | Producer 21 received reply: Otvet
request-1  | Producer Id: 21, request: 5
request-1  | Producer 21 received reply: Otvet
```
