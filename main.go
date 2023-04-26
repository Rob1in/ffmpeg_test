package main

import (
	"context"
	"fmt"
	"github.com/edaniels/golog"
	"github.com/edaniels/gostream"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/rimage"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/test"
	"go.viam.com/utils/rpc"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image"
	"math"
	"sort"
	"testing"
	"time"
)

func main() {
	logger := golog.NewDevelopmentLogger("client")
	var imgs []image.Image

	//****** UPDATE THIS PART ********
	robot, err := client.New(
		context.Background(),
		"xxxxxxxxxxxxxxxxxx.viam.cloud",
		logger,
		client.WithDialOptions(rpc.WithCredentials(rpc.Credentials{
			Type:    utils.CredentialsTypeRobotLocationSecret,
			Payload: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		})),
	)
	//************ END ************

	if err != nil {
		logger.Error(err)
	}
	logger.Info("Resources:")
	logger.Info(robot.ResourceNames())
	defer robot.Close(context.Background())

	s, err := getStreamFromRobot(robot)
	if err != nil {
		logger.Error(err)
	}
	result := testing.Benchmark(func(b *testing.B) {
		BenchmarkFFMPEG(b, s, &imgs)
	})

	err = saveOneOutput(imgs[0])
	if err != nil {
		logger.Error(err)
	}

	oc := getOccurrences(imgs)
	mean := meanMapValues(oc)
	avg := averageTimePerOp(result)
	ni := time.Duration(mean) * avg
	//plotHistogram(oc)
	fmt.Println("On average, getting a new image every", ni, " (", frequency(ni), "Hz )")
}

func getStreamFromRobot(robot *client.RobotClient) (gostream.VideoStream, error) {
	ffmpegCamComponent, err := camera.FromRobot(robot, "ffmpeg-cam")
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return ffmpegCamComponent.Stream(ctx)
}

func meanMapValues(m map[string]int) float64 {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return float64(sum) / float64(len(m))
}

func BenchmarkFFMPEG(b *testing.B, s gostream.VideoStream, imgs *[]image.Image) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img, _, err := s.Next(ctx)
		test.That(b, err, test.ShouldBeNil)
		*imgs = append(*imgs, img)
	}
}

func averageTimePerOp(result testing.BenchmarkResult) time.Duration {
	return result.T / time.Duration(result.N)
}

// frequency is rounded at the second decimal
func frequency(d time.Duration) float64 {
	return math.Round(float64(time.Second)/float64(d)*100) / 100
}

func plotHistogram(m map[string]int) {
	var values plotter.Values
	for _, v := range m {
		values = append(values, float64(v))
	}
	sort.Float64s(values)
	p := plot.New()
	p.Title.Text = "Histogram of duplicates"
	p.X.Label.Text = "# of duplicates"
	p.Y.Label.Text = "Frequency"
	hist, err := plotter.NewHist(values, int(math.Sqrt(float64(len(m)))))
	if err != nil {
		fmt.Println(err)
		return
	}
	p.Add(hist)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "./histogram.png"); err != nil {
		fmt.Println(err)
	}
}
func printMapValues(m map[string]int) {
	for _, v := range m {
		fmt.Println(v)
	}
}

func getOccurrences(imgs []image.Image) map[string]int {
	imgsB := getBytes(imgs)
	counts := make(map[string]int)
	for _, img := range imgsB {
		counts[string(img)]++
	}
	return counts
}

func getBytes(imgs []image.Image) [][]byte {
	var imgsBytes [][]byte
	for _, img := range imgs {
		if lazy, ok := img.(*rimage.LazyEncodedImage); ok {
			imgsBytes = append(imgsBytes, lazy.RawData())
		}
	}
	return imgsBytes
}
func saveOneOutput(img image.Image) error {
	if lazy, ok := img.(*rimage.LazyEncodedImage); ok {
		img, err := rimage.DecodeImage(context.Background(), lazy.RawData(), lazy.MIMEType())
		if err != nil {
			return err
		}
		return rimage.SaveImage(img, "./from_stream.jpeg")
	}
	return rimage.SaveImage(img, "./from_stream.jpeg")
}
