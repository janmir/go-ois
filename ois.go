//"If we have seen further it is by standing on the shoulders of giants."

package ois

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

//Olympus .
type Olympus struct {
	client     *gorequest.SuperAgent
	cameraMode int8
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
	_connectionMode = "/get_connectmode.cgi"
	_switchMode     = "/switch_cammode.cgi"
	_doMiscActions  = "/exec_takemisc.cgi"
	_doMotion       = "/exec_takemotion.cgi"
	_setProperty    = "/set_camprop.cgi" //POST

	//Modes
	_shutter = iota
	_play
	_liveview

	//Misc
	_pollingInterval = 5000 //ms
)

var (
	ticker *time.Ticker
)

//New creates a new instance of OIS
func New() *Olympus {
	ol := &Olympus{}

	//init the client object
	ol.client = gorequest.New()

	//set debug
	if _debug {
		ol.client.SetDebug(true)
	}

	//poll camera to check connection
	ticker = time.NewTicker(time.Millisecond * _pollingInterval)
	go func() {
		for t := range ticker.C {
			logger("Tick")
			logger(t)

			//call the connection checker
			ol.Connect()
			logger("Tock...")
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
	res, body, errors := ol.client.Get(_domain + _connectionMode).
		End()
	catchHTTPError("", res, errors)

	//log
	logger(body)

	return ol
}

//Mode sets the camera mode
func (ol *Olympus) Mode() *Olympus {
	//set flag
	//shutter
	//play
	//liveview
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
	res, body, errors := ol.client.Get(_domain + _imagePath).
		End()
	catchHTTPError("", res, errors)

	makeImage(filename, body)

	return ol
}

//Thumbnail retrieves thumbnail version of an image
func (ol *Olympus) Thumbnail(filename string) *Olympus {
	res, body, errors := ol.client.Get(_domain + _getThumbnail).
		Query("DIR=" + _imagePath + "/" + filename).
		End()
	catchHTTPError("", res, errors)

	makeImage("th_"+filename, body)

	return ol
}

//AutoFocus sets the  auto-focus point
func (ol *Olympus) AutoFocus(x, y int) *Olympus {
	dimen := ""

	xx, yy := fmt.Sprintf("%04d", x), fmt.Sprintf("%04d", y)
	dimen = xx + "x" + yy

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
	//check mode
	switch ol.cameraMode {
	case _shutter:
		{
			//switch_cammode.cgi?mode=play
			//exec_shutter.cgi?com=2nd1strelease
		}
	case _liveview:
		{
			// > GET /switch_cammode.cgi?mode=shutter HTTP/1.1
			// > GET /exec_takemotion.cgi?com=starttake HTTP/1.1
			// > get last taken image
			// > GET /exec_takemisc.cgi?com=getrecview HTTP/1.1
		}
	case _play:
	default:
	}
	//play
	//shutter
	//liveview

	//update filename of last image
	*out = ""

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

func makeImage(filaname, data string) {
	err := ioutil.WriteFile(filaname, []byte(data), 0666)
	catch(err)
}

func logger(arg interface{}) {
	if _debug {
		log.Println(arg)
	}
}

//Snips
//udp -> https://stackoverflow.com/questions/27176523/udp-in-golang-listen-not-a-blocking-call?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa
//gstreamer -> https://schneide.blog/2015/03/03/streaming-images-from-your-application-to-the-web-with-gstreamer-and-icecast-part-1/
