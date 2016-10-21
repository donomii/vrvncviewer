// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux windows

// An app that draws a green triangle on a red background.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool.
//
// See http://godoc.org/golang.org/x/mobile/cmd/gomobile to install gomobile.
//
// Get the basic example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/basic
//   $ gomobile build golang.org/x/mobile/example/basic # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/basic
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/basic && basic
package main

import "github.com/pkg/profile"
import "image/color"
import (
    "strings"
    "net"
    "errors"
    "encoding/binary"
    "log"
    "runtime"

    "golang.org/x/mobile/app"
    "golang.org/x/mobile/event/lifecycle"
    "golang.org/x/mobile/event/paint"
    "golang.org/x/mobile/event/size"
    "golang.org/x/mobile/event/touch"
    "golang.org/x/mobile/exp/app/debug"
    "golang.org/x/mobile/exp/f32"
    "golang.org/x/mobile/exp/gl/glutil"
    "golang.org/x/mobile/gl"
    "fmt"
    "os"
    "time"
    "image"
    //"math/rand"
    _ "image/png"
    "github.com/donomii/sceneCamera"
)
import "github.com/go-gl/mathgl/mgl32"
        import "golang.org/x/mobile/exp/sensor"

var clientWidth=uint(1024)
var clientHeight=uint(768)
var u8Pix []uint8
var (
    startDrawing bool
    imageData image.Image
    imageBounds image.Rectangle
    images   *glutil.Images
    fps      *debug.FPS
    program  gl.Program
    position gl.Attrib
    u_Texture gl.Uniform
    a_TexCoordinate gl.Attrib
    colour gl.Attrib
    buf      gl.Buffer
    tbuf      gl.Buffer
    
    screenWidth int
    screenHeight int

    green  float32
    red  float32
    blue  float32
    touchX float32
    touchY float32
    selection int
    gallery []string
    reCalcNeeded bool
    prevTime int64
)

var scanOn = true
var vMeta map[string]vertexMeta
var triBuff[]byte
var vTrisf map[string][]float32
var vBuffs map[string]gl.Buffer

var vCols map[string][]byte
var vColsf map[string][]float32
var vColBuffs map[string]gl.Buffer

var trans  mgl32.Mat4
var theatreCamera  mgl32.Mat4
var transU gl.Uniform
var recursion int = 4
var threeD bool = false
var polyCount int
var clock float32 = 0.0
var Tex gl.Texture
var sceneCam *sceneCamera.SceneCamera

var viewAngle [3]float32


var texAlignData = f32.Bytes(binary.LittleEndian,
    0.0, 0.0, // top left
    0.0, 1.0, // top left
    1.0, 0.0, // top left
    0.0, 1.0, // top left
    1.0, 1.0, // top left
    1.0, 0.0, // top left
)


var triangleData = f32.Bytes(binary.LittleEndian,
    -1.0, 1.0, 0.0, // top lef
    -1.0, -1.0, 0.0, // bottom left
    1.0, 1.0, 0.0, // bottom right
    -1.0, -1.0, 0.0, // bottom right
    1.0, -1.0, 0.0, // top left
    1.0, 1.0, 0.0, // bottom right
//
    //0.0, 1.0, 0.0, // top left
    //0.0, 0.0, 0.0, // bottom left
    //0.0, 0.0, 0.2, // bottom right
    //0.0, 1.0, 0.2, // top right
    //0.0, 1.0, 0.0, // top left
    //0.0, 0.0, 0.2, // bottom right
)

var texData = f32.Bytes(binary.LittleEndian,
    0.0, 1.0, 1.0, // top left
    1.0, 0.0, 0.0, // bottom left
    1.0, 0.0, 0.0, // bottom right
    0.0, 0.0, 1.0, // bottom right
    1.0, 0.0, 0.0, // top left
    0.0, 0.0, 0.0, // bottom right
//
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.0, // bottom left
    0.0, 0.0, 1.0, // bottom right
    0.0, 1.0, 1.0, // top right
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.9, // bottom right
    0.0, 0.0, 0.9, // bottom right
    0.0, 1.0, 0.9, // top right
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.9, // bottom right
)


func do_profile() {
    defer profile.Start(profile.MemProfile).Stop()
    //defer profile.Start(profile.TraceProfile).Stop()
    //defer profile.Start(profile.CPUProfile).Stop()
    time.Sleep(60*time.Second)
}

func main() {
    log.Printf("Starting main...")
    sceneCam = sceneCamera.New()
    runtime.GOMAXPROCS(2)
    app.Main(func(a app.App) {
        log.Printf("Starting app...")
        reCalcNeeded = true
        var glctx gl.Context
        var sz size.Event
        sensor.Notify(a)
        theatreCamera = mgl32.Ident4()
        trans = mgl32.Ident4()
        trans = trans.Mul4(mgl32.Translate3D(0.0, 0.0, 1.0))
        if threeD {
            trans = compose(trans, mgl32.Scale3D(1.6, 0.6,1.0))
        }
        theatreCamera = mgl32.LookAt(0.0, 0.0, 0.6, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0)
        for e := range a.Events() {
            switch e := a.Filter(e).(type) {
            case sensor.Event:
                  delta := e.Timestamp - prevTime
                  prevTime = e.Timestamp
                  scale := float32(36000000.0/float32(delta))
                  sceneCam.ProcessEvent(e)


                  var sora_vec mgl32.Vec3   //The real sora
                  sora_vec = mgl32.Vec3{float32(e.Data[1])/scale, -float32(e.Data[0])/scale,float32(-e.Data[2])/scale/float32(3.14)}

                  if threeD {
                  } else {
                      theatreCamera = theatreCamera.Mul4(mgl32.Translate3D(sora_vec[1]/scale, -sora_vec[0]/scale, 0.0))
                  }
            case lifecycle.Event:
                switch e.Crosses(lifecycle.StageVisible) {
                case lifecycle.CrossOn:
                    glctx, _ = e.DrawContext.(gl.Context)
                    onStart(glctx)
                    sensor.Enable(sensor.Gyroscope, 10 * time.Millisecond)
                    a.Send(paint.Event{})
                case lifecycle.CrossOff:
                    sensor.Disable(sensor.Gyroscope)
                    onStop(glctx)
                    glctx = nil
                }
            case size.Event:
                sz = e
                reCalcNeeded = true
                screenWidth = sz.WidthPx
                screenHeight = sz.HeightPx
                touchX = float32(sz.WidthPx /2)
                touchY = float32(sz.HeightPx * 9/10)
                if (sz.Orientation == size.OrientationLandscape) {
                    //threeD = true
                } else {
                    threeD = false
                }
            case paint.Event:
                if glctx == nil || e.External {
                    // As we are actively painting as fast as
                    // we can (usually 60 FPS), skip any paint
                    // events sent by the system.
                    continue
                }

                onPaint(glctx, sz)
                a.Publish()
                // Drive the animation by preparing to paint the next frame
                // after this one is shown.
                time.Sleep(100 * time.Millisecond)
                a.Send(paint.Event{})
            case touch.Event:
                theatreCamera = mgl32.LookAt(0.0, 0.0, 0.1, 0.0, 0.0, -0.5, 0.0, 1.0, 0.0)
                if e.Type == touch.TypeBegin {
                    reCalcNeeded = true
                    selection++
                    if selection +1  > len(gallery) {
                        selection=0
                    }
                }
            }
        }
    })
}

var connectCh chan bool


func externalIP() (string, error) {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // interface down
        }
        if iface.Flags&net.FlagLoopback != 0 {
            continue // loopback interface
        }
        addrs, err := iface.Addrs()
        if err != nil {
            return "", err
        }
        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            if ip == nil || ip.IsLoopback() {
                continue
            }
            ip = ip.To4()
            if ip == nil {
                continue // not an ipv4 address
            }
            return ip.String(), nil
        }
    }
    return "", errors.New("are you connected to the network?")
}

func scanHosts() {
    connectCh = make(chan bool, 30)
    for i:=1; i<30; i++ {
        connectCh <- true
    }
    ip, _ := externalIP()
    ip_chunks := strings.Split(ip, ".")
    classC := strings.Join(ip_chunks[:3], ".")
    //log.Printf("IP: %v\n", classC)
    for j:=1;j<255;j++ {
        if scanOn {
            pasteText(50.0, 0, ip_chunks[0], u8Pix, false)
            pasteText(50.0, 64, ip_chunks[1], u8Pix, false)
            pasteText(50.0, 128, ip_chunks[2], u8Pix, false)
            pasteText(50.0, 192, fmt.Sprintf("%v", j), u8Pix, false)
            testIP := fmt.Sprintf("%v.%v", classC, j)
            //log.Printf("testIP: %v\n", testIP)
            <-connectCh
            //fmt.Printf("%v:5900\n",testIP)
            go run_vnc(fmt.Sprintf("%v:5900",testIP))
            <-connectCh
            //fmt.Printf("http://%v:8080/\n",testIP)
            go http_mjpeg(fmt.Sprintf("http://%v:8080/",testIP))
        }
    }
    time.Sleep(500*time.Millisecond)
    //fmt.Println("Finished scan")
    scanHosts()
}

func onStart(glctx gl.Context) {
    log.Printf("Onstart callback...")
    dim := clientWidth*clientHeight*4
    u8Pix = make([]uint8, dim, dim)
    go scanHosts()
    fmt.Printf("Waiting on connnection\n")
    //for {
        //time.Sleep(50 * time.Millisecond)
        //if startDrawing {
            //break
        //}
    //}
    //time.Sleep(3 * time.Second)
    var err error
    program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
    if err != nil {
        log.Printf("error creating GL program: %v", err)
        os.Exit(1)
        return
    }


    position = glctx.GetAttribLocation(program, "position")
    a_TexCoordinate = glctx.GetAttribLocation(program, "a_TexCoordinate")
    transU = glctx.GetUniformLocation(program, "transform")
    u_Texture = glctx.GetUniformLocation(program, "u_Texture")
    fmt.Println("Creating buffers")

    buf = glctx.CreateBuffer()
    glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
    fmt.Printf("triangleData: %V\n", triangleData)
    glctx.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

    tbuf = glctx.CreateBuffer()
    glctx.BindBuffer(gl.ARRAY_BUFFER, tbuf)
    fmt.Printf("texAlignData: %V\n", texAlignData)
    glctx.BufferData(gl.ARRAY_BUFFER, texAlignData, gl.STATIC_DRAW)


    Tex = glctx.CreateTexture()
    glctx.BindTexture(gl.TEXTURE_2D, Tex)

    glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);


    //images = glutil.NewImages(glctx)
    //fps = debug.NewFPS(images)
}

func onStop(glctx gl.Context) {
    log.Printf("Stopping...")
    //glctx.DeleteProgram(program)
    //glctx.DeleteBuffer(buf)
    fps.Release()
    images.Release()
}

func transpose( m mgl32.Mat4) mgl32.Mat4{
    var r mgl32.Mat4
    for i, v := range []int{0,4,8,12,1,5,9,13,2,6,10,14,3,7,11,15} {
        r[i] = m[v]
    }
    //fmt.Println(r)
    return r
}

func pasteText(tSize float64, ypos int, text string, u8Pix []uint8, transparent bool) {
    img := DrawStringRGBA(50.0, color.RGBA{255,255,255,255}, text)
    po2 := uint(NextPo2(img.Bounds().Max.X)*2)
    //log.Printf("Chose texture size: %v\n", po2)
    wordBuff := paintTexture (img, nil, po2)
    startDrawing = true
    bpp := uint(4)  //bytes per pixel

    h:= img.Bounds().Max.Y
    w:= uint(img.Bounds().Max.X)
    for i:=uint(0);i<uint(h); i++ {
        for j := uint(0);j<w; j++ {
            if (wordBuff[i*po2*4 + j*4]>128) || !transparent {
                u8Pix[(uint(ypos)+i)*clientWidth*bpp+j*bpp] = wordBuff[i*po2*4 + j*4]
                u8Pix[(uint(ypos)+i)*clientWidth*bpp+j*bpp +1] = wordBuff[i*po2*4 + j*4 +1]
                u8Pix[(uint(ypos)+i)*clientWidth*bpp+j*bpp +2] = wordBuff[i*po2*4 + j*4 +2]
                u8Pix[(uint(ypos)+i)*clientWidth*bpp+j*bpp +3] = wordBuff[i*po2*4 + j*4 +3]
            }
        }
    }
}



func onPaint(glctx gl.Context, sz size.Event) {
    glctx.Enable(gl.BLEND)
    glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
    glctx.Enable( gl.DEPTH_TEST );
    glctx.DepthFunc( gl.LEQUAL );
    glctx.DepthMask(true)
    glctx.Clear(gl.COLOR_BUFFER_BIT|gl.DEPTH_BUFFER_BIT)
    glctx.ClearColor(0, 0, 0, 1)
    glctx.UseProgram(program)


    
    glctx.TexImage2D(gl.TEXTURE_2D, 0, int(clientWidth), int(clientHeight), gl.RGBA, gl.UNSIGNED_BYTE, u8Pix)


    var view mgl32.Mat4
    if threeD {
        view = compose3(mgl32.Perspective(55, float32(screenWidth)/float32(screenHeight), 0.1, 2048.0), sceneCam.ViewMatrix(), trans)
    } else {
        view = compose(theatreCamera, trans)
    }
    glctx.UniformMatrix4fv(transU, view[0:16])

    glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
    glctx.EnableVertexAttribArray(position)
    glctx.VertexAttribPointer(position, 3, gl.FLOAT, false, 0, 0)


    glctx.BindBuffer(gl.ARRAY_BUFFER, tbuf)
    glctx.EnableVertexAttribArray(a_TexCoordinate)
    glctx.VertexAttribPointer(a_TexCoordinate, 2, gl.FLOAT, false, 0, 0)

    glctx.ActiveTexture(gl.TEXTURE0);
    // Bind the texture to this unit.
    glctx.BindTexture(gl.TEXTURE_2D, Tex);
    // Tell the texture uniform sampler to use this texture in the shader by binding to texture unit 0.
    glctx.Uniform1i(u_Texture, 0);

    glctx.Viewport(0,0, sz.WidthPx/2, sz.HeightPx)
    glctx.DrawArrays(gl.TRIANGLES, 0, 6)
    glctx.Viewport(sz.WidthPx/2,0, sz.WidthPx/2, sz.HeightPx)
    glctx.DrawArrays(gl.TRIANGLES, 0, 6)
    glctx.DisableVertexAttribArray(position)
}

type vertexMeta struct {
    coordsPerVertex int
    vertexCount     int
}

const (
    coordsPerVertex = 3
    vertexCount     = 3
)


const vertexShader = `#version 100 
uniform mat4 transform;

attribute vec2 a_TexCoordinate; // Per-vertex texture coordinate information we will pass in.
varying vec2 v_TexCoordinate;   // This will be passed into the fragment shader.

    attribute vec4 position;
    varying vec4 color;
    void main() {
        gl_Position = transform * position;
        color = vec4(1.0,1.0,1.0,1.0);
        // Pass through the texture coordinate.
        v_TexCoordinate = a_TexCoordinate;
    }
`

const fragmentShader = `#version 100
precision mediump float;
varying vec4 color;
uniform sampler2D u_Texture;    // The input texture.
varying vec2 v_TexCoordinate; // Interpolated texture coordinate per fragment.
void main() {
    //gl_FragColor = color;
    gl_FragColor = texture2D(u_Texture, v_TexCoordinate);
}`



func compose (a, b mgl32.Mat4) mgl32.Mat4 {
return a.Mul4(b)
}

func compose3 (a, b, c mgl32.Mat4) mgl32.Mat4 {
    t := b.Mul4(c)
return a.Mul4(t)
}

func checkGlErr(glctx gl.Context) {
    err := glctx.GetError()
    if (err>0) {
        fmt.Printf("GLerror: %v\n", err)
        panic("GLERROR")
    }
}
