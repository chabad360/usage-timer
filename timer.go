package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mqu/go-notify"
)

const appID = "me.chabad360.timer"
const maxButton = 2

func main() {
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	errorCheck(err)

	application.Connect("activate", func() { startWindow(application) })

	application.Connect("shutdown", func() {
		log.Println("application shutdown")
	})

	os.Exit(application.Run(os.Args))
}

func startWindow(application *gtk.Application) {
	log.Println("application activate")

	builder, err := gtk.BuilderNewFromFile("ui/timer.glade")
	errorCheck(err)

	win, err := getWindow(builder)
	errorCheck(err)

	win.Show()

	application.AddWindow(win)

	getAboutButton(builder, win)

	quitBtn := getQuitButton(builder, application)

	obj, err := builder.GetObject("time")
	errorCheck(err)

	timeInput := obj.(*gtk.Entry)

	obj, err = builder.GetObject("time_left")
	errorCheck(err)

	timeLeft := obj.(*gtk.ProgressBar)

	moreBtn, err := getButton(builder, "more_button")
	errorCheck(err)

	startBtn, err := getButton(builder, "start_button")
	errorCheck(err)

	startBtn.Connect("clicked", func() {
		startButton(timeInput, timeLeft, startBtn, quitBtn, moreBtn, win)
	})

}

func startButton(timeInput *gtk.Entry, timeLeft *gtk.ProgressBar, startBtn *gtk.Button, quitBtn *gtk.Button, moreBtn *gtk.Button, win *gtk.Window) {
	input, err := timeInput.GetText()
	errorCheck(err)
	if time, err := strconv.ParseFloat(input, 64); err == nil {
		if response := showAsk("Would you like to start a timer for "+input+" minutes?", win); response == gtk.ResponseType(-8) {
			timeInput.SetSensitive(false)
			startBtn.SetSensitive(false)
			quitBtn.SetSensitive(false)
			str := strconv.FormatFloat(time, 'g', 4, 64) + "m1s"
			startTimer(str, timeLeft, moreBtn)
		}
	} else {
		showError("Time must be a number", win)
	}
}

func errorCheck(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func getWindow(builder *gtk.Builder) (*gtk.Window, error) {
	obj, err := builder.GetObject("main_window")
	errorCheck(err)

	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, errors.New("not a *gtk.Window")
}

func getButton(builder *gtk.Builder, name string) (*gtk.Button, error) {
	obj, err := builder.GetObject(name)
	errorCheck(err)

	if btn, ok := obj.(*gtk.Button); ok {
		return btn, nil
	}
	return nil, errors.New("not a *gtk.Button")
}

func getQuitButton(builder *gtk.Builder, application *gtk.Application) *gtk.Button {
	quitBtn, err := getButton(builder, "quit_button")
	errorCheck(err)

	quitBtn.Connect("clicked", func() {
		application.Quit()
	})

	return quitBtn
}

func getAboutButton(builder *gtk.Builder, window *gtk.Window) *gtk.Button {
	aboutBtn, err := getButton(builder, "about_button")
	errorCheck(err)

	aboutBtn.Connect("clicked", func() {
		showAbout(window)
	})

	return aboutBtn
}

func sendNotification(title string, text string, image string) {
	notify.Init(appID)
	hello := notify.NotificationNew(title, text, image)
	hello.Show()
}

func startTimer(minutes string, bar *gtk.ProgressBar, moreBtn *gtk.Button) {
	min, err := time.ParseDuration(minutes)
	errorCheck(err)
	secs := int(min.Seconds() - 1)
	setFraction(bar, float64(0), float64(0), secs)

	endTime := time.Now().Add(min)

	extraTime, _ := time.ParseDuration("5m")
	buttonUse := 0
	t := make(chan time.Time, 100)

	moreBtn.Connect("clicked", func() {
		endTime = endTime.Add(extraTime)
		t <- endTime
		buttonUse++
		moreBtn.SetSensitive(false)
	})

	t <- endTime
	go timer(bar, moreBtn, min, buttonUse, t)
}

func timer(bar *gtk.ProgressBar, moreBtn *gtk.Button, min time.Duration, buttonUse int, t chan time.Time) {
	secs := int(min.Seconds() - 1)
	percent := float64(1 / min.Seconds())
	progress := float64(0)
	endTime := <-t

	for range time.Tick(time.Second) {
		select { // Needed to keep endTime from blocking, annoyingly increases cyclomatic complexity.
		case endTime = <-t:
		default:
		}

		timeRemaining := endTime.Sub(time.Now())
		secs = int(timeRemaining.Seconds())
		progress = setFraction(bar, progress, percent, secs)

		if secs == 300 {
			sendNotification("Usage Timer", "5 Minutes Left!", "Warning")
			if buttonUse < maxButton {
				moreBtn.SetSensitive(true)
			}
		} else if secs == 120 {
			sendNotification("Usage Timer", "2 Minutes Left!", "Warning")
		} else if secs <= 0 {
			sendNotification("Usage Timer", "Countdown reached!", "Error")
			if err := exec.Command("/usr/bin/poweroff").Run(); err != nil {
				sendNotification("Error!", "Failed to poweroff!", "error")
			}
			break
		}

	}
}

func setFraction(bar *gtk.ProgressBar, progress float64, percent float64, secs int) float64 {
	percent = float64((1 - progress) / float64(secs))
	progress = progress + percent

	bar.SetText(fmt.Sprintf("%02d:%02d:%02d Left", int(secs/(60*60)%24), int((secs/60)%60), int(secs%60)))
	bar.SetFraction(progress)

	return progress
}

func showAsk(text string, window *gtk.Window) gtk.ResponseType {
	askDialog := gtk.MessageDialogNew(window, gtk.DialogFlags(1), gtk.MessageType(2), gtk.ButtonsType(4), "Are You Sure?")
	askDialog.FormatSecondaryText(text)
	askDialog.SetDefaultResponse(gtk.ResponseType(-9))

	response := askDialog.Run()
	askDialog.Close()

	return gtk.ResponseType(response)
}

func showError(text string, window *gtk.Window) {
	errorDialog := gtk.MessageDialogNew(window, gtk.DialogFlags(1), gtk.MessageType(3), gtk.ButtonsType(1), "Error!")
	errorDialog.FormatSecondaryText(text)
	errorDialog.Run()

	errorDialog.Close()
}

func showAbout(window *gtk.Window) {
	aboutDialog, err := gtk.AboutDialogNew()
	errorCheck(err)

	aboutDialog.SetAuthors([]string{"Mendel Greenberg"})
	aboutDialog.SetCopyright("(c) 2020 Mendel Greenberg")
	aboutDialog.SetLicenseType(gtk.License(3))
	aboutDialog.SetLogoIconName("time")
	aboutDialog.SetName("Usage Timer")
	aboutDialog.SetVersion("0.1")

	aboutDialog.Show()
}
