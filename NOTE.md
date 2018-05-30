# Snips
- udp 
    - https://stackoverflow.com/questions/27176523/udp-in-golang-listen-not-a-blocking-call?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa
- gstreamer 
    - https://schneide.blog/2015/03/03/streaming-images-from-your-application-to-the-web-with-gstreamer-and-icecast-part-1/
- ffmpeg pipes 
    - https://ffmpeg.org/ffmpeg-protocols.html#pipe
    - https://stackoverflow.com/questions/49798803/creating-an-image-from-a-video-frame
- mjpeg
    - https://github.com/c0va23/go-mjpeg_simple_server
- image to byte
    - https://stackoverflow.com/questions/39577318/encoding-an-image-to-jpeg-in-go?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa

# liveview
1. GET /switch_cammode.cgi?mode=play HTTP/1.1
2. GET /switch_cammode.cgi?mode=rec&lvqty=0640x0480 HTTP/1.1
3. GET /exec_takemisc.cgi?com=startliveview&port=28488 HTTP/1.1
4. GET /exec_takemisc.cgi?com=stopliveview HTTP/1.1
> shoot
5. GET /exec_takemotion.cgi?com=starttake HTTP/1.1
6. GET /exec_takemisc.cgi?com=getrecview HTTP/1.1

# Shutter mode
1. GET /switch_cammode.cgi?mode=play HTTP/1.1
2. GET /switch_cammode.cgi?mode=shutter HTTP/1.1
> shoot
3. GET /exec_shutter.cgi?com=1st2ndpush HTTP/1.1
4. GET /exec_shutter.cgi?com=2nd1strelease HTTP/1.1
