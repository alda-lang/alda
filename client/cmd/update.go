package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"alda.io/client/generated"
	"alda.io/client/help"
	"alda.io/client/json"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"alda.io/client/text"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

var assumeYes bool
var desiredVersion string

func init() {
	updateCmd.Flags().BoolVarP(
		&assumeYes,
		"yes",
		"y",
		false,
		"Do not prompt for confirmation before updating",
	)

	updateCmd.Flags().StringVar(
		&desiredVersion,
		"version",
		"",
		"The version to update to (default: latest)",
	)
}

func displayVersionInfo(json *json.Container) {
	version, ok := json.Search("version").Data().(string)
	if !ok {
		version = "<unknown version>"
	}

	date, ok := json.Search("date").Data().(string)
	if !ok {
		date = "<unknown date>"
	}

	changelog, ok := json.Search("changelog").Data().(string)
	if !ok {
		changelog = "<changelog missing>"
	}

	fmt.Println(
		text.Boxed(
			fmt.Sprintf(
				`%s (%s)

%s
%s`,
				aurora.Bold("Alda "+version),
				date,
				aurora.Bold("Changelog:"),
				text.Indent(1, changelog),
			),
		),
	)
}

type releaseAsset struct {
	assetName string
	assetType string
	assetUrl  string
}

func downloadAssets(
	outdir string, version string, assets []releaseAsset,
) error {
	fmt.Printf(
		"%s",
		aurora.Bold(fmt.Sprintf("Downloading Alda %s...\n", version)),
	)

	maxAssetNameLength := -1
	for _, asset := range assets {
		nameLength := len(asset.assetName)
		if nameLength > maxAssetNameLength {
			maxAssetNameLength = nameLength
		}
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(assets))

	progress := mpb.New(
		mpb.WithWidth(25),
		mpb.WithRefreshRate(100*time.Millisecond),
		mpb.WithWaitGroup(&waitGroup),
	)

	var downloadError error

	for _, asset := range assets {
		client := &http.Client{Timeout: 10 * time.Second}
		response, err := client.Get(asset.assetUrl)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Error().Err(err).Msg("Unable to read response body")
			} else {
				log.Error().
					Bytes("body", responseBody).
					Int("status", response.StatusCode).
					Msg("Asset HTTP response")
			}

			return fmt.Errorf("unexpected HTTP response (%d)", response.StatusCode)
		}

		bar := progress.Add(
			response.ContentLength,
			mpb.NewBarFiller("[·· ]"),
			mpb.PrependDecorators(
				decor.Name(
					fmt.Sprintf(
						"  %s%s",
						asset.assetName,
						strings.Repeat(" ", maxAssetNameLength-len(asset.assetName)),
					),
				),
			),
			mpb.AppendDecorators(
				decor.CurrentKibiByte("% .2f"),
			),
		)

		proxyReader := bar.ProxyReader(response.Body)
		defer proxyReader.Close()

		out, err := os.Create(filepath.Join(outdir, asset.assetName))
		if err != nil {
			return err
		}
		defer out.Close()

		go func() {
			defer waitGroup.Done()

			if _, err := io.Copy(out, proxyReader); err != nil {
				downloadError = err
			}
		}()
	}

	progress.Wait()

	return downloadError
}

func installAsset(indir string, outdir string, asset releaseAsset) error {
	if asset.assetType != "executable" {
		return fmt.Errorf(
			"unexpected asset type: %s (%s)", asset.assetType, asset.assetName,
		)
	}

	fmt.Printf(
		"%s\n",
		aurora.Bold(fmt.Sprintf("Installing %s...", asset.assetName)),
	)

	// Naïvely, you would think we could simply replace `alda` with the new
	// version that we downloaded. However, this doesn't always work, especially
	// given that we're currently _running_ the old version in this very process.
	//
	// We're doing something safer here, following the advice here:
	// https://stackoverflow.com/a/7198760/2338327
	//
	// (That StackOverflow question is specifically about Windows, but this is a
	// safer method in general, and it works regardless of OS.)

	inpath := filepath.Join(indir, asset.assetName)
	outpath := filepath.Join(outdir, asset.assetName)

	oldRenamed := system.RenamedExecutable(outpath)

	// Move the old (and probably currently running) executable, giving it a new
	// name that also designates it as an old version to be cleaned up.
	//
	// NOTE: The Alda CLI checks for the existence of these old executable
	// versions and cleans them up on each run.
	log.Debug().
		Str("old-name", outpath).
		Str("new-name", oldRenamed).
		Msg("Renaming existing executable.")

	err := os.Rename(outpath, oldRenamed)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Now that we've moved the existing executable, we create a new one at the
	// same location and copy the new version into it.

	log.Debug().
		Str("source", inpath).
		Msg("Opening source.")

	in, err := os.Open(inpath)
	if err != nil {
		return err
	}
	defer in.Close()

	log.Debug().
		Str("destination", outpath).
		Msg("Opening destination.")

	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer out.Close()

	log.Debug().Msg("Copying file.")

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	// The copy was successful, so now we can remove the temporary file that we
	// copied in. (It isn't really a problem if this fails, because it's a temp
	// file, so we're ignoring `err` here.)
	_ = os.Remove(inpath)

	if runtime.GOOS != "windows" {
		// Here, we make the file executable. The code below should be equivalent to
		// running `chmod +x ...` on the file, i.e. ensuring that the "execute" bit
		// is present for each permissions digit (owner, group, others) in this
		// octal number representing the file permissions.
		//
		// Reference:
		// https://en.wikipedia.org/wiki/File-system_permissions#Numeric_notation
		//
		// The execute bit adds 1 to the total, so to ensure that each digit
		// includes the execute bit, we do a bitwise OR operation with 0111 as the
		// operand.
		fileInfo, err := out.Stat()
		if err != nil {
			return err
		}
		perms := fileInfo.Mode().Perm()

		log.Debug().
			Str("filename", outpath).
			Str("perms", fmt.Sprintf("%#o", perms)).
			Msg("Before making file executable")

		if err := os.Chmod(outpath, perms|0111); err != nil {
			return err
		}

		fileInfo, err = out.Stat()
		if err != nil {
			return err
		}
		perms = fileInfo.Mode().Perm()

		log.Debug().
			Str("filename", outpath).
			Str("perms", fmt.Sprintf("%#o", perms)).
			Msg("After making file executable")
	}

	return nil
}

func installVersion(json *json.Container) error {
	version, ok := json.Search("version").Data().(string)
	if !ok {
		version = "<unknown version>"
	}

	osAndArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	assetsInfo, ok := json.Search("assets", osAndArch).Data().([]interface{})
	if !ok {
		return fmt.Errorf(
			"no assets found for version: %s, os/arch: %s", version, osAndArch,
		)
	}

	assets := []releaseAsset{}

	for _, assetInfo := range assetsInfo {
		errUnexpectedAsset := fmt.Errorf("unexpected asset: %#v", assetInfo)

		m, ok := assetInfo.(map[string]interface{})
		if !ok {
			return errUnexpectedAsset
		}

		assetName, ok := m["name"].(string)
		if !ok {
			return errUnexpectedAsset
		}

		assetType, ok := m["type"].(string)
		if !ok {
			return errUnexpectedAsset
		}

		assetUrl, ok := m["url"].(string)
		if !ok {
			return errUnexpectedAsset
		}

		assets = append(assets, releaseAsset{
			assetName: assetName,
			assetType: assetType,
			assetUrl:  assetUrl,
		})
	}

	tmpdir, err := ioutil.TempDir("", "alda-update")
	if err != nil {
		return err
	}

	if err := downloadAssets(tmpdir, version, assets); err != nil {
		return err
	}

	aldaPath, err := system.AldaExecutablePath()
	if err != nil {
		return err
	}

	for _, asset := range assets {
		fmt.Println()
		if err := installAsset(tmpdir, filepath.Dir(aldaPath), asset); err != nil {
			return err
		}
	}

	return nil
}

func installDesiredVersion(version string) error {
	fmt.Println("Fetching information about Alda releases...")
	fmt.Println()

	response, err := (&http.Client{Timeout: 30 * time.Second}).Get(
		aldaApiUrl("/release/" + version),
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200: // OK to proceed
	case 404:
		return help.UserFacingErrorf(
			`The requested Alda version, %s, was not found.`,
			aurora.Bold(version),
		)
	default:
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error().Err(err).Msg("Unable to read Alda API response body")
		} else {
			log.Error().
				Bytes("body", responseBody).
				Int("status", response.StatusCode).
				Msg("Alda API response")
		}

		return fmt.Errorf("unexpected Alda API response (%d)", response.StatusCode)
	}

	json, err := json.ParseJSONBuffer(response.Body)
	if err != nil {
		return err
	}

	displayVersionInfo(json)
	fmt.Println()

	if !assumeYes && !text.PromptForConfirmation(
		fmt.Sprintf(
			"Alda %s is currently installed. Install version %s?",
			aurora.Bold(generated.ClientVersion),
			aurora.Bold(version),
		),
		true, // default to "yes"
	) {
		fmt.Println("Aborting.")
		return nil
	}

	fmt.Println()

	return installVersion(json)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to the latest version of Alda",
	RunE: func(_ *cobra.Command, args []string) error {
		if desiredVersion != "" {
			return installDesiredVersion(desiredVersion)
		}

		return fmt.Errorf("TODO: implement checking for updates")
	},
}
