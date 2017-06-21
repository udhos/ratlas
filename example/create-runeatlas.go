package main

import (
  "fmt"
  "github.com/vrav/ratlas"
)

func main() {
  runes := []rune(" ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890`~!@#$%^&*()[]{}/=?+\\|-_.>,<'\";:�")
  fontName := "Vera.ttf"
  atlas := ratlas.New(fontName, 288.0, 72.0, 2048, 2048, 16, runes)
  atlas.SaveGobFile(fmt.Sprintf("%s.gob", fontName))
  atlas.SaveImageFiles()
}
