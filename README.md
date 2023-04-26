# ffmpeg_test
Script to test ffmpeg/rtsp camera speed.



## Results

Calling successively `stream.Next()` will return multiple times the same image at high pace.
This script aims at measuring the “real” latency, meaning how fast can we get a new image.

To be precise, this value is the average time to get an image multiplied by the average number of times a same image is returned. 

Example: 
If `stream.Next()` takes on average 3ms and on average the same image is returned 13.7 times, the real latency is 3ms*13.7 = 41.1 ms (24.3 Hz).

If you are concerned if the average number of times “makes sense”, you can plot a histogram of the distribution by calling `plotHistogram()`. From what I have seen, the variance is low. 
The stastical significancy of the average time per operation should be guaranteed by the native Go `Benchmark` implementation. 

## Important notes
The script is written for the current implementation of `ffmpeg` component in `rdk` (v0.2.34). It takes into account that the GoSDK “intercepts” `rimage.LazyEncodedImage` (and decodes it). 
Future changes on `rdk` might break it.
