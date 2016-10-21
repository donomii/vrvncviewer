package main

import "math"
import "image"
import "github.com/donomii/go-vnc"
import "net"
import "time"
import "log"
import "golang.org/x/net/context"

var im *image.NRGBA


func NextPo2(n int) int {
    return int(math.Pow(2,math.Ceil(math.Log2(float64(n)))))
}

func run_vnc (server_port string) {
// Establish TCP connection to VNC server.
nc, err := net.DialTimeout("tcp", server_port, 100*time.Millisecond)
if err != nil {
  //log.Fatalf("Error connecting to VNC host. %v", err)
    connectCh <- true
    return
}
    scanOn = false

    // Negotiate connection with the server.
    vcc := vnc.NewClientConfig("aaaaaaaa")
    vc, err := vnc.Connect(context.Background(), nc, vcc)
    log.Println("Connected to server",server_port)
    //go requestUpdates(vc)
    if err != nil {
      scanOn=true
      log.Fatalf("Error negotiating connection to VNC host. %v", err)
    }
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
            for i:=uint(0);i<uint(v.Height); i++ {
                start := (uint(v.Y)+i)*clientWidth*3 + uint(v.X)*3
                for j := uint(0);j<uint(v.Width);j++ {
                    c := v.Enc.(*vnc.RawEncoding).Colors[i*uint(v.Width)+j]
                    u8Pix[start+j*3] = uint8(c.R)
                    u8Pix[start+j*3+1] = uint8(c.G)
                    u8Pix[start+j*3+2] = uint8(c.B)
                }
            }
        }
      }
    }
    scanOn = true
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
