package camera

import (
	"fmt"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
)

const halfPi = math.Pi / 2.0

var (
	AxisFront = mgl.Vec3{0, 0, 1}
	AxisUp    = mgl.Vec3{0, 1, 0}
	AxisRight = mgl.Vec3{1, 0, 0}
)

type Camera interface {
	View() mgl.Mat4
	Projection() mgl.Mat4
	Position() mgl.Vec3
}

type FixedCamera struct {
	view       mgl.Mat4
	projection mgl.Mat4
	position   mgl.Vec3
}

func (c FixedCamera) Projection() mgl.Mat4 { return c.projection }
func (c FixedCamera) View() mgl.Mat4       { return c.view }
func (c FixedCamera) Position() mgl.Vec3   { return c.position }

type EulerCamera struct {
	projection mgl.Mat4
	eye        mgl.Vec3

	yaw   float64
	pitch float64

	center mgl.Vec3
	up     mgl.Vec3
	right  mgl.Vec3
}

func (c *EulerCamera) Position() mgl.Vec3 {
	return c.eye
}

// Perspective computes the projection matrix and saves it
func (c *EulerCamera) SetPerspective(fovy, aspect, near, far float32) {
	c.projection = mgl.Perspective(fovy, aspect, near, far)
}

func (c *EulerCamera) updateVectors() {
	// TODO: Read up on http://learnopengl.com/#!Getting-started/Camera
	// Borrowed from:
	// - https://github.com/JoeyDeVries/LearnOpenGL/blob/master/includes/learnopengl/camera.h
	// - https://github.com/mmchugh/planetary/blob/master/src/helpers/camera.cpp
	c.center = mgl.Vec3{
		float32(math.Cos(c.pitch) * math.Cos(c.yaw)),
		float32(math.Sin(c.pitch)),
		float32(math.Cos(c.pitch) * math.Sin(c.yaw)),
	}.Normalize()

	// Reset the up vector
	c.right = c.center.Cross(AxisUp).Normalize()
	c.up = c.right.Cross(c.center)
}

// Rotate adjusts the direction vectors by a delta vector of {pitch, yaw, roll}.
// Roll is ignored for now.
func (c *EulerCamera) Rotate(delta mgl.Vec3) {
	c.yaw += float64(delta.Y())
	c.pitch += float64(delta.X())

	// Limit vertical rotation to avoid gimbal lock
	if c.pitch > halfPi {
		c.pitch = halfPi
	} else if c.pitch < -halfPi {
		c.pitch = -halfPi
	}

	c.updateVectors()
}

// RotateTo adjusts the yaw and pitch to face a point.
func (c *EulerCamera) RotateTo(center mgl.Vec3) {
	// TODO: https://math.stackexchange.com/questions/470112/calculate-camera-pitch-yaw-to-face-point
}

// Move adjusts the position of the camera by a delta vector relative to the camera is facing.
func (c *EulerCamera) Move(delta mgl.Vec3) {
	c.eye = c.eye.Add(c.right.Mul(delta[0])).Add(c.up.Mul(delta[1])).Add(c.center.Mul(delta[2]))
}

// View returns the transform matrix from world space into camera space
func (c *EulerCamera) View() mgl.Mat4 {
	return mgl.LookAtV(c.eye, c.eye.Add(c.center), c.up)
}

// Projection returns the projection matrix for the camera perspective
func (c *EulerCamera) Projection() mgl.Mat4 {
	return c.projection
}

// String returns a string representation of the camera for debugging.
func (c *EulerCamera) String() string {
	return fmt.Sprintf(`Camera:
	eye:    %v
	center: %v
	up:     %v
	pitch, yaw: %v, %v`+"\n", c.eye, c.center, c.up, c.pitch, c.yaw)
}

func NewQuatCamera() *QuatCamera {
	return &QuatCamera{
		rotation: mgl.QuatIdent(),
	}
}

// QuatCamera is a Camera implementation using quaternion for rotation.
type QuatCamera struct {
	projection mgl.Mat4
	position   mgl.Vec3
	rotation   mgl.Quat
}

func (c *QuatCamera) Position() mgl.Vec3 {
	return c.position
}

// Perspective computes the projection matrix and saves it
func (c *QuatCamera) SetPerspective(fovy, aspect, near, far float32) {
	c.projection = mgl.Perspective(fovy, aspect, near, far)
}

func (c *QuatCamera) Rotate(delta mgl.Vec3) {
	if delta[0] != 0 {
		// Pitch (about the X axis)
		q := mgl.QuatRotate(delta[0], AxisRight).Normalize()
		c.rotation = c.rotation.Mul(q).Normalize()
	}
	if delta[1] != 0 {
		// Yaw (about the Y axis)
		q := mgl.QuatRotate(delta[1], AxisUp).Normalize()
		c.rotation = q.Mul(c.rotation).Normalize()
	}
	// TODO: Roll
}

// RotateTo adjusts the yaw and pitch to face a point.
func (c *QuatCamera) RotateTo(center mgl.Vec3) {
	direction := center.Sub(c.position).Normalize()
	right := direction.Cross(AxisUp)
	up := right.Cross(direction)

	c.rotation = mgl.QuatLookAtV(c.position, center, up)
}

// Move adjusts the position of the camera by a delta vector relative to the camera is facing.
func (c *QuatCamera) Move(delta mgl.Vec3) {
	c.position = c.position.Add(c.rotation.Rotate(delta))
}

// MoveTo adjusts the absolute position of the camera
func (c *QuatCamera) MoveTo(position mgl.Vec3) {
	c.position = position
}

// Lerp will interpolate between the desired position/center by amount.
func (c *QuatCamera) Lerp(position mgl.Vec3, center mgl.Vec3, amount float32) {
	direction := center.Sub(position).Normalize()
	right := direction.Cross(AxisUp)
	up := right.Cross(direction)

	targetRot := mgl.QuatLookAtV(position, center, up)
	c.rotation = mgl.QuatNlerp(c.rotation, targetRot, amount)
	c.position = c.position.Add(position.Sub(c.position).Mul(amount))
}

// View returns the transform matrix from world space into camera space
func (c *QuatCamera) View() mgl.Mat4 {
	// FIXME: Is there a way to get this matrix from the quat+position directly?
	return mgl.LookAtV(c.position, c.position.Add(c.Center()), c.Up())
}

// Projection returns the projection matrix for the camera perspective
func (c *QuatCamera) Projection() mgl.Mat4 {
	return c.projection
}

// Center returns the direction vector of the camera
func (c *QuatCamera) Up() mgl.Vec3 {
	return c.rotation.Rotate(AxisUp)
}

// Center returns the direction vector of the camera
func (c *QuatCamera) Center() mgl.Vec3 {
	axisBack := mgl.Vec3{}.Sub(AxisFront)
	return c.rotation.Rotate(axisBack)
}

// String returns a string representation of the camera for debugging.
func (c *QuatCamera) String() string {
	return fmt.Sprintf(`Camera:
	position: %v
	rotation: %v
	center:   %v
	up:       %v`+"\n", c.position, c.rotation, c.Center(), c.Up())
}
