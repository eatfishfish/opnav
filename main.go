package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	cood "main/coordtransform"

	"github.com/chromedp/chromedp"
)

type PartDistance struct {
	PartDistance int    `json:"partDistance"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	TravelType   int    `json:"travelType"`
	RoadName     string `json:"roadName"`
	RoadType     int    `json:"roadType"`
	NextRoadName string `json:"nextRoadName"`
	NextRoadType int    `json:"nextRoadType"`
	Point        string `json:"point"`
	Direction    string `json:"direction"`
	PartDesc     string `json:"partDesc"`
	Shapepoint   string `json:"shapepoint"`
}

type UidInfo struct {
	Uid           int    `json:"uid"`
	Cellid        int    `json:"cellid"`
	District      int    `json:"district"`
	Distance      int    `json:"distance"`
	Rtistat       int    `json:"rtistat"`
	Carspeedlimit int    `json:"carspeedlimit"`
	Tkspeedlimit  int    `json:"tkspeedlimit"`
	Busspeedlimit int    `json:"busspeedlimit"`
	Shapepoint    string `json:"shapepoint"`
}

type RouteInfo struct {
	Desc            string         `json:"desc"`
	TotalDistance   int            `json:"totalDistance"`
	TotalTime       int            `json:"totalTime"`
	PointNum        int            `json:"pointNum"`
	Prefer          int            `json:"prefer"`
	TravelId        int            `json:"travelId"`
	Routeindex      int            `json:"routeindex"`
	Routeid         int            `json:"routeid"`
	Trafficlight    int            `json:"trafficlight"`
	Feature         int            `json:"feature"`
	Etatime         int            `json:"etatime"`
	Avoidtrafficjam int            `json:"avoidtrafficjam"`
	Routebv         int            `json:"routebv"`
	Tag             string         `json:"tag"`
	Toll            int            `json:"toll"`
	Etc             int            `json:"etc"`
	TravelDesc      []PartDistance `json:"travelDesc"`
	UidInfo         []UidInfo      `json:"uidInfo"`
	Csinfo          []interface{}  `json:"csinfo"`
}

type CareRoute struct {
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	Count        int         `json:"count"`
	RouteInfo    []RouteInfo `json:"routeInfo"`
}

type ResultLL struct {
	Result       [][2]string `json:"result"`
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
}

// -------get params-------
type Origin struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Destination struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Condition struct {
	Plan int `json:"plan"`
}

type ParamsCare struct {
	Origin       Origin        `json:"origin"`
	Destinations []Destination `json:"destinations"`
	Condition    Condition     `json:"condition"`
}

// ----------------mapbox-------

type MapBoxRoute struct {
	Routes    []RouteMB  `json:"routes"`
	WayPoints []WayPoint `json:"waypoints"`
	Code      string     `json:"code"`
	Uuid      string     `json:"uuid"`
}

type RouteMB struct {
	WeightTypical   float32  `json:"weight_typical"`
	DurationTypical float32  `json:"duration_typical"`
	WeightName      string   `json:"weight_name"`
	Weight          string   `json:"weight"`
	Duration        float32  `json:"duration"`
	Distance        float32  `json:"distance"`
	Legs            []Leg    `json:"legs"`
	Geometry        Geometry `json:"geometry"`
}

type Leg struct {
	ViaWayPoints    []interface{} `json:"via_waypoints"`
	Admins          []interface{} `json:"admins"`
	Annotation      Annotation    `json:"annotation"`
	WeightTypical   float32       `json:"weight_typical"`
	DurationTypical float32       `json:"duration_typical"`
	Steps           []Step        `json:"steps"`
	Distance        float32       `json:"distance"`
	Summary         string        `json:"summary"`
}

type Annotation struct {
	Maxspeed []Speed `json:"maxspeed"`
}

type Speed struct {
	Speed float32 `json:"speed"`
	Unit  string  `json:"unit"`
}

type Step struct {
	Geometry          Geometry `json:"geometry"`
	Distance          float32  `json:"distance"`
	BannerInstruction []string `json:"bannerInstructions"`
	Duration          float32  `json:"duration"`
	DurationTypical   float32  `json:"duration_typical"`
}

type Geometry struct {
	Coordinates [][2]float32 `json:"coordinates"`
	Tp          string       `json:"type"`
}

type WayPoint struct {
	Distance float32    `json:"distance"`
	Name     string     `json:"name"`
	Location [2]float32 `json:"location"`
}

type Coordinate struct {
	Coord [2]float32 `json:"_"`
}

var ak string = "4e647c0d9638c45c94188e534"
var port string = "9999"

func init() {
	file := "./" + "maplog" + ".txt"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)

	cfg := InitConfig("./config.txt")
	ak = cfg["ak"]
	port = cfg["port"]
}

func getWayPointsFromCareland(longStart, latStart, longEnd, latEnd int64) *MapBoxRoute {

	ctx := context.Background()

	req, err := http.NewRequest("GET", "https://api.careland.com.cn/api/v2/navi/routeplan", nil)
	if err != nil {
		log.Fatal(err)
	}

	var param ParamsCare
	param.Origin.X = float32(longStart)
	param.Origin.Y = float32(latStart)
	var des Destination
	des.X = float32(longEnd)
	des.Y = float32(latEnd)

	param.Destinations = append(param.Destinations, des)
	param.Condition.Plan = 1
	byBuf, _ := json.Marshal(param)

	params := req.URL.Query()
	params.Add("params", string(byBuf))
	params.Add("ak", ak)
	req.URL.RawQuery = params.Encode()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Println("ok")
	} else {
		return nil
	}

	var cr CareRoute
	err = json.Unmarshal(body, &cr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if len(cr.RouteInfo) > 0 && len(cr.RouteInfo[0].TravelDesc) > 0 {
		cr.RouteInfo[0].TravelDesc[0].X = int(longStart)
		cr.RouteInfo[0].TravelDesc[0].Y = int(latStart)
	}

	mp := ConvertCareRoute2MapBoxRoute(&cr)

	log.Printf("%+v", mp)

	return mp
}

func ConvertCareRoute2MapBoxRoute(cr *CareRoute) *MapBoxRoute {
	if len(cr.RouteInfo) == 0 {
		return nil
	}

	var mr MapBoxRoute
	var rb RouteMB
	rb.Distance = float32(cr.RouteInfo[0].TotalDistance)
	rb.Duration = float32(cr.RouteInfo[0].TotalTime)

	var l Leg
	var an Annotation
	var Coord [2]float32

	for _, e := range cr.RouteInfo[0].TravelDesc {
		if len(e.Shapepoint) == 0 {
			break
		}

		var sp Step
		setpShapepoint := fmt.Sprintf("%v,%v;%s", e.X, e.Y, e.Shapepoint)
		shapepoints := strings.Split(setpShapepoint, ";")
		speed_idx := 0
		sp.Distance = float32(e.PartDistance)
		if e.PartDistance == 0 {
			sp.Distance = rb.Distance
		}

		sp.BannerInstruction = make([]string, 0)
		sp.Duration = float32(rb.Duration)

		ret := convertCareLL(setpShapepoint, 0, 2)
		for i, ll := range ret.Result {
			if ll[0] == "" || ll[1] == "" {
				continue
			}

			long, _ := strconv.ParseFloat(ll[0], 32)
			lat, _ := strconv.ParseFloat(ll[1], 32)

			longW, latW := cood.GCJ02toWGS84(long, lat)
			Coord[0] = float32(longW)
			Coord[1] = float32(latW)
			sp.Geometry.Coordinates = append(sp.Geometry.Coordinates, Coord)

			//calc speed limit
			var currentLng float64
			var currentLat float64

			temp := strings.Split(shapepoints[i], ",")
			if len(temp) == 2 {
				currentLng, _ = strconv.ParseFloat(temp[0], 32)
				currentLat, _ = strconv.ParseFloat(temp[1], 32)
			}

			speed := 0
			for _, uid := range cr.RouteInfo[0].UidInfo {
				speedShapepoints := strings.Split(uid.Shapepoint, ";")
				if len(speedShapepoints) == 2 {
					sll := strings.Split(speedShapepoints[0], ",")
					var longS float64
					var latS float64
					if len(sll) == 2 {
						longS, _ = strconv.ParseFloat(sll[0], 32)
						latS, _ = strconv.ParseFloat(sll[1], 32)
					}

					var longE float64
					var latE float64
					sll = strings.Split(speedShapepoints[1], ",")
					if len(sll) == 2 {
						longE, _ = strconv.ParseFloat(sll[0], 32)
						latE, _ = strconv.ParseFloat(sll[1], 32)
					}

					if math.Abs(currentLng-longS) <= math.Abs(longS-longE) &&
						math.Abs(currentLat-latS) <= math.Abs(latS-latE) {
						speed = uid.Carspeedlimit
						break
					}
				}
			}

			msp := Speed{Unit: "km/h"}
			msp.Speed = float32(speed)
			an.Maxspeed = append(an.Maxspeed, msp)
			if speed_idx > 0 && an.Maxspeed[speed_idx-1].Speed == 0.0 && speed > 0 {
				an.Maxspeed[speed_idx-1].Speed = float32(speed)
			} else if speed_idx > 0 && an.Maxspeed[speed_idx-1].Speed != 0.0 && speed == 0 {
				an.Maxspeed[speed_idx] = an.Maxspeed[speed_idx-1]
			}
			speed_idx++

			if speed > 1 {
				sp.Duration = sp.Distance / ((1.0 / 3.6) * float32(speed))
			}
		}

		if len(sp.Geometry.Coordinates) > 0 {
			sp.Geometry.Tp = "LineString"
			l.Steps = append(l.Steps, sp)
		}
	}

	l.Annotation = an
	rb.Legs = append(rb.Legs, l)
	mr.Routes = append(mr.Routes, rb)

	mr.Code = "Ok"
	mr.Uuid = fmt.Sprint(rand.Int63())

	return &mr
}

func getWayPointsFromAmap(longStart, latStart, longEnd, latEnd float32) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	callFn := fmt.Sprintf("GetTest(%v,%v,%v,%v)", longStart, latStart, longEnd, latEnd)

	// run task list
	var res interface{}
	var resRoute interface{}
	err := chromedp.Run(ctx,
		chromedp.Navigate(`E:\\source\\opserver\\static\\nav.html`), // navigate to random page
		chromedp.EvaluateAsDevTools(callFn, &res),
		chromedp.WaitVisible(`#box2`),
		chromedp.EvaluateAsDevTools(`GetValue()`, &resRoute),
	)
	if err != nil {
		log.Fatal(err)
	}

	wp := [][2]float64{}

	log.Println(resRoute)
	ret := resRoute.([]interface{})
	for _, e := range ret {
		v := e.(map[string]interface{})
		log.Println(v["lng"])
		f := (v["lng"]).(float64)
		log.Println(f)

		test := [2]float64{(v["lng"]).(float64), (v["lat"]).(float64)}
		wp = append(wp, test)
	}

	log.Println(ret)
}

func main() {

	httpSvr()
}

func GetRoute(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetRoute = %+v\n", r.URL.Path)

	temp := strings.Split(r.URL.Path, "/")
	ll := temp[len(temp)-1]
	v := strings.Split(ll, ";")
	if len(v) <= 1 {
		return
	}

	start := strings.Split(v[0], ",")
	end := strings.Split(v[len(v)-1], ",")
	longStart, _ := strconv.ParseFloat(start[0], 32)
	latStart, _ := strconv.ParseFloat(start[1], 32)
	longEnd, _ := strconv.ParseFloat(end[0], 32)
	latEnd, _ := strconv.ParseFloat(end[1], 32)

	//getWayPointsFromAmap(float32(longStart), float32(latStart), float32(longEnd), float32(latEnd))

	wgs84ll := fmt.Sprintf("%v,%v;%v,%v", longStart, latStart, longEnd, latEnd)
	ret := convertCareLL(wgs84ll, 1, 0)
	if ret.ErrorMessage == "ok" {
		clongStart, _ := strconv.ParseInt(ret.Result[0][0], 10, 32)
		clatStart, _ := strconv.ParseInt(ret.Result[0][1], 10, 32)
		clongEnd, _ := strconv.ParseInt(ret.Result[1][0], 10, 32)
		clatEnd, _ := strconv.ParseInt(ret.Result[1][1], 10, 32)
		mp := getWayPointsFromCareland(clongStart, clatStart, clongEnd, clatEnd)
		if mp != nil {
			var wp WayPoint
			wp.Location[0] = float32(longStart)
			wp.Location[1] = float32(latStart)
			mp.WayPoints = append(mp.WayPoints, wp)
			wp.Location[0] = float32(longEnd)
			wp.Location[1] = float32(latEnd)
			mp.WayPoints = append(mp.WayPoints, wp)
		}

		bybuf, err := json.Marshal(mp)
		if err == nil {
			w.Write(bybuf)
		}
	}

	//care := `{"origin":{"x":410864332,"y":81373600},"destinations":[{"x":410365780,"y":81509815}],"condition":{"plan":1,"avoid":2,"forbid":32},"vehicle":{"height":3.1,"width":3.2,"weight":27,"axles":2,"type":134217728,"licenseplatetype":8,"selfweight":10,"load":20,"weels":0,"seats":5,"licenseplate":"粤B12345"}}`
}

func convertCareLL(coords string, tp1, tp2 int) *ResultLL {
	url := `https://api.careland.com.cn/api/v2/pub/geoconv?`
	url = url + fmt.Sprintf("ak=%s", ak) + fmt.Sprintf("&coords=%s", coords) + fmt.Sprintf("&from=%d&to=%d", tp1, tp2)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		//fmt.Println("ok")
	} else {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)

	var res ResultLL
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &res
}

func httpSvr() {
	http.HandleFunc("/directions/v5/mapbox/driving-traffic/", GetRoute)
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("ListenAndServe error %+v", err)
	}
}
