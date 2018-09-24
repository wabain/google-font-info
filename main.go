package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/golang/protobuf/proto"

	"github.com/wabain/google-font-info/internal/google_fonts"
	"github.com/wabain/google-font-info/pkg/freetype_ffi"
)

func main() {
	workDir := flag.String("workdir", ".font-cache", "The working directory")
	fontRepo := flag.String("fontrepo", "https://github.com/google/fonts.git", "The Google Fonts repository")
	fontBranch := flag.String("fontbranch", "master", "The Google Fonts repository branch")

	// TODO(wabain): currently no manifest is actually generated
	manifest := flag.String("manifest", "fonts.json", "The location to write the manifest file [TODO]")

	flag.Parse()

	if err := run(*workDir, *fontRepo, *fontBranch, *manifest); err != nil {
		fmt.Fprintf(os.Stderr, "google-font-info: %v\n", err)
		os.Exit(1)
	}
}

func run(workDir, fontRepo, fontBranch, manifestPath string) error {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return err
	}

	repoPath := path.Join(workDir, "fonts")

	if err := download(repoPath, fontRepo, fontBranch); err != nil {
		return err
	}

	fontFamilies, err := loadFontFamilyMeta(repoPath, manifestPath)
	if err != nil {
		return err
	}

	metrics, err := getFontFamilyMetrics(fontFamilies)
	if err != nil {
		return err
	}

	for i, fontFamilyMetrics := range metrics {
		metadata := fontFamilies[i]

		fmt.Printf("Family %s, created %s by %s\n",
			metadata.Name,
			metadata.DateAdded,
			metadata.Designer)
		fmt.Println("Files:")

		for j, fontFaceMetrics := range fontFamilyMetrics {
			if fontFaceMetrics == nil {
				continue
			}
			fontMeta := metadata.Fonts[j]
			fmt.Printf("    %s (%s)\n", fontMeta.Path, fontMeta.FullName)
			fmt.Printf("    %#v\n", fontFaceMetrics)
		}
	}

	return nil
}

func download(repoPath, fontRepo, fontBranch string) error {
	repoExists := true

	if _, err := os.Stat(path.Join(repoPath, ".git")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		repoExists = false
	}

	if repoExists {
		return nil // XXX should check validity/update
	}

	cmd := exec.Command("git", "clone", "-b", fontBranch, fontRepo, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %v", err)
	}

	return nil
}

func loadFontFamilyMeta(repoPath, manifestPath string) ([]FontFamilyMetadata, error) {
	fontFamilies := []FontFamilyMetadata{}

	licenseDirs := []string{"apache", "ofl", "ufl"}
	for _, license := range licenseDirs {
		fontEntries, err := ioutil.ReadDir(path.Join(repoPath, license))
		if err != nil {
			return nil, err
		}

		for _, fontEntry := range fontEntries {
			fontDir := path.Join(repoPath, license, fontEntry.Name())
			metadata, err := readFontMetadata(fontDir)
			if err != nil {
				return nil, err
			}

			if metadata != nil {
				fontFamilies = append(fontFamilies, *metadata)
			}
		}
	}

	return fontFamilies, nil
}

type FontFamilyMetadata struct {
	Name      string
	DateAdded string
	Designer  string
	Aliases   []string
	Fonts     []FontMetadata
}

type FontMetadata struct {
	FullName string
	Path     string
	Weight   int32
	Style    string
}

func readFontMetadata(fontDir string) (*FontFamilyMetadata, error) {
	metadataPath := path.Join(fontDir, "METADATA.pb")
	inBytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	in := string(inBytes)

	metadata := &google_fonts.FamilyProto{}
	if err := proto.UnmarshalText(in, metadata); err != nil {
		return nil, fmt.Errorf("parsing %s: %v", metadataPath, err)
	}

	metadataOut := FontFamilyMetadata{
		Name:      *metadata.Name,
		DateAdded: *metadata.DateAdded,
		Designer:  *metadata.Designer,
		Aliases:   metadata.Aliases,
		Fonts:     []FontMetadata{},
	}

	for _, fontMeta := range metadata.Fonts {
		metadataOut.Fonts = append(metadataOut.Fonts, FontMetadata{
			FullName: *fontMeta.FullName,
			Path:     path.Join(fontDir, *fontMeta.Filename),
			Weight:   *fontMeta.Weight,
			Style:    *fontMeta.Style,
		})
	}

	return &metadataOut, nil
}

func getFontFamilyMetrics(fontFamilies []FontFamilyMetadata) ([][]*freetype_ffi.FaceMetrics, error) {
	freetype, err := freetype_ffi.FreetypeInit()
	if err != nil {
		return nil, err
	}
	defer freetype_ffi.FreetypeDone(freetype)

	results := [][]*freetype_ffi.FaceMetrics{}

	for _, familyMeta := range fontFamilies {
		familyMetrics := []*freetype_ffi.FaceMetrics{}

		for _, fontMeta := range familyMeta.Fonts {
			face, err := freetype_ffi.GetFaceMetrics(freetype, fontMeta.Path)
			if err == nil {
				familyMetrics = append(familyMetrics, face)
			} else {
				familyMetrics = append(familyMetrics, nil)
				fmt.Fprintf(os.Stderr,
					"google-font-info: reading %s: %v\n",
					fontMeta.Path,
					err)
			}
		}

		results = append(results, familyMetrics)
	}

	return results, nil
}
