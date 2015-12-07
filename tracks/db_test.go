package tracks

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/users"

	"gopkg.in/mgo.v2/bson"
)

func TestBD(t *testing.T) {
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	// подключаемся к хранилищу
	db, err := InitDB(mdb)
	if err != nil {
		t.Fatal("Connect error:", err)
	}

	const deviceID = "test0123456789"
	var (
		groupID = users.SampleGroupID
		points  = []*geo.Point{
			{37.57351, 55.715084},
			{37.595061, 55.736077},
			{37.589248, 55.765944},
			{37.587119, 55.752658},
			{37.627804, 55.752541},
			{37.642815, 55.74711},
			{37.689024, 55.724109},
		}
		currentPoint          = points[0]                       // текущая точка
		destinationPointIndex = 1                               // индекс точки назначения
		destinationPoint      = points[destinationPointIndex]   // точка назначения
		currentTime           = time.Now().Add(time.Hour * -12) // время начала
		interval              time.Duration                     // интервал времени
	)
	const (
		MaxSpeed   = 4.5 * 1000 / 60 / 60 // максимальная скорость передвижения — 4.5 км/ч
		MaxBearing = 45.0                 // максимальное отклонение
	)

	for {
		track := &TrackData{
			DeviceID: deviceID,
			GroupID:  groupID,
			Time:     currentTime,
			Point:    currentPoint,
			Type:     uint8(rand.Int31n(5)),
		}
		if err := db.Add(track); err != nil {
			t.Fatal(err)
		}
		// случайный интервал до 10 минут
		interval = time.Duration(rand.Int63n(int64(time.Minute * 10)))
		currentTime = currentTime.Add(interval) // увеличиваем время
		// вычисляем расстояние с учетом прошедшего времени
		dist := rand.Float64() * MaxSpeed * float64(interval.Seconds())
		// случайное отклонение от заданного направления
		bearing := currentPoint.BearingTo(destinationPoint) + rand.Float64()*MaxBearing - (MaxBearing / 2)
		// перемещаемся на заданное расстояние в заданном направлении
		currentPoint = currentPoint.Move(dist, bearing)
		// проверяем, что расстояние не меньше 150 метров
		if currentPoint.Distance(destinationPoint) < 150.0 {
			if len(points) == destinationPointIndex+1 {
				log.Println("Мы достигли конечной точки назначения")
				break
			}
			log.Println("Мы достигли точки назначения", destinationPointIndex)
			currentPoint = destinationPoint
			destinationPointIndex++
			destinationPoint = points[destinationPointIndex]
		}
		if currentTime.After(time.Now()) {
			log.Println("Время истекло")
			break
		}
	}

	var lastId bson.ObjectId
	for {
		// fmt.Println("lastID:", lastId.Hex())
		tracks, err := db.Get(groupID, deviceID, 5, lastId.Hex())
		if err != nil {
			t.Fatal(err)
		}
		if len(tracks) == 0 {
			fmt.Println("END")
			break
		}
		fmt.Println(len(tracks),
			tracks[0].Time.Format("15:04:05"), tracks[0].ID.Hex(), "-",
			tracks[len(tracks)-1].Time.Format("15:04:05"), tracks[len(tracks)-1].ID.Hex())
		lastId = tracks[len(tracks)-1].ID
	}

	track, err := db.GetLast(groupID, deviceID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("last:", track.Time.Format("15:04:05"), track.ID.Hex(), track.Point)
	jsondata, err := json.MarshalIndent(track, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(jsondata))

	deviceIDs, err := db.GetDevicesID(groupID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("device ids:", strings.Join(deviceIDs, ", "))

	_, err = db.GetDay(groupID, deviceID)
	if err != nil {
		t.Fatal(err)
	}
}
