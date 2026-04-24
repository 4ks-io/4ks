package function

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestValidateObjectSizeRejectsOversizeFiles(t *testing.T) {
	t.Parallel()

	err := validateObjectSize(maxUploadBytes + 1)
	if err == nil {
		t.Fatal("expected oversize file to be rejected")
	}
}

func TestLoadValidatedImageBytesRejectsUnsupportedMIME(t *testing.T) {
	t.Parallel()

	_, _, err := loadValidatedImageBytes(bytes.NewReader([]byte("not-an-image")), "image/webp")
	if err == nil {
		t.Fatal("expected unsupported MIME type to be rejected")
	}
	if !strings.Contains(err.Error(), "unsupported MIME type") {
		t.Fatalf("expected unsupported MIME type error, got %v", err)
	}
}

func TestLoadValidatedImageBytesRejectsDecompressionBombLikeDimensions(t *testing.T) {
	t.Parallel()

	payload := minimalPNG(8000, 6000)
	_, _, err := loadValidatedImageBytes(bytes.NewReader(payload), "image/png")
	if err == nil {
		t.Fatal("expected large decoded dimensions to be rejected")
	}
	if !strings.Contains(err.Error(), "pixel count exceeds limit") {
		t.Fatalf("expected pixel count limit error, got %v", err)
	}
}

func TestDecodeImageForProcessingAcceptsValidPNG(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode returned error: %v", err)
	}

	decoded, format, err := decodeImageForProcessing(bytes.NewReader(buf.Bytes()), "image/png")
	if err != nil {
		t.Fatalf("decodeImageForProcessing returned error: %v", err)
	}
	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}
	if decoded.Bounds().Dx() != 2 || decoded.Bounds().Dy() != 2 {
		t.Fatalf("expected decoded bounds 2x2, got %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}

func minimalPNG(width, height uint32) []byte {
	var out bytes.Buffer
	out.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	writePNGChunk(&out, "IHDR", pngIHDR(width, height))
	writePNGChunk(&out, "IEND", nil)
	return out.Bytes()
}

func pngIHDR(width, height uint32) []byte {
	var data bytes.Buffer
	_ = binary.Write(&data, binary.BigEndian, width)
	_ = binary.Write(&data, binary.BigEndian, height)
	data.WriteByte(8)
	data.WriteByte(2)
	data.WriteByte(0)
	data.WriteByte(0)
	data.WriteByte(0)
	return data.Bytes()
}

func writePNGChunk(out *bytes.Buffer, chunkType string, data []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	out.Write(length[:])
	out.WriteString(chunkType)
	out.Write(data)

	crc := crc32.NewIEEE()
	_, _ = crc.Write([]byte(chunkType))
	_, _ = crc.Write(data)

	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	out.Write(sum[:])
}
