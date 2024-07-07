package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	db "github.com/shashwatsahai/go-scrapper/db"
)

var (
	pageSize int = 10
	dbConn   *sql.DB
	wgInsert sync.WaitGroup // WaitGroup for insert operations
)

type placeReviews struct {
	place_id string
	reviews  []string
}

var results []placeReviews

func fetchPageResults(dbConn *sql.DB, page int) (*sql.Rows, error) {
	fmt.Println(page)
	offset := pageSize * (page - 1)
	query := fmt.Sprintf(`SELECT place_id, address FROM hotels ORDER BY id LIMIT %d OFFSET %d`, pageSize, offset)
	rows, err := dbConn.Query(query)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func fetchUrls() ([]string, []string) {
	var placeIds []string
	var adds []string

	var err error
	dbConn, err = db.ConnectDB()
	if err != nil {
		panic(err)
	}

	var page int = 1

	for {
		var hasResults bool = false
		rows, err := fetchPageResults(dbConn, page)

		if err != nil {
			panic(err)
		}

		for rows.Next() {
			hasResults = true
			var placeId, add sql.NullString

			if err := rows.Scan(&placeId, &add); err != nil {
				panic(err)
			}

			if placeId.Valid {
				if len(placeIds) < 2 {
					placeIds = append(placeIds, fmt.Sprintf("https://www.google.com/maps/place/?q=place_id:%s", placeId.String))
				}

			}

			if add.Valid {
				if len(placeIds) < 2 {
					adds = append(adds, add.String)
				}
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}

		rows.Close()

		if !hasResults {
			fmt.Println("No more results, closing on page", page)
			break
		}

		page++
	}

	return placeIds, adds
}

func clickButtonByText(page *rod.Page) {
	// Find and click on the button based on its text content
	reviewsButton := page.MustElementX("//button[.//div[contains(text(), 'Reviews') and contains(@class, 'fontTitleSmall')]]")
	reviewsButton.MustClick()
	page.MustWaitDOMStable()
}

func scrapeReviews(page *rod.Page, index int, placeUrl string, wg *sync.WaitGroup) []string {
	defer wg.Done()

	fmt.Printf("Scraping reviews from page %d\n", index)

	page.WaitLoad()
	page.MustWaitDOMStable()

	html, err := page.HTML()
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	var texts []string
	doc.Find(".MyEned .wiI7pd").Each(func(i int, s *goquery.Selection) {
		reviewText := s.Text()
		texts = append(texts, reviewText)
		fmt.Printf("Review %d: %s\n", i+1, reviewText)

		go func(review, placeID string) {
			defer wgInsert.Done() // Signal completion of insert operation
			fmt.Println("\n", placeID, " = ", review)
			err := insertReview(review, placeID)
			if err != nil {
				log.Printf("Failed to insert review for place ID %s: %v\n", placeID, err)
			}
		}(reviewText, placeUrl)
	})

	// insertReviews(texts, placeUrl)
	return texts
}

func processUrls(placeIds []string, browser *rod.Browser, wg *sync.WaitGroup) {
	var textsPlaces [][]string
	for i, placeUrl := range placeIds {
		wg.Add(1)
		go func(url string, idx int) {
			defer wg.Done()

			fmt.Printf("Opening URL: %s\n", url)
			p := browser.MustPage("")
			defer p.MustClose()

			p.MustNavigate(url).MustWaitLoad()
			p.WaitLoad()

			clickButtonByText(p)
			scrapedTexts := scrapeReviews(p, idx, placeUrl, wg)
			textsPlaces = append(textsPlaces, scrapedTexts)
			// rs := placeReviews{
			// 	placeUrl
			// }
			results = append(results, placeReviews{place_id: placeUrl, reviews: scrapedTexts})
			fmt.Printf("Scraping from %s done\n", url)
		}(placeUrl, i)
	}
}
func main() {
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")
	defer page.MustClose()

	placeIds, _ := fetchUrls()
	var wg sync.WaitGroup

	processUrls(placeIds, browser, &wg)
	wg.Wait()

	wgInsert.Wait()
	fmt.Println("All scraping operations completed.")
}

func insertReview(review, placeID string) error {
	wgInsert.Add(1)
	defer wgInsert.Done()
	query := "INSERT INTO hotels_reviews (reviews, place_id) VALUES ($1, $2)"
	_, err := dbConn.Exec(query, review, placeID)
	if err != nil {
		return err
	}
	return nil
}
