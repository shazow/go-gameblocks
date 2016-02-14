package loader

import (
	"image"
	image_draw "image/draw"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/gl"
)

type Textures interface {
	Load(...string) error
	Get2D(string) gl.Texture
	GetCube(string) gl.Texture
	Close() error
}

func TextureLoader(glctx gl.Context) Textures {
	return &textureLoader{
		glctx:  glctx,
		images: map[string]*image.RGBA{},
	}
}

type textureLoader struct {
	glctx  gl.Context
	images map[string]*image.RGBA
}

func (loader *textureLoader) Close() error {
	// TODO: ...
	return nil
}

func (loader *textureLoader) loadAsset(name string) (*image.RGBA, error) {
	imgFile, err := asset.Open(name)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	image_draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, image_draw.Src)
	return rgba, nil
}

func (loader *textureLoader) Load(names ...string) error {
	for _, name := range names {
		img, err := loader.loadAsset(name)
		if err != nil {
			return err
		}
		loader.images[name] = img
	}
	return nil
}

func (loader *textureLoader) Get2D(name string) gl.Texture {
	glctx := loader.glctx
	tex := glctx.CreateTexture()
	img := loader.images[name]

	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_2D, tex)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	glctx.TexImage2D(
		gl.TEXTURE_2D,
		0,
		img.Rect.Size().X,
		img.Rect.Size().Y,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		img.Pix)

	return tex
}

func (loader *textureLoader) GetCube(name string) gl.Texture {
	glctx := loader.glctx
	tex := glctx.CreateTexture()
	img := loader.images[name]

	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_CUBE_MAP, tex)

	target := gl.TEXTURE_CUBE_MAP_POSITIVE_X
	for i := 0; i < 6; i++ {
		// TODO: Load atlas, not the same image
		glctx.TexImage2D(
			gl.Enum(target+i),
			0,
			img.Rect.Size().X,
			img.Rect.Size().Y,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			img.Pix,
		)
	}

	glctx.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	glctx.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	// Not available in GLES 2.0 :(
	//gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	return tex
}
