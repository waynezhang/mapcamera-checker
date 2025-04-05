package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jipok/go-persist"
)

var conditions = map[int]string{
	2: "新同品",
	3: "美品",
	5: "良品",
	6: "並品",
}

var selltypes = map[int]string{
	1: "新品",
	2: "中古",
}

var sellstatuses = map[int]string{
	1: "在庫あり",
	4: "SOLD OUT",
}

type item struct {
	Genpin_id    int
	Genpin_name  string
	Mapcode      string
	Salesprice   int
	Conditionid  int
	Selltypeid   int
	Sellstatusid int
	Newstockflag int
}

type mpResponse struct {
	Response struct {
		NumFound int
		Docs     []item
	}
}

var logger = log.New(os.Stdout, "", log.Ltime)

func retrieve(keyword string) ([]item, error) {
	// curl "https://www.mapcamera.com/ec/api/itemsearch?igngkeyword=1&siteid=1&limit=100&page=1&devicetype=pc&format=searchresult&keyword=$KEYWORD" \
	//     -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15' \
	//     -H 'Referer: https://www.mapcamera.com' \
	//     -H 'Sec-Fetch-Dest: document' \
	//     -H 'Priority: u=0, i'

	baseURL := "https://www.mapcamera.com/ec/api/itemsearch"

	params := url.Values{}
	params.Add("igngkeyword", "1")
	params.Add("siteid", "1")
	params.Add("limit", "100")
	params.Add("page", "1")
	params.Add("page", "1")
	params.Add("devicetype", "pc")
	params.Add("format", "searchresult")

	requestURL := baseURL + "?" + params.Encode() + "&" + keyword

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15")
	req.Header.Set("Referer", "https://www.mapcamera.com")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Priority", "u=0, i")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result mpResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return result.Response.Docs, nil
}

func filter(keyword string, items []item) []item {
	_ = os.Mkdir("./data", 0755)
	db, err := persist.OpenSingleMap[string]("./data/" + keyword)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Store.Close()

	ret := []item{}
	for _, ite := range items {
		key := ite.Mapcode
		last_status, _ := db.Get(key)
		status := sellstatuses[ite.Sellstatusid]
		if status == "SOLD OUT" {
			if last_status != "" {
				// old item sold out
				db.Delete(key)
			}
		} else {
			if last_status == "" {
				// new item
				if status == "" {
					status = strconv.Itoa(ite.Sellstatusid)
				}
				db.Set(key, status)
				ret = append(ret, ite)
			}
		}
	}

	return ret
}

func notify(ite item) {
	topic := os.Getenv("NTFY_TOPIC")
	if topic == "" {
		logger.Println("No NTFY_TOPIC set")
		return
	}

	ntfyURL := "https://ntfy.sh/" + topic

	condition := conditions[ite.Conditionid]
	selltype := selltypes[ite.Selltypeid]

	title := ite.Genpin_name
	message := fmt.Sprintf("New arrived %s / %s / %s / ¥%d", ite.Genpin_name, condition, selltype, ite.Salesprice)

	req, err := http.NewRequest("POST", ntfyURL, strings.NewReader(message))
	if err != nil {
		log.Printf("Failed to create notification request: %v", err)
		return
	}

	req.Header.Set("Title", title)
	req.Header.Set("Tags", "camera,mapcamera")
	req.Header.Set("Click", fmt.Sprintf("https://www.mapcamera.com/item/%s", ite.Mapcode))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Println("Failed to send notification due to " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Notification service returned non-OK status: %d\n", resp.StatusCode)
	}
}

func healthcheck() {
	logger.Println("Pinging healthcheck.io")

	healthcheck_slug := os.Getenv("HC_SLUG")
	if healthcheck_slug == "" {
		logger.Println("No HC_SLUG set")
		return
	}

	_, err := http.Get("https://hc-ping.com/" + healthcheck_slug)
	if err != nil {
		logger.Println("Failed to send healthcheck request due to " + err.Error())
	}
}

func run(keyword string) {
	items, err := retrieve(keyword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for items: %v\n", err)
		os.Exit(1)
	}

	filtered := filter(keyword, items)

	logger.Printf("Found %d/%d items\n", len(filtered), len(items))
	for _, item := range filtered {
		fmt.Printf("- %s: ¥%d\n", item.Genpin_name, item.Salesprice)
		notify(item)
	}

	healthcheck()
}

func main() {
	var keyword string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [keyword]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  keyword: keyword to search\n")
	}

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	keyword = args[0]

	logger.Printf("Checking keyword: %s\n", keyword)

	run(keyword)
}
