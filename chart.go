package main

import (
	"github.com/dustin/go-humanize"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"github.com/vdobler/chart/txtg"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path"
)

type ChartWriter interface {
	Write(*Chart)
}

func point2Chart(in []*Point) []chart.XYErrValue {
	out := make([]chart.XYErrValue, 0, len(in))

	for _, v := range in {
		out = append(out, v)
	}

	return out
}

type Chart struct {
	c    chart.Chart
	name string
}

type ImageWriter struct{}
type TermWriter struct{}

func (im *ImageWriter) Write(c *Chart) {
	os.MkdirAll("data", os.ModePerm)

	fp, err := os.Create(path.Join("data", c.name+".png"))
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	img := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	bg := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(img, img.Bounds(), bg, image.ZP, draw.Src)

	//row, col := d.Cnt/d.N, d.Cnt%d.N
	igr := imgg.AddTo(img, 0, 0, 1024, 768, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.c.Plot(igr)
	png.Encode(fp, img)
}

func (tm *TermWriter) Write(c *Chart) {
	tgr := txtg.New(100, 40)
	c.c.Plot(tgr)
	os.Stdout.Write([]byte(tgr.String() + "\n\n\n"))
}

// return a,b in solution to y = ax + b such that root mean square distance
// between trend line and original points is minimized.
func linearRegression(v []*Point) (a, b float64) {
	var (
		Sx, Sy, Sxx, Syy, Sxy float64
	)
	n := float64(len(v))

	for i, p := range v {
		x, y := float64(i)+1.0, p.YVal()
		Sx = Sx + x
		Sy = Sy + y
		Sxx = Sxx + x*x
		Syy = Syy + y*y
		Sxy = Sxy + x*y
	}

	det := Sxx*n - Sx*Sx
	a = (Sxy*n - Sy*Sx) / det
	b = (Sxx*Sy - Sx*Sxy) / det
	return a, b
}

func trendline(v []*Point) []*Point {
	a, b := linearRegression(v)
	trend := make([]*Point, 0, len(v))

	for i, p := range v {
		y := a*float64(i) + b
		trend = append(trend, &Point{X: p.XVal(), Y: y})
	}

	return trend
}

func movingAverage(v []*Point, window int) []*Point {
	sum := func(y []*Point) (float64, float64) {
		xa, ya := 0.0, 0.0
		for _, p := range y {
			ya += p.YVal()
			xa += p.XVal()
		}
		return xa, ya
	}

	if window < 1 {
		panic("window must be 1 or larger")
	}

	div := float64(window)
	n := len(v)
	res := make([]*Point, 0, n)

	if window > n {
		// this will simply return the same output
		window = n
	}

	for i := window; i < n+1; i++ {
		x, y := sum(v[i-window : i])
		x /= div
		y /= div

		// not quite correct but good enough.
		// should stretch the curve from v[0].x to v[n-1].x
		if i == window {
			x = v[0].XVal()
		} else if i == n {
			x = v[n-1].XVal()
		}
		res = append(res, &Point{Y: y, X: x})
	}

	return res
}

func TimeChart(title, xlabel, ylabel string, data []*Point, drawTrend, extrapolation, drawMovingAvg bool) *Chart {
	c := &chart.ScatterChart{Title: title}
	c.XRange.Label = xlabel
	c.YRange.Label = ylabel
	c.XRange.Time = true
	c.XRange.TicSetting.Mirror = 1

	style := chart.AutoStyle(4, true)
	c.AddDataGeneric(ylabel, point2Chart(data), chart.PlotStyleLinesPoints, style)

	if drawMovingAvg {
		window := 5

		if len(data) < window {
			window = (len(data) / 5)
			if window < 1 {
				window = 1
			}
		}

		mav := movingAverage(data, window)
		style = chart.AutoStyle(1, false)
		c.AddDataGeneric("Moving Average", point2Chart(mav), chart.PlotStyleLines, style)
	}

	if drawTrend {
		style = chart.AutoStyle(6, false)
		linreg := trendline(data)
		c.AddDataGeneric("Linear Regression", point2Chart(linreg), chart.PlotStyleLines, style)
	}

	return &Chart{
		c:    c,
		name: safeFilename(title + " time chart"),
	}
}

func PieChart(title string, labels []string, data []float64) *Chart {
	c := &chart.PieChart{Title: title}
	c.AddDataPair("Tables", labels, data)

	c.FmtVal = func(value, sum float64) (s string) {
		return humanize.Bytes(uint64(value))
	}
	c.Inner = 0.3

	return &Chart{
		c:    c,
		name: safeFilename(title + " pie chart"),
	}
}
