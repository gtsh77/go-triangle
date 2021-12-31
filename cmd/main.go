package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime"
	"unsafe"

	gl "opengl-triangle-go/pkg/gl/v3.3-core/gl"

	"github.com/veandco/go-sdl2/sdl"
	"go.uber.org/zap"
)

var (
	SCR_WIDTH    int32 = 800
	SCR_HEIGHT   int32 = 600
	vertex, frag *uint8
	l, _         = zap.NewProduction()
	log          = l.Sugar()
	err          error
	window       *sdl.Window
	VBO, VS, FS  uint32
	P, VAO       uint32
	ctx          sdl.GLContext
	ok           int32
	vertices     []float32 = []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.0, 0.5, 0.0,
	}
	vShader = `
		#version 330 core
		layout (location = 0) in vec3 aPos;

		void main()
		{
			gl_Position = vec4(aPos.x, aPos.y, aPos.z, 1.0);
		}
		`
	fShader = `
		#version 330 core
		out vec4 FragColor;

		void main()
		{
			FragColor = vec4(1.0, 0.5, 0.2, 1.0);
		}
		`
)

func main() {
	//lock thread
	runtime.LockOSThread()

	//init SDL
	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		return
	}

	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE); err != nil {
		return
	}
	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3); err != nil {
		return
	}
	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2); err != nil {
		return
	}
	if err = sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8); err != nil {
		return
	}

	//create Window
	if window, err = sdl.CreateWindow("LearnOpenGL", 0, 0, SCR_WIDTH, SCR_HEIGHT, sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE); err != nil {
		log.With("create-window", err).Fatal()
		return
	}

	//create GL Context
	if ctx, err = window.GLCreateContext(); err != nil {
		log.With("create-context", err).Fatal()
		return
	}

	//init GL
	if err = gl.Init(); err != nil {
		log.With("gl-init", err).Fatal()
		return
	}

	//alloc buffer for VAO
	gl.GenVertexArrays(1, &VAO)
	//alloc vertices buffer on GPU
	gl.GenBuffers(1, &VBO)
	//bind VAO (to auto-store VBO operations)
	gl.BindVertexArray(VAO)
	//bind to ARRAY_BUFFER
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	//transfer vertices to GPU
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, unsafe.Pointer(&vertices[0]), gl.STATIC_DRAW)
	//unbind VAO (to stop auto-store VBO operations)
	gl.BindVertexArray(0)

	//alloc buffer for v-shader
	VS = gl.CreateShader(gl.VERTEX_SHADER)
	//set source code pointer
	b, _ := gl.Strs(vShader + "\x00")
	gl.ShaderSource(VS, 1, b, nil)
	//compile shader
	gl.CompileShader(VS)
	//check if compiled - OK
	if gl.GetShaderiv(VS, gl.COMPILE_STATUS, &ok); ok == 0 {
		log.With("vs-compile", "can't compile").Fatal()
	}

	//alloc buffer for f-shader
	FS = gl.CreateShader(gl.VERTEX_SHADER)
	//set source code pointer
	a, _ := gl.Strs(fShader + "\x00")
	gl.ShaderSource(FS, 1, a, nil)
	//compile shader
	gl.CompileShader(FS)
	//check if compiled - OK
	if gl.GetShaderiv(FS, gl.COMPILE_STATUS, &ok); ok == 0 {
		log.With("fs-compile", "can't compile").Fatal()
	}

	//v-hint (stride float32 4 bytes *3 elements in v-array)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	//enable
	gl.EnableVertexAttribArray(0)

	out := gl.Str("outFragmentColor\x00")

	//create GPU programm
	P = gl.CreateProgram()
	//attach compiled shaders
	gl.AttachShader(P, VS)
	gl.AttachShader(P, FS)
	gl.BindFragDataLocation(P, 0, out)
	//link programm
	gl.LinkProgram(P)
	//check if linked - OK
	var plog [100]uint8
	var wbuf *bytes.Buffer = new(bytes.Buffer)
	if gl.GetProgramiv(P, gl.LINK_STATUS, &ok); ok == 0 {
		gl.GetProgramInfoLog(P, 100, nil, &plog[0])
		binary.Write(wbuf, binary.LittleEndian, plog)
		log.With("link-programm", fmt.Errorf("%s\n", wbuf)).Fatal()
	}
	//cleanup
	gl.DeleteShader(VS)
	gl.DeleteShader(FS)

	//use programm
	gl.UseProgram(P)
	//activate VAO-1
	gl.BindVertexArray(VAO)
	//draw to frame buffer
	gl.DrawArrays(gl.TRIANGLES, 0, 3)

}

// package main

// import (
// 	"fmt"
// 	"math"
// 	"runtime"
// 	"time"
// 	"unsafe"

// 	gl "opengl-triangle-go/pkg/gl/v3.2-core/gl"

// 	"github.com/veandco/go-sdl2/sdl"
// )

// func createprogram() uint32 {
// 	// VERTEX SHADER
// 	vs := gl.CreateShader(gl.VERTEX_SHADER)
// 	vs_source, _ := gl.Strs(vertexShaderSource + "\x00")
// 	gl.ShaderSource(vs, 1, vs_source, nil)
// 	gl.CompileShader(vs)
// 	var vs_status int32
// 	gl.GetShaderiv(vs, gl.COMPILE_STATUS, &vs_status)
// 	fmt.Printf("Compiled Vertex Shader: %v\n", vs_status)

// 	// FRAGMENT SHADER
// 	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
// 	fs_source, _ := gl.Strs(fragmentShaderSource + "\x00")
// 	gl.ShaderSource(fs, 1, fs_source, nil)
// 	gl.CompileShader(fs)
// 	var fstatus int32
// 	gl.GetShaderiv(fs, gl.COMPILE_STATUS, &fstatus)
// 	fmt.Printf("Compiled Fragment Shader: %v\n", fstatus)

// 	// CREATE PROGRAM
// 	program := gl.CreateProgram()
// 	gl.AttachShader(program, vs)
// 	gl.AttachShader(program, fs)
// 	fragoutstring := gl.Str("outColor" + "\x00")
// 	// defer gl.GLStringFree(fragoutstring)
// 	gl.BindFragDataLocation(program, uint32(0), fragoutstring)

// 	gl.LinkProgram(program)
// 	var linkstatus int32
// 	gl.GetProgramiv(program, gl.LINK_STATUS, &linkstatus)
// 	fmt.Printf("Program Link: %v\n", linkstatus)

// 	return program
// }

// var uniRoll float32 = 0.0
// var uniYaw float32 = 1.0
// var uniPitch float32 = 0.0
// var uniscale float32 = 0.3
// var yrot float32 = 20.0
// var zrot float32 = 0.0
// var xrot float32 = 0.0
// var UniScale int32

// func main() {
// 	var window *sdl.Window
// 	var context sdl.GLContext
// 	var event sdl.Event
// 	var running bool
// 	var err error
// 	runtime.LockOSThread()
// 	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
// 		panic(err)
// 	}
// 	defer sdl.Quit()
// 	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED,
// 		sdl.WINDOWPOS_UNDEFINED,
// 		winWidth, winHeight, sdl.WINDOW_OPENGL)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer window.Destroy()
// 	context, err = window.GLCreateContext()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer sdl.GLDeleteContext(context)

// 	gl.Init()
// 	gl.Viewport(0, 0, winWidth, winHeight)
// 	// OPENGL FLAGS
// 	gl.ClearColor(0.0, 0.1, 0.0, 1.0)
// 	gl.Enable(gl.DEPTH_TEST)
// 	gl.DepthFunc(gl.LESS)
// 	gl.Enable(gl.BLEND)
// 	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

// 	// VERTEX BUFFER
// 	var vertexbuffer uint32
// 	gl.GenBuffers(1, &vertexbuffer)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, vertexbuffer)
// 	gl.BufferData(gl.ARRAY_BUFFER, len(triangle_vertices)*4, unsafe.Pointer(&triangle_vertices[0]), gl.STATIC_DRAW)

// 	// COLOUR BUFFER
// 	var colourbuffer uint32
// 	gl.GenBuffers(1, &colourbuffer)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
// 	gl.BufferData(gl.ARRAY_BUFFER, len(triangle_colours)*4, unsafe.Pointer(&triangle_colours[0]), gl.STATIC_DRAW)

// 	// GUESS WHAT
// 	program := createprogram()

// 	// VERTEX ARRAY
// 	var VertexArrayID uint32
// 	gl.GenVertexArrays(1, &VertexArrayID)
// 	gl.BindVertexArray(VertexArrayID)
// 	gl.EnableVertexAttribArray(0)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, vertexbuffer)
// 	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

// 	// VERTEX ARRAY HOOK COLOURS
// 	gl.EnableVertexAttribArray(1)
// 	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
// 	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)

// 	//UNIFORM HOOK
// 	unistring := gl.Str("scaleMove" + "\x00")
// 	UniScale = gl.GetUniformLocation(program, unistring)
// 	fmt.Printf("Uniform Link: %v\n", UniScale+1)

// 	gl.UseProgram(program)

// 	running = true
// 	for running {
// 		for event = sdl.PollEvent(); event != nil; event =
// 			sdl.PollEvent() {
// 			switch t := event.(type) {
// 			case *sdl.QuitEvent:
// 				running = false
// 			case *sdl.MouseMotionEvent:

// 				xrot = float32(t.Y) / 2
// 				yrot = float32(t.X) / 2
// 				fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
// 			}
// 		}
// 		drawgl()
// 		window.GLSwap()
// 	}
// }

// func drawgl() {
// 	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

// 	uniYaw = yrot * (math.Pi / 180.0)
// 	yrot = yrot - 1.0
// 	uniPitch = zrot * (math.Pi / 180.0)
// 	zrot = zrot - 0.5
// 	uniRoll = xrot * (math.Pi / 180.0)
// 	xrot = xrot - 0.2

// 	gl.Uniform4f(UniScale, uniRoll, uniYaw, uniPitch, uniscale)
// 	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
// 	gl.DrawArrays(gl.TRIANGLES, int32(0), int32(len(triangle_vertices)*4))

// 	time.Sleep(50 * time.Millisecond)

// }

// const (
// 	winTitle           = "OpenGL Shader"
// 	winWidth           = 640
// 	winHeight          = 480
// 	vertexShaderSource = `
// #version 330
// layout (location = 0) in vec3 Position;
// layout(location = 1) in vec3 vertexColor;
// uniform vec4 scaleMove;
// out vec3 fragmentColor;
// void main()
// {
// // YOU CAN OPTIMISE OUT cos(scaleMove.x) AND sin(scaleMove.y) AND UNIFORM THE VALUES IN
//     vec3 scale = Position.xyz * scaleMove.w;
// // rotate on z pole
//    vec3 rotatez = vec3((scale.x * cos(scaleMove.x) - scale.y * sin(scaleMove.x)), (scale.x * sin(scaleMove.x) + scale.y * cos(scaleMove.x)), scale.z);
// // rotate on y pole
//     vec3 rotatey = vec3((rotatez.x * cos(scaleMove.y) - rotatez.z * sin(scaleMove.y)), rotatez.y, (rotatez.x * sin(scaleMove.y) + rotatez.z * cos(scaleMove.y)));
// // rotate on x pole
//     vec3 rotatex = vec3(rotatey.x, (rotatey.y * cos(scaleMove.z) - rotatey.z * sin(scaleMove.z)), (rotatey.y * sin(scaleMove.z) + rotatey.z * cos(scaleMove.z)));
// // move
// vec3 move = vec3(rotatex.xy, rotatex.z - 0.2);
// // terrible perspective transform
// vec3 persp = vec3( move.x  / ( (move.z + 2) / 3 ) ,
// 		   move.y  / ( (move.z + 2) / 3 ) ,
// 		     move.z);

//     gl_Position = vec4(persp, 1.0);
//     fragmentColor = vertexColor;
// }
// `
// 	fragmentShaderSource = `
// #version 330
// out vec4 outColor;
// in vec3 fragmentColor;
// void main()
// {
// 	outColor = vec4(fragmentColor, 1.0);
// }
// `
// )

// var triangle_vertices = []float32{
// 	-1.0, -1.0, -1.0,
// 	-1.0, -1.0, 1.0,
// 	-1.0, 1.0, 1.0,
// 	1.0, 1.0, -1.0,
// 	-1.0, -1.0, -1.0,
// 	-1.0, 1.0, -1.0,
// 	1.0, -1.0, 1.0,
// 	-1.0, -1.0, -1.0,
// 	1.0, -1.0, -1.0,
// 	1.0, 1.0, -1.0,
// 	1.0, -1.0, -1.0,
// 	-1.0, -1.0, -1.0,
// 	-1.0, -1.0, -1.0,
// 	-1.0, 1.0, 1.0,
// 	-1.0, 1.0, -1.0,
// 	1.0, -1.0, 1.0,
// 	-1.0, -1.0, 1.0,
// 	-1.0, -1.0, -1.0,
// 	-1.0, 1.0, 1.0,
// 	-1.0, -1.0, 1.0,
// 	1.0, -1.0, 1.0,
// 	1.0, 1.0, 1.0,
// 	1.0, -1.0, -1.0,
// 	1.0, 1.0, -1.0,
// 	1.0, -1.0, -1.0,
// 	1.0, 1.0, 1.0,
// 	1.0, -1.0, 1.0,
// 	1.0, 1.0, 1.0,
// 	1.0, 1.0, -1.0,
// 	-1.0, 1.0, -1.0,
// 	1.0, 1.0, 1.0,
// 	-1.0, 1.0, -1.0,
// 	-1.0, 1.0, 1.0,
// 	1.0, 1.0, 1.0,
// 	-1.0, 1.0, 1.0,
// 	1.0, -1.0, 1.0}

// var triangle_colours = []float32{
// 	0.583, 0.771, 0.014,
// 	0.609, 0.115, 0.436,
// 	0.327, 0.483, 0.844,
// 	0.822, 0.569, 0.201,
// 	0.435, 0.602, 0.223,
// 	0.310, 0.747, 0.185,
// 	0.597, 0.770, 0.761,
// 	0.559, 0.436, 0.730,
// 	0.359, 0.583, 0.152,
// 	0.483, 0.596, 0.789,
// 	0.559, 0.861, 0.639,
// 	0.195, 0.548, 0.859,
// 	0.014, 0.184, 0.576,
// 	0.771, 0.328, 0.970,
// 	0.406, 0.615, 0.116,
// 	0.676, 0.977, 0.133,
// 	0.971, 0.572, 0.833,
// 	0.140, 0.616, 0.489,
// 	0.997, 0.513, 0.064,
// 	0.945, 0.719, 0.592,
// 	0.543, 0.021, 0.978,
// 	0.279, 0.317, 0.505,
// 	0.167, 0.620, 0.077,
// 	0.347, 0.857, 0.137,
// 	0.055, 0.953, 0.042,
// 	0.714, 0.505, 0.345,
// 	0.783, 0.290, 0.734,
// 	0.722, 0.645, 0.174,
// 	0.302, 0.455, 0.848,
// 	0.225, 0.587, 0.040,
// 	0.517, 0.713, 0.338,
// 	0.053, 0.959, 0.120,
// 	0.393, 0.621, 0.362,
// 	0.673, 0.211, 0.457,
// 	0.820, 0.883, 0.371,
// 	0.982, 0.099, 0.879}
