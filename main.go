package main

import (
	"embed"
	"fmt"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"image/color"
	krumpinDb "krumpin/db"
	"krumpin/models"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed resources/skills.txt
var skillsFile embed.FS

var diceOptions = []string{"d4", "d8", "d10", "d12", "d20"}

func rollDice(dice string) (int, error) {
	trimmed := strings.TrimPrefix(dice, "d")
	sides, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid dice type: %s", dice)
	}
	return rand.Intn(sides) + 1, nil
}

func flashAmber(bgRect *canvas.Rectangle, w fyne.Window) {
	original := bgRect.FillColor
	amber := color.NRGBA{R: 255, G: 191, B: 0, A: 255}

	bgRect.FillColor = amber
	w.Canvas().Refresh(bgRect)

	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if i%2 == 0 {
			bgRect.FillColor = amber
		} else {
			bgRect.FillColor = original
		}
		w.Canvas().Refresh(bgRect)
	}
	time.Sleep(300 * time.Millisecond)
	bgRect.FillColor = original
	w.Canvas().Refresh(bgRect)
}

func showRollResultDialog(result int, maxSides int, skillName string, parentWindow fyne.Window) {
	msg := fmt.Sprintf("Rolling %%s for %s... Result: %d", skillName, result)
	resultLabel := widget.NewLabel(fmt.Sprintf(msg, ""))

	bgColor := theme.Current().Color(theme.ColorNameBackground, theme.VariantDark)
	bgRect := canvas.NewRectangle(bgColor)
	content := container.NewMax(bgRect, container.NewCenter(resultLabel))

	d := dialog.NewCustom("Dice Roll", "Close", content, parentWindow)
	d.Show()

	if result == maxSides {
		go flashAmber(bgRect, parentWindow)
	}
}

func newSkillWidget(skill *models.CharacterSkill, parentWindow fyne.Window) *fyne.Container {
	rollButton := widget.NewButton(skill.Skill.Name, func() {
		if skill.DiceVal == "" {
			dialog.ShowInformation("Dice Roll", fmt.Sprintf("No dice selected for %s.", skill.Skill.Name), parentWindow)
			return
		}
		result, err := rollDice(skill.DiceVal)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		trimmed := strings.TrimPrefix(skill.DiceVal, "d")
		maxSides, _ := strconv.Atoi(trimmed)

		showRollResultDialog(result, maxSides, skill.Skill.Name, parentWindow)
	})

	diceSelect := widget.NewSelect(diceOptions, func(selected string) {
		skill.DiceVal = selected
	})
	if skill.DiceVal != "" {
		diceSelect.SetSelected(skill.DiceVal)
	} else {
		skill.DiceVal = diceOptions[0]
		diceSelect.SetSelected(diceOptions[0])
	}

	var infoButton *InfoCircle
	if skill.Skill.Description != "" {
		infoButton = NewInfoCircle(skill.Skill.Description, parentWindow)
	} else {
		// Provide a placeholder so layout remains consistent.
		infoButton = NewInfoCircle("", parentWindow)
		infoButton.Hide() // Hide if no tooltip is needed.
	}

	skillContainer := container.NewHBox(
		infoButton,
		rollButton,
		diceSelect,
	)
	return skillContainer
}

type customTheme struct{}

func (c *customTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // light grey background
	case theme.ColorNameButton, theme.ColorNameDisabledButton:
		return color.NRGBA{R: 100, G: 150, B: 250, A: 255} // soft blue for buttons
	case theme.ColorNameForeground:
		return color.Black

	}
	return theme.DefaultTheme().Color(n, v)
}

func (c *customTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

func (c *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *customTheme) Size(s fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(s)
}

func main() {
	a := app.NewWithID("uk.co.diggidydale.krumpintracker")
	w := a.NewWindow("Krump'in Character Tracker")

	mainContainer := container.NewVBox()

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter Character Name")
	nameLabel := widget.NewLabel("Character Name:")
	charNameContainer := container.NewVBox(nameLabel, nameEntry)
	mainContainer.Add(charNameContainer)

	log.Printf("about to initialise db")
	err := krumpinDb.InitialiseDb()
	if err != nil {
		dialog.ShowError(err, w)
	}
	log.Printf("db should be initialised")
	baseSkills, _ := krumpinDb.LoadBaseSkills()

	skillsContainer := container.NewVBox()
	skillsTitle := widget.NewLabelWithStyle("Skills (Click baseSkill name to roll dice):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	skillsContainer.Add(skillsTitle)

	for _, baseSkill := range baseSkills {
		var skill *models.CharacterSkill
		skill.Skill = *baseSkill
		skill.DiceVal = diceOptions[0]

		skillUI := newSkillWidget(skill, w)
		skillsContainer.Add(skillUI)
	}
	mainContainer.Add(skillsContainer)
	centered := container.NewCenter(mainContainer)
	w.SetContent(centered)

	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
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

// Ensure InfoCircle implements desktop.Hoverable.
var _ desktop.Hoverable = (*InfoCircle)(nil)

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
