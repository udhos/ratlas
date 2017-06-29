## fontdraw-simple
This short program opens an OpenGL window using GLFW and draws some example text within the window. In this case, the rune atlas was saved to a gob, and its image was converted to a SDF representation (which allows for cleaner scaling given one input font size). Because the source atlas was generated at 4X size, note the call to ScaleNumbers(). To create an SDF atlas yourself, see package [isdf](https://github.com/vrav/isdf).

In action:

![img1](http://i.imgur.com/8dxkNgz.png)

## fontdraw-immediate
This slightly more robust example shows immediate mode drawing, where the mesh used to render text is generated every frame and streamed to the GPU. It's more costly than retained mode rendering, but it's easy to work with. The text has been animated to show off one potential usability advantage to regenerating the text mesh every frame.

In action:

![img1](http://i.imgur.com/JNAPaXg.png)
