package main

import (
	"fmt"
	gl "github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/draw"
	_ "image/png"
	"os"
)

type Texture struct {
	texture uint32
	Size    image.Point
}

func (t *Texture) Handle() uint32 {
	return t.texture
}

func (t *Texture) DeleteTexture() {
	gl.DeleteTextures(1, &t.texture)
}

func (t *Texture) BindTexture(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D, t.texture)
}

func (t *Texture) UnbindTexture(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func NewTexture(file string, linear bool) (err error, texture *Texture, img image.Image) {
	var imgFile *os.File
	if imgFile, err = os.Open(file); err != nil {
		return err, nil, img
	}
	defer imgFile.Close()

	if img, _, err = image.Decode(imgFile); err != nil {
		return err, nil, img
	}

	var rgba *image.RGBA

	rgba = image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return fmt.Errorf("unsupported stride"), nil, img
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var t uint32
	gl.GenTextures(1, &t)
	texture = &Texture{t, rgba.Rect.Size()}

	texture.BindTexture(0)
	if linear {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	} else {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return nil, texture, img
}
