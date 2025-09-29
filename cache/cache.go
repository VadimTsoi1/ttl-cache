package cache

import (
	"sync"
	"time"
)

// Item - единица хранения
// Value - что сохраняем
// ExpireAt - момент времени когда запись считается просроченной
// Если ExpireAt в прошлом - значение должно быть удалено и не возвращаться
type Item struct {
	Value    string
	ExpireAt time.Time
}

// Cache - потокобезопасный in-memory кэш с TTL и фоновым клинапом
// Встроенный map в Go не потокобезопасен, используем мьютекс
type Cache struct {
	mu sync.RWMutex

	items map[string]Item

	// Параметры фоновой очистки
	cleanupInterval time.Duration // как часто запускать чистку
	ticker          *time.Ticker  // будильник
	stopCh          chan struct{} // сигнал остановить фон
	running         bool          // запущена ли фоновая горутина
}

// NewCache - "конструктор", гарантирует, что карта инициализирована
// Если cleanupInterval > 0, можно вызвать c.CleanUp() чтобы включить фон
func NewCache(cleanupInterval time.Duration) *Cache {
	return &Cache{
		items:           make(map[string]Item),
		cleanupInterval: cleanupInterval,
		stopCh:          make(chan struct{}),
	}
}

// Set кладет значение с заданным TTL
// ttl <= 0: запись не хранится(удаляем ключ если он был)
func (c *Cache) Set(key, value string, ttl time.Duration) {
	if ttl <= 0 {
		c.Delete(key)
		return
	}
	exp := time.Now().Add(ttl)

	c.mu.Lock() // запись
	c.items[key] = Item{Value: value, ExpireAt: exp}
	c.mu.Unlock()
}

// Get возвращает (value, true) если ключ существует и не просрочен
// Если запись просрочена - кдалеям и возвращаем ("", false)
func (c *Cache) Get(key string) (string, bool) {
	now := time.Now()

	c.mu.Lock()
	it, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return "", false
	}
	if !it.ExpireAt.IsZero() && now.After(it.ExpireAt) {
		delete(c.items, key)
		c.mu.Unlock()
		return "", false
	}
	val := it.Value
	c.mu.Unlock()
	return val, true
}

// Delete удаляет ключ если он существует
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Exists - проверить наличие живой записи
func (c *Cache) Exists(key string) bool {
	_, ok := c.Get(key)
	return ok
}

// Keys - вернуть текущие ключи
func (c *Cache) Keys() []string {
	c.mu.RLock()
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	c.mu.RUnlock()
	return keys
}

// cleanExpired - разовая очистка просроченных записей
func (c *Cache) cleanExpired() {
	now := time.Now()
	c.mu.Lock()
	for k, it := range c.items {
		if !it.ExpireAt.IsZero() && now.After(it.ExpireAt) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

// CleanUp - запускает фоновую горутину которая раз в cleanupInterval удаляет просроченные записи
func (c *Cache) CleanUp() {
	if c.cleanupInterval <= 0 {
		return // нет интервала - нечего запускать
	}

	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.ticker = time.NewTicker(c.cleanupInterval)
	c.running = true
	c.mu.Unlock()
	// Фонововая горутина слушаем тики и сиггнал остановки
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.cleanExpired()
			case <-c.stopCh:
				c.mu.Lock()
				if c.ticker != nil {
					c.ticker.Stop()
				}
				c.running = false
				c.mu.Unlock()
				return
			}
		}
	}()
}

// Close - корректно останавливает фоновые процессы и совобождает ресурсы
func (c *Cache) Close() {
	c.mu.Lock()
	running := c.running
	c.mu.Unlock()

	if running {
		close(c.stopCh)
		c.mu.Lock()
		c.stopCh = make(chan struct{})
		c.mu.Unlock()
	}
}
