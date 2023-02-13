package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/metadb"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

type locationMessage struct {
	DeviceIDs struct {
		ID string `json:"device_id"`
	} `json:"end_device_ids"`
	UplinkMessage struct {
		DecodedPayload struct {
			Latitude  float64 `json:"Latitude"`
			Longitude float64 `json:"Longitud"`
			Location  string  `json:"Location"`
			Alarm     string  `json:"ALARM_status"`
			Batv      float64 `json:"BatV"`
		} `json:"decoded_payload"`
	} `json:"uplink_message"`
}

func buildMapURL(locations []float64) string {
	// Create the initial URL string
	url := "https://www.openstreetmap.org/export/embed.html?marker="

	// Iterate through the locations and add the latitude and longitude to the URL
	for i := 0; i < len(locations); i += 2 {
		url += fmt.Sprintf("%.6f,%.6f", locations[i], locations[i+1])
	}
	// Return the generated URL
	return url
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
	Index     int       `json:"index"`
}

func greaterThanDistance(location1, location2 *Location, meters int) bool {
	const R = 6371e3 // Earth's radius in meters

	lat1 := location1.Latitude * math.Pi / 180
	lat2 := location2.Latitude * math.Pi / 180
	dLat := (lat2 - lat1)
	dLon := (location2.Longitude - location1.Longitude) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance > float64(meters)
}

func getLocationByIndex(index int, db db.Database) (*Location, error) {
	location := Location{}
	data, err := db.ReadTx().Get([]byte("loc_" + strconv.Itoa(index)))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &location); err != nil {
		return nil, err
	}
	return &location, nil
}

func main() {
	username := flag.String("user", "tomassa-gps@ttn", "the things network application username")
	appkey := flag.String("key", "", "the things network application key")
	dataDir := flag.String("dataDir", "./", "data directory for storing the database")
	port := flag.Int("port", 8080, "port to listen to")
	logLevel := flag.String("logLevel", "info", "log level")
	distanceThreshold := flag.Int("distance", 100, "distance threshold in meters, to consider a new location")
	flag.Parse()
	log.Init(*logLevel, "stdout")

	log.Infof("creating/opening database")
	db, err := metadb.New(db.TypePebble, filepath.Join(*dataDir, "database"))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("creating http router")
	router := httprouter.HTTProuter{}
	if err := router.Init("0.0.0.0", *port); err != nil {
		log.Fatal(err)
	}
	log.Infof("creating api service")
	api, err := apirest.NewAPI(&router, "/api")
	if err != nil {
		log.Fatal(err)
	}
	if err := api.RegisterMethod("/locations", "GET", apirest.MethodAccessTypePublic, func(a *apirest.APIdata, h *httprouter.HTTPContext) error {
		locations := []Location{}
		db.Iterate([]byte("loc_"), func(key, value []byte) bool {
			location := Location{}
			if err := json.Unmarshal(value, &location); err != nil {
				log.Warnf("could not decode location: %s", err)
				return true
			}
			if len(key) < 1 {
				log.Warnf("could not decode location index %s", key)
				return true
			}
			index, err := strconv.Atoi(string(key))
			if err != nil {
				log.Warnf("could not decode location index: %s", err)
				return true
			}
			location.Index = index
			locations = append(locations, location)
			return true
		})
		data, err := json.Marshal(&locations)
		if err != nil {
			return err
		}
		return h.Send(data, apirest.HTTPstatusCodeOK)
	}); err != nil {
		log.Fatal(err)
	}

	index := 0
	lastIndexBytes, err := db.ReadTx().Get([]byte("index"))
	if err != nil {
		log.Warnf("could not get last index, assuming 0")
	} else {
		index, err = strconv.Atoi(string(lastIndexBytes))
		if err != nil {
			log.Fatalf("could not decode last index: %s", err)
		}
	}

	var broker = "tcp://eu1.cloud.thethings.network:1883"
	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("trackerd")
	options.SetUsername(*username)
	options.SetPassword(*appkey)

	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "#"
	token = client.Subscribe(topic, 0,
		func(client mqtt.Client, msg mqtt.Message) {
			jmsg := locationMessage{}
			if err := json.Unmarshal(msg.Payload(), &jmsg); err != nil {
				log.Errorw(err, "could not decode message")
				return
			}
			log.Debugf("message for topic %s: %+v\n", msg.Topic(), jmsg)
			lat := jmsg.UplinkMessage.DecodedPayload.Latitude
			lon := jmsg.UplinkMessage.DecodedPayload.Longitude
			log.Infow("new location received",
				"latitude", jmsg.UplinkMessage.DecodedPayload.Latitude,
				"longitude", jmsg.UplinkMessage.DecodedPayload.Longitude,
				"device", jmsg.DeviceIDs.ID,
				"url", buildMapURL([]float64{jmsg.UplinkMessage.DecodedPayload.Latitude, jmsg.UplinkMessage.DecodedPayload.Longitude}))
			if lat == 0 || lon == 0 {
				log.Warnf("received invalid location")
				return
			}
			location := Location{
				Latitude:  lat,
				Longitude: lon,
				Timestamp: time.Now(),
			}

			// check if location is too close to previous location
			if index > 0 {
				previousLocation, err := getLocationByIndex(index-1, db)
				if err != nil {
					log.Warnf("could not get previous location: %s", err)
				} else {
					if !greaterThanDistance(&location, previousLocation, *distanceThreshold) {
						log.Infof("location is too close to previous location, ignoring")
						return
					}
				}
			}

			data, err := json.Marshal(&location)
			if err != nil {
				log.Errorf("could not encode location: %s", err)
				return
			}
			wtx := db.WriteTx()
			if err := wtx.Set([]byte(fmt.Sprintf("loc_%d", index)), data); err != nil {
				log.Errorf("could not store location: %s", err)
				return
			}
			if err := wtx.Set([]byte("index"), []byte(fmt.Sprintf("%d", index))); err != nil {
				log.Errorf("could not store index: %s", err)
				return
			}
			wtx.Commit()
			index++
		})

	router.AddRawHTTPHandler("/we*", "GET", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		http.ServeFile(w, r, path)
	})

	token.Wait()
	log.Infow("subscribed", "topic", topic)

	select {}
}
