package main

import (
	"fmt"
	gl "github.com/go-gl/gl/v4.1-core/gl"
	"image"
	"image/draw"
	_ "image/png"
	"os"
)

var (
	// fragment shader
	fragShaderFont = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
  uniform sampler2D tex1;
  uniform vec4 color;
  uniform vec4 bg;
  void main()
  {
    vec4 col0 = texture(tex1, out_uvs);
    colourOut = col0.r * color;
    // Porter duff gl.ONE, gl.ONE_MINUS_SRC_ALPHA
    colourOut = vec4(colourOut.r + bg.r * (1-colourOut.a), colourOut.g + bg.g * (1-colourOut.a), colourOut.b + bg.b * (1-colourOut.a), colourOut.a + bg.a * (1-colourOut.a));
  }`
)

var gWidths = [...]int{
	19,  // 32
	24,  // 33
	35,  // 34
	47,  // 35
	47,  // 36
	75,  // 37
	86,  // 38
	19,  // 39
	31,  // 40
	31,  // 41
	31,  // 42
	47,  // 43
	26,  // 44
	38,  // 45
	26,  // 46
	35,  // 47
	47,  // 48
	47,  // 49
	47,  // 50
	47,  // 51
	47,  // 52
	47,  // 53
	47,  // 54
	47,  // 55
	47,  // 56
	47,  // 57
	26,  // 58
	26,  // 59
	47,  // 60
	47,  // 61
	47,  // 62
	36,  // 63
	75,  // 64
	68,  // 65
	65,  // 66
	71,  // 67
	79,  // 68
	65,  // 69
	59,  // 70
	79,  // 71
	82,  // 72
	39,  // 73
	34,  // 74
	70,  // 75
	62,  // 76
	93,  // 77
	76,  // 78
	81,  // 79
	60,  // 80
	82,  // 81
	69,  // 82
	54,  // 83
	69,  // 84
	69,  // 85
	64,  // 86
	99,  // 87
	65,  // 88
	58,  // 89
	66,  // 90
	33,  // 91
	35,  // 92
	33,  // 93
	47,  // 94
	47,  // 95
	22,  // 96
	43,  // 97
	49,  // 98
	41,  // 99
	51,  // 100
	42,  // 101
	31,  // 102
	47,  // 103
	53,  // 104
	27,  // 105
	25,  // 106
	49,  // 107
	27,  // 108
	77,  // 109
	53,  // 110
	49,  // 111
	51,  // 112
	50,  // 113
	34,  // 114
	36,  // 115
	32,  // 116
	51,  // 117
	45,  // 118
	64,  // 119
	40,  // 120
	42,  // 121
	43,  // 122
	33,  // 123
	33,  // 124
	33,  // 125
	47,  // 126
	47,  // 127
	47,  // 128
	47,  // 129
	19,  // 130
	47,  // 131
	38,  // 132
	94,  // 133
	47,  // 134
	47,  // 135
	29,  // 136
	109, // 137
	54,  // 138
	25,  // 139
	109, // 140
	47,  // 141
	66,  // 142
	47,  // 143
	47,  // 144
	19,  // 145
	19,  // 146
	38,  // 147
	38,  // 148
	32,  // 149
	47,  // 150
	94,  // 151
	32,  // 152
	86,  // 153
	36,  // 154
	25,  // 155
	71,  // 156
	47,  // 157
	43,  // 158
	58,  // 159
	19,  // 160
	24,  // 161
	47,  // 162
	47,  // 163
	47,  // 164
	47,  // 165
	33,  // 166
	51,  // 167
	34,  // 168
	78,  // 169
	34,  // 170
	41,  // 171
	47,  // 172
	38,  // 173
	53,  // 174
	28,  // 175
	30,  // 176
	47,  // 177
	34,  // 178
	34,  // 179
	22,  // 180
	51,  // 181
	54,  // 182
	26,  // 183
	25,  // 184
	34,  // 185
	37,  // 186
	41,  // 187
	75,  // 188
	75,  // 189
	75,  // 190
	34,  // 191
	68,  // 192
	68,  // 193
	68,  // 194
	68,  // 195
	68,  // 196
	68,  // 197
	90,  // 198
	71,  // 199
	65,  // 200
	65,  // 201
	65,  // 202
	65,  // 203
	39,  // 204
	39,  // 205
	39,  // 206
	39,  // 207
	79,  // 208
	76,  // 209
	81,  // 210
	81,  // 211
	81,  // 212
	81,  // 213
	81,  // 214
	47,  // 215
	81,  // 216
	69,  // 217
	69,  // 218
	69,  // 219
	69,  // 220
	58,  // 221
	61,  // 222
	54,  // 223
	43,  // 224
	43,  // 225
	43,  // 226
	43,  // 227
	43,  // 228
	43,  // 229
	64,  // 230
	41,  // 231
	42,  // 232
	42,  // 233
	42,  // 234
	42,  // 235
	27,  // 236
	27,  // 237
	27,  // 238
	27,  // 239
	48,  // 240
	53,  // 241
	49,  // 242
	49,  // 243
	49,  // 244
	49,  // 245
	49,  // 246
	47,  // 247
	49,  // 248
	51,  // 249
	51,  // 250
	51,  // 251
	51,  // 252
	42,  // 253
	51,  // 254
	42,  // 255
}

//
// Char
//

type Char struct {
	Index int
	X     int
	Y     int
}

//
// String
//

type String struct {
	Chars []Char
	Size  image.Point
	vbo   *VBO
}

func (s *String) DeleteString() {
	if s.vbo != nil {
		s.vbo.DeleteVBO()
	}
}

func (s *String) Draw(f *Font, color [4]float32, bg [4]float32, mat Matrix2x3, scale float32, offsetX float32, offsetY float32) (err error) {
	if s.vbo == nil {
		s.createVertexBuffer(f)
	}
	if f.program == nil {
		var attribs []string = []string{
			"position",
			"uvs",
		}
		if f.program, err = LoadShaderProgram(vertShader, fragShaderFont, attribs); err != nil {
			return (err)
		}
	}

	f.program.UseProgram()
	f.texture.BindTexture(0)
	s.vbo.Bind()

	var matrix_font Matrix2x3 = mat.Scale(scale, scale)
	matrix_font = matrix_font.Translate(offsetX, offsetY)
	f.program.ProgramUniformMatrix4fv("ModelviewMatrix", matrix_font.Array())
	f.program.ProgramUniform1i("tex1", 0)
	f.program.ProgramUniform4fv("color", color)
	f.program.ProgramUniform4fv("bg", bg)

	if err = f.program.ValidateProgram(); err != nil {
		return err
	}

	s.vbo.DrawQuads(len(s.Chars))

	f.texture.UnbindTexture(0)
	s.vbo.Unbind()
	f.program.UnuseProgram()

	return nil
}

func (s *String) createVertexBuffer(f *Font) {
	s.vbo = NewVBO()

	n := len(s.Chars)

	verts := make([]float32, n*20)
	indices := make([]uint32, n*6)

	/*
		verts := [...]float32{
			x, y, 0.0, 0, 0,
			x + w, y, 0.0, 1, 0,
			x + w, y + h, 0.0, 1, 1,
			x, y + h, 0.0, 0, 1,
		}

		indices := [...]uint32{
			0, 1, 2,
			2, 3, 0,
		}
	*/
	var curX float32 = 0
	i := 0
	ii := 0
	var jj uint32 = 0
	var dv = float32(f.cell_size) / float32(f.texture.Size.Y)
	for j := 0; j < n; j++ {
		var c Char = s.Chars[j]
		var x float32 = curX
		var y float32 = 0
		var w float32 = float32(gWidths[c.Index])
		var h float32 = float32(f.cell_size)
		var u float32 = float32(c.X*f.cell_size) / float32(f.texture.Size.X)
		var v float32 = float32(c.Y*f.cell_size) / float32(f.texture.Size.Y)
		var du = float32(w) / float32(f.texture.Size.X)

		verts[i+0] = x
		verts[i+1] = y
		verts[i+2] = 0
		verts[i+3] = u
		verts[i+4] = v
		i += 5

		verts[i+0] = x + w
		verts[i+1] = y
		verts[i+2] = 0
		verts[i+3] = u + du
		verts[i+4] = v
		i += 5

		verts[i+0] = x + w
		verts[i+1] = y + h
		verts[i+2] = 0
		verts[i+3] = u + du
		verts[i+4] = v + dv
		i += 5

		verts[i+0] = x
		verts[i+1] = y + h
		verts[i+2] = 0
		verts[i+3] = u
		verts[i+4] = v + dv
		i += 5

		indices[ii+0] = 0 + jj
		indices[ii+1] = 1 + jj
		indices[ii+2] = 2 + jj
		indices[ii+3] = 2 + jj
		indices[ii+4] = 3 + jj
		indices[ii+5] = 0 + jj
		ii += 6
		jj += 4

		curX += w
	}

	s.vbo.Load(&verts[0], 20*n, &indices[0], 6*n)
}

//
// Font
//

type Font struct {
	texture   *Texture
	program   *Program
	rows      int
	cell_size int
}

func (f *Font) DeleteFont() {
	if f.texture != nil {
		f.texture.DeleteTexture()
	}
	if f.program != nil {
		f.program.DeleteProgram()
	}
}

func (f *Font) BindTexture(unit uint32) {
	f.texture.BindTexture(unit)
}

func (f *Font) UnbindTexture(unit uint32) {
	f.texture.UnbindTexture(unit)
}

func (f *Font) NewString(s string) *String {
	var result *String = &String{make([]Char, len(s)), image.Point{0, 0}, nil}
	var width int = 0
	for i := 0; i < len(s); i++ {
		var ascii = int(s[i])
		var index = ascii - 32
		var xoff = index % f.rows
		var yoff = index / f.rows
		width += gWidths[index]

		//fmt.Printf("ascii: %d, x: %d, y: %d\n", ascii, xoff, yoff)
		result.Chars[i].Index = index
		result.Chars[i].X = xoff
		result.Chars[i].Y = yoff

	}
	result.Size = image.Point{width, f.cell_size}
	return result
}

func NewFont(file string, rows int) (err error, font *Font) {
	var imgFile *os.File
	if imgFile, err = os.Open(file); err != nil {
		return err, nil
	}
	defer imgFile.Close()

	var img image.Image
	if img, _, err = image.Decode(imgFile); err != nil {
		return err, nil
	}

	gray := image.NewGray(img.Bounds())
	if gray.Stride != gray.Rect.Size().X {
		return fmt.Errorf("unsupported stride"), nil
	}
	draw.Draw(gray, gray.Bounds(), img, image.Point{0, 0}, draw.Src)

	var t uint32
	gl.GenTextures(1, &t)
	var texture *Texture = &Texture{t, gray.Rect.Size()}

	texture.BindTexture(0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.R8,
		int32(gray.Rect.Size().X),
		int32(gray.Rect.Size().Y),
		0,
		gl.RED,
		gl.UNSIGNED_BYTE,
		gl.Ptr(gray.Pix))

	font = &Font{texture, nil, rows, gray.Rect.Size().X / rows}

	return nil, font
}
