/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/application"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	gc "github.com/untillpro/gochips"
)

var (
	gToken string
	gURL   string
)

type gitTrackerGotify struct {
	notfications chan string
	not          map[string]chan string
}

// note: websocket stream
func (wcn *gitTrackerGotify) GetLastCommit(repoURL string, repoPath string) (lastCommit string, ok bool) {
	gotifyClient := getGotifyClient(gURL)
	authInfo := auth.TokenAuth(gToken)
	appsResponse, err := gotifyClient.Application.GetApps(nil, authInfo)
	gc.PanicIfError(err)
	app := findApp(appsResponse.Payload, repoURL)
	if app == nil {
		// create application
		createAppParams := application.NewCreateAppParams().WithBody(&models.Application{
			Name:        repoURL,
			Description: "Created by cder " + time.Now().Format(time.RFC3339),
		})
		createAppResponse, err := gotifyClient.Application.CreateApp(createAppParams, authInfo)
		gc.PanicIfError(err)
		app = createAppResponse.Payload
		printPushVerCommand(app)
	}

	appMessagesParams := message.NewGetAppMessagesParams()
	appMessagesParams.ID = int64(app.ID)
	limit := int64(1)
	appMessagesParams.Limit = &limit
	appMessagesResponse, err := gotifyClient.Message.GetAppMessages(appMessagesParams, authInfo)
	gc.PanicIfError(err)
	ok = len(appMessagesResponse.Payload.Messages) != 0
	if ok {
		lastCommit = appMessagesResponse.Payload.Messages[0].Title
	}
	return
}

func getGotifyClient(rawURL string) *client.GotifyREST {
	parsedURL, err := url.Parse(rawURL)
	gc.PanicIfError(err)
	return gotify.NewClient(parsedURL, &http.Client{})
}

func findApp(apps []*models.Application, appName string) *models.Application {
	for _, app := range apps {
		if app.Name == appName {
			return app
		}
	}
	return nil
}

func printPushVerCommand(app *models.Application) {
	// $ curl "https://push.example.de/message?token=<apptoken>" -F "title=<version>" -F "message=<url>"
	parsedURL, err := url.Parse(gURL)
	gc.PanicIfError(err)
	parsedURL.Path = path.Join(parsedURL.Path, "message")
	{
		pq, err := url.ParseQuery(parsedURL.RawQuery)
		gc.PanicIfError(err)
		pq.Add("token", app.Token)
		parsedURL.RawQuery = pq.Encode()
	}
	log.Println("Command for push versions: curl \"" + parsedURL.String() + "\" -F \"title=<version>\" -F \"message=<url>\"")
}
