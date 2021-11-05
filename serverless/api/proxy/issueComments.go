package serverless

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cryogenicplanet/depp/serverless/shared"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	owner := query.Get("owner")
	repo := query.Get("repo")
	issue, err := strconv.Atoi(query.Get("issue"))

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Could not parse issue number")
		return
	}

	token, err := shared.GetInstallationToken(owner)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return

	}

	tokenCtx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	tc := oauth2.NewClient(tokenCtx, ts)

	tokenClient := github.NewClient(tc)

	issueData, _, err := tokenClient.Issues.Get(tokenCtx, owner, repo, issue)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not get issue data")
		return

	}

	link, err := url.Parse(issueData.GetCommentsURL())

	if err != nil {
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not get parse comment url")
		return

	}

	response, err := tokenClient.BareDo(tokenCtx, &http.Request{URL: link})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not get comments")
		return

	}

	// issue, _, err := client.Issues.Get(tokenCtx, "cryogenicplanet", "cryogenicplanet.github.io", 49)

	// fmt.Fprintln(w, issue.GetCommentsURL())

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not read response")
		return
	}

	headers := response.Header.Clone()

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Write(body)
}
