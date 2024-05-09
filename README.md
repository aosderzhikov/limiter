# limiter

## Использование 

При инициализации лимитера ему необходимо указать драйвер хранилища, это место, где лимитер будет смотреть текущий счетчик и обновлять его.
По умолчанию используется драйвер для работы в памяти приложения. Также можно подключить драйвер работающий по `http` из пакета `driver`.
Есть возможность реализовать свой драйвер, который будет работать с `redis`, `postgres` и тд. Для этого драйвер должен реализовывать интерфейс `CounterStorage`.


```go
	defDriver := DefaultCounterStorage(3, 1*time.Second)
	l := NewLimiter(defDriver)

	var err error
    var resp http.Response
	for i := range 4 {
		err = l.Do(nil, func() error {
			// выполняем необходимый запрос/действие
            resp, err := client.Do(req)
            if err != nil {
                return err
            }
            return nil
		})
	}
	if err != nil {
        if errors.Is(err, ErrLimitExceed) {
            // ошибка превышения лимита
        }

		// ошибка не связана с превышением лимита 
	}
```

Если требуется больше контроля, то можно работать напрямую с драйвером:

```go
    allowed, err := l.Allowed(ctx)
    if err != nil {
        // ...
    }
    if !allowed {
        // лимит превышен
    }

    // лимит не превышен делаем действие
    resp, err := client.Do(req)
    if err != nil {
        // ...
    }
    // действие успешно выполнено инкрементим
    err = l.Increment(ctx)
    if err != nil {
        // не получилось инкерементировать, обрабатываем
    }
    //
```