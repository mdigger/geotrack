package pairing

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Dictionary описывает словарь символов, из которых может генерироваться код для активации.
// Для простоты я не стал проверять словарь на то, что он содержит юникодные символы, описывающиеся
// несколькими байтами, поэтому для корректной работы рекомендуется, чтобы в словаре использовались
// только печатные ASCII символы.
type Dictionary string

// Generate возвращает случайный набор символов из словаря заданной длинны.
func (d Dictionary) Generate(length int) string {
	response := make([]byte, length)
	for i := range response {
		response[i] = d[rand.Intn(len(d))] // заполняем случайным набором из словаря
	}
	return string(response)
}

// Предопределенные словари для генерации уникальных кодов активации.
var (
	DictNumber  Dictionary = "0123456789"                              // только цифры
	DictAlfa               = DictNumber + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // цифры и буквы
	DictDefault            = DictNumber                                // словарь по умолчанию
)
