package ois

import (
	"os"
	"strconv"
	"testing"
)

func TestLiveView(t *testing.T) {
	max := 10
	o := New()
	c := make(chan []byte, 3)

	o.LiveViewStart(c)

	i := 0
	for {
		for b := range c {
			name := "img_" + strconv.Itoa(i) + ".jpg"
			makeImage(name, b)
			i++

			if i >= max {
				o.LiveViewStop()
				os.Exit(0)
			}
		}
	}
}

func TestConnect(t *testing.T) {
	o := New()
	//time.Sleep(time.Second * 60)
	o.Connect()
}

func TestShutdown(t *testing.T) {
	o := New()
	o.Shutdown()
}

func TestList(t *testing.T) {
	o := New()
	o.List()
}

func TestGetImage(t *testing.T) {
	o := New()
	o.Image("P5100027.JPG")
}

func TestGetResize(t *testing.T) {
	o := New()
	o.Resize("P5100027.ORF", Res.r1024)
}

func TestGetThumbnail(t *testing.T) {
	o := New()
	o.Thumbnail("P5100027.ORF")
}
