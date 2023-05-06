package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func inputHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Web Crawler</title>
    <link rel = "stylesheet" href = "/static/styles.css">
</head>
<body>
    <header>
        <div class = "container">
            <h1>PDC Project - Web Crawler</h1>
        </div>
    </header>
    <div class = "container">
        <form action = "/results" method = "post">
            <label for = "item">Enter the item you want to scrape:</label>
            <input type = "text" id = "item" name = "item" required>
            <button type = "submit">Scrape</button>
        </form>
    </div>
	<div class = "container">
        <p>Made with love by Saim, Ashnah and Maaz</p>
        
    </div>
</body>
</html>`)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	item := r.FormValue("item")
	if item == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	results := Scrape(item)

	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Web Crawler</title>
    <link rel = "stylesheet" href = "/static/styles.css">
</head>
<body>
    <header>
        <div class = "container">
            <h1>Naheed scraper</h1>
        </div>
    </header>
    <div class = "container">
        <div class = "table-container">
            <table>
                <tr>
                    <th>URL</th>
                    <th>Image</th>
                    <th>Name</th>
                    <th>Price</th>
                </tr>
                `)
	for _, result := range results {
		fmt.Fprintf(w, `<tr>
							<td><a href = "%s">%s</a></td>
							<td><img src="%s" width="100" height="100"></td>
							<td>%s</td>
							<td>%s</td>
						</tr>`,
			result.url, result.url, result.image, result.name, result.price)
	}
	fmt.Fprint(w, `</table>
    </div>
</div>
</body>
</html>
`)

}

func main() {
	http.HandleFunc("/", inputHandler)
	http.HandleFunc("/results", resultsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	log.Fatal(http.ListenAndServe(":3000", nil))

}

func setLimit(item string) int {
	url := "https://www.naheed.pk/catalogsearch/result/index/?p=1&q=" + item
	totalPages := ""
	totalItems := "It will be filled later"

	fin := 0

	c := colly.NewCollector(colly.Async(true))

	c.OnHTML("div.toolbar.toolbar-products", func(e *colly.HTMLElement) {
		totalItems = e.ChildText(".toolbar-number")

		if len(totalItems) < 3 {
			fin = 1
			return
		}

		totalPages = strings.TrimSpace(totalItems[3:])
		x, _ := strconv.Atoi(totalPages)

		if x%32 == 0 {
			fin = x / 32
		} else {
			fin = (x / 32) + 1
		}
	})

	c.Visit(url)
	c.Wait()
	return fin
}

// defining a data structure to store the scraped data
type Product struct {
	url, image, name, price string
}

func Scrape(item string) []Product {
	startTime := time.Now()

	// initializing the slice of structs that will contain the scraped data
	var Products []Product

	// the first pagination URL to scrape
	//pageToScrape := "https://www.naheed.pk/catalogsearch/result/index/?p=1&q=" + item
	url := "https://www.naheed.pk/"

	// // max pages to scrape
	limit := setLimit(item)
	// initializing a Colly instance
	c := colly.NewCollector(colly.Async(true))

	c.Limit(&colly.LimitRule{
		Parallelism: 5,
		//Parallelism: 10,
		//Parallelism: limit,

	})
	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	// scraping the product data
	c.OnHTML("div.images-container", func(e *colly.HTMLElement) {
		product := Product{}

		product.url = e.ChildAttr("a", "href")
		product.image = e.ChildAttr("img", "src")
		product.price = e.ChildText(".price")
		product.name = e.ChildText(".product-item-link")

		product.price = "Rs." + strings.Split(product.price, "Rs.")[1]
		Products = append(Products, product)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	// visiting each page sequentially
	for i := 1; i <= limit; i++ {
		c.Visit(url + "catalogsearch/result/index/?p=" + fmt.Sprint(i) + "&q=" + item)
	}

	c.Wait()

	// opening the CSV file
	file, err := os.Create("products_" + item + ".csv")
	if err != nil {
		log.Fatalln("Failed to create output CSV file", err)
	}
	defer file.Close()

	// initializing a file writer
	writer := csv.NewWriter(file)

	// defining the CSV headers
	headers := []string{
		"url",
		"image",
		"name",
		"price",
	}
	// writing the column headers
	writer.Write(headers)

	for _, p := range Products {
		record := []string{
			p.url,
			p.image,
			p.name,
			p.price,
		}

		writer.Write(record) // writing a new CSV record
	}

	writer.Flush()

	endTime := time.Now()

	totalTime := endTime.Sub(startTime)
	fmt.Println("Total time taken:", totalTime)

	return Products
}
