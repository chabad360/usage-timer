package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "me.chabad360.timer"

func main() {
	// Create a new application.
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	errorCheck(err)

	// Connect function to application startup event, this is not required.
	application.Connect("startup", func() {
		log.Println("application startup")
	})

	// Connect function to application activate event
	application.Connect("activate", func() {
		sendNotification("test", "testy", application)
		log.Println("application activate")

		builder, err := gtk.BuilderNewFromFile("ui/timer.glade")
		errorCheck(err)

		// Get the object with the id of "main_window".
		obj, err := builder.GetObject("main_window")
		errorCheck(err)

		// Verify that the object is a pointer to a gtk.ApplicationWindow.
		win, err := isWindow(obj)
		errorCheck(err)

		// Show the Window and all of its components.
		win.Show()
		application.AddWindow(win)

		quitBtn, err := getButton(builder, "quit_button")
		errorCheck(err)

		quitBtn.Connect("clicked", func() {
			application.Quit()
		})

		obj, err = builder.GetObject("time")
		errorCheck(err)

		time := obj.(*gtk.Entry)

		startBtn, err := getButton(builder, "start_button")
		errorCheck(err)

		startBtn.Connect("clicked", func() {
			time, err := time.GetText()
			errorCheck(err)
			startTimer(time)
		})

		mapMenuButtons(builder, application)
	})
	// Connect function to application shutdown event, this is not required.
	application.Connect("shutdown", func() {
		log.Println("application shutdown")
	})

	// Launch the application
	os.Exit(application.Run(os.Args))
}

func mapMenuButtons(bldr *gtk.Builder, app *gtk.Application) {
	menuQuitBtn, err := getMenuItem(bldr, "quit_menu")
	errorCheck(err)

	menuQuitBtn.Connect("activate", func() {
		app.Quit()
	})

	menuAbtBtn, err := getMenuItem(bldr, "about_menu")
	errorCheck(err)

	menuAbtBtn.Connect("activate", func() {
		obj, err := bldr.GetObject("about_window")
		errorCheck(err)

		abt := obj.(*gtk.AboutDialog)
		abt.Show()
	})
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

func getButton(bldr *gtk.Builder, name string) (*gtk.Button, error) {
	obj, err := bldr.GetObject(name)
	errorCheck(err)

	if btn, ok := obj.(*gtk.Button); ok {
		return btn, nil
	}
	return nil, errors.New("not a *gtk.Button")
}

func getMenuItem(bldr *gtk.Builder, name string) (*gtk.MenuItem, error) {
	obj, err := bldr.GetObject(name)
	errorCheck(err)

	if mnu, ok := obj.(*gtk.MenuItem); ok {
		return mnu, nil
	}
	return nil, errors.New("not a *gtk.MenuItem")
}

func sendNotification(title string, text string, app *gtk.Application) {
	notif := glib.NotificationNew(title)
	notif.SetBody(title)
	app.SendNotification(appID, notif)
}

func startTimer(minutes string) {
	min, err := time.ParseDuration(minutes)
	errorCheck(err)

	endTime := time.Now().Add(min)

	for range time.Tick(1 * time.Second) {
		timeRemaining := endTime.Sub(time.Now())

		if timeRemaining.Seconds() <= 0 {
			fmt.Println("Countdown reached!")
			break
		}

	}
}
