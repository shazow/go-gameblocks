package gameblocks

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

// StartMobile launches the Engine using the x/mobile backend, should be run
// from the main thread.
func StartMobile(engine Engine) {
	app.Main(func(a app.App) {
		var glctx gl.Context
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					if err := engine.Start(glctx); err != nil {
						panic(err)
					}
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					engine.Stop()
					glctx = nil
				}
			case paint.Event:
				if glctx == nil || e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}
				engine.Draw()
				a.Publish()
				// Drive the animation by preparing to paint the next frame
				// after this one is shown.
				a.Send(paint.Event{})
			case size.Event:
				engine.Resize(e)
			case touch.Event:
				engine.Touch(e)
			case key.Event:
				engine.Press(e)
			}
		}
	})
}
