package pairing

import (
	"sync"
	"time"
)

var (
	KeyLength    = 4                // длина ключа
	KeyExpired   = time.Minute * 30 // время жизни ключа
	MaxIterCount = 1000             // максимальное количество попыток генерации нового ключа
)

const initialCount = 100 // изначально выделяем память для хранения стольких одновременных ключей

// keyInfo содержит информацию об устройстве и времени генерации ключа.
type keyInfo struct {
	DeviceID string    // уникальный идентификатор устройства
	Key      string    // уникальный ключ
	Time     time.Time // время генерации ключа
}

// Pairs описывает список ключей для спаривания устройств.
type Pairs struct {
	Dictionary                     // словарь букв ключа для генерации
	devices    map[string]*keyInfo // справочник ключей для устройств
	keys       map[string]*keyInfo // справочник устройств по сгенерированным ключам
	mu         sync.Mutex
}

// Generate возвращает новый уникальный ключ для спаривания устройства.
//
// Если ключ для этого устройства уже был сгенерирован, то старый ключ удаляется и становится
// не действительным, а создается новый ключ, привязанный к этому устройству. Так же автоматически
// удаляются те ключи, которые уже устарели. Если новый ключ не удается получить за заданное
// количество попыток, то возвращается пустое значение ключа, так что необходима проверка.
//
// Параллельное выполнение нескольких функций генерации блокируется. Но, т.к. это достаточно
// быстрый процесс, то обычно это никак не сказывается на производительности.
func (p *Pairs) Generate(deviceID string) (key string) {
	p.mu.Lock() // одновременно выполняется только одна копия
	// инициализируем списки ключей и словарь, если они не были инициализированы до этого
	if p.devices == nil {
		p.devices = make(map[string]*keyInfo, initialCount)
	}
	if p.keys == nil {
		p.keys = make(map[string]*keyInfo, initialCount)
	}
	if len(p.Dictionary) == 0 {
		p.Dictionary = DictDefault // инициализируем словарь, если он не инициализирован
	}
	// проверяем, что для данного устройства нет сгенерированного ключа
	if kInfo, ok := p.devices[deviceID]; ok {
		delete(p.keys, kInfo.Key) // удаляем ключ из списка
		delete(p.devices, kInfo.DeviceID)
		// log.Printf("Delete key for %q", deviceID)
	}
	// делаем несколько попыток генерации нового уникального ключа
	for i := 0; i < MaxIterCount; i++ {
		key = p.Dictionary.Generate(KeyLength) // генерируем случайный ключ по словарю
		// проверяем, что этот ключ сейчас не используется
		if kInfo, ok := p.keys[key]; ok {
			if time.Since(kInfo.Time) < KeyExpired {
				continue // время жизни ключа еще не истекло — пробуем дальше
			}
			// ключ используется, но устарел — удаляем записи о нем
			delete(p.keys, kInfo.Key) // удаляем ключ из списка
			delete(p.devices, kInfo.DeviceID)
			// log.Printf("Delete expired key %q", key)
		}
		// сгенерированный ключ можно использовать как новый
		kInfo := &keyInfo{
			DeviceID: deviceID,
			Key:      key,
			Time:     time.Now(),
		}
		// заносим его в справочник ключей для устройств
		p.devices[deviceID] = kInfo
		p.keys[key] = kInfo
		// log.Printf("Add new key %q for device %q", key, deviceID)
		break
	}
	p.mu.Unlock()
	return
}

// GetDeviceID возвращает уникальный идентификатор устройства, связанный с указанным ключем
// активации. При этом запись об этом устройстве из базы удаляется. Если такого устройства не
// найдено или ключ просрочен, то возвращается пустая строка.
func (p *Pairs) GetDeviceID(key string) (deviceID string) {
	p.mu.Lock()
	if kInfo, ok := p.keys[key]; ok {
		delete(p.keys, kInfo.Key)
		delete(p.devices, kInfo.DeviceID)
		if time.Since(kInfo.Time) < KeyExpired {
			deviceID = kInfo.DeviceID
		}
	}
	p.mu.Unlock()
	return
}
