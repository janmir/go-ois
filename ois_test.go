package ois

import (
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	o := New()
	time.Sleep(time.Second * 60)
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
	o.Image("P5210187.JPG")
}

func TestGetThumbnail(t *testing.T) {
	o := New()
	o.Thumbnail("P5210187.JPG")
}
