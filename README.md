# Практическая работа 2: UDP RTT

Go-программы разделены на `protocol` (пакеты), `transport` (UDP-сокет) и `telemetry` (метрики).

Запуск одной серии в двух терминалах:

```powershell
go run -buildvcs=false ./cmd/server -mode baseline
go run -buildvcs=false ./cmd/client -series baseline -count 50 -interval 200ms
```

Для других серий замените `baseline` на `delay_50`, `delay_100`, `jitter`, `loss_5` или `combined`. Между сериями перезапускайте сервер. Клиент дописывает измерения в `docs/latency_samples.csv`; перед полным повтором используйте новый путь через `-csv`, чтобы не смешать серии.

Тесты: `go test -buildvcs=false ./...`. Графики и отчёт из CSV: `go run -buildvcs=false ./cmd/report`.
