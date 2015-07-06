package main

import (
	gl "github.com/go-gl/gl/v4.1-core/gl"
	"unsafe"
)

type VBO struct {
	vao        uint32
	vboVerts   uint32
	vboIndices uint32
}

func (v *VBO) DeleteVBO() {
	if v.vboVerts != 0 {
		gl.DeleteBuffers(1, &v.vboVerts)
	}
	if v.vboIndices != 0 {
		gl.DeleteBuffers(1, &v.vboIndices)
	}
	if v.vao != 0 {
		gl.DeleteVertexArrays(1, &v.vao)
	}
}

func (v *VBO) Bind() {
	gl.BindVertexArray(v.vao)
	gl.EnableVertexAttribArray(gPosition_attr)
	gl.EnableVertexAttribArray(gUVs_attr)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.vboIndices)
}

func (v *VBO) Unbind() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	gl.DisableVertexAttribArray(gPosition_attr)
	gl.DisableVertexAttribArray(gUVs_attr)
	gl.BindVertexArray(0)
}

func (v *VBO) Draw() {
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	//gl.DrawElements(gl.LINE_STRIP, 6, gl.UNSIGNED_INT, nil)
}

func (v *VBO) DrawQuads(nquads int) {
	gl.DrawElements(gl.TRIANGLES, int32(nquads*6), gl.UNSIGNED_INT, nil)
	//gl.DrawElements(gl.LINE_STRIP, 6, gl.UNSIGNED_INT, nil)
}

func (v *VBO) Load(verts *float32, vsize int, indices *uint32, isize int) {
	// calculate the memory size of floats used to calculate total memory size of float arrays
	var floatSize int = int(unsafe.Sizeof(float32(1.0)))
	var intSize int = int(unsafe.Sizeof(uint32(1)))

	gl.BindBuffer(gl.ARRAY_BUFFER, v.vboVerts)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.vboIndices)

	// load our data up and bind it to the 'position' shader attribute
	gl.BufferData(gl.ARRAY_BUFFER, floatSize*vsize, unsafe.Pointer(verts), gl.STATIC_DRAW)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, intSize*isize, unsafe.Pointer(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(gPosition_attr, 3, gl.FLOAT, false, 20, gl.PtrOffset(0))
	gl.VertexAttribPointer(gUVs_attr, 2, gl.FLOAT, false, 20, gl.PtrOffset(12))

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func NewVBO() (vbo *VBO) {
	// create and bind the required VAO object
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// create a VBO to hold the vertex data
	var vboVerts uint32
	var vboIndices uint32
	gl.GenBuffers(1, &vboVerts)
	gl.GenBuffers(1, &vboIndices)

	vbo = &VBO{vao, vboVerts, vboIndices}
	return vbo
}

func NewVBOQuad(x float32, y float32, w float32, h float32) (vbo *VBO, err error) {
	vbo = NewVBO()

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

	vbo.Load(&verts[0], len(verts), &indices[0], len(indices))

	return vbo, nil
}
