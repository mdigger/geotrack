package lbs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Cell описывает информацию о базовой станции и уровне сигнала.
type Cell struct {
	Area uint16 // the base station cell number
	ID   uint32 // base station number
	DBM  int16  // signal strength ((dbm + 110 = rxlev + 110 = watch sign strength)
}

// Request описывает информацию о запросе в формате LBS.
type Request struct {
	MCC   uint16 // country code  (250 - Россия, 255 - Украина, Беларусь - 257)
	MNC   uint32 // operator code
	Cells []*Cell
}

// Parse разбирает строку с информацией в формате LBS и возвращает его описание.
func Parse(s string) (*Request, error) {
	splitted := strings.Split(s, "-") // разделяем на элементы
	if len(splitted) < 7 {
		return nil, errors.New("agps - wrong data (len < 7)")
	}
	mcc, err := strconv.ParseUint(splitted[3], 16, 16)
	if err != nil {
		return nil, fmt.Errorf("bad MCC: %s", splitted[3])
	}
	mnc, err := strconv.ParseUint(splitted[4], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("bad MNC: %s", splitted[4])
	}
	cells := make([]*Cell, (len(splitted)-5)/3)
	for i := range cells {
		area, err := strconv.ParseUint(splitted[5+i*3], 16, 16)
		if err != nil {
			return nil, fmt.Errorf("bad Area: %s", splitted[5+i*3])
		}
		id, err := strconv.ParseUint(splitted[6+i*3], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("bad Cell ID: %s", splitted[6+i*3])
		}
		dbm, err := strconv.ParseUint(splitted[7+i*3], 16, 16)
		if err != nil {
			return nil, fmt.Errorf("bad DBM: %s", splitted[7+i*3])
		}
		cells[i] = &Cell{
			Area: uint16(area),
			ID:   uint32(id),
			DBM:  int16(dbm - 220),
		}
	}
	return &Request{
		MCC:   uint16(mcc),
		MNC:   uint32(mnc),
		Cells: cells,
	}, nil
}
