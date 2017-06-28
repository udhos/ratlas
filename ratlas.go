// Package ratlas generates rune atlas images given a slice of runes and a TTF font.
// Atlas and atlas item information can also be saved and loaded.
// Values are stored as float32 for ease of use with OpenGL.
package ratlas

import (
  "encoding/gob"
  "bytes"
  "sort"
  "fmt"
  "os"
  "io/ioutil"
  
  "image"
  "image/draw"
  "image/png"
  
  "golang.org/x/image/font"
  "github.com/golang/freetype/truetype"
  "golang.org/x/image/math/fixed"
)

func fixedFloat(v fixed.Int26_6) float32 {
  return float32(v)/64.0
}

// AtlasItem contains all the information needed to draw a specific rune within an Atlas.
type AtlasItem struct {
  Rune rune
  Advance float32
  BearingX float32
  Descent float32
  PercentPosX float32
  PercentPosY float32
  PercentWidth float32
  PercentHeight float32
  Width int
  Height int
  Node *node
  ImageIndex int
}

type Atlas struct {
  Face font.Face
  FontPt float64
  Pad int
  
  Items map[rune]*AtlasItem
  Images []draw.Image
}

// atlasItems implements Sort interface for slice of AtlasItem
// This is needed for the somewhat redundant packing algorithm
type atlasItems []*AtlasItem
func (slice atlasItems) Len() int {
    return len(slice)
}
func (slice atlasItems) Less(i, j int) bool {
    // return slice[i].Height + slice[i].Width * slice[i].Height > slice[j].Height + slice[j].Width * slice[j].Height
    return slice[i].Height > slice[j].Height
}
func (slice atlasItems) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

// Node contains 2D bin packing implementation for sorting glyphs into atlas image
type node struct {
  Used bool
  Right *node
  Down *node
  X, Y, W, H int
}
func (atlas Atlas) containsNilNodes() bool {
  for _, atlasItem := range atlas.Items {
    if atlasItem.Node == nil {
      return true
    }
  }
  return false
}
func (atlas Atlas) getNilNodes() atlasItems {
  var itemSlice atlasItems
  for _, atlasItem := range atlas.Items {
    if atlasItem.Node == nil {
      itemSlice = append(itemSlice, atlasItem)
    }
  }
  return itemSlice
}
func fitAtlasItems(items atlasItems, w, h int) {
  root := &node{ X: 0, Y: 0, W: w, H: h}
  for _, item := range items {
    if node := root.findNode(item.Width, item.Height); node != nil {
      item.Node = node.splitNode(item.Width, item.Height)
    }
  }
}
func (root *node) findNode(w, h int) *node {
  if (root.Used) {
    rightFind := root.Right.findNode(w, h)
    if rightFind != nil {
      return rightFind
    } else {
      return root.Down.findNode(w, h);
    }
  } else if ((w <= root.W) && (h <= root.H)) {
    return root;
  } else {
    return nil;
  }
}
func (this *node) splitNode(w, h int) *node {
  this.Used = true;
  this.Down  = &node{ Used: false, X: this.X, Y: this.Y + h, W: this.W, H: this.H - h,  }
  this.Right = &node{ Used: false, X: this.X + w, Y: this.Y, W: this.W - w, H: h,  }
  return this
}

func (atlas *Atlas) GobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(atlas.FontPt)
    if err != nil {
        return nil, err
    }
    err = encoder.Encode(atlas.Pad)
    if err != nil {
        return nil, err
    }
    err = encoder.Encode(atlas.Items)
    if err != nil {
        return nil, err
    }
    return w.Bytes(), nil
}
func (atlas *Atlas) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&atlas.FontPt)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&atlas.Pad)
    if err!=nil {
        return err
    }
    return decoder.Decode(&atlas.Items)
}

func (atlas *Atlas) createGob() ([]byte, error) {
  buffer := new(bytes.Buffer)
  enc := gob.NewEncoder(buffer)
  err := enc.Encode(atlas)
  if err != nil {
      return nil, fmt.Errorf("ratlas: encode error: %v", err)
  }
  return buffer.Bytes(), nil
}
func (atlas *Atlas) readGob(b []byte) error {
  buffer := bytes.NewBuffer(b)
  dec := gob.NewDecoder(buffer)
  err := dec.Decode(atlas)
  if err != nil {
    return fmt.Errorf("ratlas: decode error: %v", err)
  }
  return nil
}

// SaveGobFile dumps atlas info to a gob file.
func (atlas *Atlas) SaveGobFile(fileName string) error {
  outGob, err := os.Create(fileName)
  if err != nil {
    return fmt.Errorf("ratlas: couldn't create file %s: %v", fileName, err)
  }
  defer outGob.Close()
  
  gobBytes, err := atlas.createGob()
  if err != nil {
    return err
  }
  
  numBytes, err := outGob.Write(gobBytes)
  if err != nil {
    return fmt.Errorf("ratlas: couldn't write file %s: %v", fileName, err)
  }
  fmt.Println("ratlas: wrote", fileName, numBytes)
  return nil
}

// LoadGobFile populates an empty atlas per the contents of an exported gob file.
func (atlas *Atlas) LoadGobFile(fileName string) error {
  b, err := ioutil.ReadFile(fileName)
  if err != nil {
    return fmt.Errorf("ratlas: couldn't read file %s: %v", fileName, err)
  }
  
  err = atlas.readGob(b)
  if err != nil {
    return err
  }
  
  fmt.Println("ratlas: loaded", fileName)
  return nil
}

// SaveImageFiles dumps all generated atlas images to disk.
func (atlas *Atlas) SaveImageFiles(name string) error {
  for i, img := range atlas.Images {
    outFilename := fmt.Sprintf("%s-%d.png", name, i)
    outFile, err := os.Create(outFilename)
    if err != nil {
      return fmt.Errorf("ratlas: couldn't create file %s: %v", outFilename, err)
    }
    defer outFile.Close()
    err = png.Encode(outFile, img)
    if err != nil {
      return fmt.Errorf("ratlas: couldn't encode png:", err)
    }
    fmt.Println("ratlas: wrote", outFilename)
  }
  return nil
}

// LoadImageFiles loads a slice of strings that point to image files to load into the atlas.
func (atlas *Atlas) LoadImageFiles(imageFilenames []string) error {
  for _, imageFilename := range imageFilenames {
    inFile, err := os.Open(imageFilename)
    if err != nil {
      return fmt.Errorf("ratlas: couldn't open file %s: %v", imageFilename, err)
    }
    defer inFile.Close()
    
    img, formatString, err := image.Decode(inFile)
    if err != nil {
      return fmt.Errorf("ratlas: couldn't decode image: %v", err)
    }
    
    dimg, ok := img.(draw.Image)
    if !ok {
      return fmt.Errorf("ratlas: couldn't create drawable image from %s\n", imageFilename)
    }
    
    atlas.Images = append(atlas.Images, dimg)
    fmt.Printf("ratlas: loaded %s as %s\n", imageFilename, formatString)
  }
  return nil
}

// ReloadFont parses TTF data in order to generate a font.Face for the atlas.
func (atlas *Atlas) ReloadFont(ttfData *[]byte) error {
  // parse file bytes into font
  f, err := truetype.Parse(*ttfData)
  if err != nil {
    return fmt.Errorf("ratlas: couldn't parse font: %v", err)
  }
  
  opts := &truetype.Options{
    Size: atlas.FontPt,
    DPI: 72.0,
    Hinting: font.HintingNone,
    GlyphCacheEntries: 512,
    SubPixelsX: 4,
    SubPixelsY: 1,
  }
  face := truetype.NewFace(f, opts)
  fmt.Println("ratlas: loaded and parsed TTF data")
  
  atlas.Face = face
  
  return nil
}

// ScaleNumbers scales the numbers within an Atlas and its AtlasItem(s), for example, if a loaded image was scaled since saving the atlas info.
func (atlas *Atlas) ScaleNumbers(v float32) {
  atlas.FontPt *= float64(v)
  atlas.Pad = int(float32(atlas.Pad)*v)
  
  for _, atlasItem := range atlas.Items {
    atlasItem.Advance *= v
    atlasItem.BearingX *= v
    atlasItem.Descent *= v
    atlasItem.Width = int(float32(atlasItem.Width)*v)
    atlasItem.Height = int(float32(atlasItem.Height)*v)
    
    atlasItem.Node.X = int(float32(atlasItem.Node.X)*v)
    atlasItem.Node.Y = int(float32(atlasItem.Node.Y)*v)
    atlasItem.Node.W = atlasItem.Width
    atlasItem.Node.H = atlasItem.Height
  }
  fmt.Println("ratlas: scaled atlas numbers by", v)
}

// Kern returns a float32 of the kern distance between two runes.
func (atlas *Atlas) Kern(a, b rune) float32 {
  return fixedFloat(atlas.Face.Kern(a, b))
}

// Ascent returns a float32 of the distance from the top of a line to its baseline.
func (atlas *Atlas) Ascent() float32 {
  faceMetrics := atlas.Face.Metrics()
  return fixedFloat(faceMetrics.Ascent)
}

// Height returns a float32 of the recommended amount of vertical space between two lines of text.
func (atlas *Atlas) Height() float32 {
  faceMetrics := atlas.Face.Metrics()
  return fixedFloat(faceMetrics.Height)
}

// Descent returns a float32 of the distance from the bottom of a line to its baseline.
func (atlas *Atlas) Descent() float32 {
  faceMetrics := atlas.Face.Metrics()
  return fixedFloat(faceMetrics.Descent)
}

// New returns a Atlas of a given TTF data, image dimensions, and a given slice of runes.
func New(ttfData *[]byte, fontPt float64, imgWidth, imgHeight, pad int, runes []rune) Atlas {
  // create atlas
  var atlas Atlas
  atlas.FontPt = fontPt
  atlas.ReloadFont(ttfData)
  atlas.Pad = pad
  atlas.Items = make(map[rune]*AtlasItem)
  
  // cycle through runes and add each
  for _, r := range runes {
    // _, ok := atlas.Face.GlyphAdvance(r)
    // if !ok {
    //   fmt.Println("not ok\n")
    //   continue
    // }
    
    var atlasItem AtlasItem
    atlasItem.Rune = r
      
    bounds, advance, _ := atlas.Face.GlyphBounds(r)
    minX := bounds.Min.X.Floor()
    minY := bounds.Min.Y.Floor()
    maxX := bounds.Max.X.Ceil()
    maxY := bounds.Max.Y.Ceil()
    atlasItem.Advance = fixedFloat(advance)
    // fmt.Printf("%s {%v, %v} {%v, %v} %v\n", string(r), minX, minY, maxX, maxY, atlasItem.Advance)
    
    atlasItem.BearingX = fixedFloat(bounds.Min.X) - float32(pad)
    atlasItem.Descent = float32(maxY) + (fixedFloat(bounds.Min.Y) - float32(minY)) + float32(pad)
    // ^ not sure if tiny middle add is needed, still WIP
    // fmt.Printf("%s x %v, descent %v\n", string(r), atlasItem.BearingX, atlasItem.Descent)
    
    glyphWidth := maxX - minX + pad*2
    glyphHeight := maxY - minY + pad*2
    
    atlasItem.Width = glyphWidth
    atlasItem.Height = glyphHeight
    
    atlasItem.PercentWidth = float32(glyphWidth) / float32(imgWidth)
    atlasItem.PercentHeight = float32(glyphHeight) / float32(imgHeight)
    
    atlas.Items[atlasItem.Rune] = &atlasItem
  }
  
  // while we have glyphs that aren't on a sheet, create new sheets for them
  for atlas.containsNilNodes() {
    // create new atlas image sheet
    atlas.Images = append(atlas.Images, image.NewGray(image.Rect(0, 0, imgWidth, imgHeight)))
    imageIndex := len(atlas.Images) - 1
    draw.Draw(atlas.Images[imageIndex], atlas.Images[imageIndex].Bounds(), image.Black, image.Point{}, draw.Src)
    
    // sort nil nodes per AtlasItems sort implementation
    itemSlice := atlas.getNilNodes()
    sort.Sort(itemSlice)
    
    // give each rune a position within an image sheet
    // if it doesn't fit on current sheet, node remains nil
    fitAtlasItems(itemSlice, imgWidth, imgHeight)
    
    // attempt to copy AtlasItems into atlas sheet, until full
    for _, atlasItem := range itemSlice {
      if atlasItem.Node == nil {
        break
      }
      atlasItem.ImageIndex = imageIndex
      
      bounds, _, _ := atlas.Face.GlyphBounds(atlasItem.Rune)
      minX := bounds.Min.X.Floor()
      // maxX := bounds.Max.X.Ceil()
      minY := bounds.Min.Y.Floor()
      // maxY := bounds.Max.Y.Ceil()
      
      // create glyph image
      dst := image.NewGray(image.Rect(0, 0, atlasItem.Width, atlasItem.Height))
      draw.Draw(dst, dst.Bounds(), image.Black, image.Point{}, draw.Src)
      
      // render glyph to free standing glyph image
      d := &font.Drawer{
        Dst: dst,
        Src: image.White,
        Face: atlas.Face,
      }
      d.Dot = fixed.P(-minX+pad, -minY+pad)
    	dr, mask, maskp, _, _ := d.Face.Glyph(d.Dot, atlasItem.Rune)
      draw.DrawMask(d.Dst, dr, d.Src, image.Point{}, mask, maskp, draw.Over)
      
      // copy glyph image to atlas image
      draw.Draw(atlas.Images[imageIndex], image.Rect(atlasItem.Node.X, atlasItem.Node.Y, atlasItem.Node.X+atlasItem.Width, atlasItem.Node.Y+atlasItem.Height), dst, image.Point{}, draw.Src)
      
      atlasItem.PercentPosX = float32(atlasItem.Node.X) / float32(imgWidth)
      atlasItem.PercentPosY = float32(atlasItem.Node.Y) / float32(imgHeight)
    }
  }
  
  return atlas
}
