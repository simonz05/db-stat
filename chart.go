package main

import (
	"image"
	"image/draw"
	"image/color"
	"image/png"
	"github.com/vdobler/chart"
	"github.com/vdobler/chart/txtg"
	"github.com/vdobler/chart/imgg"
	"os"
	"path"
)

type ChartWriter interface {
	Write(chart.Chart)
}

func point2Chart(in []*Point) []chart.XYErrValue {
	out := make([]chart.XYErrValue, 0, len(in))

	for _, v := range in {
		out = append(out, v)
	}

	return out
}

type outputType int

const (
    termOutput    outputType = iota 
    imageOutput                      
)

type Chart struct {
	c chart.Chart
	name string
}

func (c *Chart) ImageWrite() {
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

func (c *Chart) TermWrite() {
	tgr := txtg.New(100, 40)
	c.c.Plot(tgr)
	os.Stdout.Write([]byte(tgr.String() + "\n\n\n"))
}

func (c *Chart) Write(t outputType) {
	switch t {
	case termOutput:
		c.TermWrite()
	case imageOutput:
		c.ImageWrite()
	default:
	}
}

func TimeChart(title, xlabel, ylabel string, data []*Point) *Chart {
	c := &chart.ScatterChart{Title: title}
	c.XRange.Label = xlabel
	c.YRange.Label = ylabel
	c.XRange.Time = true
	c.XRange.TicSetting.Mirror = 1

	//style := chart.Style{
	//	Symbol: '+', 
	//	SymbolColor: color.NRGBA{0x00, 0x00, 0xff, 0xff}, 
	//	LineStyle: chart.SolidLine}
	style := chart.AutoStyle(4, true)
	c.AddDataGeneric(ylabel, point2Chart(data), chart.PlotStyleLinesPoints, style)

	return &Chart{
		c: c,
		name: safeFilename(title+" time chart"),
	}
}
