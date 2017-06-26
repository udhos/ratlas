# ratlas
Golang text glyph/rune atlas image generator.

Generates rune atlas images given a TTF font.

![img1](http://i.imgur.com/QMN1bIV.png)

## Usage Examples
To build the examples, enter the example directory and type:
```
go build create-runeatlas.go
go build fontdraw-simple.go
```
See the source of each sample application for detailed example usage.

The pattern for the atlas creation function is as follows:
```
func New(ttfData *[]byte, fontPt float64, imgWidth, imgHeight, pad int, runes []rune) Atlas
```
This will create a ratlas.Atlas from the given font bytedata, at the given font size, on images of the specified dimensions, using the runes specified.

## License

MIT, see [LICENSE.md](http://github.com/vrav/isdf/blob/master/LICENSE.md) for details.
