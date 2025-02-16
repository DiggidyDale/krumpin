package main

import (
	"embed"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"krumpin/db"
	"krumpin/models"
	kTheme "krumpin/theme"
	"krumpin/utils"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed resources/skills.txt
var skillsFile embed.FS

var diceOptions = []string{"d4", "d8", "d10", "d12", "d20"}

func rollDice(skill binding.String) (int, error) {
	val, _ := skill.Get()
	trimmed := strings.TrimPrefix(val, "d")
	sides, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid dice type: %s", skill)
	}

	result := rand.Intn(sides) + 1
	if result == sides {
		index := utils.FindIndex(diceOptions, val)
		if index != -1 && len(diceOptions)-1 != index {
			_ = skill.Set(diceOptions[index+1])
		}
	}
	return result, nil
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
		val, _ := skill.DiceVal.Get()
		if val == "" {
			dialog.ShowInformation("Dice Roll", fmt.Sprintf("No dice selected for %s.", skill.Skill.Name), parentWindow)
			return
		}
		trimmed := strings.TrimPrefix(val, "d")
		result, err := rollDice(skill.DiceVal)
		if err != nil {
			dialog.ShowError(err, parentWindow)
			return
		}
		maxSides, _ := strconv.Atoi(trimmed)

		showRollResultDialog(result, maxSides, skill.Skill.Name, parentWindow)
	})

	val, _ := skill.DiceVal.Get()
	diceSelect := widget.NewSelect(diceOptions, func(selected string) {
		val = selected
	})
	if val != "" {
		diceSelect.SetSelected(val)
	} else {
		_ = skill.DiceVal.Set(diceOptions[0])
		diceSelect.SetSelected(diceOptions[0])
	}

	var infoButton *kTheme.InfoCircle
	if skill.Skill.Description != "" {
		infoButton = kTheme.NewInfoCircle(skill.Skill.Description, parentWindow)
	} else {
		// Provide a placeholder so layout remains consistent.
		infoButton = kTheme.NewInfoCircle("", parentWindow)
		infoButton.Hide() // Hide if no tooltip is needed.
	}

	skillContainer := container.NewHBox(
		infoButton,
		rollButton,
		diceSelect,
	)
	return skillContainer
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
	err := db.InitialiseDb()
	if err != nil {
		dialog.ShowError(err, w)
	}
	log.Printf("db should be initialised")
	baseSkills, _ := db.LoadBaseSkills()

	skillsContainer := container.NewVBox()
	skillsTitle := widget.NewLabelWithStyle("Skills (Click baseSkill name to roll dice):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	skillsContainer.Add(skillsTitle)

	for _, baseSkill := range baseSkills {
		if baseSkill != nil {
			skill := &models.CharacterSkill{}
			skill.Skill = *baseSkill
			skill.DiceVal = binding.NewString()
			err := skill.DiceVal.Set(diceOptions[0])
			if err != nil {
				dialog.ShowError(err, w)
			}

			skillUI := newSkillWidget(skill, w)
			skillsContainer.Add(skillUI)
		}
	}
	mainContainer.Add(skillsContainer)
	centered := container.NewCenter(mainContainer)
	w.SetContent(centered)

	w.Resize(fyne.NewSize(400, 300))
	w.ShowAndRun()
}
