package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocolly/colly"
)

// defining a data structure to store the scraped data
type Product struct {
	url, image, name, price string
}

// it verifies if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
func inputHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Web Crawler	</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <div class="container">
            <h1>PDC Project - Web Crawler</h1>
        </div>
    </header>
    <div class="container">
        <form action="/results" method="post">
            <label for="item">Enter the item you want to scrape:</label>
            <input type="text" id="item" name="item" required>
            <button type="submit">Scrape</button>
        </form>
    </div>
	<div class="container">
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

	//p := DefaultParser{}
	results := Scrape(item)

	fmt.Print(results)
	// err := writeResultsToCSV(results, "output.csv")

	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Web Crwaler</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <header>
        <div class="container">
            <h1>Naheed scraper</h1>
        </div>
    </header>
    <div class="container">
        <div class="table-container">
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
	    <td>%s</td>
	    <td>%s</td>
	    <td>%s</td>
	    <td>%s</td>
	    <td>%d</td>
	</tr>`,
			result.url, result.image, result.name, result.price)
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
	log.Fatal(http.ListenAndServe(":9090", nil))

}

func Scrape(item string) []Product {

	// fmt.Print("Enter what do you want to search: ")
	// fmt.Scan(&item)

	startTime := time.Now()

	// initializing the slice of structs that will contain the scraped data
	var Products []Product

	// initializing the list of pages to scrape with an empty slice
	pagesToScrape := []string{}

	// the first pagination URL to scrape
	pageToScrape := "https://www.naheed.pk/catalogsearch/result/index/?p=1&q=hats"
	url := "https://www.naheed.pk/"

	// initializing the list of pages discovered with a pageToScrape
	pagesDiscovered := []string{pageToScrape}

	// // max pages to scrape
	limit := 50
	// initializing a Colly instance
	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{
		// limit the parallel requests to 4 request at a time
		Parallelism: 10,
	})
	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	//iterating over the list of pagination links to implement the crawling logic
	c.OnHTML("a.page-numbers", func(e *colly.HTMLElement) {
		// discovering a new page
		newPaginationLink := e.Attr("href")

		// if the page discovered is new
		if !contains(pagesToScrape, newPaginationLink) {
			// if the page discovered should be scraped
			if !contains(pagesDiscovered, newPaginationLink) {
				pagesToScrape = append(pagesToScrape, newPaginationLink)
			}
			pagesDiscovered = append(pagesDiscovered, newPaginationLink)
		}
	})

	// scraping the product data
	c.OnHTML("div.images-container", func(e *colly.HTMLElement) {
		product := Product{}

		product.url = e.ChildAttr("a", "href")
		product.image = e.ChildAttr("img", "src")
		product.price = e.ChildText(".price")
		product.name = e.ChildText(".product-item-link")

		fmt.Println("Url:", product.url)
		fmt.Println("Image:", product.image)
		fmt.Println("Price:", product.price)
		fmt.Println("Name:", product.name)
		Products = append(Products, product)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL)
	})

	// visiting the first page
	c.Visit(pageToScrape)
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

	// adding each Pokemon product to the CSV output file
	for _, Product := range Products {
		// converting a PokemonProduct to an array of strings
		record := []string{
			Product.url,
			Product.image,
			Product.name,
			Product.price,
		}

		// writing a new CSV record
		writer.Write(record)
	}
	defer writer.Flush()

	endTime := time.Now()

	totalTime := endTime.Sub(startTime)
	fmt.Println("Total time taken:", totalTime)

	return Products
}
