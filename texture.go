package main

import (
	"bufio"
	"golang.org/x/image/bmp"
	"image"
	"image/color"
	"os"
	"path/filepath"
)

func saveTextures(outPath string, mdl *Mdl) error {
	var (
		err      error
		filePath string
		file     *os.File
		writer   *bufio.Writer
	)
	var palette = make([]color.Color, 256)

	for _, tex := range mdl.Textures {
		func() {
			filePath = filepath.Join(outPath, tex.Name.String())

			if err = os.RemoveAll(filePath); err != nil {
				printError(err)
				return
			}

			file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				printError(err)
				return
			}
			defer file.Close()

			writer = bufio.NewWriter(file)
			defer writer.Flush()

			texIndices, texPalette := &tex.Indices, &tex.Pallets
			width, height := int(tex.Width), int(tex.Height)

			for p := 0; p < 256*3; p += 3 {
				palette[p/3] = color.RGBA{
					R: (*texPalette)[p],
					G: (*texPalette)[p+1],
					B: (*texPalette)[p+2],
					A: 0xff}
			}

			img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					img.SetColorIndex(x, y, (*texIndices)[y*width+x])
				}
			}

			err = bmp.Encode(writer, img)
			if err != nil {
				printError(err)
			}
		}()
	}

	return nil
}
