package ratlas

import (
  "encoding/gob"
  "bytes"
  "sort"
  "fmt"
  "log"
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
  Node *Node
  ImageIndex int
}

type Atlas struct {
  Face font.Face
  FontFile string
  FontPt float64
  FontRes float64
  Pad int
  
  Items map[rune]*AtlasItem
  Images []draw.Image
}

// AtlasItems implements Sort interface for slice of AtlasItem
// This is needed for the somewhat redundant packing algorithm
type AtlasItems []*AtlasItem
func (slice AtlasItems) Len() int {
    return len(slice)
}
func (slice AtlasItems) Less(i, j int) bool {
    // return slice[i].Height + slice[i].Width * slice[i].Height > slice[j].Height + slice[j].Width * slice[j].Height
    return slice[i].Height > slice[j].Height
}
func (slice AtlasItems) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

// Node contains 2D bin packing implementation for sorting glyphs into atlas image
type Node struct {
  Used bool
  Right *Node
  Down *Node
  X, Y, W, H int
}
func (atlas Atlas) ContainsNilNodes() bool {
  for _, atlasItem := range atlas.Items {
    if atlasItem.Node == nil {
      return true
    }
  }
  return false
}
func (atlas Atlas) GetNilNodes() AtlasItems {
  var itemSlice AtlasItems
  for _, atlasItem := range atlas.Items {
    if atlasItem.Node == nil {
      itemSlice = append(itemSlice, atlasItem)
    }
  }
  return itemSlice
}
func FitAtlasItems(items AtlasItems, w, h int) {
  root := &Node{ X: 0, Y: 0, W: w, H: h}
  for _, item := range items {
    if node := root.FindNode(item.Width, item.Height); node != nil {
      item.Node = node.SplitNode(item.Width, item.Height)
    }
  }
}
func (root *Node) FindNode(w, h int) *Node {
  if (root.Used) {
    rightFind := root.Right.FindNode(w, h)
    if rightFind != nil {
      return rightFind
    } else {
      return root.Down.FindNode(w, h);
    }
  } else if ((w <= root.W) && (h <= root.H)) {
    return root;
  } else {
    return nil;
  }
}
func (node *Node) SplitNode(w, h int) *Node {
  node.Used = true;
  node.Down  = &Node{ Used: false, X: node.X, Y: node.Y + h, W: node.W, H: node.H - h,  }
  node.Right = &Node{ Used: false, X: node.X + w, Y: node.Y, W: node.W - w, H: h,  }
  return node
}

func (atlas *Atlas) GobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(atlas.FontFile)
    if err != nil {
        return nil, err
    }
    err = encoder.Encode(atlas.FontPt)
    if err != nil {
        return nil, err
    }
    err = encoder.Encode(atlas.FontRes)
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
    err := decoder.Decode(&atlas.FontFile)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&atlas.FontPt)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&atlas.FontRes)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&atlas.Pad)
    if err!=nil {
        return err
    }
    return decoder.Decode(&atlas.Items)
}

func (atlas *Atlas) createGob() []byte {
  buffer := new(bytes.Buffer)
  enc := gob.NewEncoder(buffer)
  err := enc.Encode(atlas)
  if err != nil {
      log.Println("Encode error:", err)
  }
  return buffer.Bytes()
}
func (atlas *Atlas) readGob(b []byte) {
  buffer := bytes.NewBuffer(b)
  dec := gob.NewDecoder(buffer)
  err := dec.Decode(atlas)
  if err != nil {
    log.Println("Decode error:", err)
  }
}

func (atlas *Atlas) SaveGobFile(outFile string) {
  outGob, err := os.Create(outFile)
  if err != nil {
    log.Println("Couldn't create file:", err)
    return
  }
  defer outGob.Close()
  numBytes, err := outGob.Write(atlas.createGob())
  if err != nil {
    log.Println("Couldn't write file:", err)
    return
  }
  log.Println("Wrote", outFile, numBytes)
}
func (atlas *Atlas) LoadGobFile(fileName string) {
  b, err := ioutil.ReadFile(fileName)
  if err != nil {
    log.Println("Couldn't read file:", err)
    return
  }
  atlas.readGob(b)
  log.Println("Loaded", fileName)
}

func (atlas *Atlas) SaveImageFiles() {
  for i, img := range atlas.Images {
    outFilename := fmt.Sprintf("%s-%d.png", atlas.FontFile, i)
    outFile, err := os.Create(outFilename)
    if err != nil {
      log.Println("Couldn't create file:", err)
      return
    }
    defer outFile.Close()
    err = png.Encode(outFile, img)
    if err != nil {
      log.Println("Couldn't encode png:", err)
      return
    }
    log.Println("Wrote", outFilename)
  }
}
func (atlas *Atlas) LoadImageFiles(imageFilenames []string) {
  for _, imageFilename := range imageFilenames {
    inFile, err := os.Open(imageFilename)
    if err != nil {
      log.Println("Couldn't open file:", err)
      return
    }
    defer inFile.Close()
    
    img, formatString, err := image.Decode(inFile)
    if err != nil {
      log.Println("Couldn't decode image:", err)
      return
    }
    
    dimg, ok := img.(draw.Image)
    if !ok {
      log.Printf("Couldn't create drawable image from %s\n", imageFilename)
      return
    }
    
    atlas.Images = append(atlas.Images, dimg)
    log.Printf("Loaded %s as %s\n", imageFilename, formatString)
  }
}

func (atlas *Atlas) ReloadFont() {
  fontFile := atlas.FontFile
  
  // load font file bytes
  bytes, err := ioutil.ReadFile(fontFile)
  if err != nil {
    log.Println("Couldn't read file:", err)
    return
  }
  
  // parse file bytes into font
  f, err := truetype.Parse(bytes)
  if err != nil {
    log.Println("Couldn't parse font:", err)
    return
  }
  
  opts := &truetype.Options{
    Size: atlas.FontPt,
    DPI: atlas.FontRes,
    Hinting: font.HintingNone,
    GlyphCacheEntries: 512,
    SubPixelsX: 4,
    SubPixelsY: 1,
  }
  face := truetype.NewFace(f, opts)
  log.Println("Loaded and parsed", fontFile)
  
  atlas.Face = face
}

func (atlas *Atlas) ScaleNumbers(v float32) {
  atlas.FontPt *= float64(v)
  atlas.FontRes *= float64(v)
  atlas.Pad = int(float32(atlas.Pad)*v)
  
  for _, atlasItem := range atlas.Items {
    atlasItem.Advance *= v
    atlasItem.BearingX *= v
    atlasItem.Descent *= v
    atlasItem.Width = int(float32(atlasItem.Width)*v)
    atlasItem.Height = int(float32(atlasItem.Width)*v)
    
    atlasItem.Node.X = int(float32(atlasItem.Node.X)*v)
    atlasItem.Node.Y = int(float32(atlasItem.Node.Y)*v)
    atlasItem.Node.W = atlasItem.Width
    atlasItem.Node.H = atlasItem.Height
  }
  log.Println("Scaled Atlas numbers by", v)
}

// FontAtlasFromRunes returns a Atlas of a given font for a given slice of runes.
func New(fontFileName string, fontPt, fontRes float64, imgWidth, imgHeight, pad int, runes []rune) Atlas {
  // create atlas
  var atlas Atlas
  atlas.FontFile = fontFileName
  atlas.FontPt = fontPt
  atlas.FontRes = fontRes
  atlas.ReloadFont()
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
    // ^ not sure if tiny middle add is needed
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
  for atlas.ContainsNilNodes() {
    // create new atlas image sheet
    atlas.Images = append(atlas.Images, image.NewGray(image.Rect(0, 0, imgWidth, imgHeight)))
    imageIndex := len(atlas.Images) - 1
    draw.Draw(atlas.Images[imageIndex], atlas.Images[imageIndex].Bounds(), image.Black, image.Point{}, draw.Src)
    
    // sort nil nodes per AtlasItems sort implementation
    itemSlice := atlas.GetNilNodes()
    sort.Sort(itemSlice)
    
    // give each rune a position within an image sheet
    // if it doesn't fit on current sheet, node remains nil
    FitAtlasItems(itemSlice, imgWidth, imgHeight)
    
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
