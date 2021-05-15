package deglacer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/MH4GF/notion-deglacer/notion"
	"github.com/kjk/notionapi"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/sync/errgroup"
)

func Run(argv []string) error {
	if err := initialize(); err != nil {
		return err
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var eg errgroup.Group

	srv := &http.Server{Addr: ":" + port, Handler: http.HandlerFunc(index)}
	eg.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	sig := <-c
	log.Printf("received signal %s, shutting down\n", sig)
	eg.Go(func() error {
		return srv.Shutdown(context.Background())
	})
	return eg.Wait()
}

var (
	oldNotionClient    *notionapi.Client
	notionClient       *notion.Client
	slackCli           *slack.Client
	slackSigningSecret string
)

func initialize() error {
	notionToken := os.Getenv("NOTION_TOKEN")
	oldNotionClient = &notionapi.Client{}
	notionClient = &notion.Client{
		AuthToken: notionToken,
	}
	slackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET")
	if slackSigningSecret == "" {
		return errors.New("env SLACK_SIGNING_SECRET required")
	}
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return errors.New("env SLACK_TOKEN is empty")
	}
	slackCli = slack.New(slackToken)
	return nil
}

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello")
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv.Write(body)
		if err := sv.Ensure(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ev, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch ev.Type {
		case slackevents.URLVerification:
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case slackevents.CallbackEvent:
			if ev.InnerEvent.Type != slackevents.LinkShared {
				fmt.Fprint(w, "ok")
				return
			}
			inEv, ok := ev.InnerEvent.Data.(*slackevents.LinkSharedEvent)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			go unfurl(inEv)
			fmt.Fprint(w, "ok")
			return
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func unfurl(ev *slackevents.LinkSharedEvent) {
	unfurls := make(map[string]slack.Attachment, len(ev.Links))

	for _, link := range ev.Links {
		if !strings.HasSuffix(link.Domain, ".notion.so") {
			continue
		}
		u, err := url.Parse(link.URL)
		if err != nil {
			log.Println(err)
			continue
		}

		// official notion api does not provide retrieving pageID
		// because notionapi library is only used to extract pageID
		// notionapi can't parse query parameter
		u.RawQuery = ""
		u.Fragment = ""
		pageID := notionapi.ExtractNoDashIDFromNotionURL(u.String())

		page, err := notionClient.RetrievePage(pageID)
		if err != nil {
			log.Println(err)
			continue
		}

		title := page.PageTitle()
		if title == "" {
			log.Println("title is not found")
		}
		fmt.Println(title)

		// TODO: Add the text in the first block of the page
		unfurls[link.URL] = slack.Attachment{
			Title:     title,
			TitleLink: link.URL,
			Footer:    "Notion",
		}
	}

	if len(unfurls) == 0 {
		return
	}

	_, _, err := slackCli.PostMessage(ev.Channel, slack.MsgOptionUnfurl(ev.MessageTimeStamp.String(), unfurls))
	if err != nil {
		log.Println(err)
	}
}
