package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime)
	mongourl := flag.String("mongo", "mongodb://localhost/watch", "mongoDB connection URL")
	radiofilter := flag.String("radio", "gsm", "filter for radio (comma separated)")
	countryfilter := flag.String("country", "250", "filter for country (comma separated)")
	minSamples := flag.Int64("minsample", 0, "filter for min samples count")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Import LBS database data\n")
		fmt.Fprintf(os.Stderr, "%s [-params] datafile.csv\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		return
	}
	filename := flag.Arg(0)

	// устанавливаем соединение с сервером MongoDB
	log.Printf("Connecting to MongoDB %q...", *mongourl)
	mdb, err := mongo.Connect(*mongourl)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return
	}
	defer mdb.Close()

	db, err := lbs.InitDB(mdb)
	if err != nil {
		log.Printf("Error initializing indexes: %v", err)
		return
	}
	coll := db.GetCollection(lbs.CollectionName)
	defer db.FreeCollection(coll)

	bulk := coll.Bulk()
	bulk.Unordered()

	// разбираем фильтры и формируем соответствующие справочники
	var (
		filterRadio   = make(map[string]bool)
		filterCountry = make(map[uint16]bool)
	)
	for _, radio := range strings.Split(*radiofilter, ",") {
		filterRadio[strings.ToLower(strings.TrimSpace(radio))] = true
	}
	for _, country := range strings.Split(*countryfilter, ",") {
		mcc, err := strconv.ParseUint(country, 10, 16)
		if err != nil {
			continue
		}
		filterCountry[uint16(mcc)] = true
	}
	if len(filterRadio) > 0 || len(filterCountry) > 0 {
		log.Printf("Filters country - %q, radio - %q",
			strings.Join(strings.Split(*countryfilter, ","), ", "),
			strings.Join(strings.Split(*radiofilter, ","), ", "))
	}

	log.Printf("Reading data from CSV %q...", filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		return
	}
	defer file.Close()

	var counter, lines uint64 // счетчики
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error parsing CSV file: %v", err)
			return
		}
		lines++
		if lines == 1 {
			r.FieldsPerRecord = len(record) // устанавливаем количество полей
			continue                        // пропускаем первую строку с заголовком в CSV-файле
		}
		fmt.Fprintf(os.Stderr, "\r* find %8d | skipped %8d records ", counter, lines-1-counter)

		radio := strings.ToLower(record[0])
		if len(filterRadio) > 0 && !filterRadio[radio] {
			continue // игнорируем записи с неподдерживаемым типом радио
		}
		samples, err := strconv.ParseInt(record[9], 10, 32)
		if err != nil {
			log.Printf("[%d] bad Samples: %s", lines, record[9])
			continue
		}
		if samples < *minSamples {
			continue // не импортируем данные с маленьким количеством подтверждений
		}
		mcc, err := strconv.ParseUint(record[1], 10, 16)
		if err != nil {
			log.Printf("[%d] bad MCC: %s", lines, record[1])
			continue
		}
		if len(filterCountry) > 0 && !filterCountry[uint16(mcc)] {
			continue // игнорируем записи с неподдерживаемым типом радио
		}
		mnc, err := strconv.ParseUint(record[2], 10, 16)
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
		distance, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			log.Printf("[%d] bad range:", lines, record[8])
			continue
		}
		key := lbs.Key{
			RadioType:         radio,
			MobileCountryCode: uint16(mcc),
			MobileNetworkCode: uint16(mnc),
			LocationAreaCode:  uint16(area),
			CellId:            uint32(cell),
		}
		data := lbs.Data{
			Location: geo.NewPoint(lon, lat),
			Accuracy: distance,
		}

		// created, err := strconv.ParseInt(record[11], 10, 64)
		// if err != nil {
		// 	log.Printf("[%d] bad Created: %s", lines, record[11])
		// 	continue
		// }
		// updated, err := strconv.ParseInt(record[12], 10, 64)
		// if err != nil {
		// 	log.Printf("[%d] bad Updated: %s", lines, record[12])
		// 	continue
		// }

		bulk.Upsert(key, bson.M{"$set": data})
		counter++
	}
	fmt.Fprintln(os.Stderr, "")

	if counter == 0 {
		log.Println("No record for import. Exit...")
		return
	}

	// если это не обновление, то подчищаем старые (не обновленные) данные
	if !strings.Contains(filename, "diff") {
		log.Println("Deleting old data...")
		deleteResult, err := coll.RemoveAll(nil)
		if err != nil {
			log.Printf("MongoDB deleting old data error: %v", err)
			return
		}
		if deleteResult.Removed > 0 {
			log.Printf("Deleted %d records", deleteResult.Removed)
		}
	}

	log.Printf("Bulk importing to MongoDB [%d records]...", counter)
	bulkResult, err := bulk.Run()
	if err != nil {
		log.Printf("MongoDB bulk insert error: %v", err)
		return
	}
	if bulkResult.Modified > 0 {
		log.Printf("Modified %d records", bulkResult.Modified)
	}

	total, err := coll.Count()
	if err != nil {
		log.Printf("MongoDB total counting error: %v", err)
		return
	}
	log.Printf("Total unique records in DB: %d", total)
}
