package webpbin

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/belphemur/go-binwrapper"
)

// Config holds the configuration for webpbin operations
type Config struct {
	SkipDownload   bool
	Dest           string
	LibwebpVersion string
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	version := "1.6.0"
	return &Config{
		SkipDownload:   false,
		Dest:           getPath(version),
		LibwebpVersion: version,
	}
}

// getPath returns the path for the given version
func getPath(version string) string {
	return filepath.Join(map[string]string{
		"windows": filepath.Join(os.Getenv("APPDATA")),
		"darwin":  filepath.Join(os.Getenv("HOME"), ".cache"),
		"linux":   filepath.Join(os.Getenv("HOME"), ".cache"),
	}[runtime.GOOS], "webp", version, "bin")
}

// SetSkipDownload sets the skip download flag
func (c *Config) SetSkipDownload(skip bool) {
	c.SkipDownload = skip
}

// SetVendorPath sets the vendor path
func (c *Config) SetVendorPath(path string) {
	c.Dest = path
}

// SetLibVersion sets the libwebp version and updates the destination path
func (c *Config) SetLibVersion(version string) {
	c.LibwebpVersion = version
	c.Dest = getPath(version)
}

// LoadDefaultFromENV loads default configuration from environment variables
func (c *Config) LoadDefaultFromENV() error {
	if os.Getenv("SKIP_DOWNLOAD") == "true" {
		c.SkipDownload = true
	}

	if path := os.Getenv("VENDOR_PATH"); path != "" {
		c.Dest = path
	}

	if version := os.Getenv("LIBWEBP_VERSION"); version != "" {
		c.SetLibVersion(version)
	}

	return nil
}

// DetectUnsupportedPlatforms detects platforms without prebuilt binaries (alpine and arm).
// For these platforms libwebp tools should be built manually.
// See https://github.com/belphemur/go-webpbin/blob/master/docker/Dockerfile and https://github.com/belphemur/go-webpbin/blob/master/docker/Dockerfile.arm for details
func (c *Config) DetectUnsupportedPlatforms() {
	if runtime.GOARCH == "arm" {
		c.SkipDownload = true
	} else if runtime.GOOS == "linux" {
		output, err := os.ReadFile("/etc/issue")

		if err == nil && bytes.Contains(bytes.ToLower(output), []byte("alpine")) {
			c.SkipDownload = true
		}
	}
}

// DetectUnsupportedPlatforms detects platforms without prebuilt binaries using default config.
// This function is kept for backward compatibility.
func DetectUnsupportedPlatforms() {
	config := NewConfig()
	config.DetectUnsupportedPlatforms()
}

// createBinWrapper creates a BinWrapper with the given configuration
func createBinWrapper(config *Config) *binwrapper.BinWrapper {
	base := "https://storage.googleapis.com/downloads.webmproject.org/releases/webp/"

	b := binwrapper.NewBinWrapper().AutoExe()

	// Load defaults from environment and detect unsupported platforms
	config.LoadDefaultFromENV()
	config.DetectUnsupportedPlatforms()

	if !config.SkipDownload {
		b.Src(
			binwrapper.NewSrc().
				URL(base + "libwebp-" + config.LibwebpVersion + "-mac-arm64.tar.gz").
				Os("darwin").
				Arch("arm64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-mac-x86-64.tar.gz").
					Os("darwin").
					Arch("x64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-linux-x86-32.tar.gz").
					Os("linux").
					Arch("x86")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-linux-x86-64.tar.gz").
					Os("linux").
					Arch("x64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-linux-aarch64.tar.gz").
					Os("linux").
					Arch("arm64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-linux-aarch64.tar.gz").
					Os("linux").
					Arch("aarch64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-windows-x64.zip").
					Os("win32").
					Arch("x64")).
			Src(
				binwrapper.NewSrc().
					URL(base + "libwebp-" + config.LibwebpVersion + "-windows-x86.zip").
					Os("win32").
					Arch("x86"))
	}

	return b.Strip(2).Dest(config.Dest)
}

func createReaderFromImage(img image.Image) (io.Reader, error) {
	enc := &png.Encoder{
		CompressionLevel: png.NoCompression,
	}

	var buffer bytes.Buffer
	err := enc.Encode(&buffer, img)

	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func version(b *binwrapper.BinWrapper) (string, error) {
	b.Reset()
	err := b.Run("-version")

	if err != nil {
		return "", err
	}

	version := string(b.StdOut())
	version = strings.Replace(version, "\n", "", -1)
	version = strings.Replace(version, "\r", "", -1)
	return version, nil
}
