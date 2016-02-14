package loader

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/gl"
)

// TODO: Need a ShaderRegistry of somekind, ideally with support for default
// scene values vs per-shape values and attribute checking.
// TODO: Should each NewShader be a struct embedding a Program?

type Shader interface {
	Use()
	Close() error
	Attrib(string) gl.Attrib
	Uniform(string) gl.Uniform
	Context() gl.Context
}

func NewShader(glctx gl.Context, vertAsset, fragAsset string) (Shader, error) {
	program, err := LoadProgram(glctx, vertAsset, fragAsset)
	if err != nil {
		return nil, err
	}

	return &shader{
		glctx:    glctx,
		program:  program,
		attribs:  map[string]gl.Attrib{},
		uniforms: map[string]gl.Uniform{},
	}, nil
}

type shader struct {
	glctx   gl.Context
	program gl.Program

	attribs  map[string]gl.Attrib
	uniforms map[string]gl.Uniform
}

func (shader *shader) Context() gl.Context {
	return shader.glctx
}

func (shader *shader) Attrib(name string) gl.Attrib {
	v, ok := shader.attribs[name]
	if !ok {
		v = shader.glctx.GetAttribLocation(shader.program, name)
		shader.attribs[name] = v
	}
	return v
}

func (shader *shader) Uniform(name string) gl.Uniform {
	v, ok := shader.uniforms[name]
	if !ok {
		v = shader.glctx.GetUniformLocation(shader.program, name)
		shader.uniforms[name] = v
	}
	return v
}

func (shader *shader) Use() {
	shader.glctx.UseProgram(shader.program)
}

func (shader *shader) Close() error {
	shader.glctx.DeleteProgram(shader.program)
	return nil
}

type Shaders interface {
	Load(...string) error
	Get(string) Shader
	Reload() error
	Close() error
}

func ShaderLoader(glctx gl.Context) *shaderLoader {
	return &shaderLoader{
		glctx:   glctx,
		shaders: map[string]*shader{},
	}
}

type shaderLoader struct {
	glctx   gl.Context
	shaders map[string]*shader
}

func (loader *shaderLoader) Load(names ...string) error {
	for _, name := range names {
		s, err := NewShader(
			loader.glctx,
			fmt.Sprintf("%s.v.glsl", name),
			fmt.Sprintf("%s.f.glsl", name),
		)
		if err != nil {
			return err
		}
		loader.shaders[name] = s.(*shader)
	}
	return nil
}

func (loader *shaderLoader) Get(name string) Shader {
	return loader.shaders[name]
}

func (loader *shaderLoader) Reload() error {
	for k, shader := range loader.shaders {
		err := LoadShaders(
			loader.glctx,
			shader.program,
			fmt.Sprintf("%s.v.glsl", k),
			fmt.Sprintf("%s.f.glsl", k),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (loader *shaderLoader) Close() error {
	for _, shader := range loader.shaders {
		shader.Close()
	}
	return nil
}

func loadAsset(name string) ([]byte, error) {
	f, err := asset.Open(name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func loadShader(glctx gl.Context, shaderType gl.Enum, assetName string) (gl.Shader, error) {
	// Borrowed from golang.org/x/mobile/exp/gl/glutil
	src, err := loadAsset(assetName)
	if err != nil {
		return gl.Shader{}, err
	}

	shader := glctx.CreateShader(shaderType)
	if shader.Value == 0 {
		return gl.Shader{}, fmt.Errorf("glutil: could not create shader (type %v)", shaderType)
	}
	glctx.ShaderSource(shader, string(src))
	glctx.CompileShader(shader)
	if glctx.GetShaderi(shader, gl.COMPILE_STATUS) == 0 {
		defer glctx.DeleteShader(shader)
		return gl.Shader{}, fmt.Errorf("shader compile: %s", glctx.GetShaderInfoLog(shader))
	}
	return shader, nil
}

func LoadShaders(glctx gl.Context, program gl.Program, vertexAsset, fragmentAsset string) error {
	vertexShader, err := loadShader(glctx, gl.VERTEX_SHADER, vertexAsset)
	if err != nil {
		return err
	}
	fragmentShader, err := loadShader(glctx, gl.FRAGMENT_SHADER, fragmentAsset)
	if err != nil {
		glctx.DeleteShader(vertexShader)
		return err
	}

	if glctx.GetProgrami(program, gl.ATTACHED_SHADERS) > 0 {
		for _, shader := range glctx.GetAttachedShaders(program) {
			glctx.DetachShader(program, shader)
		}
	}

	glctx.AttachShader(program, vertexShader)
	glctx.AttachShader(program, fragmentShader)
	glctx.LinkProgram(program)

	// Flag shaders for deletion when program is unlinked.
	glctx.DeleteShader(vertexShader)
	glctx.DeleteShader(fragmentShader)

	if glctx.GetProgrami(program, gl.LINK_STATUS) == 0 {
		defer glctx.DeleteProgram(program)
		return fmt.Errorf("LoadShaders: %s", glctx.GetProgramInfoLog(program))
	}
	return nil
}

// LoadProgram reads shader sources from the asset repository, compiles, and
// links them into a program.
func LoadProgram(glctx gl.Context, vertexAsset, fragmentAsset string) (program gl.Program, err error) {
	log.Println("LoadProgram:", vertexAsset, fragmentAsset)

	program = glctx.CreateProgram()
	if program.Value == 0 {
		return gl.Program{}, fmt.Errorf("glutil: no programs available")
	}

	err = LoadShaders(glctx, program, vertexAsset, fragmentAsset)
	return
}
