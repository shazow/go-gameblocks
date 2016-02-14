package gameblocks

import (
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/shazow/gameblocks/control"
	"github.com/shazow/gameblocks/loader"
)

type Vector interface {
	Position() mgl.Vec3
	Direction() mgl.Vec3
}

type WorldContext struct {
	Bindings control.Bindings
	Shaders  loader.Shaders
	Textures loader.Textures
}

type World interface {
	Scene

	Reset()
	Tick(time.Duration) error
	Focus() Vector

	Start(WorldContext) error
}

func FixedVector(position mgl.Vec3, direction mgl.Vec3) Vector {
	return &fixedVector{
		position:  position,
		direction: direction,
	}
}

type fixedVector struct {
	position  mgl.Vec3
	direction mgl.Vec3
}

func (v fixedVector) Position() mgl.Vec3  { return v.position }
func (v fixedVector) Direction() mgl.Vec3 { return v.direction }

type stubWorld struct {
	Scene
}

func (w stubWorld) Reset()                     {}
func (w stubWorld) Tick(d time.Duration) error { return nil }
func (w stubWorld) Focus() Vector              { return fixedVector{} }
func (w stubWorld) Start(_ WorldContext) error { return nil }

func StubWorld(scene Scene) World {
	return stubWorld{scene}
}
