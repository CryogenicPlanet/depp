package shared

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v39/github"
)

func GetInstallationToken(owner string) (*github.InstallationToken, error) {
	keyData := os.Getenv("GITHUB_PRIVATE_KEY")
	id := os.Getenv("GITHUB_APP_ID")

	tr := http.DefaultTransport

	ctx := context.Background()

	intId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	// Wrap the shared transport for use with the app ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewAppsTransport(tr, intId, []byte(keyData))
	if err != nil {
		return nil, err
	}

	// Use installation transport with github.com/google/go-github
	jwtClient := github.NewClient(&http.Client{Transport: itr})

	if err != nil {
		return nil, err
	}

	// fmt.Println("Owner", owner)

	installs, _, err := jwtClient.Apps.ListInstallations(ctx, nil)

	if err != nil {
		return nil, err
	}

	var installId int64

	for _, install := range installs {
		// fmt.Printf("Install %+v\n", install)

		if strings.EqualFold(install.Account.GetLogin(), owner) {

			installId = install.GetID()
			break
		}
	}
	if installId == 0 {
		panic("Could not find install")
	}

	token, _, err := jwtClient.Apps.CreateInstallationToken(ctx, installId, nil)

	return token, err

}
