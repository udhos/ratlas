package main

import (
  "runtime"
  "fmt"
  "log"
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

func screenX(x float32) float32 {
  return 2.0 * (x / float32(windowWidth)) - 1.0
}

func screenY(y float32) float32 {
  return 2.0 * (y / float32(windowHeight)) - 1.0
}

func runeQuad(item *ratlas.AtlasItem, posX, posY, scale float32) []float32 {
  w, h := float32(item.Width), float32(item.Height)
  
  return []float32{
    // X Y   U V
    screenX(posX), screenY(posY),                      item.PercentPosX, item.PercentPosY+item.PercentHeight,
    screenX(posX + w*scale), screenY(posY + h*scale),  item.PercentPosX+item.PercentWidth, item.PercentPosY,
    screenX(posX), screenY(posY + h*scale),            item.PercentPosX, item.PercentPosY,
    
    screenX(posX), screenY(posY),                      item.PercentPosX, item.PercentPosY+item.PercentHeight,
    screenX(posX + w*scale), screenY(posY),            item.PercentPosX+item.PercentWidth, item.PercentPosY+item.PercentHeight,
    screenX(posX + w*scale), screenY(posY + h*scale),  item.PercentPosX+item.PercentWidth, item.PercentPosY,
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
  
  var mesh []float32
  textRunes := []rune(exampleText)
  textLength := len(textRunes)
  leftMargin := float32(10.0)
  scale := float32(0.5)
  
  var posX, posY float32
  posX, posY = leftMargin, windowHeight - float32(atlas.FontPt)*scale
  lineHeight := atlas.Height()*scale
  
  for i, r := range textRunes {
    // seek ahead to see if word requires wrapping; if so, wrap
    if r != ' ' && i > 0 && textRunes[i-1] == ' ' {
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
      if aheadPos > windowWidth {
        // word will spill off edge of windowWidth; wrap line
        posY -= lineHeight
        posX = leftMargin
      }
    }
    
    item, ok := atlas.Items[r]
    if !ok {
      if r == '\n' {
        posY -= lineHeight
        posX = leftMargin
      }
      continue
    }
    
    quad := runeQuad(item, posX + item.BearingX*scale, posY - item.Descent*scale, scale)
    mesh = append(mesh, quad...)
    posX += item.Advance * scale
    if i < textLength - 1 {
      posX += atlas.Kern(textRunes[i], textRunes[i+1]) * scale
    }
  }
  
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
  
  scaleUniform := gl.GetUniformLocation(program, gl.Str("scale\x00"))
  gl.Uniform1f(scaleUniform, scale)
  
  // load the texture into GL
  texture, err := newTexture(atlas.Images[0])
  if err != nil {
    log.Fatalln(err)
  }
  
  // generate vao/vbo
  var vao uint32
  gl.GenVertexArrays(1, &vao)
  gl.BindVertexArray(vao)
  
  var vbo uint32
  gl.GenBuffers(1, &vbo)
  gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
  gl.BufferData(gl.ARRAY_BUFFER, len(mesh)*4, gl.Ptr(mesh), gl.STATIC_DRAW)
  
  posAttrib := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
  gl.EnableVertexAttribArray(posAttrib)
  gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
  
  texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("texCoord\x00")))
  gl.EnableVertexAttribArray(texCoordAttrib)
  gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
  
  numVertices := int32(len(mesh)/4)
  
  // enable blending
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  
  // bind once here
  gl.BindVertexArray(vao)
  gl.ActiveTexture(gl.TEXTURE0)
  gl.BindTexture(gl.TEXTURE_2D, texture)
  
  
	// previousTime := glfw.GetTime()
  for !window.ShouldClose() {
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    
    gl.DrawArrays(gl.TRIANGLES, 0, numVertices)
    
    glfw.PollEvents()
    window.SwapBuffers()
    
    // time.Sleep(time.Second/time.Duration(windowFps) - time.Since(t))
  }
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
out vec2 uv;

void main() {
  uv = texCoord;
  gl_Position = vec4(position, 0.0, 1.0);
}
` + "\x00"

var fragmentShader = `
#version 150

in vec2 uv;
uniform sampler2D texture;
uniform float scale;
out vec4 outputColor;

void main() {
  float distance = texture2D(texture, uv).r;
  float smoothing = 1.0/24.0/scale;
  float alpha = smoothstep(0.5 - smoothing, 0.5 + smoothing, distance);
  // float alpha = texture2D(texture, uv).r;
  outputColor = vec4(1.0, 1.0, 1.0, alpha);
}
` + "\x00"

var exampleText = `  This example text is rendered as a mesh created with screen-coordinate quads for each character of text. The size and position of each rune glyph is determined by information accessible in a properly initialized rune atlas, such as glyph advance, descent, and kerning.
  
  Proportional glyph positions and dimensions in the atlas are saved as PercentPosX, PercentWidth, etc for ease of use in UV generation.

  Simple word wrapping is also implemented in this example. Cheers!
`
