package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"googlemaps.github.io/maps"
)

//GoogleAPIKey is used for all Google APIs
const GoogleAPIKey = "AIzaSyB9ufsi7c5Mq7tf9XASClIMdhk8zisieTY"

//MapQuestAPIKey is used to access the MapQuest API
const MapQuestAPIKey = "IMPZpoduZk7sdw3Hp9ZhphmMfPmBsA6P"

//IndexWingoldCoordinate is the Coordinate for Index Exchange on Wingold Avenue
const IndexWingoldCoordinate = "43.6065827,-79.6563887"

//Coordinate is used to locate a map point
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

//MapQuestGeoCoding holds JSON from MapQuest GeoCoding
type MapQuestGeoCoding struct {
	Info struct {
		Statuscode int `json:"statuscode"`
		Copyright  struct {
			Text         string `json:"text"`
			ImageURL     string `json:"imageUrl"`
			ImageAltText string `json:"imageAltText"`
		} `json:"copyright"`
		Messages []interface{} `json:"messages"`
	} `json:"info"`
	Options struct {
		MaxResults        int  `json:"maxResults"`
		ThumbMaps         bool `json:"thumbMaps"`
		IgnoreLatLngInput bool `json:"ignoreLatLngInput"`
	} `json:"options"`
	Results []struct {
		ProvidedLocation struct {
			Location string `json:"location"`
		} `json:"providedLocation"`
		Locations []struct {
			Street             string `json:"street"`
			AdminArea6         string `json:"adminArea6"`
			AdminArea6Type     string `json:"adminArea6Type"`
			AdminArea5         string `json:"adminArea5"`
			AdminArea5Type     string `json:"adminArea5Type"`
			AdminArea4         string `json:"adminArea4"`
			AdminArea4Type     string `json:"adminArea4Type"`
			AdminArea3         string `json:"adminArea3"`
			AdminArea3Type     string `json:"adminArea3Type"`
			AdminArea1         string `json:"adminArea1"`
			AdminArea1Type     string `json:"adminArea1Type"`
			PostalCode         string `json:"postalCode"`
			GeocodeQualityCode string `json:"geocodeQualityCode"`
			GeocodeQuality     string `json:"geocodeQuality"`
			DragPoint          bool   `json:"dragPoint"`
			SideOfStreet       string `json:"sideOfStreet"`
			LinkID             string `json:"linkId"`
			UnknownInput       string `json:"unknownInput"`
			Type               string `json:"type"`
			LatLng             struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"latLng"`
			DisplayLatLng struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"displayLatLng"`
			MapURL string `json:"mapUrl"`
		} `json:"locations"`
	} `json:"results"`
}

func (m *MapQuestGeoCoding) Locate(Key string, address string) {
	response, err := http.Get("http://www.mapquestapi.com/geocoding/v1/address?key=" + Key + "&location=" + url.QueryEscape(address))
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, m)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	io.Copy(os.Stdout, response.Body)

}

func (m *MapQuestGeoCoding) Coordinate() Coordinate {

	return Coordinate{m.Results[0].Locations[0].LatLng.Lat, m.Results[0].Locations[0].LatLng.Lng}
}

//GoogleGeoCoding is the Google Struct to store GoogleGeoCoding information about an address
type GoogleGeoCoding struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Bounds struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"bounds"`
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PartialMatch bool     `json:"partial_match"`
		PlaceID      string   `json:"place_id"`
		Types        []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

func (c *Coordinate) String() string {
	return strconv.FormatFloat(c.Latitude, 'f', 6, 64) + ", " + strconv.FormatFloat(c.Longitude, 'f', 6, 64)
}

//GeoIP is the struct use to hold the JSON resulting from an IP Geolocation
type GeoIP struct {
	// The right side is the name of the JSON variable
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	Zipcode     string  `json:"zipcode"`
	Lat         float64 `json:"latitude"`
	Long        float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
	AreaCode    int     `json:"area_code"`
}

func getPublicIP() string {
	response, err := http.Get("http://whatismyip.akamai.com/")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	// Provide a domain name or IP address
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	address := strings.TrimSuffix(string(body), "\n")
	return address
}

//LatLong return the Latitude and Longitude as a string
func (g *GeoIP) LatLong() string {
	return strconv.FormatFloat(g.Lat, 'f', 6, 64) + ", " + strconv.FormatFloat(g.Long, 'f', 6, 64)
}

//Coordinate returns the Coordinate found in the GeoIP struct
func (g *GeoIP) Coordinate() Coordinate {
	c := Coordinate{g.Lat, g.Long}
	return c
}

//Locate  Find the Coordinate for an IP Address
func (g *GeoIP) Locate(address string) {
	response, err := http.Get("https://freegeoip.net/json/" + address)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, &g)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	io.Copy(os.Stdout, response.Body)

}

//LocateSelf locate the address of the current computer based on the public IP
func (g *GeoIP) LocateSelf() {
	g.Locate(getPublicIP())
}

//Show display the GeoIP struct to the command line
func (g *GeoIP) Show() {
	fmt.Println("\nIP Geo-Location:")
	fmt.Println("IP address:\t", g.IP)
	fmt.Println("Country Code:\t", g.CountryCode)
	fmt.Println("Country Name:\t", g.CountryName)
	fmt.Println("Zip Code:\t", g.Zipcode)
	fmt.Println("Latitude:\t", g.Lat)
	fmt.Println("Longitude:\t", g.Long)
	fmt.Println("Metro Code:\t", g.MetroCode)
	fmt.Println("Area Code:\t", g.AreaCode)

}

//FindCenterCoordinate finds the center coordinate on the map from a list of coordinates
func FindCenterCoordinate(Coordinates []Coordinate) Coordinate {

	const RADIANS = math.Pi / 180
	const DEGREES = 180 / math.Pi

	c := Coordinate{0.0, 0.0}

	numCoordinates := len(Coordinates)
	if numCoordinates == 0 {
		return c

	}
	numCoordinatesF := float64(numCoordinates)

	X := 0.0
	Y := 0.0
	Z := 0.0

	for _, c := range Coordinates {
		lat := c.Latitude * RADIANS
		lon := c.Longitude * RADIANS

		X += math.Cos(lat) * math.Cos(lon)
		Y += math.Cos(lat) * math.Sin(lon)
		Z += math.Sin(lat)
	}

	X /= numCoordinatesF
	Y /= numCoordinatesF
	Z /= numCoordinatesF

	lon := math.Atan2(Y, X)
	hyp := math.Sqrt(X*X + Y*Y)
	lat := math.Atan2(Z, hyp)

	c.Latitude = lat * DEGREES
	c.Longitude = lon * DEGREES

	return c

}

//LocateByStreetAddress returns the coordinate from a street address
func LocateByStreetAddress(address string) Coordinate {
	urlString := "https://maps.googleapis.com/maps/api/geocode/json?address=" + url.QueryEscape(address) + ",+CA&key=" + GoogleAPIKey
	response, err := http.Get(urlString)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	// Decode our JSON results
	var g GoogleGeoCoding
	// Unmarshal the JSON byte slice to a GeoIP struct
	err = json.Unmarshal(body, &g)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(g)
	defer response.Body.Close()
	io.Copy(os.Stdout, response.Body)
	var c Coordinate
	if len(g.Results) > 0 {
		c = Coordinate{g.Results[0].Geometry.Location.Lat, g.Results[0].Geometry.Location.Lng}
	}
	return c
}

func testFindCenter() {
	var coords []Coordinate
	coords = append(coords, Coordinate{43.704372, -79.464364})
	coords = append(coords, Coordinate{43.701208, -79.452106})
	coords = append(coords, Coordinate{43.706893, -79.453391})
	coords = append(coords, Coordinate{43.698679, -79.462161})

	center := FindCenterCoordinate(coords) //43.70278812400493 -79.45800556932514
	fmt.Println(center)

}

func testFindDistance() {
	var geo GeoIP

	geo.LocateSelf()
	geo.Show()
	origin := geo.LatLong()
	destination := IndexWingoldCoordinate
	c, err := maps.NewClient(maps.WithAPIKey(GoogleAPIKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	var origins []string
	var destinations []string
	origins = append(origins, origin)
	destinations = append(destinations, destination)

	r := &maps.DistanceMatrixRequest{
		Origins:      origins,
		Destinations: destinations,
		Mode:         "driving",
	}
	route, err := c.DistanceMatrix(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	distance := route.Rows[0].Elements[0].Distance
	fmt.Println(distance.HumanReadable)
	fmt.Println(distance.Meters)

}

func testGetDirections() {
	var geo GeoIP
	c, err := maps.NewClient(maps.WithAPIKey(GoogleAPIKey))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	geo.LocateSelf()
	origin := geo.LatLong()
	destination := IndexWingoldCoordinate
	r := &maps.DirectionsRequest{
		Origin:      origin,
		Destination: destination,
	}
	route, _, err := c.Directions(context.Background(), r)
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	fmt.Println(route)
}

func testGeocodeStreetAddressFromFileChannels(path string) {

	streets := make(chan string)
	latlongs := make(chan Coordinate)

	go AddressScanner(path, streets)
	go Locater(latlongs, streets)
	CenterFinder(latlongs)

}

func AddressScanner(path string, out chan<- string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		address := scanner.Text()
		fmt.Println(address)
		out <- address
	}
	close(out)
}

func Locater(out chan<- Coordinate, in <-chan string) {
	for a := range in {
		fmt.Println(a)
		c := LocateByStreetAddress(a)
		fmt.Println(c)
		out <- c
	}
	close(out)
}

func CenterFinder(in <-chan Coordinate) {
	for c := range in {
		fmt.Println(c)
		// coords = append(coords, c)
	}

}

func testGeocodeStreetAddressFromFile(path string) {

	var coords []Coordinate
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	n := 0
	for scanner.Scan() {
		address := scanner.Text()
		coord := LocateByStreetAddress(address)
		coords = append(coords, coord)
		n++
	}
	fmt.Println("----------- Center --------------")
	fmt.Println(n)
	fmt.Println(len(coords))
	center := FindCenterCoordinate(coords) //46.97790754920849 -86.69303559892391
	fmt.Println(center)
}

func testGeocodeStreetAddress() {
	var wg sync.WaitGroup
	var lock sync.Mutex
	var coords []Coordinate
	wg.Add(4)
	go func() {
		defer wg.Done()
		lock.Lock()
		defer lock.Unlock()
		coords = append(coords, LocateByStreetAddress("120 Little Creek Road, Mississauga, Ontario"))
	}()
	go func() {
		defer wg.Done()
		lock.Lock()
		defer lock.Unlock()
		coords = append(coords, LocateByStreetAddress("74 Wingold Avenue, North York, Ontario"))
	}()
	go func() {
		defer wg.Done()
		lock.Lock()
		defer lock.Unlock()
		coords = append(coords, LocateByStreetAddress("Square One, Mississauga"))
	}()
	go func() {
		defer wg.Done()
		lock.Lock()
		defer lock.Unlock()
		coords = append(coords, LocateByStreetAddress("7 Gaylord Place, St. Albert"))
	}()
	wg.Wait()
	center := FindCenterCoordinate(coords) //46.97790754920849 -86.69303559892391
	fmt.Println(center)
}

func main() {
	// testGeocodeStreetAddress()
	// testGeocodeStreetAddressFromFile("./addresses.txt")
	// testGeocodeStreetAddressFromFileChannels("./addresses.txt")
	var m MapQuestGeoCoding
	m.Locate(MapQuestAPIKey, "120 Little Creek Road, Mississauga, ON")
	c := m.Coordinate()
	fmt.Println(c)
}
