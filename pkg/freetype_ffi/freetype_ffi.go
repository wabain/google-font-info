package freetype_ffi

// Use the C freetype library to get font metrics.

// #cgo pkg-config: freetype2
// #include <ft2build.h>
// #include FT_FREETYPE_H
import "C"
import (
	"fmt"
	"unsafe"
)

type Freetype struct {
	lib C.FT_Library
}

// FaceMetrics Some basic metrics extracted use Freetype
type FaceMetrics struct {
	EmSize  uint16
	Ascent  int16
	Descent int16
	Height  int16
}

func FreetypeInit() (*Freetype, error) {
	var lib C.FT_Library

	ftErr := C.FT_Init_FreeType(&lib)
	if ftErr != 0 {
		return nil, fmtFtErr(ftErr)
	}

	return &Freetype{lib}, nil
}

func FreetypeDone(freetype *Freetype) error {
	ftErr := C.FT_Done_FreeType((*freetype).lib)
	return fmtFtErr(ftErr)
}

// GetFaceMetrics Use the Freetype2 library to get font metric information
func GetFaceMetrics(freetype *Freetype, fontPath string) (*FaceMetrics, error) {
	cFontPath := C.CString(fontPath)
	if cFontPath == nil {
		return nil, fmt.Errorf("Malloc failure")
	}
	defer C.free(unsafe.Pointer(cFontPath))

	var face C.FT_Face

	ftErr := C.FT_New_Face((*freetype).lib, cFontPath, 0, &face)
	if ftErr != 0 {
		return nil, fmtFtErr(ftErr)
	}
	defer C.FT_Done_Face(face)

	metrics := FaceMetrics{
		EmSize:  uint16(face.units_per_EM),
		Ascent:  int16(face.ascender),
		Descent: int16(-face.descender),
		Height:  int16(face.height),
	}

	return &metrics, nil
}

func fmtFtErr(ftErr C.FT_Error) error {
	if ftErr == 0 {
		return nil
	}
	return fmt.Errorf("Freetype error %d", ftErr)
}
