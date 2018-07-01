package main

import (
	"github.com/olesho/crawler"
	"os"
	"regexp"
	"fmt"
	"strings"
)

func main() {
	s, err := crawler.NewMysqlStorage(&crawler.MysqlConfig{
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
	})
	if err != nil {
		panic(err)
	}

	queue := crawler.NewHttqQueue(2, map[string]string{
		"Host": "www.jewishusedbooks.com",
		"User-Agent": "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:61.0) Gecko/20100101 Firefox/61.0",
		"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.5",
		"Accept-Encoding": "gzip, deflate",
		"Cookie": `visid_incap_140802=9ESw/PqPQ/GtPgC8lJnEG5VZCFsAAAAAQ0IPAAAAAACAKhmFAbkOumjCY4UsMsbxeeZ/UGvcEmZT; __utma=213681167.1674149596.1527273880.1529996013.1530255672.10; __utmz=213681167.1527273880.1.1.utmcsr=(direct)|utmccn=(direct)|utmcmd=(none); ASPSESSIONIDACBBAAAQ=ENCFJMLCDPOEPNLBOANEEJAH; incap_ses_472_140802=slonG+1lqwDQwL+JBOKMBjbZNVsAAAAATzzx+rSOaBGOh03u6A4wXg==; StartPageCookie=1; EndPageCookie=1; mysql=SELECT%20a.idProduct%2Ca.SKU%2Ca.relatedKeys%2Crating%2Ca.Description%2Ca.pubstatus%2C%20%20%20%20%20%20%20a.DescriptionLong%2Ca.ListPrice%2C%20%20%20%20%20%20%20a.Price%2Ca.ImageUrl1%2Ca.ImageUrl1watermark%2Ca.Stock%2C%20%20%20%20%20%20%20a.fileName%2Ca.noShipCharge%20FROM%20%20%20products%20a%2C%20categories_products%20b%20WHERE%20%20a.stock%3E0%20and%20a.idProduct%20%3D%20b.idProduct%20and%20a.ParentID%20is%20null%20AND%20%20%20%20b.idCategory%20%3D%2028%20AND%20%20%20%20a.active%20%3D%20-1%20ORDER%20BY%20a.idproduct%20desc; TheUrlCookie=prodlist.asp%3FidCategory%3D28; __utmb=213681167.4.10.1530255672; __utmc=213681167; __utmt=1`,
		"Connection": "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Cache-Control": "max-age=0",
	})

	c, err := crawler.New(s, &receiver{}, queue)
	if err != nil {
		panic(err)
	}

	availableTasks := c.Tasks()
	var taskId int64
	if len(availableTasks) > 0 {
		taskId = availableTasks[0].GetID()
	} else {
		taskId, err = c.Task(0, "http://www.jewishusedbooks.com/", `^(http:\/\/www.jewishusedbooks.com\/)`)
		if err != nil {
			panic(err)
		}
	}

	c.SetFilterFunc(taskId, func(url string) bool {
		return ( strings.Contains(url, "http://www.jewishusedbooks.com/prodview.asp?idProduct") ||
			strings.Contains(url, "http://www.jewishusedbooks.com/prodlist.asp?idCategory") ||
			strings.Contains(url, "http://www.jewishusedbooks.com/prodimages/") ) && (
			!strings.Contains(url, "http://www.jewishusedbooks.com/prodimages/thumb") &&
				!strings.Contains(url, "ShowAll=True&ShowAll=True") )
	})

	c.Run(taskId)
	<- c.Done()
}

type receiver struct {
}

var matchProductRule *regexp.Regexp = regexp.MustCompile(`^http:\/\/www.jewishusedbooks.com\/prodview.asp\?idProduct=([0-9]+)`)
var matchImageRule *regexp.Regexp = regexp.MustCompile(`^http:\/\/www.jewishusedbooks.com\/prodimages\/(.+)`)

func (r *receiver) Receive(url string, taskId int64, data []byte, requestError error) bool {
	if strings.Contains(string(data), "Request unsuccessful. Incapsula incident ID") {
		fmt.Println("Request unsuccessful: ", url)
		return false
	}
	fmt.Println("Successful: ", url)

	if matchProductRule.MatchString(url) {
		strs := matchProductRule.FindStringSubmatch(url)
		fmt.Println("Written: ", strs[1])
		err := writeFile("./results/" + strs[1], data)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	if matchImageRule.MatchString(url) {
		strs := matchImageRule.FindStringSubmatch(url)
		fmt.Println("Written: ", strs[1])
		err := writeFile("./images/" + strs[1], data)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}

	return true
}

func writeFile(fileName string, data []byte) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}