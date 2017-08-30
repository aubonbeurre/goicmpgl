package main

// speed-up compilation:
// go install -a github.com/go-gl/gl/v4.1-core/gl
// go install -a github.com/go-gl/glfw3/v3.1/glfw

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/aubonbeurre/glplus"
	gl "github.com/go-gl/gl/v4.1-core/gl"
	glfw "github.com/go-gl/glfw3/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jessevdk/go-flags"
)

var (
	// vertex shader
	vertShader = `#version 330
  in vec4 position;
  in vec2 uvs;
  out vec4 out_pos;
  out vec2 out_uvs;
  uniform mat3 ModelviewMatrix;
  void main()
  {
		gl_Position = vec4(ModelviewMatrix * vec3(position.xy, 1.0), 0.0).xywz;
  	out_uvs = uvs;
  }`

	// fragment shader
	fragShaderTex0 = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
  uniform sampler2D tex1;
  uniform float blend;
  void main()
  {
    vec4 col0 = texture(tex1, out_uvs);
    //col0 = vec4(col0.r * col0.a, col0.g * col0.a, col0.b * col0.a, col0.a);
    col0 = col0 * blend;
    colourOut = col0;
  }`

	sFragmentShaderGrid = `#version 330
  uniform vec4 color1;
  uniform vec4 color2;
  uniform vec3 grid;
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;

  void main(void)
  {
  	// grid
  	float blocksize = grid.z;
  	int x = int(out_uvs.x * grid.x / blocksize);
  	int y = int(out_uvs.y * grid.y / blocksize);
  	int evenodd = (x + y) & 1;
  	colourOut = color1 * evenodd + (1 - evenodd) * color2;
  }`

	gOffX        float32 // texture coordinates
	gOffY        float32
	gMouseX      float32 // framebuffer coordinates
	gMouseY      float32
	gMouseDown           = false
	gHelp                = true
	gZoom        float32 = 1
	gBlend       float32 = 1
	gRetinaScale float32 = 1

	gImage1 image.Image
	gImage2 image.Image

	gDiffFlag = false

	sProgram2Src = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
uniform sampler2D decalA;
uniform sampler2D decalB;
uniform float diffBlend;
void main()
{
	vec4 srcColorA = texture(decalA, out_uvs);
	vec4 srcColorB = texture(decalB, out_uvs);
	vec4 diff = abs(srcColorA - srcColorB);
	if(diff != vec4(0.0,0.0,0.0,0.0)) diff = vec4(1.0,0.0,0.0,1.0);
	diff = srcColorA * 0.1 + diff * 0.9;
	colourOut = srcColorA * (1.0 - diffBlend) + diff * diffBlend;
}
`

	sProgram3Src = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
uniform sampler2D decalA;
uniform sampler2D decalB;
uniform float diffBlend;
void main()
{
	vec4 srcColorA = texture(decalA, out_uvs);
	vec4 srcColorB = texture(decalB, out_uvs);
	vec4 diff = abs(srcColorA - srcColorB);
	if(diff != vec4(0.0,0.0,0.0,0.0)) diff = vec4(1.0,0.0,0.0,1.0);
	diff = srcColorB * 0.1 + diff * 0.9;
	colourOut = diff * (1.0 - diffBlend) + srcColorB * diffBlend;
}
`

	sProgram4Src = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
uniform sampler2D decalA;
uniform sampler2D decalB;
uniform float diffBlend;
void main()
{
	vec4 srcColorA = texture(decalA, out_uvs);
	vec4 srcColorB = texture(decalB, out_uvs);
	vec4 diff = abs(srcColorA - srcColorB);
	vec4 diff1 = diff;
	if(diff1 != vec4(0.0,0.0,0.0,0.0)) diff1 = vec4(1.0,0.0,0.0,1.0);
	diff1 = srcColorA * 0.1 + diff1 * 0.9;
	float d = distance(srcColorA, srcColorB);
	vec4 diff2 = vec4(1.0,1.0-d,1.0-d,1.0);
	colourOut = diff1 * (1.0 - diffBlend) + diff2 * diffBlend;
}
`

	sProgram5Src = `#version 330
  in vec4 out_pos;
  in vec2 out_uvs;
  out vec4 colourOut;
uniform sampler2D decalA;
uniform sampler2D decalB;
uniform float diffBlend;
void main()
{
	vec4 srcColorA = texture(decalA, out_uvs);
	vec4 srcColorB = texture(decalB, out_uvs);
	vec4 diff = abs(srcColorA - srcColorB);
	vec4 diff1 = diff;
	if(diff1 != vec4(0.0,0.0,0.0,0.0)) diff1 = vec4(1.0,0.0,0.0,1.0);
	diff1 = srcColorB * 0.1 + diff1 * 0.9;
	float d = distance(srcColorA, srcColorB);
	vec4 diff2 = vec4(1.0,1.0-d,1.0-d,1.0);
	colourOut = diff2 * (1.0 - diffBlend) + diff1 * diffBlend;
}
`
)

var gOpts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

// handle GLFW errors by printing them out
func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

// key events are a way to get input from GLFW.
// here we check for the escape key being pressed. if it is pressed,
// request that the window be closed
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeyZ && action == glfw.Press {
		gZoom = 1
		gOffX = 0
		gOffY = 0
		gBlend = 1
	} else if key == glfw.KeyUp && (action == glfw.Press || action == glfw.Repeat) {
		gBlend += 0.05
		if gBlend > 1 {
			gBlend = 1
		}
	} else if key == glfw.KeyDown && (action == glfw.Press || action == glfw.Repeat) {
		gBlend -= 0.05
		if gBlend < 0 {
			gBlend = 0
		}
	} else if key == glfw.KeyLeftBracket && (action == glfw.Press || action == glfw.Repeat) {
		gZoom /= 2
	} else if key == glfw.KeyRightBracket && (action == glfw.Press || action == glfw.Repeat) {
		gZoom *= 2
	} else if key == glfw.Key1 && (action == glfw.Press || action == glfw.Repeat) {
		gBlend = 0
	} else if key == glfw.Key2 && (action == glfw.Press || action == glfw.Repeat) {
		gBlend = 1
	} else if key == glfw.Key3 && (action == glfw.Press || action == glfw.Repeat) {
		gBlend = 0.5
	} else if key == glfw.KeyH && (action == glfw.Press || action == glfw.Repeat) {
		gHelp = !gHelp
	}
}

func winToFb(xpos float64, ypos float64) (x float32, y float32) {
	return float32(xpos) * gRetinaScale, float32(ypos) * gRetinaScale
}

func fbToTex(xpos float32, ypos float32) (x float32, y float32) {
	return xpos/gZoom - gOffX, ypos/gZoom - gOffY
}

func texToFb(xpos float32, ypos float32) (x float32, y float32) {
	return (xpos + gOffX) * gZoom, (ypos + gOffY) * gZoom
}

func fbColor(xtex float32, ytex float32) (col color.Color) {
	var xtexi, ytexi = int(math.Floor(float64(xtex) + .5)), int(math.Floor(float64(ytex) + .5))
	if gDiffFlag {
		if gBlend == 0 {
			col = gImage1.At(xtexi, ytexi)
		} else {
			col = gImage2.At(xtexi, ytexi)
		}
	} else {
		col = gImage1.At(xtexi, ytexi)
	}
	return col
}

func mouseDownCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	//var xpos, ypos float64 = w.GetCursorPosition()
	//gMouseX, gMouseY = winToFb(xpos, ypos)

	if action == glfw.Press {
		gMouseDown = true

		var xtex, ytex float32 = fbToTex(gMouseX, gMouseY)
		var col = fbColor(xtex, ytex)

		var r, g, b, a uint32 = col.RGBA()
		fmt.Printf("X,Y: %.2f %.2f RGBA: 0x%x 0x%x 0x%x 0x%x\n", xtex, ytex, r, g, b, a)
	} else {
		gMouseDown = false
	}
}

func mouseMoveCallback(w *glfw.Window, xpos float64, ypos float64) {
	var x, y float32 = winToFb(xpos, ypos)
	if gMouseDown {
		gOffX += (x - gMouseX) / gZoom
		gOffY += (y - gMouseY) / gZoom
	}
	gMouseX = x
	gMouseY = y
	//var xtex, ytex float32 = fbToTex(x, y)
	//fmt.Printf("%.2f %.2f\n", xtex, ytex)
}

func mouseWheelCallback(w *glfw.Window, xoff float64, yoff float64) {
	//fmt.Printf("x=%.2f y=%.2f\n", xoff, yoff)
	if yoff < 0 {
		gOffX += gMouseX / gZoom
		gOffY += gMouseY / gZoom
		gZoom /= 2
	} else if yoff > 0 {
		gZoom *= 2
		gOffX -= gMouseX / gZoom
		gOffY -= gMouseY / gZoom
	}
}

func downloadImage(url string) (path string, err error) {
	var f *os.File
	if f, err = ioutil.TempFile("", ""); err != nil {
		return "", err
	}

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	f.Close()
	return f.Name(), nil
}

func main() {
	var parser = flags.NewParser(&gOpts, flags.Default)

	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		os.Exit(1)
	}

	if len(args) < 1 || len(args) > 2 {
		panic(fmt.Errorf("Too many or not enough arguments"))
	}

	gDiffFlag = len(args) == 2

	// make sure that we display any errors that are encountered
	//glfw.SetErrorCallback(errorCallback)

	// the GLFW library has to be initialized before any of the methods
	// can be invoked
	if err = glfw.Init(); err != nil {
		panic(err)
	}

	// to be tidy, make sure glfw.Terminate() is called at the end of
	// the program to clean things up by using `defer`
	defer glfw.Terminate()

	// hints are the way you configure the features requested for the
	// window and are required to be set before calling glfw.CreateWindow().

	// desired number of samples to use for mulitsampling
	//glfw.WindowHint(glfw.Samples, 4)

	// request a OpenGL 4.1 core context
	if runtime.GOOS == "darwin" {
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
	}
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	var window *glfw.Window
	window, err = glfw.CreateWindow(1024, 768, "goicmpgl", nil, nil)
	if err != nil {
		// we legitimately cannot recover from a failure to create
		// the window in this sample, so just bail out
		panic(err)
	}

	// set the callback function to get all of the key input from the user
	window.SetKeyCallback(keyCallback)
	window.SetMouseButtonCallback(mouseDownCallback)
	window.SetScrollCallback(mouseWheelCallback)
	window.SetCursorPosCallback(mouseMoveCallback)

	// GLFW3 can work with more than one window, so make sure we set our
	// new window as the current context to operate on
	window.MakeContextCurrent()

	// disable v-sync for max FPS if the driver allows it
	//glfw.SwapInterval(0)

	// make sure that GLEW initializes all of the GL functions
	if err = gl.Init(); err != nil {
		panic(err)
	}

	var attribs = []string{
		"position",
		"uvs",
	}

	// compile our shaders
	var progTex0 *glplus.Program
	if progTex0, err = glplus.LoadShaderProgram(vertShader, fragShaderTex0, attribs); err != nil {
		panic(err)
	}
	defer progTex0.DeleteProgram()

	var progGrid *glplus.Program
	if progGrid, err = glplus.LoadShaderProgram(vertShader, sFragmentShaderGrid, attribs); err != nil {
		panic(err)
	}
	defer progGrid.DeleteProgram()

	var diffProg2 *glplus.Program
	if diffProg2, err = glplus.LoadShaderProgram(vertShader, sProgram2Src, attribs); err != nil {
		panic(err)
	}
	defer diffProg2.DeleteProgram()

	var diffProg3 *glplus.Program
	if diffProg3, err = glplus.LoadShaderProgram(vertShader, sProgram3Src, attribs); err != nil {
		panic(err)
	}
	defer diffProg3.DeleteProgram()

	var diffProg4 *glplus.Program
	if diffProg4, err = glplus.LoadShaderProgram(vertShader, sProgram4Src, attribs); err != nil {
		panic(err)
	}
	defer diffProg4.DeleteProgram()

	var diffProg5 *glplus.Program
	if diffProg5, err = glplus.LoadShaderProgram(vertShader, sProgram5Src, attribs); err != nil {
		panic(err)
	}
	defer diffProg5.DeleteProgram()

	var image1path = args[0]
	if strings.HasPrefix(image1path, "http") {
		if image1path, err = downloadImage(image1path); err != nil {
			panic(err)
		}
	}

	var texture *glplus.Texture
	if texture, gImage1, err = glplus.NewTexture(image1path, false, false); err != nil {
		panic(err)
	}
	defer texture.DeleteTexture()

	var texture2 *glplus.Texture
	if gDiffFlag {
		var image2path = args[1]
		if strings.HasPrefix(image2path, "http") {
			if image2path, err = downloadImage(image2path); err != nil {
				panic(err)
			}
		}

		if texture2, gImage2, err = glplus.NewTexture(image2path, false, false); err != nil {
			panic(err)
		}
		defer texture2.DeleteTexture()

		if texture.Size.X != texture2.Size.X || texture.Size.Y != texture2.Size.Y {
			fmt.Println("WARNING: image dimensions differ!")
		} else {
			fmt.Printf("image dimensions: %dx%d\n", texture.Size.X, texture.Size.Y)
		}
	} else {
		fmt.Printf("image dimensions: %dx%d\n", texture.Size.X, texture.Size.Y)
	}

	var font *glplus.Font
	if font, err = glplus.NewFont("FreeSerif.ttf"); err != nil {
		panic(err)
	}
	defer font.DeleteFont()

	var help1 = font.NewString("1: show only A")
	defer help1.DeleteString()
	var help2 = font.NewString("2: show only B")
	defer help2.DeleteString()
	var help3 = font.NewString("3: show diff A&B")
	defer help3.DeleteString()
	var helph = font.NewString("h: toggle this help")
	defer helph.DeleteString()
	var helparrows = font.NewString("<up>,<down>: go from A to B")
	defer helparrows.DeleteString()
	var helpzoom = font.NewString("[]: zoom in/out (also mouse wheel)")
	defer helpzoom.DeleteString()
	var helpclear = font.NewString("Z: reset zoom/view")
	defer helpclear.DeleteString()
	var helpescape = font.NewString("ESC: quit")
	defer helpescape.DeleteString()

	var vbo *glplus.VBO
	vbo = glplus.NewVBOQuad(0, 0, float32(texture.Size.X), float32(texture.Size.Y))
	defer vbo.DeleteVBO()

	var cnt float32

	// while there's no request to close the window
	for !window.ShouldClose() {
		cnt++

		// get the texture of the window because it may have changed since creation
		width, height := window.GetFramebufferSize()
		wwidth, _ := window.GetSize()
		gRetinaScale = float32(width) / float32(wwidth)

		//fmt.Printf("x=%d y=%d wx=%d wy=%d\n", width, height, wwidth, wheight)

		if cnt >= float32(width) {
			cnt = 0
		}

		var matrix = mgl32.Translate2D(-1, 1)
		matrix = matrix.Mul3(mgl32.Scale2D(2.0/float32(width), -2.0/float32(height)))

		// clear it all out
		gl.Viewport(0, 0, int32(width), int32(height))
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)

		var matrix3 = matrix.Mul3(mgl32.Scale2D(gZoom, gZoom))
		matrix3 = matrix3.Mul3(mgl32.Translate2D(gOffX, gOffY))

		// draw the grid
		if true {
			vbo.Bind()
			progGrid.UseProgram()

			color1 := [4]float32{.4, .4, .4, 1}
			color2 := [4]float32{.9, .9, .9, 1}
			grid := [3]float32{float32(texture.Size.X), float32(texture.Size.Y), 8 / gZoom}
			//fmt.Printf("%.2f %.2f %.2f %.2f\n", grid[0], grid[1], grid[2], grid[3])
			progGrid.ProgramUniformMatrix3fv("ModelviewMatrix", matrix3)
			progGrid.ProgramUniform4fv("color1", color1)
			progGrid.ProgramUniform4fv("color2", color2)
			progGrid.ProgramUniform3fv("grid", grid)

			if err = progGrid.ValidateProgram(); err != nil {
				panic(err)
			}

			vbo.Draw()

			vbo.Unbind()
			progGrid.UnuseProgram()
		}

		// draw the texture
		if !gDiffFlag {
			vbo.Bind()
			progTex0.UseProgram()
			texture.BindTexture(0)

			progTex0.ProgramUniformMatrix3fv("ModelviewMatrix", matrix3)
			progTex0.ProgramUniform1i("tex1", 0)
			progTex0.ProgramUniform1f("blend", gBlend)

			if err = progTex0.ValidateProgram(); err != nil {
				panic(err)
			}

			vbo.Draw()

			vbo.Unbind()
			progTex0.UnuseProgram()
			texture.UnbindTexture(0)
		} else {
			var diffBlend = gBlend

			var diffProg *glplus.Program

			if diffBlend < 0.25 {
				diffBlend *= 4
				diffProg = diffProg2
			} else if diffBlend < 0.5 { // 0.25 -> 0.5
				diffBlend = 4*diffBlend - 1
				diffProg = diffProg4
			} else if diffBlend < 0.75 { // 0.5 -> 0.75
				diffBlend = 4*diffBlend - 2
				diffProg = diffProg5
			} else { // 0.75 -> 1.0=
				diffBlend = 4*diffBlend - 3
				diffProg = diffProg3
			}

			vbo.Bind()
			diffProg.UseProgram()
			texture.BindTexture(0)
			texture2.BindTexture(1)

			diffProg.ProgramUniformMatrix3fv("ModelviewMatrix", matrix3)
			diffProg.ProgramUniform1i("decalA", 0)
			diffProg.ProgramUniform1i("decalB", 1)
			diffProg.ProgramUniform1f("diffBlend", diffBlend)

			if err = diffProg.ValidateProgram(); err != nil {
				panic(err)
			}

			vbo.Draw()

			vbo.Unbind()
			diffProg.UnuseProgram()
			texture.UnbindTexture(0)
			texture2.UnbindTexture(1)
		}

		// font
		if gHelp {
			color := [...]float32{0, 0, 1, 1}
			bg := [...]float32{0.5, 0.5, 0.5, 0.5}
			var line float32
			if err = helph.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128); err != nil {
				panic(err)
			}
			line++
			help1.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			help2.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			help3.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			helparrows.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			helpzoom.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			helpclear.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
			line++
			helpescape.Draw(font, color, bg, matrix, 0.5, 20, 100+line*128)
		}

		// swapping OpenGL buffers and polling events has been decoupled
		// in GLFW3, so make sure to invoke both here
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}
