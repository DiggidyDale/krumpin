package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	theme2 "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type CustomTheme struct {
}

func (c *CustomTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme2.ColorNameBackground:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // light grey background
	case theme2.ColorNameButton, theme2.ColorNameDisabledButton:
		return color.NRGBA{R: 100, G: 150, B: 250, A: 255} // soft blue for buttons
	case theme2.ColorNameForeground:
		return color.Black

	}
	return theme2.DefaultTheme().Color(n, v)
}

func (c *CustomTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme2.DefaultTheme().Font(s)
}

func (c *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme2.DefaultTheme().Icon(name)
}

func (c *CustomTheme) Size(s fyne.ThemeSizeName) float32 {
	return theme2.DefaultTheme().Size(s)
}

type InfoCircle struct {
	widget.BaseWidget

	tooltipText  string
	parentWindow fyne.Window
	popup        *widget.PopUp
}

// NewInfoCircle creates a new InfoCircle with the provided tooltip text.
func NewInfoCircle(tooltipText string, win fyne.Window) *InfoCircle {
	ic := &InfoCircle{
		tooltipText:  tooltipText,
		parentWindow: win,
	}
	ic.ExtendBaseWidget(ic)
	return ic
}

// CreateRenderer implements fyne.Widget.
func (ic *InfoCircle) CreateRenderer() fyne.WidgetRenderer {
	// Create a circle with a border.
	circle := canvas.NewCircle(color.NRGBA{R: 200, G: 200, B: 250, A: 255})
	circle.StrokeColor = color.Black
	circle.StrokeWidth = 1

	// Create a label with "?".
	label := canvas.NewText("?", color.Black)
	label.Alignment = fyne.TextAlignCenter

	objects := []fyne.CanvasObject{circle, label}

	return &infoCircleRenderer{
		ic:      ic,
		circle:  circle,
		label:   label,
		objects: objects,
	}
}

// MinSize returns the minimum size of the InfoCircle.
func (ic *InfoCircle) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}

type infoCircleRenderer struct {
	ic      *InfoCircle
	circle  *canvas.Circle
	label   *canvas.Text
	objects []fyne.CanvasObject
}

// Layout positions the circle and label inside the InfoCircle.
func (r *infoCircleRenderer) Layout(size fyne.Size) {
	r.circle.Resize(size)
	r.circle.Move(fyne.NewPos(0, 0))
	r.label.Resize(size)
	r.label.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size of the renderer.
func (r *infoCircleRenderer) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}

// Refresh redraws the widget.
func (r *infoCircleRenderer) Refresh() {
	canvas.Refresh(r.ic)
}

// BackgroundColor returns the background color.
func (r *infoCircleRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

// Objects returns the renderer’s objects.
func (r *infoCircleRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy is a no-op.
func (r *infoCircleRenderer) Destroy() {}

// MouseIn is called when the mouse enters the InfoCircle.
func (ic *InfoCircle) MouseIn(*desktop.MouseEvent) {
	if ic.tooltipText == "" {
		return
	}
	// Create a label for the tooltip.
	tooltip := widget.NewLabel(ic.tooltipText)
	tooltipContainer := container.NewPadded(tooltip)
	// Create a popup anchored to the canvas.
	ic.popup = widget.NewPopUp(tooltipContainer, ic.parentWindow.Canvas())
	// Position the tooltip near the InfoCircle.
	pos := ic.Position() // position relative to parent container
	min := ic.MinSize()  // size of the info circle
	ic.popup.ShowAtPosition(fyne.NewPos(pos.X+min.Width, pos.Y))
}

// MouseOut is called when the mouse leaves the InfoCircle.
func (ic *InfoCircle) MouseOut() {
	if ic.popup != nil {
		ic.popup.Hide()
		ic.popup = nil
	}
}

// MouseMoved satisfies the desktop.Hoverable interface.
func (ic *InfoCircle) MouseMoved(*desktop.MouseEvent) {}

// Ensure InfoCircle implements desktop.Hoverable.
var _ desktop.Hoverable = (*InfoCircle)(nil)
