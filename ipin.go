package ipingo

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
)

func GetNormalizedPNG(filename string) ([]byte, error) {
	pngheader := []byte("\x89PNG\r\n\x1a\n")

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	oldPNG, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(oldPNG, pngheader) {
		return nil, fmt.Errorf("file:%v has no pngheader", filename)
	}

	idatAcc := []byte{}
	breakLoop := false

	width := 0
	height := 0

	newPNG := pngheader[:]
	chunkPos := len(newPNG)

	// For each chunk in the PNG file
	for chunkPos < len(oldPNG) {
		skip := false

		// Reading chunk
		chunkLengthBytes := oldPNG[chunkPos : chunkPos+4]
		chunkLength := unpackBingEndianInt32(chunkLengthBytes)
		chunkType := oldPNG[chunkPos+4 : chunkPos+8]
		chunkData := oldPNG[chunkPos+8 : chunkPos+8+chunkLength]
		chunkCRCBytes := oldPNG[chunkPos+chunkLength+8 : chunkPos+chunkLength+12]
		chunkCRC := unpackBingEndianInt32(chunkCRCBytes)
		chunkPos += chunkLength + 12

		// Parsing the header chunk
		if string(chunkType) == "IHDR" {
			width = unpackBingEndianInt32(chunkData[0:4])
			height = unpackBingEndianInt32(chunkData[4:8])
		}
		// Parsing the image chunk
		if string(chunkType) == "IDAT" {
			// Store the chunk data for later decompression
			idatAcc = append(idatAcc, chunkData...)
			skip = true
		}

		// Removing CgBI chunk
		if string(chunkType) == "CgBI" {
			skip = true
		}

		// Add all accumulated IDATA chunks
		if string(chunkType) == "IEND" {
			// Uncompressing the image chunk
			chunkData, err = decompress(idatAcc)
			if err != nil {
				// The PNG image is normalized
				return nil, err
			}

			chunkType = []byte("IDAT")

			// Swapping red & blue bytes for each pixel
			newdata := []byte{}
			for y := 0; y < height; y++ {
				i := len(newdata)
				newdata = append(newdata, chunkData[i])
				for x := 0; x < width; x++ {
					i = len(newdata)
					newdata = append(newdata, chunkData[i+2])
					newdata = append(newdata, chunkData[i+1])
					newdata = append(newdata, chunkData[i+0])
					newdata = append(newdata, chunkData[i+3])
				}
			}

			// Compressing the image chunk
			chunkData = newdata
			chunkData, err = compress(chunkData)
			if err != nil {
				return nil, err
			}
			chunkLength = len(chunkData)
			chunkCRC := crc32.ChecksumIEEE(chunkType)
			chunkCRC = crc32.Update(chunkCRC, crc32.IEEETable, chunkData)
			chunkCRC = uint32((uint64(chunkCRC) + 0x100000000) % 0x100000000)
			breakLoop = true
		}

		if !skip {
			newPNG = append(newPNG, packBingEndianInt32(chunkLength)...)
			newPNG = append(newPNG, chunkType...)
			if chunkLength > 0 {
				newPNG = append(newPNG, chunkData...)
			}
			newPNG = append(newPNG, packBingEndianInt32(chunkCRC)...)
		}

		if breakLoop {
			break
		}
	}

	return newPNG, nil
}

func packBingEndianInt32(i int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

func unpackBingEndianInt32(b []byte) int {
	return int(binary.BigEndian.Uint32(b))
}

func decompress(src []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(src))
	defer r.Close()
	var out bytes.Buffer

	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func compress(src []byte) ([]byte, error) {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	defer w.Close()

	if _, err := w.Write(src); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return in.Bytes(), nil
}
