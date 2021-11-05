package serverless

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cryogenicplanet/depp/serverless/shared"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	owner := query.Get("owner")
	repo := query.Get("repo")
	commentBody := query.Get("commentBody")
	commentId := query.Get("commentId")
	issue := query.Get("issue")

	if owner == "" || repo == "" || commentBody == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Invalid params")
		return
	}

	if commentId == "" && issue == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Invalid params")
		return
	}

	issueNumber, issueErr := strconv.Atoi(issue)
	commentNumber, commentErr := strconv.ParseInt(commentId, 10, 64)

	if issueErr != nil && commentErr != nil {
		fmt.Println(issueErr, commentErr)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Could not parse issue number or comment id")
		return
	}

	if issueErr != nil {
		issueNumber = -1
	}

	if commentErr != nil {
		commentNumber = -1
	}

	token, err := shared.GetInstallationToken(owner)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
	}

	tokenCtx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	tc := oauth2.NewClient(tokenCtx, ts)

	tokenClient := github.NewClient(tc)

	if issueNumber != -1 {
		_, _, err = tokenClient.Issues.CreateComment(tokenCtx, owner, repo, issueNumber, &github.IssueComment{Body: &commentBody})
	} else {
		_, _, err = tokenClient.Issues.EditComment(tokenCtx, owner, repo, commentNumber, &github.IssueComment{Body: &commentBody})
	}

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Could not get create/edit comment", err)
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Success")
}
