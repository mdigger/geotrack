package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mdigger/geolocate"
)

// ParseLBS разбирает строку с информацией в формате LBS и возвращает его описание.
// Первым параметром указывается тип радио (gsm, lte, cdma, wcdam и так далее). Вторым — строка
// с данными LBS. В ответ возвращает сформированную для запроса данных структуру.
func ParseLBS(radio, lbsStr string) (*geolocate.Request, error) {
	switch radio {
	case "", "gsm", "lte", "cdma", "wcdma":
	default:
		return nil, fmt.Errorf("bad radio type: %s", radio)
	}
	splitted := strings.Split(lbsStr, "-") // разделяем на элементы
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
	cellTowers := make([]geolocate.CellTower, (len(splitted)-5)/3)
	for i := range cellTowers {
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
		cellTowers[i] = geolocate.CellTower{
			CellId:            uint32(id),
			LocationAreaCode:  uint16(area),
			MobileCountryCode: uint16(mcc),
			MobileNetworkCode: uint16(mnc),
			SignalStrength:    int16(dbm - 220),
		}
	}
	return &geolocate.Request{
		RadioType:             radio,
		HomeMobileCountryCode: uint16(mcc),
		HomeMobileNetworkCode: uint16(mnc),
		ConsiderIp:            false,
		CellTowers:            cellTowers,
	}, nil
}
