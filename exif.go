package main

import (
	"io"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ExifData — извлечённые метаданные
type ExifData struct {
	CameraMake  string  `json:"camera_make,omitempty"`
	CameraModel string  `json:"camera_model,omitempty"`
	DateTime    string  `json:"datetime,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	ISO         int     `json:"iso,omitempty"`
	FNumber     string  `json:"f_number,omitempty"`
	Exposure    string  `json:"exposure,omitempty"`
}

// ExtractExif пытается прочитать EXIF из потока.
// Если данных нет или файл не поддерживается — возвращает nil без ошибки.
func ExtractExif(r io.Reader) *ExifData {
	x, err := exif.Decode(r)
	if err != nil {
		// Не изображение или нет EXIF — это нормально
		return nil
	}

	data := &ExifData{}

	if tag, err := x.Get(exif.Make); err == nil {
		if v, err := tag.StringVal(); err == nil {
			data.CameraMake = v
		}
	}

	if tag, err := x.Get(exif.Model); err == nil {
		if v, err := tag.StringVal(); err == nil {
			data.CameraModel = v
		}
	}

	if t, err := x.DateTime(); err == nil {
		data.DateTime = t.Format(time.RFC3339)
	}

	if tag, err := x.Get(exif.PixelXDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			data.Width = v
		}
	}

	if tag, err := x.Get(exif.PixelYDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			data.Height = v
		}
	}

	// GPS-координаты
	if lat, lon, err := x.LatLong(); err == nil {
		data.Latitude = lat
		data.Longitude = lon
	}

	// ISO
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if v, err := tag.Int(0); err == nil {
			data.ISO = v
		}
	}

	return data
}