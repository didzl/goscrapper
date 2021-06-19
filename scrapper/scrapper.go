package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//scrape indeed by term
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/%EC%B7%A8%EC%97%85?q="+term+"&limit=50"
	var jobs []jobResult
	c :=make(chan []jobResult)
	total :=getPages(baseURL)
	
	for i:=0; i<total; i++ {
		go getPage(i, baseURL,c)
	}

	for i:=0; i<total; i++{
		jobResult := <-c
		jobs = append(jobs, jobResult...)		
	}

	writeJobs(jobs)
	fmt.Println("DONE!", len(jobs))
}


func getPages(url string) int {
	pages:= 0
	res, err := http.Get(url)

	checkErr(err)
	checkStatus(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	doc.Find(".pagination-list").Each(func (i int, s *goquery.Selection)  {
		pages = s.Find("a").Length()
	})

	return pages
}


func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatus(res *http.Response)  {
	if res.StatusCode !=200 {
		log.Fatalln("Request failed with status:", res.StatusCode)
	}
}


func getPage(page int, url string, mainC chan<- []jobResult) {
	var jobs []jobResult

	c:= make(chan jobResult)

	pageURL:= url+ "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting ", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatus(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i:=0; i <searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

type jobResult struct {
	id string
	title string
	location string
	salary string
	summary string

}

//CleanString
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}


func extractJob(card *goquery.Selection, c chan<- jobResult) {
	id, _ := card.Attr("data-jk")
	title:= CleanString(card.Find(".title>a").Text())		
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())

	c <- jobResult{
		id:id, 
		title:title,
		location:location, 
		salary:salary, 
		summary:summary,
	}

}

//write csv
func writeJobs(jobs []jobResult) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w:= csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"LINK", "TITLE", "LOCATION", "SALARY", "SUMMARY"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs{
		jobSlice := []string{"https://kr.indeed.com/viewjob?jk="+job.id, job.title, job.location,job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}