package webpbin

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/image/webp"
)

func init() {
	DetectUnsupportedPlatforms()
	copyFile("test-data/Example.jpg", "source.jpg")
	copyFile("test-data/sample.webp", "source.webp")
}

func downloadFile(url, target string) {
	_, err := os.Stat(target)

	if err != nil {
		resp, err := http.Get(url)

		if err != nil {
			fmt.Printf("Error while downloading test image: %v\n", err)
			panic(err)
		}

		defer resp.Body.Close()

		f, err := os.Create(target)

		if err != nil {
			panic(err)
		}

		defer f.Close()

		_, err = io.Copy(f, resp.Body)

		if err != nil {
			panic(err)
		}
	}
}

func copyFile(source, target string) {
	_, err := os.Stat(target)

	if err != nil {
		src, err := os.Open(source)

		if err != nil {
			fmt.Printf("Error while opening source file: %v\n", err)
			panic(err)
		}

		defer src.Close()

		dst, err := os.Create(target)

		if err != nil {
			panic(err)
		}

		defer dst.Close()

		_, err = io.Copy(dst, src)

		if err != nil {
			panic(err)
		}
	}
}

func TestEncodeImage(t *testing.T) {
	c := NewCWebP(nil)
	f, err := os.Open("source.jpg")
	assert.Nil(t, err)
	img, err := jpeg.Decode(f)
	assert.Nil(t, err)
	c.InputImage(img)
	c.OutputFile("target.webp")
	err = c.Run()
	assert.Nil(t, err)
	validateWebp(t)
}

func TestEncodeReader(t *testing.T) {
	c := NewCWebP(nil)
	f, err := os.Open("source.jpg")
	assert.Nil(t, err)
	c.Input(f)
	c.OutputFile("target.webp")
	err = c.Run()
	assert.Nil(t, err)
	validateWebp(t)
}

func TestEncodeFile(t *testing.T) {
	c := NewCWebP(nil)
	c.InputFile("source.jpg")
	c.OutputFile("target.webp")
	err := c.Run()
	assert.Nil(t, err)
	validateWebp(t)
}

func TestEncodeWriter(t *testing.T) {
	f, err := os.Create("target.webp")
	assert.Nil(t, err)
	defer f.Close()

	c := NewCWebP(nil)
	c.InputFile("source.jpg")
	c.Output(f)
	err = c.Run()
	assert.Nil(t, err)
	f.Close()
	validateWebp(t)
}

func TestRunWithContextCancel(t *testing.T) {
	c := NewCWebP(nil)
	pr, pw := io.Pipe()
	defer pw.Close()

	c.Input(pr)
	c.OutputFile("target_cancel.webp")
	defer os.Remove("target_cancel.webp")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- c.RunWithContext(ctx)
	}()

	// give cwebp time to start and block waiting for stdin input
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(10 * time.Second):
		t.Fatal("RunWithContext did not return after context cancellation")
	}
}

func TestRunWithContextTimeout(t *testing.T) {
	c := NewCWebP(nil)
	pr, pw := io.Pipe()
	defer pw.Close()

	c.Input(pr)
	c.OutputFile("target_timeout.webp")
	defer os.Remove("target_timeout.webp")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := c.RunWithContext(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestVersionCWebP(t *testing.T) {
	c := NewCWebP(nil)
	r, err := c.Version()
	assert.Nil(t, err)

	if _, ok := os.LookupEnv("DOCKER_ARM_TEST"); !ok {
		assert.Equal(t, "1.6.0", r)
	}
}

func validateWebp(t *testing.T) {
	defer os.Remove("target.webp")
	fSource, err := os.Open("source.jpg")
	assert.Nil(t, err)
	imgSource, err := jpeg.Decode(fSource)
	assert.Nil(t, err)
	fTarget, err := os.Open("target.webp")
	assert.Nil(t, err)
	defer fTarget.Close()
	imgTarget, err := webp.Decode(fTarget)
	assert.Nil(t, err)
	assert.Equal(t, imgSource.Bounds(), imgTarget.Bounds())
}
