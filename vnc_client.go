package main

import "image"
import "fmt"
import "github.com/donomii/go-vnc"
//import "github.com/donomii/glim"
import "net"
import "time"
import "log"
import "golang.org/x/net/context"

var im *image.NRGBA


func run_vnc (server_port string, timeout int) {
    // Establish TCP connection to VNC server.
    nc, err := net.DialTimeout("tcp", server_port, time.Duration(timeout)*time.Millisecond)
    if err != nil {
      //log.Fatalf("Error connecting to VNC host. %v", err)
        connectCh <- true
        return
    }
    //scanOn = false

    // Negotiate connection with the server.
    vcc := vnc.NewClientConfig("aaaaaaaa")
    vc, err := vnc.Connect(context.Background(), nc, vcc)
    log.Println("Connected to server",server_port)
    //go requestUpdates(vc)
    if err != nil {
        log.Printf("Error: %s : %v", server_port, err)
        status = fmt.Sprintf("!! %s : %v !!", server_port, err)
    } else {
        if !(clientWidth == uint(vc.FramebufferWidth())) || !(clientHeight == uint(vc.FramebufferHeight())) {
            clientWidth, clientHeight = uint(vc.FramebufferWidth()), uint(vc.FramebufferHeight())
            log.Printf("Resizing to (%v,%v)", clientWidth, clientHeight)
            im = image.NewNRGBA(image.Rectangle{image.Point{0,0}, image.Point{int(clientWidth),int(clientHeight)}})
            dim := clientWidth*clientHeight*4
            u8Pix = make([]uint8, dim, dim)
        }
        //log.Printf("Using size (%v,%v)", clientWidth, clientHeight)

        //go do_profile()
        // Listen and handle server messages.
        go vc.ListenAndHandle()
        //The graphics buffers are ready, we can start using them, even if they are blank
        startDrawing = true
        scanOn = false

        // Process messages coming in on the ServerMessage channel.
            requestUpdate(vc)
            startDrawing = true
            for msg := range vcc.ServerMessageCh {
                requestUpdate(vc)
                startDrawing = true

                switch msg.Type() {
                case vnc.FramebufferUpdateMsg:
                rects := msg.(*vnc.FramebufferUpdate).Rects
                for _,v := range rects {
                startDrawing = true
                bpp := uint(4)
                for y:=uint(0);y<uint(v.Height); y++ {
                    start := (uint(v.Y)+y)*clientWidth*bpp + uint(v.X)*bpp
                    for j := uint(0);j<uint(v.Width);j++ {
                        sOffset := y*uint(v.Width)*bpp + j*bpp

                        u8Pix[start+j*bpp]   = v.BytePix[sOffset] //uint8(c.R)
                        u8Pix[start+j*bpp+1] = v.BytePix[sOffset+1] //uint8(c.G)
                        u8Pix[start+j*bpp+2] = v.BytePix[sOffset+2] //uint8(c.B)
                        u8Pix[start+j*bpp+3] = v.BytePix[sOffset+3] //uint8(255)


                    }
                }
            }
          }
          lockScreen = false
        }
    }
    //scanOn = true
}

func requestUpdates (vc *vnc.ClientConn) {
    time.Sleep(100*time.Millisecond)
    requestUpdate(vc)
    requestUpdates(vc)
}

func requestUpdate (vc *vnc.ClientConn) {
    if err := vc.FramebufferUpdateRequest(vnc.RFBTrue, 0, 0, uint16(clientWidth), uint16(clientHeight)); err != nil {
      //log.Printf("error requesting framebuffer update: %v", err)
    }
}
