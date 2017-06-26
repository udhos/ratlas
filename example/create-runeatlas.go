package main

import (
  "fmt"
  "io/ioutil"
  "github.com/vrav/ratlas"
)

func main() {
  runes := []rune(" ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890`~!@#$%^&*()[]{}/=?+\\|-_.>,<'\";:�")
  fontFile := "Vera.ttf"
  
  // load font file bytes
  ttfData, err := ioutil.ReadFile(fontFile)
  if err != nil {
    panic(err)
  }
  
  atlas := ratlas.New(&ttfData, 288.0, 2048, 2048, 16, runes)
  atlas.SaveGobFile(fmt.Sprintf("%s.gob", fontFile))
  atlas.SaveImageFiles()
}
