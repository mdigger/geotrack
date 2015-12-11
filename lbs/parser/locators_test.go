package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mdigger/geolocate"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
)

const (
	// googleAPIKey = "<API Key>"
	// yandexAPIKey = "<API Key>"
	googleAPIKey = "AIzaSyBDw1oDEngRh098SlWFKWDJ5k7BFrfX_WI"
	yandexAPIKey = "AFVzaFYBAAAAtuSxOwMARr9H5Bte6G1FlAeX4vHlmPYgQeEAAAAAAAAAAABiS9nunmKQsVptuGIh0rAM1K8PoQ=="
)

var testData2 = []struct {
	lbs  string
	wifi string
}{
	{
		lbs:  `867293-6568-022970357-fa-2-1e3e-5c3f-8e-1e3e-5c3d-85-1e3e-50b8-7f-1e3e-5c40-7c-1e3e-5c41-79-1e3e-5c42-78-1e3e-50bd-74`,
		wifi: `[{"m":"f09e63c1f260","i":"101","s":"238"},{"m":"24a43cb20bc0","i":"KNOW-HOW","s":"240"},{"m":"0019e1008b31","i":"Beeline_WiFi_","s":"246"},{"m":"d421223446fa","i":"MGTS_GPON_707","s":"246"},{"m":"54a050829b74","i":"nogotok","s":"246"},{"m":"e0cb4edcd819","i":"diva","s":"246"},{"m":"90f6528c0640","i":"Sonic","s":"246"}]`,
	},
	{
		lbs:  `867293-47330-022970357-fa-2-1e3e-5c3f-94-1e3e-50b8-88-1e3e-50bd-86-1e3e-5c41-83-1e3e-5c3d-82-1e3e-5c42-75-1e3e-50b8-88`,
		wifi: `[{"m":"00026fdc4564","i":"WOKKER WiFi P","s":"224"},{"m":"00026fdc4565","i":"WOKKER WiFi F","s":"226"},{"m":"dc9fdb324d1b","i":"5th-Avenue","s":"238"},{"m":"107bef5587c4","i":"auto","s":"240"},{"m":"60a44c85fb88","i":"nogotok","s":"240"},{"m":"04219157b31b","i":"Shoko","s":"246"},{"m":"e0cb4edcd819","i":"diva","s":"248"}]`,
	},
	{
		lbs:  `867293-10905-022970357-fa-2-1e3e-5c3f-98-1e3e-50b8-8b-1e3e-50bd-86-1e3e-5c41-81-1e3e-5c3d-7d-1e3e-5c42-76-1e3e-50b8-8b`,
		wifi: `[{"m":"00026fdc4564","i":"WOKKER WiFi P","s":"214"},{"m":"d4ca6debbcf1","i":"tms","s":"238"},{"m":"04219157b31b","i":"Shoko","s":"240"},{"m":"dc9fdb324d1b","i":"5th-Avenue","s":"240"},{"m":"60a44c85fb88","i":"nogotok","s":"246"},{"m":"00219157b31b","i":"MTS/TASCOM","s":"246"},{"m":"2cab255c4ce9","i":"diva 1","s":"250"}]`,
	},
	{
		lbs:  `867293-22828-022970357-fa-2-1e3e-5c3f-92-1e3e-50b8-7d-1e3e-50bd-73-1e3e-5c42-71-1e3e-50b8-7d-1e3e-50bd-74-1e3e-5c42-72`,
		wifi: `[{"m":"28285de817fe","i":"TochkKrasoti","s":"224"},{"m":"e0cb4edcd819","i":"diva","s":"236"}]`,
	},
	{
		lbs:  `867293-15496-021782910-fa-1-18c0-81c-a0-18c0-819-a1-18c0-817-99-18c0-87d-91-18c0-81a-8e-18c0-818-8b-18d8-b77-84`,
		wifi: `[{"m":"6a285d874738","i":"Dom5","s":"234"},{"m":"84c9b25a106e","i":"ice","s":"238"},{"m":"6a285da0d5c0","i":"Dom7","s":"240"},{"m":"c4e984266899","i":"Sanacia","s":"244"},{"m":"d4ca6dbbcdb1","i":"IKO TRADE","s":"246"},{"m":"ec43f6df5bf0","i":"3AOcto","s":"250"}]`,
	},
	{
		lbs:  `867293-59603-021782910-fa-1-18c0-81c-9e-18c0-819-9a-18c0-81a-95-18c0-817-92-18c0-87d-8e-18d8-b77-8e-18c0-81b-87`,
		wifi: `[{"m":"6a285da0d5c0","i":"Dom7","s":"238"},{"m":"4c5e0c624e4d","i":"Geapplic","s":"246"}]`,
	},
	{
		lbs:  `867293-59603-021782910-fa-1-18c0-81c-9e-18c0-819-9c-18c0-81a-96-18c0-817-94-18c0-87d-91-18d8-b77-8f-18c0-818-88`,
		wifi: `[{"m":"6a285d874738","i":"Dom5","s":"236"},{"m":"f835dd0cad04","i":"YOTA","s":"238"},{"m":"00265acf9ff6","i":"MASHA","s":"246"},{"m":"d4ca6dbbcdb1","i":"IKO TRADE","s":"248"},{"m":"bcc493697790","i":"WIFI-Staff","s":"248"}]`,
	},
	{
		lbs:  `867293-59603-021782910-fa-1-18c0-81c-9e-18c0-819-9d-18c0-81a-96-18c0-817-94-18c0-87d-91-18d8-b77-8e-18c0-818-86`,
		wifi: `[{"m":"6a285d874738","i":"Dom5","s":"226"},{"m":"f835dd0cad04","i":"YOTA","s":"242"},{"m":"bcc493697790","i":"WIFI-Staff","s":"244"}]`,
	},
	{
		lbs:  `867293-60389-021782910-fa-1-294-2b5d-95-294-2b5e-7c-294-2b5e-7c-294-2b5e-7c-294-2b5e-7f-294-2b5e-7f`,
		wifi: `[{"m":"f41fc2320d40","i":"MosMetro_Free","s":"240"}]`,
	},
	{
		lbs:  `867293-7592-021782910-fa-1-294-2b5e-83-294-2b5e-7a-294-2b5e-7a-294-2b5e-7a`,
		wifi: `[{"m":"ec1d7fb60098","i":"BEELINE-M","s":"222"}]`,
	},
	{
		lbs:  `867293-578-021782910-fa-1-2cd-6719-7b-2cd-6718-7f-2cd-671b-7d-18c5-cd74-7a-2cd-9fbb-7a-18c5-b4f-75-2cd-563b-7c`,
		wifi: `[{"m":"30469a0ad58d","i":"BALTIKA","s":"244"},{"m":"d4bf7f04d664","i":"AKAO 25","s":"244"},{"m":"f8c091149823","i":"Akado55","s":"246"},{"m":"f8c0911238c3","i":"DrWeb_10","s":"248"},{"m":"dc028ed19ed2","i":"MGTS_GPON_FBF","s":"248"},{"m":"30b5c24723f6","i":"bosbsn","s":"250"},{"m":"d4bf7f0c0dcc","i":"Akado_28","s":"250"}]`,
	},
	{
		lbs:  `867293-41230-022970357-fa-2-1e3e-7310-94-1e3e-730b-a0-1e3e-96e4-8d-1e3e-96e3-81-1e3e-7312-7d-1e3e-5506-7c-1e3e-7313-7c`,
		wifi: `[{"m":"60a44c664b74","i":"fnbic_17_11","s":"232"}]`,
	},
	{
		lbs:  `867293-4919-022970357-fa-2-1e3e-7310-96-1e3e-730b-9b-1e3e-96e4-93-1e3e-96e3-8c-1e3e-7313-83-1e3e-5506-7f-1e3e-7312-78`,
		wifi: `[{"m":"0014d1bdef5e","i":"TRENDnet651(B","s":"248"}]`,
	},
	{
		lbs:  `867293-55511-022970357-fa-2-1e3e-50eb-95-1e3e-50dc-9c-1e3e-545f-7a-1e3e-50e6-79-1e3e-5c42-77-1e3e-545c-75-1e3e-50f1-70`,
		wifi: `[{"m":"0a18d6a9ccbf","i":"SBERBANK","s":"218"},{"m":"4ad9e7218499","i":"SBERBANK","s":"234"},{"m":"48f8b30a0e79","i":"Cisco97316","s":"236"},{"m":"48ee0ce2bb53","i":"68rus","s":"246"},{"m":"88e3ab7b1ad4","i":"Volutex LLC","s":"248"}]`,
	},
	{
		lbs:  `867293-28579-022970357-fa-2-1e3e-6259-a6-1e3e-1841-aa-1e3e-625b-93-1e3e-625a-82-1e3e-446b-7e-1e3e-73db-7e-1e3e-50bc-7a`,
		wifi: `[{"m":"4c5e0c7c0744","i":"MosGorTrans_F","s":"240"},{"m":"0019e10070b1","i":"Beeline_WiFi_","s":"244"},{"m":"0016caf537e1","i":"Beeline_WiFi_","s":"246"},{"m":"0019e1013420","i":"Beeline_WiFi","s":"250"},{"m":"0019e1013421","i":"Beeline_WiFi_","s":"250"}]`,
	},
	{
		lbs:  `867293-3444-022970357-fa-2-1e3f-5100-88-1e3f-5103-8d-1e3f-5c7e-7c-1e3e-510b-79-1e3f-524d-78-1e3f-5103-8d-1e3f-5c7e-7c`,
		wifi: `[{"m":"000c42b7a01b","i":"TMS","s":"222"},{"m":"14d14d4a52ab","i":"vabisabi","s":"242"},{"m":"14d64d4a52ab","i":"MTS/Tascom","s":"244"}]`,
	},
}

func TestLocators2(t *testing.T) {
	google, err := geolocate.New(geolocate.Google, googleAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	yandex, err := geolocate.New(geolocate.Yandex, yandexAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	mozilla, err := geolocate.New(geolocate.Mozilla, "test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("")
	for _, s := range testData2 {
		// fmt.Println("LBS: ", s.lbs)
		req1, err := ParseLBS("gsm", s.lbs)
		if err != nil {
			t.Fatal(err)
		}
		req2 := *req1
		var names []string
		req2.WifiAccessPoints, names, err = ParseWiFi(s.wifi)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("WIFI:", strings.Join(names, ", "))
		resp, err := google.Get(*req1)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "Google", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "Google", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = google.Get(req2)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "*Google", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "*Google", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = yandex.Get(*req1)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "Yandex", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "Yandex", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = yandex.Get(req2)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "*Yandex", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "*Yandex", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = mozilla.Get(*req1)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "Mozilla", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "Mozilla", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = mozilla.Get(req2)
		if err != nil {
			fmt.Printf("%9s error: %v\n", "*Mozilla", err)
			t.Error(err)
		} else {
			fmt.Printf("%9s [%.7f,%.7f] %.2f\n", "*Mozilla", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		fmt.Println("-----------------------------------------")
	}
}

func TestLocators(t *testing.T) {
	google, err := geolocate.New(geolocate.Google, googleAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	mozilla, err := geolocate.New(geolocate.Mozilla, "test")
	if err != nil {
		t.Fatal(err)
	}
	yandex, err := geolocate.New(geolocate.Yandex, yandexAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	mongodb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		t.Fatal(err)
	}
	defer mongodb.Close()
	lbs, err := lbs.InitDB(mongodb)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range testData {
		req, err := ParseLBS("gsm", s)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := google.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Google", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Google", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = yandex.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Yandex", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Yandex", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = mozilla.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Mozilla", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Mozilla", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = lbs.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Internal", err)
			t.Error(err)
		}
		fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Internal", resp.Location.Lon, resp.Location.Lat, resp.Accuracy)
		fmt.Println("-----------------------------------------")
	}
}
