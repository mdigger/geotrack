package lbs

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/mdigger/geotrack/geo"
)

// Filters описывает фильтры для выборки информации при импорте. Только данные,
// совпадающие с данным фильром будут импортированы.
type Filters struct {
	Radio   map[string]bool // типы сетей
	Country map[uint16]bool // коды стран
}

type Data struct {
	Point     *geo.Point // координаты
	Timestamp time.Time  // временная метка
}

// ImportCSV импортирует данные о вышках сотовых станций и их координта из CSV-файла
// в хранилище.
func (db *DB) ImportCSV(filename string, filter *Filters) {
	coll := db.GetCollection(CollectionName)
	defer db.FreeCollection(coll)
	bulk := coll.Bulk()
	bulk.Unordered()

	log.Printf("Reading data from CSV %q...", filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	var (
		counter, lines uint32       // счетчик
		timestamp      = time.Now() // bson.NewObjectIdWithTime(time.Now())
	)
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error parsing CSV file:", err)
			return
		}
		lines++
		if lines == 1 {
			r.FieldsPerRecord = len(record) // устанавливаем количество полей
			continue                        // пропускаем первую строку с заголовком в CSV-файле
		}

		radio := record[0]
		if filter != nil && !filter.Radio[radio] {
			continue // игнорируем записи с неподдерживаемым типом радио
		}
		mcc, err := strconv.ParseUint(record[1], 10, 16)
		if err != nil {
			log.Printf("[%d] bad MCC: %s", lines, record[1])
			continue
		}
		if filter != nil && !filter.Country[uint16(mcc)] {
			continue // игнорируем записи с неподдерживаемым типом радио
		}
		mnc, err := strconv.ParseUint(record[2], 10, 32)
		if err != nil {
			log.Printf("[%d] bad MNC: %s", lines, record[2])
			continue
		}
		area, err := strconv.ParseUint(record[3], 10, 16)
		if err != nil {
			log.Printf("[%d] bad Area: %s", lines, record[3])
			continue
		}
		cell, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			log.Printf("[%d] bad Cell: %s", lines, record[4])
			continue
		}
		lon, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Printf("[%d] bad longitude:", lines, record[6])
			continue
		}
		lat, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			log.Printf("[%d] bad latitude:", lines, record[7])
			continue
		}

		key := Key{
			Radio: radio,
			MCC:   uint16(mcc),
			MNC:   uint32(mnc),
			Area:  uint16(area),
			Cell:  uint32(cell),
		}
		data := Data{
			Point:     geo.NewPoint(lon, lat),
			Timestamp: timestamp,
		}
		bulk.Upsert(key, bson.M{"$set": data})
		counter++
		fmt.Printf("\r* Find %d from %d records ", counter, lines-1)
	}
	fmt.Printf("\r")

	if counter == 0 {
		log.Println("No record for import. Exit...")
		return
	}
	log.Printf("Bulk importing to MongoDB [%d records]...", counter)
	bulkResult, err := bulk.Run()
	if err != nil {
		log.Println("MongoDB bulk insert error:", err)
		return
	}
	if bulkResult.Modified > 0 {
		log.Printf("Modified %d records", bulkResult.Modified)
	}

	log.Println("Deleting old data...")
	deleteResult, err := coll.RemoveAll(bson.M{"_id": bson.M{"$lt": timestamp}})
	if err != nil {
		log.Println("MongoDB deleting old data error:", err)
		return
	}
	if deleteResult.Removed > 0 {
		log.Printf("Deleted %d records", deleteResult.Removed)
	}
	total, err := coll.Count()
	if err != nil {
		log.Println("MongoDB total counting error:", err)
		return
	}
	log.Printf("Total records in DB: %d", total)
}

// RecordsCount возвращает количество записей в хранилище LBS.
func (db *DB) RecordsCount() int {
	coll := db.GetCollection(CollectionName)
	defer db.FreeCollection(coll)
	total, _ := coll.Count()
	return total
}
