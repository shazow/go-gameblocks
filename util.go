package gameblocks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	_ "image/png"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type dimslice_float32 struct {
	dim   int
	slice []float32
}

func (o dimslice_float32) Slice(i, j int) interface{} { return o.slice[i:j] }
func (o dimslice_float32) Dim() int                   { return o.dim }
func (o dimslice_float32) String() string {
	return fmt.Sprintf("<float32 slice: len=%d dim=%d>", len(o.slice), o.dim)
}

type dimslice_uint8 struct {
	dim   int
	slice []uint8
}

func (o dimslice_uint8) Slice(i, j int) interface{} { return o.slice[i:j] }
func (o dimslice_uint8) Dim() int                   { return o.dim }
func (o dimslice_uint8) String() string {
	return fmt.Sprintf("<uint8 slice: len=%d dim=%d>", len(o.slice), o.dim)
}

func NewDimSlice(dim int, slice interface{}) DimSlicer {
	switch slice := slice.(type) {
	case []float32:
		return &dimslice_float32{dim, slice}
	case []uint8:
		return &dimslice_uint8{dim, slice}
	}
	panic(fmt.Sprintf("invalid slice type: %T", slice))
}

type DimSlicer interface {
	Slice(int, int) interface{}
	Dim() int
	String() string
}

// EncodeObjects converts float32 vertices into a LittleEndian byte array.
// Offset and length are based on the number of rows per dimension.
// TODO: Replace with https://github.com/lunixbochs/struc?
func EncodeObjects(offset int, length int, objects ...DimSlicer) []byte {
	//log.Println("EncodeObjects:", offset, length, objects)
	// TODO: Pre-allocate? Use a SyncPool?
	/*
		dimSum := 0 // yum!
		for _, obj := range objects {
			dimSum += obj.Dim()
		}
		v := make([]float32, dimSum*length)
	*/

	buf := bytes.Buffer{}

	for i := offset; i < length; i++ {
		for _, obj := range objects {
			data := obj.Slice(i*obj.Dim(), (i+1)*obj.Dim())
			if err := binary.Write(&buf, binary.LittleEndian, data); err != nil {
				panic(fmt.Sprintln("binary.Write failed:", err))
			}
		}
	}
	//fmt.Printf("Wrote %d vertices: %d to %d \t", shape.Len()-n, n, shape.Len())
	//fmt.Println(wrote)

	return buf.Bytes()
}

// MultiMul multiplies every non-nil Mat4 reference and returns the result. If
// none are given, then it returns the identity matrix.
func MultiMul(matrices ...*mgl.Mat4) mgl.Mat4 {
	var r mgl.Mat4
	ok := false
	for _, m := range matrices {
		if m == nil {
			continue
		}
		if !ok {
			r = *m
			ok = true
			continue
		}
		r = r.Mul4(*m)
	}
	if ok {
		return r
	}
	return mgl.Ident4()
}

func Quad(a mgl.Vec3, b mgl.Vec3) []float32 {
	return []float32{
		// First triangle
		b[0], b[1], b[2], // Top Right
		a[0], b[1], a[2], // Top Left
		a[0], a[1], a[2], // Bottom Left
		// Second triangle
		a[0], a[1], a[2], // Bottom Left
		b[0], b[1], b[2], // Top Right
		b[0], a[1], b[2], // Bottom Right
	}
}

func Upvote(tip mgl.Vec3, size float32) []float32 {
	a := tip.Add(mgl.Vec3{-size / 2, -size * 2, 0})
	b := tip.Add(mgl.Vec3{size / 2, -size, 0})
	return []float32{
		tip[0], tip[1], tip[2], // Top
		tip[0] - size, tip[1] - size, tip[2], // Bottom left
		tip[0] + size, tip[1] - size, tip[2], // Bottom right

		// Arrow handle
		b[0], b[1], b[2], // Top Right
		a[0], b[1], a[2], // Top Left
		a[0], a[1], a[2], // Bottom Left
		a[0], a[1], a[2], // Bottom Left
		b[0], b[1], b[2], // Top Right
		b[0], a[1], b[2], // Bottom Right
	}
}
