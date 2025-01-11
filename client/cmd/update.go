package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"alda.io/client/color"
	"alda.io/client/generated"
	"alda.io/client/help"
	"alda.io/client/json"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"alda.io/client/text"
	"github.com/AlecAivazis/survey/v2"
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
				color.Aurora.Bold("Alda "+version),
				date,
				color.Aurora.Bold("Changelog:"),
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

func downloadAssets(assets []releaseAsset) (assetsDir string, err error) {
	outdir, err := os.MkdirTemp("", "alda-update")
	if err != nil {
		return "", err
	}

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
			return "", err
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				log.Error().Err(err).Msg("Unable to read response body")
			} else {
				log.Error().
					Bytes("body", responseBody).
					Int("status", response.StatusCode).
					Msg("Asset HTTP response")
			}

			return "", fmt.Errorf(
				"unexpected HTTP response (%d)", response.StatusCode,
			)
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
			return "", err
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

	if downloadError != nil {
		return "", downloadError
	}

	return outdir, nil
}

func installAsset(indir string, asset releaseAsset) error {
	if asset.assetType != "executable" {
		return fmt.Errorf(
			"unexpected asset type: %s (%s)", asset.assetType, asset.assetName,
		)
	}

	aldaPath, err := system.AldaExecutablePath()
	if err != nil {
		return err
	}

	outdir := filepath.Dir(aldaPath)

	fmt.Printf(
		"%s\n",
		color.Aurora.Bold(fmt.Sprintf("Installing %s...", asset.assetName)),
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

	err = os.Rename(outpath, oldRenamed)
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

func downloadAndInstallAssets(assets []releaseAsset) error {
	assetsDir, err := downloadAssets(assets)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		fmt.Println()
		if err := installAsset(assetsDir, asset); err != nil {
			return err
		}
	}

	return nil
}

func versionString(json *json.Container) string {
	version, ok := json.Search("version").Data().(string)
	if !ok {
		return "<unknown version>"
	}

	return version
}

func dateString(json *json.Container) string {
	date, ok := json.Search("date").Data().(string)
	if !ok {
		return "<unknown date>"
	}

	return date
}

func versionAssets(json *json.Container) ([]releaseAsset, error) {
	version := versionString(json)

	osAndArch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	assetsInfo, ok := json.Search("assets", osAndArch).Data().([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"no assets found for version: %s, os/arch: %s", version, osAndArch,
		)
	}

	assets := []releaseAsset{}

	for _, assetInfo := range assetsInfo {
		errUnexpectedAsset := fmt.Errorf("unexpected asset: %#v", assetInfo)

		m, ok := assetInfo.(map[string]interface{})
		if !ok {
			return nil, errUnexpectedAsset
		}

		assetName, ok := m["name"].(string)
		if !ok {
			return nil, errUnexpectedAsset
		}

		assetType, ok := m["type"].(string)
		if !ok {
			return nil, errUnexpectedAsset
		}

		assetUrl, ok := m["url"].(string)
		if !ok {
			return nil, errUnexpectedAsset
		}

		assets = append(assets, releaseAsset{
			assetName: assetName,
			assetType: assetType,
			assetUrl:  assetUrl,
		})
	}

	return assets, nil
}

func installVersion(json *json.Container) error {
	assets, err := versionAssets(json)
	if err != nil {
		return err
	}

	fmt.Printf(
		"%s",
		color.Aurora.Bold(fmt.Sprintf("Downloading Alda %s...\n", versionString(json))),
	)

	return downloadAndInstallAssets(assets)
}

func errUnexpectedAldaApiResponse(response *http.Response) error {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("Unable to read Alda API response body")
	} else {
		log.Error().
			Bytes("body", responseBody).
			Int("status", response.StatusCode).
			Msg("Alda API response")
	}

	return fmt.Errorf(
		"unexpected Alda API response (%d)",
		response.StatusCode,
	)
}

func errUnexpectedAldaApiResponseJson(json *json.Container) error {
	return fmt.Errorf("unexpected Alda API response: %#v", json)
}

func fetchLatestReleasesInfo() (*json.Container, error) {
	fmt.Println("Fetching information about Alda releases...")
	fmt.Println()

	response, err := (&http.Client{Timeout: 30 * time.Second}).Get(
		aldaApiUrl("/releases/latest?from-version=" + generated.ClientVersion),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errUnexpectedAldaApiResponse(response)
	}

	return json.ParseJSONBuffer(response.Body)
}

func fetchReleaseInfo(version string) (*json.Container, error) {
	fmt.Println("Fetching information about Alda releases...")
	fmt.Println()

	response, err := (&http.Client{Timeout: 30 * time.Second}).Get(
		aldaApiUrl("/release/" + version),
	)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200: // OK to proceed
	case 404:
		return nil, help.UserFacingErrorf(
			`The requested Alda version, %s, was not found.`,
			color.Aurora.Bold(version),
		)
	default:
		return nil, errUnexpectedAldaApiResponse(response)
	}

	return json.ParseJSONBuffer(response.Body)
}

func promptAndInstallVersion(json *json.Container) error {
	displayVersionInfo(json)
	fmt.Println()

	if !assumeYes && !text.PromptForConfirmation(
		fmt.Sprintf(
			"Alda %s is currently installed. Install version %s?",
			color.Aurora.Bold(generated.ClientVersion),
			color.Aurora.Bold(versionString(json)),
		),
		true, // default to "yes"
	) {
		fmt.Println("Aborting.")
		return nil
	}

	fmt.Println()

	return installVersion(json)
}

func installCorrectAldaPlayerVersion() error {
	json, err := fetchReleaseInfo(generated.ClientVersion)
	if err != nil {
		return err
	}

	assets, err := versionAssets(json)
	if err != nil {
		return err
	}

	version := versionString(json)

	fmt.Printf(
		"%s",
		color.Aurora.Bold(fmt.Sprintf("Downloading alda-player %s...\n", version)),
	)

	var aldaPlayerAsset releaseAsset
	for _, asset := range assets {
		if asset.assetName == "alda-player" ||
			asset.assetName == "alda-player.exe" {
			aldaPlayerAsset = asset
		}
	}

	if aldaPlayerAsset.assetName == "" {
		return fmt.Errorf(
			"release %s seems to be missing an alda-player asset",
			version,
		)
	}

	return downloadAndInstallAssets([]releaseAsset{aldaPlayerAsset})
}

func checkForUpdates() error {
	apiResponse, err := fetchLatestReleasesInfo()
	if err != nil {
		return err
	}

	explanation, ok := apiResponse.Search("explanation").Data().(string)
	if ok {
		fmt.Println(text.Boxed(explanation))
		fmt.Println()
	}

	releasesList := apiResponse.Search("releases")
	if releasesList == nil {
		// API response JSON missing a "releases" key
		return errUnexpectedAldaApiResponseJson(apiResponse)
	}

	releases := releasesList.Children()
	if releases == nil {
		// "releases" is present, but not a list (i.e. doesn't have children)
		return errUnexpectedAldaApiResponseJson(apiResponse)
	}

	switch len(releases) {
	case 0:
		fmt.Println("Alda is up to date.")
		return nil
	case 1:
		return promptAndInstallVersion(releases[0])
	default:
		fmt.Println("There are multiple newer releases available.")
		fmt.Println()

		options := []string{}
		versions := map[string]*json.Container{}

		for _, release := range releases {
			option := fmt.Sprintf(
				"Alda %s (%s)", versionString(release), dateString(release),
			)
			options = append(options, option)
			versions[option] = release
		}

		selectedVersion := ""

		survey.AskOne(
			&survey.Select{
				Message: "Which version would you like to install?",
				Options: options,
			},
			&selectedVersion,
		)

		fmt.Println()

		return promptAndInstallVersion(versions[selectedVersion])
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to the latest version of Alda",
	RunE: func(_ *cobra.Command, args []string) error {
		if desiredVersion != "" {
			json, err := fetchReleaseInfo(desiredVersion)
			if err != nil {
				return err
			}
			return promptAndInstallVersion(json)
		}

		return checkForUpdates()
	},
}
