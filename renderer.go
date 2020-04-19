package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func CreateWindow(w, h int, title string) (*glfw.Window, error) {
	// Initialize glfw
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// Create window
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		return nil, err
	}

	// Create context
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Configure global opengl state
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	return window, nil
}

type shader struct {
	fragmentFilename string
	fragmentSource   string
	lastModified     time.Time
	vertexData       []float32
	vertexShader     uint32
	Program          uint32
	Vao              uint32
	Vbo              uint32
}

func CreateShader(fragmentFilename string) (*shader, error) {
	s := &shader{
		fragmentFilename: fragmentFilename,
		vertexData: []float32{
			-1.0, -1.0, 0.5, 0.0, 0.0,
			-1.0, 1.0, 0.5, 0.0, 1.0,
			1.0, -1.0, 0.5, 1.0, 0.0,
			1.0, 1.0, 0.5, 1.0, 1.0,
		},
	}

	gl.GenVertexArrays(1, &s.Vao)
	gl.GenBuffers(1, &s.Vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, s.Vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(s.vertexData)*4, gl.Ptr(s.vertexData), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(s.Vao)

	defaultVertexSource := `
#version 410 core
in vec3 in_pos;
void main()
{
	gl_Position = vec4( in_pos.x, in_pos.y, in_pos.z, 1.0 );
}
`
	vertexShader, err := compileShader(&defaultVertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return &shader{}, err
	}
	s.vertexShader = vertexShader

	err = s.reloadFragmentShader()
	if err != nil {
		return &shader{}, err
	}

	return s, nil
}

func compileShader(source *string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(*source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func (s *shader) loadFragmentShader() error {
	b, err := ioutil.ReadFile(s.fragmentFilename)
	if err != nil {
		return err
	}
	s.fragmentSource = string(b) + "\x00"

	return nil
}

func (s *shader) reloadFragmentShader() error {
	err := s.loadFragmentShader()
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(&s.fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, s.vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return fmt.Errorf("failed to link program: %v", log)
	}

	s.Program = program

	gl.DeleteShader(fragmentShader)

	return nil
}

func (s *shader) ReloadIfModified() error {
	f, err := os.Open(s.fragmentFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if !fi.ModTime().Equal(s.lastModified) {
		s.lastModified = fi.ModTime()
		err = s.reloadFragmentShader()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *shader) SetfUniform(name string, x float32) {
	location := gl.GetUniformLocation(s.Program, gl.Str(name+"\x00"))

	if location != -1 {
		gl.ProgramUniform1f(s.Program, location, x)
	}
}

func (s *shader) Set2fUniform(name string, x float32, y float32) {
	location := gl.GetUniformLocation(s.Program, gl.Str(name+"\x00"))

	if location != -1 {
		gl.ProgramUniform2f(s.Program, location, x, y)
	}
}
