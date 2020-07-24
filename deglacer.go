package deglacer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Songmu/kibelasync/kibela"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func Run(argv []string) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", index)
	return http.ListenAndServe(":"+port, nil)
}

var (
	kibelaCli              *kibela.Kibela
	slackCli               *slack.Client
	kibelaTeam             string
	slackVerificationToken string
)

func init() {
	// KIBELA_TOKEN and KIBELA_TEAM are required
	var err error
	kibelaCli, err = kibela.New("0.0.1+deglacer")
	if err != nil {
		log.Fatal(err)
	}
	kibelaTeam = os.Getenv("KIBELA_TEAM")
	slackVerificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")
	if slackVerificationToken == "" {
		log.Fatal("env SLACK_VERIFICATION_TOKEN required")
	}
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		log.Fatal("env SLACK_TOKEN is empty")
	}
	slackCli = slack.New(slackToken)
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
		ev, err := slackevents.ParseEvent(json.RawMessage(body),
			slackevents.OptionVerifyToken(&slackevents.TokenComparator{
				VerificationToken: slackVerificationToken,
			}))
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
			if err := callback(inEv); err != nil {
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

func callback(ev *slackevents.LinkSharedEvent) error {
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
		note, err := kibelaCli.GetNote(id)
		if err != nil {
			log.Println(err)
			continue
		}
		if m := fragmentReg.FindStringSubmatch(u.Fragment); len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			comment, err := kibelaCli.GetComment(id)
			if err != nil {
				log.Println(err)
				continue
			}
			unfurls[link.URL] = slack.Attachment{
				AuthorIcon: comment.Author.AvatarImage.URL,
				AuthorLink: fmt.Sprintf("https://%s.kibe.la/@%s", kibelaTeam, comment.Author.Account),
				AuthorName: comment.Author.Account,
				Title:      fmt.Sprintf(`comment for "%s"`, note.Title),
				TitleLink:  link.URL,
				Text:       spacesReg.ReplaceAllString(comment.Summary, " "),
				Footer:     "Kibela",
				FooterIcon: "https://cdn.kibe.la/assets/shortcut_icon-99b5d6891a0a53624ab74ef26a28079e37c4f953af6ea62396f060d3916df061.png",
				Ts:         json.Number(fmt.Sprintf("%d", comment.PublishedAt.Time.Unix())),
			}
			continue
		}
		unfurls[link.URL] = slack.Attachment{
			AuthorIcon: note.Author.AvatarImage.URL,
			AuthorLink: fmt.Sprintf("https://%s.kibe.la/@%s", kibelaTeam, note.Author.Account),
			AuthorName: note.Author.Account,
			Title:      note.Title,
			TitleLink:  link.URL,
			Text:       spacesReg.ReplaceAllString(note.Summary, " "),
			Footer:     "Kibela",
			FooterIcon: "https://cdn.kibe.la/assets/shortcut_icon-99b5d6891a0a53624ab74ef26a28079e37c4f953af6ea62396f060d3916df061.png",
			Ts:         json.Number(fmt.Sprintf("%d", note.PublishedAt.Time.Unix())),
		}
	}

	if len(unfurls) == 0 {
		return nil
	}
	_, _, err := slackCli.PostMessage(ev.Channel, slack.MsgOptionUnfurl(ev.MessageTimeStamp.String(), unfurls))
	return err
}
