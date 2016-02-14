package gameblocks

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/gl"
)

// TODO: ...
// Ref: https://github.com/krux02/turnt-octo-wallhack/blob/master/particles/ParticleSystem.go

type Emitter interface {
	Shape
	Tick(time.Duration)
	MoveTo(mgl.Vec3)
}

func RandomParticle(origin mgl.Vec3, force float32) *particle {
	return &particle{
		position: origin,
		velocity: mgl.Vec3{
			(0.5 - rand.Float32()) * force,
			rand.Float32() * force,
			(0.5 - rand.Float32()) * force,
		},
	}

}

const particleLen = 9 * 3

type particle struct {
	position mgl.Vec3
	velocity mgl.Vec3
}

func (p *particle) Vertices() []float32 {
	v := p.velocity.Normalize()
	return []float32{
		p.position[0] + v[0]*0.1,
		p.position[1] + v[1]*0.1,
		p.position[2] + v[2]*0.1,
		p.position[0] - v[0]*0.05,
		p.position[1] - v[1]*0.05,
		p.position[2] + v[2]*0.05,
		p.position[0] + v[0]*0.05,
		p.position[1] - v[1]*0.05,
		p.position[2] + v[2]*0.05,
	}
}

func (p *particle) Tick(force mgl.Vec3) {
	p.velocity = force.Add(p.velocity)
	p.position = p.position.Add(p.velocity)
}

var particleForce float32 = 0.09
var gravityForce = mgl.Vec3{0, -0.2, 0}

func ParticleEmitter(glctx gl.Context, origin mgl.Vec3, num int, rate float32) Emitter {
	bufSize := num * particleLen * vecSize
	vbo := glctx.CreateBuffer()
	glctx.BindBuffer(gl.ARRAY_BUFFER, vbo)
	glctx.BufferInit(gl.ARRAY_BUFFER, bufSize, gl.DYNAMIC_DRAW)

	return &particleEmitter{
		glctx:     glctx,
		VBO:       vbo,
		origin:    origin,
		rate:      rate,
		particles: make([]*particle, 0, num),
		num:       num,
	}
}

type particleEmitter struct {
	glctx gl.Context

	VBO       gl.Buffer
	origin    mgl.Vec3
	rate      float32
	particles []*particle
	num       int
}

func (emitter *particleEmitter) MoveTo(pos mgl.Vec3) {
	emitter.origin = pos
}

func (emitter *particleEmitter) Tick(interval time.Duration) {
	// Randomize emitting
	t := float32(interval.Seconds())
	n := int(emitter.rate * t * rand.Float32())

	extra := len(emitter.particles) + 1 + n - emitter.num
	if extra > 0 {
		// Pop oldest particles
		emitter.particles = emitter.particles[extra:]
	}

	for i := 0; i <= n; i++ {
		p := RandomParticle(emitter.origin, particleForce)
		emitter.particles = append(emitter.particles, p)
	}

	f := gravityForce.Mul(t)
	for _, particle := range emitter.particles {
		particle.Tick(f)
	}

	emitter.Buffer()
}

func (emitter *particleEmitter) Buffer() {
	data := emitter.Bytes()
	if len(data) > 0 {
		emitter.glctx.BindBuffer(gl.ARRAY_BUFFER, emitter.VBO)
		emitter.glctx.BufferSubData(gl.ARRAY_BUFFER, 0, data)
	}
}

func (emitter *particleEmitter) Len() int {
	return len(emitter.particles)
}

func (emitter *particleEmitter) Bytes() []byte {
	buf := bytes.Buffer{}
	for _, particle := range emitter.particles {
		binary.Write(&buf, binary.LittleEndian, particle.Vertices())
	}
	return buf.Bytes()
}

func (emitter *particleEmitter) Stride() int {
	return vecSize * vertexDim
}

func (emitter *particleEmitter) Draw(ctx DrawContext) {
	shader := ctx.Shader
	glctx := ctx.GL
	glctx.BindBuffer(gl.ARRAY_BUFFER, emitter.VBO)

	glctx.EnableVertexAttribArray(shader.Attrib("vertCoord"))
	glctx.VertexAttribPointer(shader.Attrib("vertCoord"), vertexDim, gl.FLOAT, false, emitter.Stride(), 0)

	glctx.DrawArrays(gl.TRIANGLES, 0, emitter.Len()*particleLen/vertexDim)

	glctx.DisableVertexAttribArray(shader.Attrib("vertCoord"))
}

func (emitter *particleEmitter) Close() error {
	emitter.glctx.DeleteBuffer(emitter.VBO)
	return nil
}
