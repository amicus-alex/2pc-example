Запуск
```
docker-compose up
```

Тест:
1. створення кастомера с балансом 100
```
curl -X POST http://localhost:8081/customer/create \  
  -H "Content-Type: application/json" \
  -d '{"user_id":"123", "balance":100}'

```
2. Успішний ордер на 10
```
curl -X POST http://localhost:8080/coordinator/order \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123", "amount":10}'

```
3. Неуспішний ордер на 100
```
curl -X POST http://localhost:8080/coordinator/order \
  -H "Content-Type: application/json" \
  -d '{"user_id":"123", "amount":100}'

```
