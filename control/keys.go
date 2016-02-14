package control

import key "golang.org/x/mobile/event/key"

type KeyBinding uint8

const (
	KeyUnknown KeyBinding = iota
	KeyCamForward
	KeyCamReverse
	KeyCamLeft
	KeyCamRight
	KeyCamUp
	KeyCamDown
	KeyLineLeft
	KeyLineRight
	KeyCameraFollow
	KeyPause
	KeyReload
	KeyDebug
)

func DefaultBindings() Bindings {
	b := map[key.Code]KeyBinding{
		key.CodeW:          KeyCamForward,
		key.CodeS:          KeyCamReverse,
		key.CodeA:          KeyCamLeft,
		key.CodeD:          KeyCamRight,
		key.CodeQ:          KeyCamUp,
		key.CodeE:          KeyCamDown,
		key.CodeRightArrow: KeyLineRight,
		key.CodeLeftArrow:  KeyLineLeft,
		key.CodeF:          KeyCameraFollow,
		key.CodeSpacebar:   KeyPause,
		key.CodeR:          KeyReload,
		key.CodeBackslash:  KeyDebug,
	}
	return &bindings{
		bindings: b,
		on:       map[KeyBinding]func(KeyBinding){},
		pressed:  map[KeyBinding]bool{},
	}
}

type Bindings interface {
	Lookup(code key.Code) KeyBinding
	On(k KeyBinding, fn func(KeyBinding))
	Press(code key.Code)
	Pressed(k KeyBinding) bool
	Release(code key.Code)
}

type bindings struct {
	bindings map[key.Code]KeyBinding
	on       map[KeyBinding]func(KeyBinding)
	pressed  map[KeyBinding]bool
}

func (b *bindings) Lookup(code key.Code) KeyBinding {
	k, ok := b.bindings[code]
	if !ok {
		return KeyUnknown
	}
	return k
}

func (b *bindings) On(k KeyBinding, fn func(KeyBinding)) {
	b.on[k] = fn
}

func (b *bindings) Press(code key.Code) {
	key := b.Lookup(code)
	b.pressed[key] = true

	fn, ok := b.on[key]
	if ok {
		fn(key)
	}
}

func (b *bindings) Release(code key.Code) {
	key := b.Lookup(code)
	b.pressed[key] = false
}

func (b *bindings) Pressed(k KeyBinding) bool {
	p, ok := b.pressed[k]
	if ok {
		return p
	}
	return false
}
