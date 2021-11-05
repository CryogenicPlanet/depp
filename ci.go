package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

var githubToken string
var ciMode bool
var repo string = "cryogenicplanet.github.io"
var owner string = "CryogenicPlanet"
var issue int = 49 // ISSUE AND PR ARE THE SAME FOR GITHUB API PURPOSES
var markdownString string

type Issue struct {
	Number int `json:"number"`
}

//  Type for this payload are defined here
//  https://github.com/actions/toolkit/blob/e2eeb0a784f4067a75f0c6cd2cc9703f3cbc7744/packages/github/src/interfaces.ts#L15
type Payload struct {
	Issues      Issue `json:"issues"`
	PullRequest Issue `json:"pull_request"`
}

type IssueComment struct {
	Body string `json:"body"`
	Id   int64  `json:"id"`
}

const DEEP_REPORT_TITLE = "# Depp Report"

func checkPrComments(client github.Client, ctx context.Context) int64 {

	issueData, _, err := client.Issues.Get(ctx, owner, repo, issue)

	check(err)

	// fmt.Printf("%+v\n", issueData)

	url, err := url.Parse(issueData.GetCommentsURL())

	check(err)

	response, err := client.BareDo(ctx, &http.Request{Method: "GET", URL: url})

	check(err)

	body, err := ioutil.ReadAll(response.Body)

	check(err)

	issueComments := []IssueComment{}

	json.Unmarshal(body, &issueComments)

	for _, comment := range issueComments {
		if strings.Contains(comment.Body, DEEP_REPORT_TITLE) {
			return comment.Id
		}
	}
	return -1
}

func makePrComment(deployUrl string) {
	setGithubRepoFromEnv()
	setIssueNumberFromEnv()

	fmt.Println("Git env is", owner, repo, issue)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	body := DEEP_REPORT_TITLE + "\n  Report Deploy URL is " + deployUrl + "\n \n" + markdownString

	commentId := checkPrComments(*client, ctx)

	if commentId == -1 {
		_, _, err := client.Issues.CreateComment(ctx, owner, repo, issue, &github.IssueComment{Body: &body})
		check(err)

	} else {
		_, _, err := client.Issues.EditComment(ctx, owner, repo, commentId, &github.IssueComment{Body: &body})
		check(err)
	}

}

func setGithubRepoFromEnv() {
	repoUrl := os.Getenv("GITHUB_REPOSITORY")

	if repoUrl == "" {
		panic("ENV GITHUB_REPOSITORY not found, do not use -ci in local environment")
	}

	splits := strings.Split(repoUrl, "/")

	owner = splits[0]
	repo = splits[1]
}

func setIssueNumberFromEnv() {
	prNumber := os.Getenv("PR_NUMBER")

	if prNumber == "" {
		panic("ENV PR_NUMBER not set, please set ENV PR_NUMBER when using -ci")
	}

	no, err := strconv.Atoi(prNumber)
	if err != nil {
		// handle error
		fmt.Println(err)
		os.Exit(2)
	}
	issue = no

}
