package mirror

import (
	"fmt"
	dockerparser "github.com/novln/docker-parser"
	"strings"
)

// Manifest for a container
type Manifest struct {
	Image         string                  `json:"-"`
	Destination   string                  `json:"-"`
	SourceImage   *dockerparser.Reference `json:"-"`
	MirrorImage   *dockerparser.Reference `json:"-"`
	SchemaVersion int                     `json:"schemaVersion"`
	MediaType     string                  `json:"mediaType"`
	Manifests     []*ManifestEntry        `json:"manifests"`
	Simple        bool                    `json:"-"`
}

// ManifestEntry one entry per architecture
type ManifestEntry struct {
	MediaType string   `json:"mediaType"`
	Size      int      `json:"size"`     // Size of image
	Digest    string   `json:"digest"`   // Digest hash
	Platform  Platform `json:"platform"` // Platform of image
	SrcImage  string   `json:"-"`        // Source image reference
	DstImage  string   `json:"-"`        // Mirror image reference
}

type Platform struct {
	OS           string `json:"os"`           // Operating System
	Architecture string `json:"architecture"` // CPU Architecture
	Variant      string `json:"variant"`      // CPU Variant
}

func MirrorImage(image, dest string) error {
	manifest := &Manifest{
		Image:       image,
		Destination: dest,
	}

	return RunTasks(
		manifest.begin,
		manifest.retrieveManifest,
		manifest.resolveImages,
		manifest.pullImages,
		manifest.tagImages,
		manifest.pushImages,
		manifest.createNewManifest,
		manifest.pushManifest,
		manifest.complete,
	)
}

// ForEach runs a function for every linux artifact in the image
func (m *Manifest) ForEach(f func(m *ManifestEntry) error) error {
	for _, manifest := range m.Manifests {
		if manifest.Platform.OS == "linux" {
			err := f(manifest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// begin the Manifest struct with src & dest images
func (m *Manifest) begin() error {
	src, err := dockerparser.Parse(m.Image)
	if err != nil {
		return err
	}
	m.SourceImage = src

	dest, err := dockerparser.Parse(strings.Replace(m.SourceImage.Remote(),
		m.SourceImage.Registry(), m.Destination, 1))
	if err != nil {
		return err
	}
	m.MirrorImage = dest

	fmt.Printf("Beginning mirror of %s\n", src.Remote())

	return nil
}

// complete the manifest
func (m *Manifest) complete() error {
	fmt.Printf("Completed mirror of %s as %s\n\n", m.SourceImage.Remote(), m.MirrorImage.Remote())
	return nil
}

// retrieveManifest the manifest from the source repository
func (m *Manifest) retrieveManifest() error {
	fmt.Printf("Retrieving manifest for %s\n", m.SourceImage.Remote())

	err := ExecJson(m, "docker", "manifest", "inspect", m.SourceImage.Remote())
	if err != nil {
		return err
	}

	// Log the found platforms
	var platforms []string
	_ = m.ForEach(func(me *ManifestEntry) error {
		platforms = append(platforms, me.Platform.Architecture+me.Platform.Variant)
		return nil
	})
	fmt.Printf("Found architectures: %s\n", strings.Join(platforms, ", "))

	return nil
}

// resolveImages manifest image names
func (m *Manifest) resolveImages() error {

	// No manifests then presume it's a traditional amd64 image
	if len(m.Manifests) == 0 {
		m.Manifests = append(m.Manifests, &ManifestEntry{
			MediaType: "",
			Size:      0,
			Digest:    "",
			Platform: Platform{
				OS:           "linux",
				Architecture: "amd64",
				Variant:      "",
			},
			SrcImage: m.SourceImage.Remote(),
			DstImage: m.MirrorImage.Remote(),
		})

		// Mark as a simple container
		m.Simple = true

		return nil
	}

	return m.ForEach(func(me *ManifestEntry) error {
		me.SrcImage = m.SourceImage.Remote() + "@" + me.Digest

		a := []string{
			m.MirrorImage.Remote(),
			me.Platform.Architecture,
		}

		// arm has v5, v7 variants we need to handle.
		// arm64 is v8 but we can ignore for now
		if me.Platform.Variant != "" && me.Platform.Architecture != "arm64" {
			a = append(a, me.Platform.Variant)
		}

		me.DstImage = strings.Join(a, "_")

		return nil
	})
}

// pullImages images in the manifest
func (m *Manifest) pullImages() error {
	fmt.Println("Pulling images")
	return m.ForEach(func(me *ManifestEntry) error {
		fmt.Printf("Pulling %s\n", me.SrcImage)
		return Exec("docker", "pull", me.SrcImage)
	})
}

// tagImages tags all images to the mirror repository
func (m *Manifest) tagImages() error {
	fmt.Println("Tagging images")
	return m.ForEach(func(me *ManifestEntry) error {
		return Exec("docker", "tag", me.SrcImage, me.DstImage)
	})
}

// pushImages images in the manifest to the destination
func (m *Manifest) pushImages() error {
	fmt.Println("Pushing images")
	return m.ForEach(func(me *ManifestEntry) error {
		fmt.Printf("Pushing %s\n", me.DstImage)
		return Exec("docker", "push", me.DstImage)
	})
}

// Create manifest appends tasks to create the new manifest
func (m *Manifest) createNewManifest() error {

	// Not applicable for simple mode
	if m.Simple {
		return nil
	}

	multiImage := m.MirrorImage.Remote()

	fmt.Printf("Creating manifest for %s\n", multiImage)

	args := []string{"manifest", "create", "-a", multiImage}

	_ = m.ForEach(func(me *ManifestEntry) error {
		args = append(args, me.DstImage)
		return nil
	})

	err := Exec("docker", args...)
	if err != nil {
		return err
	}

	// Annotate each image
	return m.ForEach(func(me *ManifestEntry) error {
		args := []string{
			"manifest", "annotate",
			"--os", me.Platform.OS,
			"--arch", me.Platform.Architecture,
		}

		if me.Platform.Architecture == "arm" {
			args = append(args, "--variant", me.Platform.Variant)
		}

		args = append(args, multiImage, me.DstImage)

		fmt.Printf("Annotating %s\n", me.DstImage)
		return Exec("docker", args...)
	})
}

func (m *Manifest) pushManifest() error {

	// Not applicable for simple mode
	if m.Simple {
		return nil
	}

	multiImage := m.MirrorImage.Remote()
	fmt.Printf("Pushing %s\n", multiImage)
	return Exec("docker", "manifest", "push", "-p", multiImage)
}
