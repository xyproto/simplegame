package main

import (
	"fmt"
	"math"
	"runtime"
	"time"

	gl "github.com/chsc/gogl/gl33"
	"github.com/veandco/go-sdl2/sdl"
)

func createProgram() gl.Uint {
	vs := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	fs := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	program := linkProgram(vs, fs)
	return program
}

func compileShader(shaderSource string, shaderType gl.Enum) gl.Uint {
	shader := gl.CreateShader(shaderType)
	source := gl.GLString(shaderSource)
	defer gl.GLStringFree(source)

	gl.ShaderSource(shader, 1, &source, nil)
	gl.CompileShader(shader)

	var status gl.Int
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		panic(fmt.Sprintf("Failed to compile shader of type %d", shaderType))
	}

	return shader
}

func linkProgram(vs, fs gl.Uint) gl.Uint {
	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)

	fragOutString := gl.GLString("outColor")
	defer gl.GLStringFree(fragOutString)
	gl.BindFragDataLocation(program, gl.Uint(0), fragOutString)

	gl.LinkProgram(program)

	var linkStatus gl.Int
	gl.GetProgramiv(program, gl.LINK_STATUS, &linkStatus)
	if linkStatus == gl.FALSE {
		panic("Failed to link program")
	}

	return program
}

func initSDL() (*sdl.Window, sdl.GLContext) {
	var window *sdl.Window
	var context sdl.GLContext
	var err error

	runtime.LockOSThread()

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}

	context, err = window.GLCreateContext()
	if err != nil {
		panic(err)
	}

	return window, context
}

func initGL() {
	gl.Init()
	gl.Viewport(0, 0, gl.Sizei(winWidth), gl.Sizei(winHeight))

	gl.ClearColor(0.0, 0.1, 0.0, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func createBuffers(vertices []gl.Float, colors []gl.Float) (gl.Uint, gl.Uint) {
	vertexBuffer := createBuffer(gl.ARRAY_BUFFER, vertices)
	colorBuffer := createBuffer(gl.ARRAY_BUFFER, colors)

	return vertexBuffer, colorBuffer
}

func createBuffer(target gl.Enum, data []gl.Float) gl.Uint {
	var buffer gl.Uint
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(target, buffer)
	gl.BufferData(target, gl.Sizeiptr(len(data)*4), gl.Pointer(&data[0]), gl.STATIC_DRAW)

	return buffer
}

func handleEvents(running *bool) {
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			*running = false
		case *sdl.MouseMotionEvent:
			xrot = float32(t.Y)

			yrot = float32(t.X)
		}
	}
}

func drawScene(program gl.Uint, vertexBuffer gl.Uint, colorBuffer gl.Uint) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, gl.FALSE, 0, gl.Pointer(nil))
	gl.EnableVertexAttribArray(0)

	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, gl.FALSE, 0, gl.Pointer(nil))
	gl.EnableVertexAttribArray(1)

	uniformTime := gl.GetUniformLocation(program, gl.GLString("time"))
	gl.Uniform1f(uniformTime, gl.Float(math.Mod(time.Now().Sub(startTime).Seconds(), 2*math.Pi)))

	gl.DrawArrays(gl.TRIANGLES, 0, 3)

	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
}


func main() {
	window, context := initSDL()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	initGL()

	program := createProgram()
	vertexBuffer, colorBuffer := createBuffers(vertices, colors)

	running := true
	for running {
		handleEvents(&running)
		drawScene(program, vertexBuffer, colorBuffer)
		window.GLSwap()
	}
}
