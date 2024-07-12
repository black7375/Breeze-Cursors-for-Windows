package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	svgBaseSize     = 32
	exportSize      = 256
	imageOutputDir  = "./export"
	cursorOutputDir = "./cursors"
	svgSource       = "./breeze/cursors/Breeze/src/svg/"
	animDelayMs     = 2
)

var cursorFiles = [][]string{
	{"all-scroll"},
	{"size_fdiag"},
	{"crosshair"},
	{"not-allowed"},
	{"size_bdiag"},
	{"pointer"},
	{"default"},
	{"progress", "progress-01", "progress-02", "progress-03", "progress-04", "progress-05", "progress-06", "progress-07", "progress-08", "progress-09", "progress-10", "progress-11", "progress-12", "progress-13", "progress-14", "progress-15", "progress-16", "progress-17", "progress-18", "progress-19", "progress-20", "progress-21", "progress-22", "progress-23"},
	{"pencil"},
	{"help"},
	{"size_hor"},
	{"size_ver"},
	{"wait", "wait-01", "wait-02", "wait-03", "wait-04", "wait-05", "wait-06", "wait-07", "wait-08", "wait-09", "wait-10", "wait-11", "wait-12", "wait-13", "wait-14", "wait-15", "wait-16", "wait-17", "wait-18", "wait-19", "wait-20", "wait-21", "wait-22", "wait-23"},
	{"text"},
	{"center_ptr"},
}

func inkscapeExportCmd(input string, output string, size int) *exec.Cmd {
	return exec.Command(
		"inkscape.exe",
		"--export-background-opacity=0",
		"--export-type=png",
		fmt.Sprintf("--export-width=%d", size),
		fmt.Sprintf("--export-filename=%s", output),
		input,
	)
}

func clickGenCmd(outputDir string, hotspotX int, hotspotY int, delay int, file ...string) *exec.Cmd {
	args := []string{"-o", outputDir,
		"-p", "windows",
		"-s", "32", "64", "96", "128", "256",
		"-x", strconv.Itoa(hotspotX),
		"-y", strconv.Itoa(hotspotY),
	}
	if delay > 0 {
		args = append(args, "-d", strconv.Itoa(delay))
	}
	args = append(args, file...)
	return exec.Command(
		"clickgen",
		args...,
	)
}

func inkscapeExportHotSpotCoordinates(input string) (x float64, y float64, err error) {
	out, err := exec.Command(
		"inkscape.exe",
		"--query-id=hotspot",
		"--query-x",
		"--query-y",
		input,
	).Output()
	if err != nil {
		return 0, 0, err
	}
	idcs := strings.Split(strings.Trim(string(out), "\n"), "\n")
	if len(idcs) != 2 {
		return 0, 0, errors.New("inkscape.exe output is invalid")
	}
	x, err = strconv.ParseFloat(idcs[0], 64)
	if err != nil {
		return 0, 0, errors.New("invalid first value")
	}
	y, err = strconv.ParseFloat(idcs[1], 64)
	if err != nil {
		return 0, 0, errors.New("invalid second value")
	}
	return x, y, nil
}

// From KDE hotspot_test
// Displace the hotspot to the right and down by 1/100th of a pixel, then
// floor. So if by some float error the hotspot is at 4.995, it will be
// displaced to 5.005, then floored to 5. This is to prevent the hotspot
// from potential off-by-one errors when the cursor is scaled.
func hotspotPx(x float64, y float64, scale float64) (int, int) {
	const hotspotDisplace = 1
	return int(math.Floor(((x*scale + hotspotDisplace) * 100) / 100)), int(math.Floor(((y*scale + hotspotDisplace) * 100) / 100))
}

func ExportSvg() error {
	for _, items := range cursorFiles {
		for _, name := range items {
			imageOutFile := filepath.Join(imageOutputDir, name+".png")

			fmt.Printf("Export: %s\n", name)

			out, err := inkscapeExportCmd(filepath.Join(svgSource, name+".svg"), imageOutFile, exportSize).CombinedOutput()
			if len(out) > 0 {
				fmt.Println(string(out))
			}
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

func ExportCursors() error {
	for _, items := range cursorFiles {
		x, y, err := inkscapeExportHotSpotCoordinates(filepath.Join(svgSource, items[0]+".svg"))
		if err != nil {
			fmt.Println(err)
			continue
		}
		hotspotX, hotspotY := hotspotPx(x, y, exportSize/svgBaseSize)

		fmt.Printf("Cursor: %s\n", items[0])

		var files []string
		for _, item := range items {
			files = append(files, filepath.Join(imageOutputDir, item+".png"))
		}

		delay := 0
		// is animation
		if len(items) > 1 {
			delay = animDelayMs
		}
		out, err := clickGenCmd(cursorOutputDir, hotspotX, hotspotY, delay, files...).CombinedOutput()
		if len(out) > 0 {
			fmt.Println(string(out))
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return nil
}

func main() {
	err := os.MkdirAll(imageOutputDir, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(cursorOutputDir, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	err = ExportSvg()
	if err != nil {
		log.Fatalln(err)
	}

	err = ExportCursors()
	if err != nil {
		log.Fatalln(err)
	}
}
