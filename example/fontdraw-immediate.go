package main

import (
  "runtime"
  "fmt"
  "math"
  "strings"
  "io/ioutil"
  
  "image"
  "image/draw"

  "github.com/go-gl/gl/v4.1-core/gl"
  "github.com/go-gl/glfw/v3.2/glfw"
  
  "github.com/vrav/ratlas"
)

const (
  windowWidth = 800
  windowHeight = 600
  windowTitle = "fontdraw-simple"
)

type TextMesh []float32

func (mesh *TextMesh) AddRune(item *ratlas.AtlasItem, posX, posY, scale, r, g, b, a float32) {
  quad := runeQuadSingleColor(item, posX + item.BearingX*scale, posY - item.Descent*scale, scale, r, g, b, a);
  *mesh = append(*mesh, quad...)
}

func (mesh *TextMesh) TextBox(atlas *ratlas.Atlas, str string, left, top, width, height, fontPt, r, g, b, a float32) {
  textRunes := []rune(str)
  textLength := len(textRunes)
  
  scale := fontPt/float32(atlas.FontPt)
  
  var posX, posY float32
  posX, posY = left, top - float32(atlas.FontPt)*scale
  right, bottom := left+width, top-height
  lineHeight := atlas.Height()*scale
  
  Outer: for i, ru := range textRunes {
    // seek ahead to see if word requires wrapping; if so, wrap
    if ru != ' ' && i > 0 && textRunes[i-1] == ' ' {
      aheadPos := posX
      for j, rr := range textRunes[i:] {
        if rr == ' ' || rr == '\n' {
          break
        }
        if item, ok := atlas.Items[rr]; ok {
          aheadPos += item.Advance * scale
          if j < textLength - 1 {
            aheadPos += atlas.Kern(textRunes[j], textRunes[j+1]) * scale
          }
        }
      }
      if aheadPos > right {
        // word will spill off edge; wrap line
        if posY < bottom {
          // fmt.Println("broke at wrap")
          break Outer;
        }
        posY -= lineHeight
        posX = left
      }
    }
    
    item, ok := atlas.Items[ru]
    if !ok {
      if ru == '\n' {
        if posY < bottom {
          // fmt.Println("broke at \\n")
          break Outer;
        }
        posY -= lineHeight
        posX = left
      }
      continue
    }
    
    if ru == ' ' {
      posX += item.Advance * scale
      continue
    }
    
    mesh.AddRune(item, posX, posY, scale, r, g, b, a)
    
    posX += item.Advance * scale
    if i < textLength - 1 {
      posX += atlas.Kern(textRunes[i], textRunes[i+1]) * scale
    }
  }
}

func screenX(x float32) float32 {
  return 2.0 * (x / float32(windowWidth)) - 1.0
}

func screenY(y float32) float32 {
  return 2.0 * (y / float32(windowHeight)) - 1.0
}

func runeQuadSingleColor(item *ratlas.AtlasItem, posX, posY, scale, r, g, b, a float32) TextMesh {
  w, h := float32(item.Width), float32(item.Height)
  
  return TextMesh{
    // X Y   U V   scale   r g b a
    screenX(posX), screenY(posY),                      item.PercentPosX, item.PercentPosY+item.PercentHeight, scale, r, g, b, a,
    screenX(posX + w*scale), screenY(posY + h*scale),  item.PercentPosX+item.PercentWidth, item.PercentPosY, scale, r, g, b, a,
    screenX(posX), screenY(posY + h*scale),            item.PercentPosX, item.PercentPosY, scale, r, g, b, a,
    
    screenX(posX), screenY(posY),                      item.PercentPosX, item.PercentPosY+item.PercentHeight, scale, r, g, b, a,
    screenX(posX + w*scale), screenY(posY),            item.PercentPosX+item.PercentWidth, item.PercentPosY+item.PercentHeight, scale, r, g, b, a,
    screenX(posX + w*scale), screenY(posY + h*scale),  item.PercentPosX+item.PercentWidth, item.PercentPosY, scale, r, g, b, a,
  }
}

// handleKeys is used as a glfw key callback
func handleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeySpace && action == glfw.Press {
		//
  }
}

func main() {
  // init glfw
  runtime.LockOSThread()
  if err := glfw.Init(); err != nil {
    panic(err)
  }
  defer glfw.Terminate()
  
  glfw.WindowHint(glfw.Resizable, glfw.False)
  glfw.WindowHint(glfw.ContextVersionMajor, 3)
  glfw.WindowHint(glfw.ContextVersionMinor, 2)
  glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
  glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
  
  window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
  if err != nil {
    panic(err)
  }
  
  window.SetKeyCallback(handleKeys)
  window.MakeContextCurrent()
  
  // init gl
  if err := gl.Init(); err != nil {
    panic(err)
  }
  version := gl.GoStr(gl.GetString(gl.VERSION))
  fmt.Println("OpenGL version", version)
  gl.ClearColor(0.2,0.2,0.3,1.0)
  
  // load font file bytes
  fontFile := "Vera.ttf"
  ttfData, err := ioutil.ReadFile(fontFile)
  if err != nil {
    panic(err)
  }
  
  // init rune atlas
  // runes := []rune(" ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890`~!@#$%^&*()[]{}/=?+\\|-_.>,<'\";:�")
  // atlas := ratlas.New(&ttfData, 72.0, 512, 512, 4, runes)
  var atlas ratlas.Atlas
  err = atlas.LoadGobFile("Vera.ttf.gob")
  if err != nil {
    panic(err)
  }
  err = atlas.LoadImageFiles([]string{"atlas0sdf.png"})
  if err != nil {
    panic(err)
  }
  atlas.ScaleNumbers(0.25)
  err = atlas.ReloadFont(&ttfData)
  if err != nil {
    panic(err)
  }
  
  // the important bit. generate a mesh of exampleText's characters.
  
  
  // loading up OpenGL.
  
  // Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
    panic(err)
  }
  gl.UseProgram(program)
  
  textureUniform := gl.GetUniformLocation(program, gl.Str("texture\x00"))
  gl.Uniform1i(textureUniform, 0)
  gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))
  
  // load the texture into GL
  texture, err := newTexture(atlas.Images[0])
  if err != nil {
    panic(err)
  }
  
  // generate vao/vbo
  var vao uint32
  gl.GenVertexArrays(1, &vao)
  gl.BindVertexArray(vao)
  
  var vbo uint32
  // make sure maxVertices is big enough to hold all the text.
  // in a more robust implementation, requiring to draw more text could call glBufferData again at a new size.
  maxRunes := 8192
  maxVertices := 3*2*maxRunes
  lenBuffer := maxVertices*9
  gl.GenBuffers(1, &vbo)
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.BufferData(gl.ARRAY_BUFFER, lenBuffer*4, gl.Ptr(nil), gl.STREAM_DRAW)
  
  posAttrib := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
  gl.EnableVertexAttribArray(posAttrib)
  gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 9*4, gl.PtrOffset(0))
  
  texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("texCoord\x00")))
  gl.EnableVertexAttribArray(texCoordAttrib)
  gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 9*4, gl.PtrOffset(2*4))
  
  scaleAttrib := uint32(gl.GetAttribLocation(program, gl.Str("glyphScale\x00")))
  gl.EnableVertexAttribArray(scaleAttrib)
  gl.VertexAttribPointer(scaleAttrib, 1, gl.FLOAT, false, 9*4, gl.PtrOffset(4*4))
  
  colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("glyphColor\x00")))
  gl.EnableVertexAttribArray(colorAttrib)
  gl.VertexAttribPointer(colorAttrib, 4, gl.FLOAT, false, 9*4, gl.PtrOffset(5*4))
  
  // setup before drawing
  
  // enable blending
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  fmt.Println("Drawing to a buffer with max runes:", maxRunes)
  
  for !window.ShouldClose() {
    time := glfw.GetTime()
    
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    
    mX, mY := window.GetCursorPos()
    mY = float64(windowHeight) - mY
    
    var frameMesh TextMesh
    
    frameMesh.TextBox(&atlas, "Test string!", float32(mX), float32(mY), 250, 250, 150.0, 1.0, 0.8, 1.0, 0.5)
    
    yLoop: for y := windowHeight; y > windowHeight-401; y -= 200 {
      for x := 0; x < windowWidth; x += 200 {
        pX, pY := float32(float64(x) + 6.0*math.Sin(float64(x)+time*3.0)), float32(float64(y) + 4.0*math.Cos(float64(y)+time))
        wW, wH := float32(windowWidth), float32(windowHeight)
        fontPt := 10.0 + 20.0 * (float32(x)/float32(windowWidth))
        frameMesh.TextBox(&atlas, exampleText, pX, pY, 200, 200, fontPt, pX/wW, pY/wH, 1.0, 0.8)
        if len(frameMesh)/9 > maxVertices {
          break yLoop
        }
      }
    }
    
    numVertices := len(frameMesh)/9
    
    // option 1: glBufferData
    // gl.BufferData(gl.ARRAY_BUFFER, len(frameMesh)*4, gl.Ptr(frameMesh), gl.STREAM_DRAW)
    
    // option 2: glBufferSubData at same size or less
    gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
    if numVertices > maxVertices {
      gl.BufferSubData(gl.ARRAY_BUFFER, 0, maxVertices*9*4, gl.Ptr(frameMesh))
    } else {
      gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(frameMesh)*4, gl.Ptr(frameMesh))
    }
    
    // bind drawing primitives
    gl.BindVertexArray(vao)
    gl.ActiveTexture(gl.TEXTURE0)
    gl.BindTexture(gl.TEXTURE_2D, texture)
    
    verticesToDraw := int32(numVertices)
    if numVertices > maxVertices {
      verticesToDraw = int32(maxVertices - maxVertices%3)
    }
    
    gl.DrawArrays(gl.TRIANGLES, 0, verticesToDraw)
    
    glfw.PollEvents()
    window.SwapBuffers()
    
    // time.Sleep(time.Second/time.Duration(windowFps) - time.Since(t))
  }
  
  gl.DeleteTextures(1, &texture)
  gl.DeleteBuffers(1, &vbo)
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

// compileShader compiles a single shader and returns its uint32 ID plus an error
func compileShader(source string, shaderType uint32) (uint32, error) {
  shader := gl.CreateShader(shaderType)
  
  csources, free := gl.Strs(source)
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

func newTexture(img draw.Image) (uint32, error) {
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
  
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

var vertexShader = `
#version 150

in vec2 position;
in vec2 texCoord;
in float glyphScale;
in vec4 glyphColor;
out vec2 uv;
out float scale;
out vec4 color;

void main() {
  uv = texCoord;
  scale = glyphScale;
  color = glyphColor;
  gl_Position = vec4(position, 0.0, 1.0);
}
` + "\x00"

var fragmentShader = `
#version 150

in vec2 uv;
in float scale;
in vec4 color;
uniform sampler2D texture;
out vec4 outputColor;

void main() {
  float distance = texture2D(texture, uv).r;
  float smoothing = 1.0/16.0/scale;
  float alpha = smoothstep(0.5 - smoothing, 0.5 + smoothing, distance);
  // float alpha = texture2D(texture, uv).r;
  outputColor = vec4(color.r, color.g, color.b, color.a * alpha);
}
` + "\x00"

var exampleText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam lectus nisl, sodales a lectus at, interdum facilisis nisi. Sed in nunc pellentesque, commodo leo at, scelerisque eros. Nulla varius pharetra ipsum eu fringilla. Integer eu mattis est, sit amet molestie turpis. Suspendisse laoreet neque lobortis rutrum suscipit. Sed pharetra auctor odio, suscipit rutrum quam posuere in. Nulla rhoncus et purus quis placerat. Fusce dictum ex eget nisl iaculis egestas. Vestibulum at ligula id ligula ultricies posuere. Donec a ipsum sed risus tincidunt cursus ac eget risus. Quisque tortor nisi, posuere id purus ut, accumsan dapibus dolor. Mauris facilisis et risus et mattis. Suspendisse ex sapien, dapibus at nisl ut, luctus euismod sem. `
