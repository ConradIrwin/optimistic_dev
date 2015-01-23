package main

import (
	"encoding/json"
	"github.com/ChimeraCoder/anaconda"
	"github.com/darkhelmet/twitterstream"
	"github.com/go-errors/errors"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Hit struct {
	User string
	Repo string
}

var tweets = make(chan string, 10)
var links = make(chan string, 10)
var hits = make(chan string, 10)

var NoRedirects = errors.Errorf("no redirects please")

func main() {

	if os.Getenv("TWITTER_CONSUMER_KEY") == "" ||
		os.Getenv("TWITTER_CONSUMER_SECRET") == "" ||
		os.Getenv("TWITTER_ACCESS_KEY") == "" ||
		os.Getenv("TWITTER_ACCESS_SECRET") == "" {
		panic(".env file does not contain consumer and access keys and secrets. See README.md")
	}

	client := twitterstream.NewClient(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_SECRET"),
		os.Getenv("TWITTER_ACCESS_KEY"), os.Getenv("TWITTER_ACCESS_SECRET"))

	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))

	go analyze()
	go lookupLinks()
	go storeHits()

	for {
		conn, err := client.Track("github")
		if err != nil {
			log.Println("Tracking failed, sleeping for 1 minute")
			time.Sleep(1 * time.Minute)
			continue
		}
		listen(conn)
	}
}

// listen waits for tweets and passes them to analyze
func listen(conn *twitterstream.Connection) {
	log.Printf("listening...")
	for {
		if tweet, err := conn.Next(); err == nil {
			tweets <- tweet.Text
		} else {
			log.Printf("Failed decoding tweet: %s", err)
			return
		}
	}
}

// analyze waits for tweets and extracts all links
func analyze() {
	log.Printf("analyzing...")

	regexp, err := regexp.Compile("https?://t\\.co/(\\w|-)+")
	if err != nil {
		panic(err)
	}
	for tweet := range tweets {
		for _, link := range regexp.FindAllString(tweet, -1) {
			links <- link
		}
	}
}

// lookupLinks converts from t.co links to github.com links
func lookupLinks() {
	log.Printf("looking up links...")

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return NoRedirects
		},
	}

	for link := range links {
		resp, err := client.Head(link)
		if !errors.Is(err, NoRedirects) {
			log.Printf("error fetching %s: %v", link, err)
			continue
		}
		location, err := resp.Location()
		if err != nil {
			log.Printf("error getting location %s: %v", link, err)
			continue
		}

		if strings.HasPrefix(location.String(), "https://github.com/") {
			path := strings.Split(strings.TrimPrefix(location.String(), "https://github.com/"), "/")

			if len(path) >= 2 {
				hits <- path[0] + "/" + path[1]
			}
		}
	}
}

var counter = make(map[string]int)
var tweeted = make(map[string]bool)

// storeHits keeps track of the mentioned github repos, and tweets about them
// because they are all truly awesome.
func storeHits() {
	readBackup("counter.json", &counter)
	readBackup("tweeted.json", &tweeted)

	log.Println(counter)

	for hit := range hits {
		counter[hit] += 1
		log.Printf("Hit! %#v %v", hit, counter[hit])
		if counter[hit] == 5 && !tweeted[hit] {
			tweet("OMG! https://github.com/" + hit + " is awesome! Thanks @" + strings.Split(hit, "/")[0] + " :)")
			tweeted[hit] = true
			writeBackup("tweeted.json", tweeted)
		}
		writeBackup("counter.json", counter)
	}
}

// writeBackup writes a JSON representation of contents to the file at filename
func writeBackup(filename string, contents interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Couldn't open %s: %#v", filename, err)
		return
	}
	err = json.NewEncoder(file).Encode(contents)
	if err != nil {
		log.Printf("Couldn't write to %s: %#v", filename, err)
	}
	err = file.Close()
	if err != nil {
		log.Printf("Couldn't close %s: %#v", filename, err)
	}
}

// readBackup reads JSON from a file created by writeBackup
func readBackup(filename string, contents interface{}) {

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Printf("Couldn't open %s: %#v", filename, err)
		return
	}
	err = json.NewDecoder(file).Decode(contents)
	if err != nil {
		log.Printf("Couldn't decode %s: %#v", filename, err)
	}
}

var lastTweet = time.Unix(0, 0)

// send a tweet, via twitter
func tweet(message string) {
	if time.Since(lastTweet) > 3*time.Hour {
		log.Printf("@optimistic_dev %#v", message)
		lastTweet = time.Now()

		api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_KEY"), os.Getenv("TWITTER_ACCESS_SECRET"))
		_, err := api.PostTweet(message, nil)
		if err != nil {
			log.Println("error posting tweet: %#v", err)
		}
	} else {
		log.Printf("WANT TO TWEET: %#v", message)
	}
}
