package serviceworker

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ServiceWorkerTemplate шаблон Service Worker
const ServiceWorkerTemplate = `
// PhantomProxy Service Worker
// Версия: 0.2.0

const PROXY_CONFIG = {
    entryPoint: '{{.EntryPoint}}',
    targetParam: '{{.TargetParam}}',
    proxyPathnames: {
        serviceWorker: '{{.SWPath}}',
        script: '{{.ScriptPath}}',
    },
    bypassPaths: ['/health', '/api', '/static'],
};

// Кэширование
const CACHE_NAME = 'phantom-cache-v1';
const OFFLINE_PAGE = '/offline.html';

// Установка SW
self.addEventListener('install', (event) => {
    console.log('[PhantomSW] Install');
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => {
            return cache.addAll([OFFLINE_PAGE]);
        })
    );
    self.skipWaiting();
});

// Активация SW
self.addEventListener('activate', (event) => {
    console.log('[PhantomSW] Activate');
    event.waitUntil(self.clients.claim());
});

// Перехват запросов
self.addEventListener('fetch', (event) => {
    const url = new URL(event.request.url);
    
    // Пропускаем собственные запросы
    if (url.pathname.startsWith(PROXY_CONFIG.proxyPathnames.serviceWorker) ||
        url.pathname.startsWith(PROXY_CONFIG.proxyPathnames.script)) {
        return;
    }
    
    // Пропускаем bypass пути
    for (const bypassPath of PROXY_CONFIG.bypassPaths) {
        if (url.pathname.startsWith(bypassPath)) {
            return;
        }
    }
    
    // Извлекаем целевой URL из параметра
    const targetUrl = url.searchParams.get(PROXY_CONFIG.targetParam);
    
    if (targetUrl) {
        // Проксирование через PhantomProxy
        const proxyUrl = PROXY_CONFIG.entryPoint + '?redirect=' + encodeURIComponent(event.request.url);
        
        event.respondWith(
            fetch(proxyUrl, {
                method: event.request.method,
                headers: event.request.headers,
                body: event.request.body,
                mode: 'cors',
                credentials: 'include'
            }).catch((error) => {
                console.error('[PhantomSW] Fetch error:', error);
                return caches.match(OFFLINE_PAGE);
            })
        );
        return;
    }
    
    // Обычный запрос - пропускаем
    event.respondWith(fetch(event.request));
});

// Обработка сообщений от страницы
self.addEventListener('message', (event) => {
    console.log('[PhantomSW] Message received:', event.data);
    
    if (event.data && event.data.type === 'SKIP_WAITING') {
        self.skipWaiting();
    }
    
    if (event.data && event.data.type === 'GET_CONFIG') {
        event.ports[0].postMessage(PROXY_CONFIG);
    }
});
`

// InjectionConfig конфигурация для инъекции
type InjectionConfig struct {
	EntryPoint string
	TargetParam string
	SWPath     string
	ScriptPath string
}

// Injector инъекция Service Worker
type Injector struct {
	config *InjectionConfig
	swTemplate *template.Template
	scriptTemplate *template.Template
}

// NewInjector создаёт новый Injector
func NewInjector(entryPoint, targetParam string) *Injector {
	config := &InjectionConfig{
		EntryPoint:  entryPoint,
		TargetParam: targetParam,
		SWPath:      "/sw.js",
		ScriptPath:  "/phantom.js",
	}
	
	swTemplate, _ := template.New("sw").Parse(ServiceWorkerTemplate)
	
	return &Injector{
		config:         config,
		swTemplate:     swTemplate,
		scriptTemplate: nil,
	}
}

// GenerateSW генерирует Service Worker JavaScript
func (i *Injector) GenerateSW() (string, error) {
	var buf strings.Builder
	err := i.swTemplate.Execute(&buf, i.config)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// InjectHTML внедряет Service Worker в HTML
func (i *Injector) InjectHTML(html []byte) []byte {
	// Поиск </body> для инъекции
	bodyTag := []byte("</body>")
	idx := lastIndex(html, bodyTag)
	
	if idx == -1 {
		return html
	}
	
	// Генерация скрипта регистрации
	script := i.generateRegistrationScript()
	
	// Вставка перед </body>
	result := make([]byte, len(html)+len(script))
	copy(result, html[:idx])
	copy(result[idx:], script)
	copy(result[idx+len(script):], bodyTag)
	
	return result
}

// generateRegistrationScript генерирует скрипт регистрации SW
func (i *Injector) generateRegistrationScript() string {
	return fmt.Sprintf(`
<script id="phantom-sw-register">
(function() {
    if ('serviceWorker' in navigator) {
        console.log('[PhantomSW] Service Worker supported');
        
        // Регистрация Service Worker
        navigator.serviceWorker.register('%s')
            .then(function(registration) {
                console.log('[PhantomSW] Registered:', registration.scope);
                
                // Отправка сообщения о готовности
                registration.active.postMessage({
                    type: 'SW_READY',
                    config: {
                        entryPoint: '%s',
                        targetParam: '%s'
                    }
                });
            })
            .catch(function(error) {
                console.log('[PhantomSW] Registration failed:', error);
            });
            
        // Контроль обновлений
        navigator.serviceWorker.ready.then(function(registration) {
            registration.addEventListener('updatefound', function() {
                console.log('[PhantomSW] Update found');
                var newWorker = registration.installing;
                newWorker.addEventListener('statechange', function() {
                    if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                        // Новая версия доступна
                        console.log('[PhantomSW] New version available');
                    }
                });
            });
        });
    } else {
        console.log('[PhantomSW] Service Worker NOT supported');
    }
})();
</script>`, i.config.SWPath, i.config.EntryPoint, i.config.TargetParam)
}

// InjectScript возвращает JavaScript для внедрения в страницу
func (i *Injector) InjectScript() string {
	return `
// PhantomProxy Client Script
(function() {
    console.log('[PhantomJS] Loaded');
    
    // Проверка поддержки Service Worker
    if ('serviceWorker' in navigator) {
        // Получение конфигурации от SW
        navigator.serviceWorker.ready.then(function(registration) {
            return new Promise(function(resolve) {
                var channel = new MessageChannel();
                channel.port1.onmessage = function(event) {
                    resolve(event.data);
                };
                registration.active.postMessage({
                    type: 'GET_CONFIG'
                }, [channel.port2]);
            });
        }).then(function(config) {
            console.log('[PhantomJS] SW Config:', config);
        });
    }
    
    // Перехват form submit для добавления параметра
    document.addEventListener('submit', function(e) {
        var form = e.target;
        if (form.tagName === 'FORM') {
            // Добавление скрытого поля с меткой
            var hiddenField = document.createElement('input');
            hiddenField.type = 'hidden';
            hiddenField.name = '_phantom';
            hiddenField.value = '1';
            form.appendChild(hiddenField);
        }
    }, true);
    
    // Логирование навигации
    window.addEventListener('beforeunload', function() {
        console.log('[PhantomJS] Page unload');
    });
})();
`
}

// HandleSWRequest обрабатывает запрос Service Worker
func (i *Injector) HandleSWRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == i.config.SWPath {
		sw, err := i.GenerateSW()
		if err != nil {
			http.Error(w, "Failed to generate SW", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Write([]byte(sw))
		return
	}
	
	if r.URL.Path == i.config.ScriptPath {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write([]byte(i.InjectScript()))
		return
	}
	
	http.NotFound(w, r)
}

// ShouldInject проверяет, нужно ли внедрять SW
func (i *Injector) ShouldInject(r *http.Request) bool {
	// Проверка HTTPS (SW требует безопасного контекста)
	if r.TLS == nil {
		// Исключение для localhost
		if r.Host == "localhost" || strings.HasPrefix(r.Host, "127.0.0.1") {
			return true
		}
		return false
	}
	
	// Проверка User-Agent на поддержку SW
	ua := r.UserAgent()
	if strings.Contains(ua, "MSIE") || strings.Contains(ua, "Trident") {
		return false // IE не поддерживает SW
	}
	
	return true
}

// lastIndex находит последнее вхождение substr в s
func lastIndex(s, substr []byte) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if string(s[i:i+len(substr)]) == string(substr) {
			return i
		}
	}
	return -1
}

// GetConfig возвращает конфигурацию
func (i *Injector) GetConfig() *InjectionConfig {
	return i.config
}

// UpdateEntryPoint обновляет точку входа
func (i *Injector) UpdateEntryPoint(entryPoint string) {
	i.config.EntryPoint = entryPoint
}
