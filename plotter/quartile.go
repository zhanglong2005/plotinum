// Copyright 2012 The Plotinum Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

package plotter

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/vg"

	"image/color"
)

var (
	// DefaultQuartMedianStyle is a fat dot.
	DefaultQuartMedianStyle = plot.GlyphStyle{
		Color:  color.Black,
		Radius: vg.Points(1.5),
		Shape:  plot.CircleGlyph{},
	}

	// DefaultQuartWhiskerStyle is a hairline.
	DefaultQuartWhiskerStyle = plot.LineStyle{
		Color:    color.Black,
		Width:    vg.Points(0.5),
		Dashes:   []vg.Length{},
		DashOffs: 0,
	}
)

// QuartPlot implements the Plotter interface, drawing
// a boxplot to represent the distribution of values.
type QuartPlot struct {
	fiveStatPlot

	// MedianStyle is the line style for the median point.
	MedianStyle plot.GlyphStyle

	// WhiskerStyle is the line style used to draw the
	// whiskers.
	WhiskerStyle plot.LineStyle
}

// NewQuartPlot returns a new QuartPlot that represents
// the distribution of the given values.  This style of
// the plot appears in Tufte's "The Visual Display of 
// Quantitative Information".
//
// An error is returned if the plot is created with
// no values.
//
// The fence values are 1.5x the interquartile before
// the first quartile and after the third quartile.  Any
// value that is outside of the fences are drawn as
// Outside points.  The adjacent values (to which the
// whiskers stretch) are the minimum and maximum
// values that are not outside the fences.
func NewQuartPlot(w vg.Length, loc float64, values Valuer) *QuartPlot {
	b := new(QuartPlot)
	b.fiveStatPlot = newFiveStat(w, loc, values)

	b.MedianStyle = DefaultQuartMedianStyle
	b.WhiskerStyle = DefaultQuartWhiskerStyle

	if len(b.Values) == 0 {
		b.Width = 0
		b.MedianStyle.Radius = 0
		b.WhiskerStyle.Width = 0
	}

	return b
}

func (b *QuartPlot) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)
	x := trX(b.Location)
	if !da.ContainsX(x) {
		return
	}

	med := plot.Point{x, trY(b.Median)}
	q1 := trY(b.Quartile1)
	q3 := trY(b.Quartile3)
	aLow := trY(b.AdjLow)
	aHigh := trY(b.AdjHigh)

	da.StrokeLine2(b.WhiskerStyle, x, aHigh, x, q3)
	da.DrawGlyph(b.MedianStyle, med)
	da.StrokeLine2(b.WhiskerStyle, x, aLow, x, q1)

	ostyle := b.MedianStyle
	ostyle.Radius = b.MedianStyle.Radius / 2
	for _, out := range b.Outside {
		y := trY(b.Value(out))
		da.DrawGlyph(ostyle, plot.Point{x, y})
	}
}

// DataRange returns the minimum and maximum x
// and y values, implementing the plot.DataRanger
// interface.
func (b *QuartPlot) DataRange() (float64, float64, float64, float64) {
	return b.Location, b.Location, b.Min, b.Max
}

// GlyphBoxes returns a slice of GlyphBoxes for the
// points and for the median line of the boxplot,
// implementing the plot.GlyphBoxer interface
func (b *QuartPlot) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	bs := make([]plot.GlyphBox, len(b.Outside)+1)
	for i, out := range b.Outside {
		bs[i].X = plt.X.Norm(b.Location)
		bs[i].Y = plt.Y.Norm(b.Value(out))
		bs[i].Rect = b.MedianStyle.Rect()
	}
	bs[len(bs)-1].X = plt.X.Norm(b.Location)
	bs[len(bs)-1].Y = plt.Y.Norm(b.Median)
	bs[len(bs)-1].Rect = plot.Rect{
		Min:  plot.Point{X: -b.Width / 2},
		Size: plot.Point{X: b.Width},
	}
	return bs
}

// OutsideLabels returns a *Labels that will plot
// a label for each of the outside points.  The
// labels are assumed to correspond to the
// points used to create the box plot.
func (b *QuartPlot) OutsideLabels(labels Labeller) (*Labels, error) {
	strs := make([]string, len(b.Outside))
	for i, out := range b.Outside {
		strs[i] = labels.Label(out)
	}
	o := quartPlotOutsideLabels{b, strs}
	ls, err := NewLabels(o)
	if err != nil {
		return nil, err
	}
	ls.XOffset += b.MedianStyle.Radius / 2
	ls.YOffset += b.MedianStyle.Radius / 2
	return ls, nil
}

type quartPlotOutsideLabels struct {
	box    *QuartPlot
	labels []string
}

func (o quartPlotOutsideLabels) Len() int {
	return len(o.box.Outside)
}

func (o quartPlotOutsideLabels) XY(i int) (float64, float64) {
	return o.box.Location, o.box.Value(o.box.Outside[i])
}

func (o quartPlotOutsideLabels) Label(i int) string {
	return o.labels[i]
}

// HorizQuartPlot is like a regular QuartPlot, however,
// it draws horizontally instead of Vertically.
type HorizQuartPlot struct{ *QuartPlot }

// MakeHorizQuartPlot returns a HorizQuartPlot,
// plotting the values in a horizontal plot
// centered along a fixed location of the y axis.
func MakeHorizQuartPlot(w vg.Length, loc float64, vs Values) HorizQuartPlot {
	return HorizQuartPlot{NewQuartPlot(w, loc, vs)}
}

func (b HorizQuartPlot) Plot(da plot.DrawArea, plt *plot.Plot) {
	trX, trY := plt.Transforms(&da)
	y := trY(b.Location)
	if !da.ContainsY(y) {
		return
	}

	med := plot.Point{trX(b.Median), y}
	q1 := trX(b.Quartile1)
	q3 := trX(b.Quartile3)
	aLow := trX(b.AdjLow)
	aHigh := trX(b.AdjHigh)

	da.StrokeLine2(b.WhiskerStyle, aHigh, y, q3, y)
	da.DrawGlyph(b.MedianStyle, med)
	da.StrokeLine2(b.WhiskerStyle, aLow, y, q1, y)

	ostyle := b.MedianStyle
	ostyle.Radius = b.MedianStyle.Radius / 2
	for _, out := range b.Outside {
		x := trX(b.Value(out))
		da.DrawGlyph(ostyle, plot.Point{x, y})
	}
}

// DataRange returns the minimum and maximum x
// and y values, implementing the plot.DataRanger
// interface.
func (b HorizQuartPlot) DataRange() (float64, float64, float64, float64) {
	return b.Min, b.Max, b.Location, b.Location
}

// GlyphBoxes returns a slice of GlyphBoxes for the
// points and for the median line of the boxplot,
// implementing the plot.GlyphBoxer interface
func (b HorizQuartPlot) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	bs := make([]plot.GlyphBox, len(b.Outside)+1)
	for i, out := range b.Outside {
		bs[i].X = plt.X.Norm(b.Value(out))
		bs[i].Y = plt.Y.Norm(b.Location)
		bs[i].Rect = b.MedianStyle.Rect()
	}
	bs[len(bs)-1].X = plt.X.Norm(b.Median)
	bs[len(bs)-1].Y = plt.Y.Norm(b.Location)
	bs[len(bs)-1].Rect = plot.Rect{
		Min:  plot.Point{Y: -b.Width / 2},
		Size: plot.Point{Y: b.Width},
	}
	return bs
}

// OutsideLabels returns a *Labels that will plot
// a label for each of the outside points.  The
// labels are assumed to correspond to the
// points used to create the box plot.
func (b *HorizQuartPlot) OutsideLabels(labels Labeller) (*Labels, error) {
	strs := make([]string, len(b.Outside))
	for i, out := range b.Outside {
		strs[i] = labels.Label(out)
	}
	o := horizQuartPlotOutsideLabels{
		quartPlotOutsideLabels{b.QuartPlot, strs},
	}
	ls, err := NewLabels(o)
	if err != nil {
		return nil, err
	}
	ls.XOffset += b.MedianStyle.Radius / 2
	ls.YOffset += b.MedianStyle.Radius / 2
	return ls, nil
}

type horizQuartPlotOutsideLabels struct {
	quartPlotOutsideLabels
}

func (o horizQuartPlotOutsideLabels) XY(i int) (float64, float64) {
	return o.box.Value(o.box.Outside[i]), o.box.Location
}
