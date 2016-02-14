package gameblocks

import (
	"log"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/shazow/gameblocks/camera"
	"github.com/shazow/gameblocks/control"
	"github.com/shazow/gameblocks/loader"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

const mouseSensitivity = 0.005
const moveSpeed = 0.1

type Point struct {
	X, Y float32
}

type Engine interface {
	Draw()
	Start(glctx gl.Context) error
	Stop()

	// Event handlers
	Touch(t touch.Event)
	Press(t key.Event)
	Resize(sz size.Event)
}

func NewEngine(w World) Engine {
	cam := camera.NewQuatCamera()
	return &engine{
		camera:       cam,
		bindings:     control.DefaultBindings(),
		world:        w,
		followOffset: mgl.Vec3{0, 7, -3},
	}
}

type engine struct {
	glctx gl.Context

	camera   *camera.QuatCamera
	bindings control.Bindings
	shaders  loader.Shaders
	textures loader.Textures
	world    World

	started  time.Time
	lastTick time.Time

	touchLoc     Point
	dragOrigin   Point
	dragging     bool
	paused       bool
	gameover     bool
	following    bool
	followOffset mgl.Vec3

	size   size.Event
	images *glutil.Images
	fps    *debug.FPS
}

func (e *engine) Follow() {
	e.following = true
}

func (e *engine) Start(glctx gl.Context) error {
	e.glctx = glctx
	e.shaders = loader.ShaderLoader(glctx)
	e.textures = loader.TextureLoader(glctx)

	err := e.world.Start(WorldContext{
		Bindings: e.bindings,
		Shaders:  e.shaders,
		Textures: e.textures,
	})
	if err != nil {
		return err
	}

	e.camera.MoveTo(e.followOffset)
	e.camera.RotateTo(e.world.Focus().Position())

	// Toggle keys
	e.bindings.On(control.KeyPause, func(_ control.KeyBinding) {
		e.paused = !e.paused
		log.Println("Paused:", e.paused)

		if e.gameover {
			e.gameover = false
			e.world.Reset()
		}
	})
	e.bindings.On(control.KeyCameraFollow, func(_ control.KeyBinding) {
		e.following = !e.following
		log.Println("Following:", e.following)
	})

	e.started = time.Now()
	e.lastTick = e.started

	e.images = glutil.NewImages(glctx)
	e.fps = debug.NewFPS(e.images)

	log.Println("Starting: ", e.world.String())
	return nil
}

func (e *engine) Stop() {
	e.fps.Release()
	e.images.Release()

	e.shaders.Close()
	e.textures.Close()
}

func (e *engine) Resize(sz size.Event) {
	e.size = sz
	x, y := float32(sz.WidthPx), float32(sz.HeightPx)
	e.touchLoc.X, e.touchLoc.Y = x/2, y/2
	e.camera.SetPerspective(0.785, x/y, 0.1, 100.0)
}

func (e *engine) Touch(t touch.Event) {
	if t.Type == touch.TypeBegin {
		e.dragOrigin = Point{t.X, t.Y}
		e.dragging = true
	} else if t.Type == touch.TypeEnd {
		e.dragging = false
		log.Println("camera=", e.camera)
	}
	e.touchLoc = Point{t.X, t.Y}
	if e.dragging {
		deltaX, deltaY := float32(e.dragOrigin.X-e.touchLoc.X), float32(e.dragOrigin.Y-e.touchLoc.Y)
		e.camera.Rotate(mgl.Vec3{deltaY * mouseSensitivity, deltaX * mouseSensitivity, 0})
		e.dragOrigin = e.touchLoc
	}
}

func (e *engine) Press(t key.Event) {
	switch t.Direction {
	case key.DirPress:
		e.bindings.Press(t.Code)
	case key.DirRelease:
		e.bindings.Release(t.Code)
	}
}

func (e *engine) Draw() {
	now := time.Now()
	interval := now.Sub(e.lastTick)
	e.lastTick = now

	// Handle key presses
	var camDelta mgl.Vec3
	if e.bindings.Pressed(control.KeyCamForward) {
		camDelta[2] -= moveSpeed
	}
	if e.bindings.Pressed(control.KeyCamReverse) {
		camDelta[2] += moveSpeed
	}
	if e.bindings.Pressed(control.KeyCamLeft) {
		camDelta[0] -= moveSpeed
	}
	if e.bindings.Pressed(control.KeyCamRight) {
		camDelta[0] += moveSpeed
	}
	if e.bindings.Pressed(control.KeyCamUp) {
		e.camera.MoveTo(e.camera.Position().Add(mgl.Vec3{0, moveSpeed, 0}))
	}
	if e.bindings.Pressed(control.KeyCamDown) {
		e.camera.MoveTo(e.camera.Position().Add(mgl.Vec3{0, -moveSpeed, 0}))
	}
	if camDelta[0]+camDelta[1]+camDelta[2] != 0 {
		e.following = false
		e.camera.Move(camDelta)
	} else if e.following {
		pos := e.world.Focus().Position()
		e.camera.Lerp(pos.Add(e.followOffset), pos, 0.1)
	}

	e.glctx.ClearColor(0, 0, 0, 1)
	e.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	e.glctx.Enable(gl.DEPTH_TEST)

	if !e.paused {
		err := e.world.Tick(interval)
		if err != nil {
			e.paused = true
			e.gameover = true
		}
	}

	frame := FrameContext{
		GL:     e.glctx,
		Camera: e.camera,
	}
	e.world.Draw(frame)

	e.glctx.Disable(gl.DEPTH_TEST)

	e.fps.Draw(e.size)
}
