// textures
package main

import (
    sysFont "golang.org/x/mobile/exp/font"
    "io/ioutil"
    "bytes"
    "golang.org/x/image/font"
    "golang.org/x/image/math/fixed"
"github.com/golang/freetype/truetype"
	"fmt"
	"log"
	"os"

	"github.com/go-gl/mathgl/mgl32"

	"golang.org/x/mobile/gl"

	"image"
	"image/color"
	"image/png"
)

type Thunk func()

var (
	rtt_frameBuff gl.Framebuffer
	rtt_tex       gl.Texture
)

func screenShot (glctx gl.Context, filename string) {
    log.Printf("Saving width: %v, height: %v\n",screenWidth, screenHeight)
    saveBuff(uint(screenWidth),uint(screenHeight), copyScreen(glctx, int(screenWidth),int(screenHeight)), filename)
}

//Copies an image to a correctly-packed texture data array.  
//
//Returns the array, modified in place.  If u8Pix is nil or texWidth is 0, it creates a new texture array and returns that.  Texture is assumed to be square.
func paintTexture (img image.Image, u8Pix []uint8, clientWidth uint) []uint8 {
    bounds := img.Bounds()
    newW := bounds.Max.X
    newH := bounds.Max.Y

    //if uint(newW) != clientWidth || uint(newH) != clientWidth {
    if (uint(newW) > clientWidth) || (uint(newH) > clientWidth) {
        panic(fmt.Sprintf("ClientWidth (%v) does not match image width(%v) and height(%v)", clientWidth, newW, newH))
    }
    if u8Pix == nil {
        dim := clientWidth*clientWidth*4 +4
        u8Pix = make([]uint8, dim, dim)
    }

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            r, g, b, a := img.At(x, y).RGBA()
            // A color's RGBA method returns values in the range [0, 65535].
            start := uint(y)*clientWidth*4 + uint(x)*4
            u8Pix[start]   = uint8(r*255/65535)
            u8Pix[start+1] = uint8(g*255/65535)
            u8Pix[start+2] = uint8(b*255/65535)
            u8Pix[start+3] = uint8(a*255/65535)
        }
    }
    return u8Pix
}

func copyScreen(glctx gl.Context, clientWidth, clientHeight int) []byte {
	buff := make([]byte, clientWidth*clientHeight*4, clientWidth*clientHeight*4)
	//fmt.Printf("reading pixels: ")
	//glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.ReadPixels(buff, 0, 0, clientWidth, clientHeight, gl.RGBA, gl.UNSIGNED_BYTE)
	checkGlErr(glctx)
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
    return buff
}

func copyFrameBuff(glctx gl.Context, clientWidth, clientHeight int) []byte {
	buff := make([]byte, clientWidth*clientHeight*4, clientWidth*clientHeight*4)
	//fmt.Printf("reading pixels: ")
	glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.ReadPixels(buff, 0, 0, clientWidth, clientHeight, gl.RGBA, gl.UNSIGNED_BYTE)
	checkGlErr(glctx)
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
    return buff
}



func saveImage(m *image.RGBA, filename string) {
	f, _ := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	png.Encode(f, m)
}

func saveBuff(texWidth, texHeight uint, buff []byte, filename string) {
	f, _ := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	m := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(texWidth), int(texHeight)}})
    if buff != nil {
		//fmt.Printf("readpixels: %V", buff)
		for y := uint(0); y < texWidth; y++ {
			for x := uint(0); x < texHeight; x++ {
				i := (x + y*texWidth) * 4
				m.Set(int(x), int(texHeight-y), color.NRGBA{uint8(buff[i]), uint8(buff[i+1]), uint8(buff[i+2]), 255})
				//if buff[i]>0 { fmt.Printf("Found colour\n") }
				//if buff[i+1]>0 { fmt.Printf("Found colour\n") }
				//if buff[i+2]>0 { fmt.Printf("Found colour\n") }
			}
		}
	}
	png.Encode(f, m)
}

func Rtt(glctx gl.Context, texWidth, texHeight int, thunk Thunk) {
	glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.Viewport(0, 0, texWidth, texHeight)
	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_2D, rtt_tex)
	//draw here the content you want in the texture
	log.Printf("Framebuffer status: %v\n", glctx.CheckFramebufferStatus(gl.FRAMEBUFFER))
	trans = mgl32.Ident4()

	//rtt_tex is now a texture with the drawn content

	glctx.Enable(gl.BLEND)
	glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	glctx.Enable(gl.DEPTH_TEST)
	glctx.DepthFunc(gl.LEQUAL)
	glctx.DepthMask(true)
	glctx.ClearColor(0, 0, 0, 1)
	glctx.UseProgram(program) //FIXME - may cause graphics glitches
	polyCount = 0
	clock += 0.001
	glctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
    thunk()

	glctx.Flush()

    buff := copyFrameBuff(glctx, texWidth, texHeight)
    saveBuff(uint(texWidth), uint(texHeight), buff, "x.png")
    glctx.BindTexture(gl.TEXTURE_2D, gl.Texture{0})
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	log.Println("Finished Rtt")
	fmt.Printf("done \n")
}

func dumpBuff(buff []uint8, size uint) {
    log.Printf("Dumping buffer with width %v\n", size)
    for y := uint(0); y < size; y++ {
        for x := uint(0); x < size; x++ {
            i := (x + y*size) * 4
            //log.Printf("Index: %v\n", i)
            if buff[i]>128 {
                fmt.Printf("I")
            } else {
                fmt.Printf("_")
            }
        }
        fmt.Println("")
    }
}

func string2Tex(glctx gl.Context, str string, tSize float64, glTex gl.Texture) {
    img := DrawStringRGBA(tSize, color.RGBA{255,255,255,255}, str)
    saveImage(img, "texttest.png")
    w:= 128 //img.Bounds().Max.X
    buff := paintTexture (img, nil, uint(w))
    dumpBuff(buff, uint(w))
    uploadTex(glctx, glTex, w, w, buff)
}

func uploadTex(glctx gl.Context, glTex gl.Texture, w,h int,  buff []uint8) {
	glctx.BindTexture(gl.TEXTURE_2D, glTex)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	glctx.TexImage2D(gl.TEXTURE_2D, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, buff)
    glctx.GenerateMipmap(gl.TEXTURE_2D);
}


//Creates a new framebuffer and texture, with the texture attached to the frame buffer
func glGenTextureFromFramebuffer(glctx gl.Context, w, h int) (gl.Framebuffer, gl.Texture) {
	f := glctx.CreateFramebuffer()
	glctx.BindFramebuffer(gl.FRAMEBUFFER, f)
	glctx.ActiveTexture(gl.TEXTURE0)
	t := glctx.CreateTexture()
	log.Printf("Texture created: %v", t)

	glctx.BindTexture(gl.TEXTURE_2D, t)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	glctx.TexImage2D(gl.TEXTURE_2D, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	glctx.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, t, 0)

	/*
	   depthbuffer := glctx.CreateRenderbuffer()
	   glctx.BindRenderbuffer(gl.RENDERBUFFER, depthbuffer)
	   glctx.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, w, h)
	   glctx.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthbuffer)
	*/

	status := glctx.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		fmt.Printf("Framebuffer status: %v\n", status)
		os.Exit(1)
	}
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	return f, t
}

func DrawStringRGBA(txtSize float64, fontColor color.RGBA, txt string) *image.RGBA {

    txtFont := LoadGameFont("f1.ttf")
    d := &font.Drawer{
        Src: image.NewUniform(fontColor), // 字体颜色
        Face: truetype.NewFace(txtFont, &truetype.Options{
            Size:    txtSize,
            DPI:     72,
            Hinting: font.HintingNone,
        }),
    }
    re := d.MeasureString(txt)
    rect := image.Rect(0, 0, int((re + 0x3f) >> 6), int(txtSize))
    rgba := image.NewRGBA(rect)
    d.Dst = rgba

    d.Dot = fixed.Point26_6{
        X: fixed.I(0),
        Y: fixed.I(rect.Max.Y),
    }
    d.DrawString(txt)
    return rgba
}

func LoadGameFont(fileName string) *truetype.Font {

        fontBytes := sysFont.Monospace()
        f := bytes.NewReader(fontBytes)
        fontBytes, err := ioutil.ReadAll(f)
        if err != nil {
            log.Println(err)
            panic(err)
        }

        txtFont, err1 := truetype.Parse(fontBytes)
        if err1 != nil {
            log.Println(err1)
            panic(err1)
        }
    return txtFont
}
