package gameblocks

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/shazow/gameblocks/loader"
	"golang.org/x/mobile/gl"
)

// TODO: Load this from an .obj file in the asset repository?

var skyboxVertices = []float32{
	-1, 1, -1,
	-1, -1, -1,
	1, -1, -1,
	1, 1, -1,
	-1, -1, 1,
	-1, 1, 1,
	1, -1, 1,
	1, 1, 1,
}

var skyboxNormals = []float32{
	-1.0, -1.0, 1.0,
	1.0, -1.0, 1.0,
	1.0, 1.0, 1.0,
	-1.0, 1.0, 1.0,
	-1.0, -1.0, -1.0,
	1.0, -1.0, -1.0,
	1.0, 1.0, -1.0,
	-1.0, 1.0, -1.0,
}

var skyboxIndices = []uint8{
	0, 1, 2, 2, 3, 0,
	4, 1, 0, 0, 5, 4,
	2, 6, 7, 7, 3, 2,
	4, 5, 7, 7, 6, 4,
	0, 3, 7, 7, 5, 0,
	1, 4, 2, 2, 4, 6,
}

func NewSkybox(shader loader.Shader, texture gl.Texture) Drawable {
	glctx := shader.Context()
	skyboxShape := NewStaticShape(glctx)
	skyboxShape.vertices = skyboxVertices
	skyboxShape.indices = skyboxIndices
	skyboxShape.Buffer()
	skyboxShape.Texture = texture

	skybox := &Skybox{
		Node: Node{
			Shape:  skyboxShape,
			shader: shader,
		},
	}

	return skybox
}

type Skybox struct {
	Node
}

func (shape *Skybox) Shader() loader.Shader {
	return shape.shader
}

func (shape *Skybox) Transform(parent *mgl.Mat4) mgl.Mat4 {
	return mgl.Ident4()
}

func (node *Skybox) Draw(ctx DrawContext) {
	cam := ctx.Camera
	shader := ctx.Shader
	glctx := ctx.GL

	shape := node.Shape.(*StaticShape)

	glctx.DepthFunc(gl.LEQUAL)
	glctx.DepthMask(false)

	// FIXME: This overrides the scene renderer, should bypass that work somehow.
	projection, view := cam.Projection(), cam.View().Mat3().Mat4()
	glctx.UniformMatrix4fv(shader.Uniform("projection"), projection[:])
	glctx.UniformMatrix4fv(shader.Uniform("view"), view[:])

	glctx.BindTexture(gl.TEXTURE_CUBE_MAP, shape.Texture)

	glctx.BindBuffer(gl.ARRAY_BUFFER, shape.VBO)
	glctx.EnableVertexAttribArray(shader.Attrib("vertCoord"))
	glctx.VertexAttribPointer(shader.Attrib("vertCoord"), vertexDim, gl.FLOAT, false, shape.Stride(), 0)

	glctx.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, shape.IBO)
	glctx.DrawElements(gl.TRIANGLES, len(shape.indices), gl.UNSIGNED_BYTE, 0)
	glctx.DisableVertexAttribArray(shader.Attrib("vertCoord"))

	glctx.DepthMask(true)
	glctx.DepthFunc(gl.LESS)
}

var floorVertices = []float32{
	-100, 0, -100,
	100, 0, -100,
	100, 0, 100,
	100, 0, 100,
	-100, 0, -100,
	-100, 0, 100,
}

var floorNormals = []float32{
	0, 1, 0,
	0, 1, 0,
	0, 1, 0,
	0, 1, 0,
	0, 1, 0,
	0, 1, 0,
}

type Floor struct {
	Node
	reflected []Drawable
}

func (scene *Floor) Draw(ctx DrawContext) {
	shader := ctx.Shader
	glctx := ctx.GL

	glctx.Enable(gl.STENCIL_TEST)
	glctx.StencilFunc(gl.ALWAYS, 1, 0xFF)
	glctx.StencilOp(gl.KEEP, gl.KEEP, gl.REPLACE)
	glctx.StencilMask(0xFF)
	glctx.DepthMask(false)
	glctx.Clear(gl.STENCIL_BUFFER_BIT)

	// Draw floor
	glctx.Uniform3fv(shader.Uniform("material.ambient"), []float32{0.2, 0.2, 0.35})
	scene.Shape.Draw(ctx)

	// Draw reflections
	glctx.StencilFunc(gl.EQUAL, 1, 0xFF)
	glctx.StencilMask(0x00)
	glctx.DepthMask(true)

	view := ctx.Camera.View()
	glctx.Uniform3fv(shader.Uniform("material.ambient"), []float32{0.6, 0.6, 0.6})
	for _, node := range scene.reflected {
		model := node.Transform(scene.transform)
		glctx.UniformMatrix4fv(shader.Uniform("model"), model[:])

		normal := model.Mul4(view).Inv().Transpose()
		glctx.UniformMatrix4fv(shader.Uniform("normalMatrix"), normal[:])

		node.Draw(ctx)
	}

	glctx.Disable(gl.STENCIL_TEST)
}

func NewFloor(shader loader.Shader, reflected ...Drawable) Drawable {
	glctx := shader.Context()
	floor := NewStaticShape(glctx)
	floor.vertices = floorVertices
	floor.normals = floorNormals
	floor.Buffer()
	flipped := mgl.Scale3D(1, -1, 1)
	return &Floor{
		Node: Node{
			Shape:     floor,
			transform: &flipped,
			shader:    shader,
		},
		reflected: reflected,
	}
}
