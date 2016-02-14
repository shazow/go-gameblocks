package gameblocks

import (
	"fmt"

	"golang.org/x/mobile/gl"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/shazow/gameblocks/camera"
	"github.com/shazow/gameblocks/loader"
)

type FrameContext struct {
	GL     gl.Context
	Camera camera.Camera

	shaderCache  map[loader.Shader]struct{}
	activeShader loader.Shader
}

func (ctx *FrameContext) bindShader(shader loader.Shader) {
	cam, glctx := ctx.Camera, ctx.GL
	projection, view, position := cam.Projection(), cam.View(), cam.Position()

	glctx.UniformMatrix4fv(shader.Uniform("cameraPos"), position[:])
	glctx.UniformMatrix4fv(shader.Uniform("view"), view[:])
	glctx.UniformMatrix4fv(shader.Uniform("projection"), projection[:])
}

func (ctx *FrameContext) DrawContext(shader loader.Shader) DrawContext {
	r := DrawContext{
		GL:     ctx.GL,
		Camera: ctx.Camera,
		Shader: shader,
	}
	if ctx.activeShader != shader {
		shader.Use()
		ctx.activeShader = shader
	}
	if ctx.shaderCache == nil {
		ctx.shaderCache = map[loader.Shader]struct{}{shader: struct{}{}}
		ctx.bindShader(shader)
	} else if _, ok := ctx.shaderCache[shader]; !ok {
		ctx.bindShader(shader)
	}
	return r
}

type DrawContext struct {
	GL        gl.Context
	Camera    camera.Camera
	Shader    loader.Shader
	Transform *mgl.Mat4
}

type Light struct {
	color    mgl.Vec3
	position mgl.Vec3
}

func (light *Light) MoveTo(position mgl.Vec3) {
	light.position = position
}

type Drawable interface {
	Draw(DrawContext)
	Transform(*mgl.Mat4) mgl.Mat4
	Shader() loader.Shader
}

// TODO: node tree with transforms
type Node struct {
	Shape
	transform *mgl.Mat4
	shader    loader.Shader
}

func (node *Node) Shader() loader.Shader {
	return node.shader
}

func (node *Node) Draw(ctx DrawContext) {
	view := ctx.Camera.View()
	model := node.Transform(ctx.Transform)
	normal := model.Mul4(view).Inv().Transpose()

	// Camera space
	ctx.GL.UniformMatrix4fv(ctx.Shader.Uniform("model"), model[:])
	ctx.GL.UniformMatrix4fv(ctx.Shader.Uniform("normalMatrix"), normal[:])

	// Bubble to shape
	node.Shape.Draw(ctx)
}

func (node *Node) Transform(parent *mgl.Mat4) mgl.Mat4 {
	return MultiMul(node.transform, parent)
}

func (node *Node) String() string {
	return fmt.Sprintf("<Shape of %d vertices; transform: %v>", node.Len(), node.transform)
}

type Scene interface {
	Add(Drawable)
	Draw(FrameContext)
	String() string
}

func NewScene() Scene {
	return &sliceScene{
		nodes: []Drawable{},
	}
}

type sliceScene struct {
	nodes     []Drawable
	transform *mgl.Mat4
}

func (scene *sliceScene) String() string {
	return fmt.Sprintf("%d nodes", len(scene.nodes))
}

func (scene *sliceScene) Add(item Drawable) {
	scene.nodes = append(scene.nodes, item)
}

func (scene *sliceScene) Draw(frame FrameContext) {
	for _, node := range scene.nodes {
		ctx := frame.DrawContext(node.Shader())
		ctx.Transform = scene.transform
		node.Draw(ctx)
	}
}
