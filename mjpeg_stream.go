package main
import "mime/multipart"
import (
    //"github.com/donomii/glim"
    //"runtime/debug"
    "bytes"
    "log"
    "image/jpeg"
    "mime"
    "strings"
    "sync"
    "time"
    "fmt"
    //mjpeg "github.com/marpie/go-mjpeg"
    "image"
    "io"
    "net/http"
    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"
    "golang.org/x/image/math/fixed"
    "image/color"
    "image/draw"
)

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, buf *bytes.Buffer) (err error) {
    // If the buffer overflows, we will get bytes.ErrTooLarge.
    // Return that as an error. Any other panic remains.
    defer func() {
        e := recover()
        if e == nil {
            err = nil
            return
        }
        if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
            err = panicErr
        } else {
            panic(e)
        }
    }()
    _, err = buf.ReadFrom(r)
    return
}

// processHttp receives the HTTP data and tries to decodes images. The images 
// are sent through a chan for further processing.
func processHttp(response *http.Response, nextChunk, freeChunks chan *bytes.Buffer, nextImg chan *image.Image, quit chan bool) {
    decoder,_ := NewDecoderFromResponse(response)
    defer response.Body.Close()
    defer close(nextImg)
    scanOn=false
    for {
        select {
        case <-quit:
            close(nextImg)
            scanOn=true
            return
        default:
            buf := <- freeChunks
            buf.Truncate(0)
            p, err := decoder.r.NextPart()
            if err == io.EOF {
                close(nextImg)
                scanOn=true
                return
            }
            if err != nil {
                log.Fatal(err)
            }
            err = readAll(p, buf)
            if err != nil {
                //log.Printf("%v", err)
                close(nextImg)
                scanOn=true
                return
            }
            if (len(freeChunks)>0) {
                nextChunk <- buf
            } else {
                log.Printf("Skip!\n")
                freeChunks <- buf
            }
            //glim.PasteText(50.0, 1, 1, int(clientWidth), int(clientHeight), fmt.Sprintf("%v", FPS), u8Pix, false)
        }
    }
}


// processHttp receives the HTTP data and tries to decodes images. The images 
// are sent through a chan for further processing.
func processChunk(nextChunk, freeChunks chan *bytes.Buffer, nextImg chan *image.Image, quit chan bool) {
    defer close(nextImg)
    scanOn=false
    for {
        //var stats debug.GCStats
        //debug.ReadGCStats(&stats)
        //log.Println(stats)
        select {
        case <-quit:
            close(nextImg)
            scanOn=true
            return
        default:
            buf := <- nextChunk
            //Discard incoming frames if there are already some frames queued
            if len(nextImg) == 0 {
                //img, err := mjpeg.Decode(response.Body)
                img, err := jpeg.Decode(bytes.NewReader(buf.Bytes()))

                if err == io.EOF {
                    close(nextImg)
                    scanOn=true
                    return
                }
                if err != nil {
                    log.Println(err)
                }
                if img != nil {
                    nextImg <- &img
                }
                lockScreen = false
                freeChunks <- buf
            }
        }
    }
}

func addLabel(img *image.Image, x, y int, label string) {
    col := color.RGBA{200, 100, 0, 255}
    point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

    im := *img
    d := &font.Drawer{
        Dst:  im.(draw.Image),
        Src:  image.NewUniform(col),
        Face: basicfont.Face7x13,
        Dot:  point,
    }
    d.DrawString(label)
}

func NewRGBA(width,height int) *image.RGBA {
    return image.NewRGBA(image.Rectangle{image.Point{0,0},image.Point{width,height}})
}

// processImage receives images through a chan, decodes them an updates the texture
func processImage(nextImg chan *image.Image, quit chan bool) {
    rgba := NewRGBA(int(clientWidth), int(clientHeight))
    for {
        scanOn=false
        //runtime.GC()
        i, ok := <-nextImg

        //addLabel(i, 100, 100, "HELLO")

        if !ok {
            break
        }
        if *i == nil {
            break
        }
        img := *i
        //fmt.Println("New Image:", img.Bounds())
        bounds := img.Bounds()
        newW := bounds.Max.X
        newH := bounds.Max.Y
        if newW != int(clientWidth) || newH != int(clientHeight) {
            clientWidth = uint(newW)
            clientHeight = uint(newH)
            fmt.Printf("Chose new width: %v, height %v\n", clientWidth, clientHeight)
            rgba = NewRGBA(int(clientWidth), int(clientHeight))
            dim := clientWidth*clientHeight*4
            u8Pix = make([]uint8, dim, dim)
        }
        //The graphics buffers are ready, we can start using them, even if they are blank
        startDrawing = true
        rect := img.Bounds()
        draw.Draw(rgba, rect, img, rect.Min, draw.Src)
        u8Pix = rgba.Pix

        //glim.RenderPara(&activeFormatter, 240,0, int(clientWidth), int(clientHeight), int(clientWidth), int(clientHeight), 0,0, u8Pix, "Connected", true, true, false)
    }
    scanOn=true
    quit <- true
}


type Decoder struct {
    r *multipart.Reader
    m sync.Mutex
}

func NewDecoder(r io.Reader, b string) *Decoder {
    d := new(Decoder)
    d.r = multipart.NewReader(r, b)
    return d
}

func NewDecoderFromResponse(res *http.Response) (*Decoder, error) {
    _, param, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
    if err != nil {
        return nil, err
    }
    return NewDecoder(res.Body, strings.Trim(param["boundary"], "-")), nil
}

func http_mjpeg(URL string, timeout int) {
    //fmt.Printf("Opening %v\n", URL)
    client := http.Client{
        Timeout: time.Duration(timeout) * time.Duration(time.Millisecond),
    }
    response, err := client.Get(URL)
    if err != nil {
        //fmt.Printf("Failed to open %v\n", URL)
        connectCh <- true
        return
    }
    fmt.Printf("Passed quick check %v\n", URL)
    response, err = http.Get(URL)
    fmt.Printf("Connected to %v\n", URL)
    nextImg := make(chan *image.Image, 0)
    nextChunk := make(chan *bytes.Buffer, 0)
    freeChunks := make(chan *bytes.Buffer, 10)
    quit := make(chan bool)
    fmt.Println("Waiting for stream...")
    go processImage(nextImg, quit)
    go processChunk(nextChunk, freeChunks, nextImg, quit)
    go processHttp(response, nextChunk, freeChunks, nextImg, quit)
    freeChunks <-  bytes.NewBuffer(make([]byte, 0, 1024*1024*10))
    freeChunks <-  bytes.NewBuffer(make([]byte, 0, 1024*1024*10))
    freeChunks <-  bytes.NewBuffer(make([]byte, 0, 1024*1024*10))
    _ = <-quit
    scanOn=true
}
