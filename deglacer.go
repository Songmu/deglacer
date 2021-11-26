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
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/Songmu/kibelasync/kibela"
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
	kibelaCli          *kibela.Kibela
	slackCli           *slack.Client
	kibelaTeam         string
  kibelaToken        string
	slackSigningSecret string
)

func initialize() error {
	// KIBELA_TOKEN and KIBELA_TEAM are required
	var err error
	kibelaCli, err = kibela.New("0.0.1+deglacer")
	if err != nil {
		return err
	}
	kibelaTeam = os.Getenv("KIBELA_TEAM")
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
		fmt.Fprintf(w, "Hello! (deglacer version: %s, rev: %s)", version, revision)
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
			if err := callback(r.Context(), inEv); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, "ok")
			return
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

var (
	noteReg     = regexp.MustCompile(`^/(?:@[^/]+|notes)/([0-9]+)`)
	fragmentReg = regexp.MustCompile(`(?i)^comment_([0-9]+)`)
	spacesReg   = regexp.MustCompile(`\s+`)
)

func callback(ctx context.Context, ev *slackevents.LinkSharedEvent) error {
	unfurls := make(map[string]slack.Attachment, len(ev.Links))

	for _, link := range ev.Links {
		if !strings.HasSuffix(link.Domain, ".kibe.la") {
			continue
		}
		u, err := url.Parse(link.URL)
		if err != nil {
			log.Println(err)
			continue
		}
		m := noteReg.FindStringSubmatch(u.Path)
		if len(m) < 2 {
			continue
		}
		id, _ := strconv.Atoi(m[1])
		note, err := kibelaCli.GetNote(ctx, id)
		if err != nil {
			log.Println(err)
			continue
		}
		var (
			author      = note.Author
			title       = note.Title
			text        = note.Summary
			publishedAt = note.PublishedAt
		)
		if m := fragmentReg.FindStringSubmatch(u.Fragment); len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			comment, err := kibelaCli.GetComment(ctx, id)
			if err != nil {
				log.Println(err)
				continue
			}
			author = comment.Author
			title = fmt.Sprintf(`comment for "%s"`, title)
			text = comment.Summary
			publishedAt = comment.PublishedAt
		}
		unfurls[link.URL] = slack.Attachment{
			// We can't use kibela's avatar URL for an icon, because it's not a public resource.
			// AuthorIcon: author.AvatarImage.URL
			AuthorLink: fmt.Sprintf("https://%s.kibe.la/@%s", kibelaTeam, author.Account),
			AuthorName: author.Account,
			Title:      title,
			TitleLink:  link.URL,
			Text:       spacesReg.ReplaceAllString(text, " "),
			Footer:     "Kibela",
			FooterIcon: "https://cdn.kibe.la/assets/shortcut_icon-99b5d6891a0a53624ab74ef26a28079e37c4f953af6ea62396f060d3916df061.png",
			Ts:         json.Number(fmt.Sprintf("%d", publishedAt.Time.Unix())),
		}
	}

	if len(unfurls) == 0 {
		return nil
	}
	_, _, err := slackCli.PostMessageContext(ctx, ev.Channel, slack.MsgOptionUnfurl(ev.MessageTimeStamp.String(), unfurls))
	return err
}
