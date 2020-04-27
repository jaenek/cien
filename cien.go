package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	windowWidth    = 800
	windowHeight   = 600
	frameTimeDelta = 1.0 / 60
	fileCheckDelta = 1.0 / 60
)

func main() {
	var filepath string

	flag.StringVar(&filepath, "p", "shader.frag", "Path to fragment shader file.")
	flag.Parse()

	runtime.LockOSThread()

	window, err := CreateWindow(windowWidth, windowHeight, "cien"+filepath)
	if err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	s, err := CreateShader(filepath)
	if err != nil {
		log.Fatalln(err)
	}

	previousFrameTime := glfw.GetTime()
	previousFileCheck := glfw.GetTime()
	for !window.ShouldClose() {
		time := glfw.GetTime()

		if (time - previousFrameTime) >= frameTimeDelta {
			previousFrameTime = time

			s.SetfUniform("fGlobalTime", float32(time))
			s.Set2fUniform("v2Resolution", windowWidth, windowHeight)

			// Render
			gl.ClearColor(0.49, 0.83, 0.91, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

			gl.BindVertexArray(s.Vao)

			gl.UseProgram(s.Program)

			gl.BindBuffer(gl.ARRAY_BUFFER, s.Vbo)

			position := gl.GetAttribLocation(s.Program, gl.Str("in_pos\x00"))
			if position != -1 {
				gl.VertexAttribPointer(uint32(position), 3, gl.FLOAT, false, 4*5, gl.PtrOffset(0))
				gl.EnableVertexAttribArray(uint32(position))
			}

			gl.BindBuffer(gl.ARRAY_BUFFER, s.Vbo)
			gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

			gl.UseProgram(0)

			// Maintenance
			window.SwapBuffers()
			glfw.PollEvents()
		}

		if (time - previousFileCheck) >= fileCheckDelta {
			err = s.ReloadIfModified()
			if err != nil {
				log.Print(err)
			}
		}
	}
}
