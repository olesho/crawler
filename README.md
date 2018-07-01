# Simple website crawler on Golang #

## Entities: ##
* Storage interface describes entity storing the state of the crawler
* MysqlStorage is MySQL(GORM) implementation of Storage
* Receiver interface describes entity asyncronously receiving results of crawling and scraping
* Crawler - the core crawler/scraper engine structure
* HttpQueue - the queue where we put URLs waiting to be crawled

## Installation: ##

```go get github.com/olesho/crawler```


## Usage example: ##
```
    s, _ := crawler.NewMysqlStorage(&crawler.MysqlConfig{User: "dbUser", Password: "5mt493fh5h4g", Host: "localhost", Port: "3306", DBName: "url_storage"})
    queue := crawler.NewHttqQueue(2, map[string]string{})
    c, _ := crawler.New(s, &receiver{}, queue)
    taskId, _ = c.Task(0, "http://bbc.co.uk/", `^(http:\/\/bbc.co.uk\/)`)
    // taskId = availableTasks[0].GetID() // if task already exists and is the only one
    c.Run(taskId)
    <- c.Done()
```