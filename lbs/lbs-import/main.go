package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
)

type Data struct {
	Point     geo.Point // координаты
	Timestamp time.Time // временная метка
}

func main() {
	log.SetFlags(log.Ltime)
	filename := flag.String("csvfile", "cell_towers.csv", "csv file with data")
	mongourl := flag.String("mongo", "mongodb://localhost/watch", "mongoDB connection URL")
	radiofilter := flag.String("radio", "GSM,UMTS", "filter for radio (comma separated)")
	countryfilter := flag.String("country", "250,255,257", "filter for country (comma separated)")
	flag.Parse()

	// устанавливаем соединение с сервером MongoDB
	log.Print("Connecting to MongoDB...")
	mdb, err := mongo.Connect(*mongourl)
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	db, err := lbs.InitDB(mdb)
	if err != nil {
		log.Println("Error initializing indexes:", err)
		return
	}

	// разбираем фильтры и формируем соответствующие справочники
	var filter = &lbs.Filters{
		Radio:   make(map[string]bool),
		Country: make(map[uint16]bool),
	}
	for _, radio := range strings.Split(*radiofilter, ",") {
		filter.Radio[strings.TrimSpace(radio)] = true
	}
	for _, country := range strings.Split(*countryfilter, ",") {
		mcc, err := strconv.ParseUint(country, 10, 16)
		if err != nil {
			continue
		}
		filter.Country[uint16(mcc)] = true
	}
	db.ImportCSV(*filename, filter)
}
