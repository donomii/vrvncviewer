# View your computer desktop in VR

View your computer's desktop in VR, using your mobile phone (and a viewer like Google cardboard).

# Use

You will need to start a server on your desktop, and VR VNC Viewer on your mobile.  Both computer and mobile must be on the same network (e.g. connected to the same WiFi station), and be on the same class C subnet.  e.g. if your computer is 192.168.1.20, your mobile must also have an ip address that starts with 192.168.1.

vrvncviewer will scan this part of the network for vnc servers (on port 5900), and mjpeg servers (mjpeg on port 8080).

## VNC server

You will need to install a VNC server, and set the password option to no authentication, or if that is not possible set the password to 'aaaaaaaa'.  There are many vnc servers, a small selection are listed here:

### Windows

* [TigerVNC](http://tigervnc.org/)
* [TightVNC](http://www.tightvnc.com/)
* [more on Wikipedia](https://en.wikipedia.org/wiki/Virtual_Network_Computing)

### Linux

* xvncserver
* vncserver
* ...

### MacOSX

MacOSX has a VNC server built in.  To enable it, follow these instructions:

Setup [VNC](https://www.dssw.co.uk/reference/vnc/index.html) for MacOSX.

## MJPEG servers

VLC and some other media players can serve videos over mjpeg.

You can also use this simple [MJPEG server](http://praeceptamachinae.com/resources/binaries/rtaVideoStreamer-demo.7z) to view your desktop.  I recommend this - it is usually faster than VNC, although the picture quality is a bit worse.

### Android app

Download and install vrvncviewer from the play store.  After you have started your VNC server, start the vrvncviewer app.  vrvncviewer will scan the local network for your computer, and display it in VR.  Tap your screen (or click with your controller) to move the VR screen to be in front of you.

Finding your computer should take a few seconds.  If vrvncviwer loses the connection, it will start scanning the network again until it finds a server.

vrvncviewer currently runs at around 3 frames per second, regardless of settings.  With a fast phone and fast network, it might be a bit higher.

## Know issues/work to do

* Improve frame rate
* Add cursor to head tracking
* Add more status messages
