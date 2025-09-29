ttl-cache (Go)

Потокобезопасный in-memory кэш с TTL и фоновой очисткой.

Репозиторий: github.com/VadimTsoi1/ttl-cache
Основной пакет: cache
Подход: map[string]Item + sync.RWMutex + time.Ticker (фон) + явный Close()
Поведение: просроченные записи удаляются «на чтении» (Get) и периодически фоном (Cleanup())

Возможности
Set(key, value, ttl) - записать значение с TTL
Get(key) - получить значение, удаляя просроченное
Delete(key) - удалить ключ
Cleanup() - включить фоновую очистку по интервалу
Close() - корректно остановить фон
Бонусы: Exists(key), Keys()

Установка
go get github.com/VadimTsoi1/ttl-cache

API
// Создать кэш. Если cleanupInterval > 0 - можно запустить Cleanup() для фоновой очистки.
func NewCache(cleanupInterval time.Duration) *Cache

// Записать значение с TTL. Политика: ttl <= 0 -> запись не хранится (ключ удаляется).
func (c *Cache) Set(key, value string, ttl time.Duration)

// Получить значение. Если запись просрочена - она удаляется и возвращается ("", false).
func (c *Cache) Get(key string) (string, bool)

// Удалить ключ.
func (c *Cache) Delete(key string)

// Запустить фоновую чистку (идемпотентно: повторный вызов ничего не делает).
func (c *Cache) Cleanup()

// Остановить фон. После Close() можно снова вызвать Cleanup().
func (c *Cache) Close()

// Бонусы
func (c *Cache) Exists(key string) bool
func (c *Cache) Keys() []string

Дизайн-решения
Потокобезопасность: sync.RWMutex защищает map.
Операции записи (Set, Delete) -> Lock/Unlock.
Get использует Lock, а не RLock, потому что может удалять протухшие записи (это запись).
TTL: храним ExpireAt time.Time.
В Set: ExpireAt = time.Now().Add(ttl).
В Get: если now > ExpireAt -> удалить и вернуть false.
Политика: ttl <= 0 - запись не хранится (ключ удаляется). Простое и предсказуемое поведение.
Фоновая очистка: Cleanup() запускает горутину с time.Ticker, которая периодически вызывает cleanExpired().
Close() останавливает тикер и завершает фон.

Тесты
go test ./...

Структура проекта
ttl-cache/
  cache/
    cache.go        # реализация
    cache_test.go   # базовые тесты
  main.go           # пример запуска 
  README.md
  go.mod