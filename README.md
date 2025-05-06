# gRPC Pub/Sub Service

## 🚀 Features
- Publisher/Subscriber with multiple subscribers per topic
- gRPC API with `Subscribe` (stream) and `Publish` (unary)
- Graceful shutdown support
- Clean architecture (validation → logic → infra)

## 🛠 Usage

### 1. Сгенерировать gRPC код:
```bash
make proto
