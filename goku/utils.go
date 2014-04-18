package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloud66/cx/term"

	"github.com/mgutz/ansi"
)

// exists returns whether the given file or directory exists or not
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// concatenates the file content and returns a new one
func appendFiles(files []string, filename string) error {
	d, err := os.Create(filename)
	defer func() {
		cerr := d.Close()
		if err == nil {
			err = cerr
		}
	}()
	if err != nil {
		return err
	}

	for _, fn := range files {
		f, err := os.Open(fn)
		defer f.Close()
		if _, err = io.Copy(d, f); err != nil {
			return err
		}
		err = d.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func must(err error) {
	if err != nil {
		printFatal(err.Error())
	}
}

func printError(message string, args ...interface{}) {
	log.Println(colorizeMessage("red", "error:", message, args...))
}

func printFatal(message string, args ...interface{}) {
	log.Fatal(colorizeMessage("red", "error:", message, args...))
}

func printWarning(message string, args ...interface{}) {
	log.Println(colorizeMessage("yellow", "warning:", message, args...))
}

func mustConfirm(warning, desired string) {
	if term.IsTerminal(os.Stdin) {
		printWarning(warning)
		fmt.Printf("> ")
	}
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		printFatal(err.Error())
	}

	if confirm != desired {
		printFatal("Confirmation did not match %q.", desired)
	}
}

func colorizeMessage(color, prefix, message string, args ...interface{}) string {
	prefResult := ""
	if prefix != "" {
		prefResult = ansi.Color(prefix, color+"+b") + " " + ansi.ColorCode("reset")
	}
	return prefResult + ansi.Color(fmt.Sprintf(message, args...), color) + ansi.ColorCode("reset")
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}

type prettyTime struct {
	time.Time
}

func (s prettyTime) String() string {
	if time.Now().Sub(s.Time) < 12*30*24*time.Hour {
		return s.Local().Format("Jan _2 15:04")
	}
	return s.Local().Format("Jan _2  2006")
}

type prettyDuration struct {
	time.Duration
}

func (a prettyDuration) String() string {
	switch d := a.Duration; {
	case d > 2*24*time.Hour:
		return a.Unit(24*time.Hour, "d")
	case d > 2*time.Hour:
		return a.Unit(time.Hour, "h")
	case d > 2*time.Minute:
		return a.Unit(time.Minute, "m")
	}
	return a.Unit(time.Second, "s")
}

func (a prettyDuration) Unit(u time.Duration, s string) string {
	return fmt.Sprintf("%2d", roundDur(a.Duration, u)) + s
}

func roundDur(d, k time.Duration) int {
	return int((d + k/2 - 1) / k)
}

func abbrev(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func ensurePrefix(val, prefix string) string {
	if !strings.HasPrefix(val, prefix) {
		return prefix + val
	}
	return val
}

func ensureSuffix(val, suffix string) string {
	if !strings.HasSuffix(val, suffix) {
		return val + suffix
	}
	return val
}

func writeSshFile(filename string, content string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	defer file.Close()

	if err != nil {
		return err
	}

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return nil
}

func downloadFile(source string, output string) error {
	out, err := os.Create(output)
	defer out.Close()
	if err != nil {
		return err
	}

	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// finds the item in the list without case sensitivity, and returns the index
// of the item that matches or begins with the given item
// if more than one match is found, it returns an error
func fuzzyFind(s []string, item string) (int, error) {
	var results []int
	for i := range s {
		// look for identical matches first
		if strings.ToLower(s[i]) == strings.ToLower(item) {
			results = append(results, i)
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}
	if len(results) > 1 {
		return 0, errors.New("More than one match found for " + item + " you might get better results by passing the environment with -e")
	}

	for i := range s {
		if strings.HasPrefix(strings.ToLower(s[i]), strings.ToLower(item)) {
			results = append(results, i)
		}
	}

	if len(results) == 0 {
		return 0, errors.New("No match found for " + item)
	}
	if len(results) > 1 {
		return 0, errors.New("More than one match found for " + item)
	}

	return results[0], nil
}

func stringsIndex(s []string, item string) int {
	for i := range s {
		if s[i] == item {
			return i
		}
	}
	return -1
}
