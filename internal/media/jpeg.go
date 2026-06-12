package media

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
)

func EncodeJPEG(image image.Image, quality int) ([]byte, error) {
	b := &bytes.Buffer{}
	if err := jpeg.Encode(b, image, &jpeg.Options{
		Quality: quality,
	}); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func estimateJPEGQuality(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	// Ensure it's a JPEG
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, fmt.Errorf("not a valid JPEG file")
	}

	idx := 2
	for idx < len(data)-1 {
		if data[idx] != 0xFF {
			idx++
			continue
		}
		marker := data[idx+1]
		if marker == 0xD9 { // End of Image
			break
		}
		// DQT Marker is 0xDB
		if marker == 0xDB {
			length := int(data[idx+2])<<8 + int(data[idx+3])
			if idx+2+length > len(data) {
				break
			}
			tableData := data[idx+4 : idx+2+length]
			return calculateQualityFromDQT(tableData), nil
		}
		idx += 2
	}
	return 0, fmt.Errorf("could not estimate quality: DQT marker missing")
}

func calculateQualityFromDQT(table []byte) int {
	// Fallback to checking the first few coefficients against IJG standards
	// A simple heuristic based on the first luminance AC coefficient:
	if len(table) < 3 {
		return 95
	}
	firstCoeff := int(table[1]) // Value at index 1 of the luma table
	if firstCoeff == 0 {
		return 100
	}
	// Approximate reverse-engineered IJG formula
	quality := 100 - firstCoeff
	if quality < 1 {
		return 1
	}
	return quality
}
