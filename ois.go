//"If we have seen further it is by standing on the shoulders of giants."

package ois

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

//Olympus .
type Olympus struct {
	client     *gorequest.SuperAgent
	cameraMode int
	live       bool
}

const (
	//Debugging flag
	_debug = true

	//URLs
	_domain    = "http://192.168.0.10"
	_imagePath = "/DCIM/100OLYMP"

	//Connections
	_connection = "Keep-Alive"
	_userAgent  = "OI.Share v2"

	//Commands
	_shutdown       = "/exec_pwoff.cgi"
	_listImages     = "/get_imglist.cgi"
	_getThumbnail   = "/get_thumbnail.cgi"
	_getResized     = "/get_resizeimg.cgi"
	_connectionMode = "/get_connectmode.cgi"
	_switchMode     = "/switch_cammode.cgi"
	_doMisc         = "/exec_takemisc.cgi"
	_doMotion       = "/exec_takemotion.cgi"
	_doShutter      = "/exec_shutter.cgi"
	_setProperty    = "/set_camprop.cgi" //POST
	_cameraInfo     = "/get_caminfo.cgi"

	//Modes
	_shutter = iota
	_play
	_liveview

	//Misc
	_pollingInterval = 5000 //ms
	_udpPort         = 28488
)

var (
	ticker *time.Ticker

	//Res size resize values
	Res = struct {
		r1024, r1600, r1920, r2048 string
	}{
		r1024: "1024",
		r1600: "1600",
		r1920: "1920",
		r2048: "2048",
	}

	//Quality rec quality level
	Quality = struct {
		q320, q640, q800, q1024, q1280 string
	}{
		q320:  "0320x0240",
		q640:  "0640x0480",
		q800:  "0800x0600",
		q1024: "1024x0768",
		q1280: "1280x0960",
	}

	udpConnection *net.UDPConn
)

//New creates a new instance of OIS
func New() *Olympus {
	ol := &Olympus{}

	//init the client object
	ol.client = gorequest.New()

	//set debug
	if false && _debug {
		ol.client.SetDebug(true)
	}

	//poll camera to check connection
	ticker = time.NewTicker(time.Millisecond * _pollingInterval)
	go func() {
		for range ticker.C {
			//call the connection checker
			ol.Connect()
		}
	}()
	//ticker.Stop()

	return ol
}

//Image defines images in list
type Image struct {
	directory, filename, size, attribute, data, time string
}

//Connect gets the camera connection mode
func (ol *Olympus) Connect() *Olympus {
	if !ol.live {
		res, _, errors := ol.client.Get(_domain + _connectionMode).
			End()
		catchHTTPError("", res, errors)

		//log
		//logger(body)
	}

	return ol
}

//Info gets the camera information i.e Camera name
func (ol *Olympus) Info(info *string) *Olympus {
	res, body, errors := ol.client.Get(_domain + _cameraInfo).
		End()
	catchHTTPError("", res, errors)

	//log
	logger(body)

	return ol
}

//Mode sets the camera mode
func (ol *Olympus) Mode(mode int, quality string) *Olympus {
	if ol.cameraMode == mode {
		return ol
	}

	q1, q2 := "", ""

	//set flag
	modeString := ""

	switch mode {
	case _play:
		modeString = "play"
	case _shutter:
		modeString = "shutter"
	case _liveview:
		modeString = "rec"
		q2 = "lvqty=" + quality
	default:
	}

	q1 = "mode=" + modeString

	res, _, errors := ol.client.Get(_domain + _switchMode).
		Query(q1).
		Query(q2).
		End()
	catchHTTPError("", res, errors)

	//set mode
	ol.cameraMode = mode

	return ol
}

//Shutdown turns off the camera
func (ol *Olympus) Shutdown() *Olympus {
	res, _, errors := ol.client.Get(_domain + _shutdown).
		End()
	catchHTTPError("", res, errors)

	return ol
}

//List returns list of all images in default directory
func (ol *Olympus) List() *Olympus {
	//change mode to play
	ol.Mode(_play, "")

	res, body, errors := ol.client.Get(_domain + _listImages).
		Query("DIR=" + _imagePath).
		End()
	catchHTTPError("", res, errors)

	//directory | filename | size | attribute | date | time
	scanner := bufio.NewScanner(strings.NewReader(body))

	//new list
	images := make([]Image, 0)

	for scanner.Scan() {
		txt := scanner.Text()
		vals := strings.Split(txt, ",")

		if len(vals) == 6 {
			images = append(images, Image{vals[0], vals[1], vals[2], vals[3], vals[4], vals[5]})
		}
	}

	logger(images)

	return ol
}

//Image retrieves an image by name
func (ol *Olympus) Image(filename string) *Olympus {
	//change mode to play
	ol.Mode(_play, "")

	res, body, errors := ol.client.Get(_domain + _imagePath + "/" + filename).
		End()
	catchHTTPError("", res, errors)

	makeImage(filename, body)

	return ol
}

//Resize returns a resized image
func (ol *Olympus) Resize(filename, size string) *Olympus {
	//change mode to play
	ol.Mode(_liveview, Quality.q640)

	res, body, errors := ol.client.Get(_domain + _getResized).
		Query("DIR=" + _imagePath + "/" + filename + "&size=" + size).
		End()
	catchHTTPError("", res, errors)

	makeImage(filename, body)

	return ol
}

//Thumbnail retrieves thumbnail version of an image
func (ol *Olympus) Thumbnail(filename string) *Olympus {
	//change mode to play
	ol.Mode(_play, "")

	res, body, errors := ol.client.Get(_domain + _getThumbnail).
		Query("DIR=" + _imagePath + "/" + filename).
		End()
	catchHTTPError("", res, errors)

	makeImage("th_"+filename, body)

	return ol
}

//AutoFocus sets the  auto-focus point
func (ol *Olympus) AutoFocus(x, y int) *Olympus {

	//set to rec/liveview mode first
	ol.Mode(_liveview, Quality.q640)

	xx, yy := fmt.Sprintf("%04d", x), fmt.Sprintf("%04d", y)
	dimen := xx + "x" + yy

	//log
	logger(dimen)

	res, body, errors := ol.client.Get(_domain + _doMotion).
		Query("com=assignafframe&point=" + dimen).
		End()
	catchHTTPError("", res, errors)

	//log
	logger(body)

	return ol
}

//Take takes a photo
func (ol *Olympus) Take(out *string) *Olympus {
	//switch to rec or shutter mode

	filename := "last.jpg"

	//check mode
	switch ol.cameraMode {
	case _shutter:
		//3. GET /exec_shutter.cgi?com=1st2ndpush HTTP/1.1
		res, _, errors := ol.client.Get(_domain + _doShutter).
			Query("com=1st2ndpush").
			End()
		catchHTTPError("", res, errors)

		//4. GET /exec_shutter.cgi?com=2nd1strelease HTTP/1.1
		res, _, errors = ol.client.Get(_domain + _doShutter).
			Query("com=2nd1strelease").
			End()
		catchHTTPError("", res, errors)

		//switch to rec mode first
		ol.Mode(_liveview, Quality.q640)

		//GET /exec_takemisc.cgi?com=getlastjpg
		res, body, errors := ol.client.Get(_domain + _doMisc).
			Query("com=getlastjpg").
			End()
		catchHTTPError("", res, errors)

		//create the image
		makeImage(filename, body)

		//update filename of last image
		*out = filename

	case _liveview:
		//5. GET /exec_takemotion.cgi?com=starttake HTTP/1.1
		res, _, errors := ol.client.Get(_domain + _doMotion).
			Query("com=starttake").
			End()
		catchHTTPError("", res, errors)

		//switch to rec mode first
		ol.Mode(_liveview, Quality.q640)

		//GET /exec_takemisc.cgi?com=getlastjpg
		//6. GET /exec_takemisc.cgi?com=getrecview HTTP/1.1
		res, body, errors := ol.client.Get(_domain + _doMisc).
			Query("com=getrecview").
			End()
		catchHTTPError("", res, errors)

		//create the image
		makeImage(filename, body)

		//update filename of last image
		*out = filename
	case _play:
		ol.Mode(_shutter, "")

		//call again
		ol.Take(out)
	default:
	}

	return ol
}

//LiveViewStart starts and ends a liveview
func (ol *Olympus) LiveViewStart(channel chan []byte) *Olympus {

	if ol.live {
		return ol
	}

	//1. GET /switch_cammode.cgi?mode=play HTTP/1.1
	ol.Mode(_play, "")

	//2. GET /switch_cammode.cgi?mode=rec&lvqty=0640x0480 HTTP/1.1
	ol.Mode(_liveview, Quality.q640)

	//3. GET /exec_takemisc.cgi?com=startliveview&port=28488 HTTP/1.1
	res, _, errs := ol.client.Get(_domain + _doMisc).
		Query("com=startliveview").
		Query("port=" + strconv.Itoa(_udpPort)).
		End()
	catchHTTPError("", res, errs)

	ol.live = true

	go func(channel chan []byte) {
		logger("Starting UDP Connection...")

		var (
			rlen int
			err  error
		)

		//UDP connection
		addr := net.UDPAddr{
			Port: _udpPort,
			IP:   net.ParseIP(_domain),
		}

		udpConnection, err = net.ListenUDP("udp", &addr)
		catch(err)
		defer ol.LiveViewStop()

		buf := make([]byte, 4000)
		jpg := make([]byte, 0)

		for ol.live {
			rlen, _, err = udpConnection.ReadFromUDP(buf)
			catch(err)

			if rlen <= 0 {
				catch(errors.New("no data in udp"))
			}

			switch buf[0] {
			case 0x90: //start
				//clear
				jpg = nil

				if index := indexSOI(buf); index >= 0 {
					jpg = append(jpg, buf[index:rlen]...)
				}
			case 0x80:
				if buf[1] == 0x60 { //mid
					jpg = append(jpg, buf[12:rlen]...)
				} else { //end
					if index := indexEOI(buf); index >= 0 {
						jpg = append(jpg, buf[12:min(rlen, index)]...)
						if jpg[0] == 0xff && jpg[1] == 0xd8 {

							//send to channel
							channel <- jpg
						}
					}
				}
			}
		}
	}(channel)

	return ol
}

//LiveViewStop stops and ends a liveview
func (ol *Olympus) LiveViewStop() *Olympus {
	if ol.live {
		logger("Stopping UDP Connection...")

		ol.live = false

		//4. GET /exec_takemisc.cgi?com=stopliveview HTTP/1.1
		res, _, errors := ol.client.Get(_domain + _doMisc).
			Query("com=stopliveview").
			End()
		catchHTTPError("", res, errors)

		//close udp
		udpConnection.Close()
	}
	return ol
}

//Utilities
func catch(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func catchHTTPError(expectedPath string, res gorequest.Response, err []error) error {
	//Standard errors
	if len(err) > 0 {
		for _, e := range err {
			catch(e)
		}
	}

	//http request errors
	if res.StatusCode != http.StatusOK {
		return errors.New("status code:" + res.Status)
	}
	path := res.Request.URL.Path
	if expectedPath != "" && path != expectedPath {
		return errors.New("request unsuccessful:" + path)
	}

	return nil
}

func makeImage(filaname string, data interface{}) {
	d := []byte{}

	switch data.(type) {
	case string:
		d = []byte(data.(string))
	default:
		d = data.([]byte)
	}

	err := ioutil.WriteFile(filaname, d, 0666)
	catch(err)
}

func logger(arg ...interface{}) {
	if _debug {
		log.Println(arg...)
	}
}

func indexSOI(buf []byte) int {
	offset := 12
	soi := []byte{0xff, 0xd8}

	//find ff d8
	if l := bytes.Split(buf[offset:], soi); len(l) == 2 {
		return len(l[0]) + offset
	}

	return -1
}

func indexEOI(buf []byte) int {
	offset := 12
	eoi := []byte{0xff, 0xd9}

	//find ff d8
	if l := bytes.Split(buf[offset:], eoi); len(l) == 2 {
		return len(l[0]) + offset + len(eoi)
	}

	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
