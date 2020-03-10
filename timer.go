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

func main() {
	// Create a new application.
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	errorCheck(err)

	// Connect function to application activate event
	application.Connect("activate", func() {
		log.Println("application activate")

		builder, err := gtk.BuilderNewFromFile("ui/timer.glade")
		errorCheck(err)

		// Get the object with the id of "main_window".
		obj, err := builder.GetObject("main_window")
		errorCheck(err)

		// Verify that the object is a pointer to a gtk.ApplicationWindow.
		win, err := isWindow(obj)
		errorCheck(err)

		win.Show()
		application.AddWindow(win)

		quitBtn, err := getButton(builder, "quit_button")
		errorCheck(err)

		quitBtn.Connect("clicked", func() {
			application.Quit()
		})

		aboutBtn, err := getButton(builder, "about_button")
		errorCheck(err)

		aboutBtn.Connect("clicked", func() {
			showAbout(win)
		})

		obj, err = builder.GetObject("time")
		errorCheck(err)

		timeInput := obj.(*gtk.Entry)

		obj, err = builder.GetObject("time_left")
		errorCheck(err)

		timeLeft := obj.(*gtk.ProgressBar)

		startBtn, err := getButton(builder, "start_button")
		errorCheck(err)

		startBtn.Connect("clicked", func() {
			input, err := timeInput.GetText()
			errorCheck(err)
			if time, err := strconv.ParseFloat(input, 64); err == nil {
				if response := showAsk("Would you like to start a timer for "+input+" minutes?", win); response == gtk.ResponseType(-8) {
					timeInput.SetSensitive(false)
					startBtn.SetSensitive(false)
					quitBtn.SetSensitive(false)
					str := strconv.FormatFloat(time, 'g', 4, 64) + "m1s"
					go startTimer(str, timeLeft)
				} else {
					fmt.Println(response)
				}
			} else {
				showError("Time must be a number", win)
			}
		})

	})
	// Connect function to application shutdown event, this is not required.
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
	})

	// Launch the application
	os.Exit(application.Run(os.Args))
}

func isWindow(obj glib.IObject) (*gtk.Window, error) {
	// Make type assertion (as per gtk.go).
	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, errors.New("not a *gtk.Window")
}

func errorCheck(e error) {
	if e != nil {
		// panic for any errors.
		log.Panic(e)
	}
}
func getButton(builder *gtk.Builder, name string) (*gtk.Button, error) {
	obj, err := builder.GetObject(name)
	errorCheck(err)

	if btn, ok := obj.(*gtk.Button); ok {
		return btn, nil
	}
	return nil, errors.New("not a *gtk.Button")
}

func getMenuItem(builder *gtk.Builder, name string) (*gtk.MenuItem, error) {
	obj, err := builder.GetObject(name)
	errorCheck(err)

	if mnu, ok := obj.(*gtk.MenuItem); ok {
		return mnu, nil
	}
	return nil, errors.New("not a *gtk.MenuItem")
}

func sendNotification(title string, text string, image string) {
	notify.Init(appID)
	hello := notify.NotificationNew(title, text, image)
	hello.Show()
}

func startTimer(minutes string, bar *gtk.ProgressBar) {
	min, err := time.ParseDuration(minutes)
	errorCheck(err)
	secs := int(min.Seconds() - 1)
	bar.SetText(fmt.Sprintf("%02d:%02d:%02d Left", int(secs/(60*60)%24), int((secs/60)%60), int(secs%60)))

	endTime := time.Now().Add(min)
	percent := float64(1 / min.Seconds())
	progress := float64(0)

	for range time.Tick(time.Second) {
		timeRemaining := endTime.Sub(time.Now())
		secs = int(timeRemaining.Seconds())
		bar.SetText(fmt.Sprintf("%02d:%02d:%02d Left", int(secs/(60*60)%24), int((secs/60)%60), int(secs%60)))
		progress = progress + percent
		bar.SetFraction(progress)

		if secs == 300 {
			sendNotification("Usage Timer", "5 Minutes Left!", "Warning")
		} else if secs == 120 {
			sendNotification("Usage Timer", "2 Minutes Left!", "Warning")
		} else if secs <= 0 {
			sendNotification("Usage Timer", "Countdown reached!", "Error")
			if err := exec.Command("/usr/bin/poweroff").Run(); err != nil {
				sendNotification("Error!", "Failed to poweroff", "error")
			}
			break
		}

	}
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

	aboutDialog.SetCopyright("(c) 2020 Mendel Greenberg")
	aboutDialog.SetLicenseType(gtk.License(3))
	aboutDialog.SetLogoIconName("time")
	aboutDialog.SetName("Usage Timer")
	aboutDialog.Show()
}
