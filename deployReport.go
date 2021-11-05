package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	openapiClient "github.com/go-openapi/runtime/client"
	"github.com/netlify/open-api/v2/go/models"
	"github.com/netlify/open-api/v2/go/porcelain"
	netlifyContext "github.com/netlify/open-api/v2/go/porcelain/context"
)

const (
	apiHostname = "api.netlify.com"
	apiPath     = "/api/v1"
	apiDebug    = false
	siteName    = "depp-report"
)

type ctxKey int

var netlifyToken string

const (
	apiClientKey ctxKey = 1 + iota
)

func httpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
			DisableKeepAlives:     true,
		},
	}
}

func newContext() context.Context {
	ctx := context.Background()

	// add OpenAPI Runtime credentials to context
	creds := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		r.SetHeaderParam("User-Agent", "test")
		r.SetHeaderParam("Authorization", "Bearer "+netlifyToken)
		return nil
	})
	ctx = netlifyContext.WithAuthInfo(ctx, creds)

	// create an OpenAPI transport
	transport := openapiClient.NewWithClient(apiHostname, apiPath, []string{"https"}, httpClient())
	transport.SetDebug(apiDebug)

	// create a Netlify api client and add to context
	//
	// client can be porcelain.New or porcelain.NewRetryable

	// client := porcelain.New(transport, strfmt.Default)
	client := porcelain.NewRetryable(transport, strfmt.Default, porcelain.DefaultRetryAttempts)
	ctx = context.WithValue(ctx, apiClientKey, client)

	return ctx
}

func getClient(ctx context.Context) *porcelain.Netlify {
	return ctx.Value(apiClientKey).(*porcelain.Netlify)
}

func deployToNetlify() string {
	ctx := newContext()
	client := getClient(ctx)

	sites, err := client.ListSites(ctx, nil)
	check(err)

	var currentSite models.Site
	siteExists := false

	for _, site := range sites {
		name := site.Name
		if name == siteName {
			currentSite = *site
			siteExists = true
			// Exits
			break
		}
	}

	if !siteExists {
		site, err := client.CreateSite(ctx, &models.SiteSetup{Site: models.Site{Name: siteName}}, false)
		check(err)
		currentSite = *site
	}

	deploy, err := client.DeploySite(ctx, porcelain.DeployOptions{Dir: "./.depp", SiteID: currentSite.ID})

	if err != nil {
		log.Fatal(err)
	}

	deployUrl := deploy.DeployURL

	return deployUrl

}
