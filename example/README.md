## fontdraw-simple
This short program opens an OpenGL window using GLFW and draws some example text within the window. In this case, the rune atlas was saved to a gob, and its image was converted to a SDF representation (which allows for cleaner scaling given one input font size). Because the source atlas was generated at 4X size, note the call to ScaleNumbers().

In action:

![img1](http://i.imgur.com/8dxkNgz.png)
