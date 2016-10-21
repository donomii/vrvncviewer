package main
import "mime/multipart"
import (
    "bytes"
    "io/ioutil"
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



// processHttp receives the HTTP data and tries to decodes images. The images 
// are sent through a chan for further processing.
func processHttp(response *http.Response, nextImg chan *image.Image, quit chan bool) {
    decoder,_ := NewDecoderFromResponse(response)
    defer response.Body.Close()
    for {
        scanOn=false
        select {
        case <-quit:
            close(nextImg)
            scanOn=true
            return
        default:
            //Discard incoming frames if there are already some frames queued
            p, err := decoder.r.NextPart()
            if err == io.EOF {
                return
            }
            if err != nil {
                log.Fatal(err)
            }
            slurp, err := ioutil.ReadAll(p)
            if err != nil {
                log.Fatal(err)
            }
            if len(nextImg) == 0 {
                //img, err := mjpeg.Decode(response.Body)
                img, err := jpeg.Decode(bytes.NewReader(slurp))
                //log.Printf("Unpacked jpeg image type: %v\n", img.(type))
switch img.(type) {
case *image.RGBA:
                log.Printf("Unpacked jpeg image type: RGBA\n")
    // i in an *image.RGBA
case *image.NRGBA:
                log.Printf("Unpacked jpeg image type: NRGBA\n")
    // i in an *image.NRBGA
}

                if err == io.EOF {
                    close(nextImg)
                    scanOn=true
                    return
                }
                if err != nil {
                    fmt.Println(err)
                }
                if img != nil {
                    nextImg <- &img
                }
            } else {
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
    return image.NewRGBA(image.Rectangle{image.Point{0,0},image.Point{800,600}})
}

// processImage receives images through a chan, decodes them an updates the texture
func processImage(nextImg chan *image.Image, quit chan bool) {
    rgba := NewRGBA(int(clientWidth), int(clientHeight))
    for {
        scanOn=false
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
            //dim := clientWidth*clientHeight*4
            //u8Pix = make([]uint8, dim, dim)
        }
        //The graphics buffers are ready, we can start using them, even if they are blank
        startDrawing = true
        rect := img.Bounds()
        draw.Draw(rgba, rect, img, rect.Min, draw.Src)
        u8Pix = rgba.Pix

        /*for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
            for x := bounds.Min.X; x < bounds.Max.X; x++ {
                r, g, b, _ := img.At(x, y).RGBA()
                // A color's RGBA method returns values in the range [0, 65535].
                start := uint(y)*clientWidth*3 + uint(x)*3
                u8Pix[start] = uint8(r*255/65535)
                u8Pix[start+1] = uint8(g*255/65535)
                u8Pix[start+2] = uint8(b*255/65535)
        }
    }*/
    //Add some kind of flashing thing to the texture so we can see that the link is still active
    //fmt.Println("Looping")
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

func http_mjpeg(URL string) {
    //fmt.Printf("Opening %v\n", URL)
    timeout := time.Duration(2000 * time.Millisecond)
    client := http.Client{
        Timeout: timeout,
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
    nextImg := make(chan *image.Image, 30)
    quit := make(chan bool)
    fmt.Println("Waiting for stream...")
    go processImage(nextImg, quit)
    go processHttp(response, nextImg, quit)
    _ = <-quit
    scanOn=true
}
