package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
)

var (
	port = flag.String("port", "8000", "listening port for handle connection")
)

func init() {
	flag.Parse()
}

func main() {
	e := echo.New()
	// e.GET("/:id", handler)
	e.GET("/", handler)

	// start server
	go e.Start(fmt.Sprintf(":%s", *port))

	// handle graceful shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

var (
	bankBaseURL = `https://s3.us-east-2.amazonaws.com/figo-interview-banks`
	bankNames   = []string{
		"MoneyBank",
		"SuperBank",
		"GoldBank",
	}
)

type transaction struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	CreatedAt *string `json:"created_at,omitempty"`
	Date      *string `json:"date,omitempty"`
	Timestamp *string `json:"timestamp,omitempty"`
}

// format(transaction) {
// 	if transaction.Date != nil {
// 		//
// 		// transaction.Date = nil
// 	}
//
// 	if transaction.TimeStamp != nil {
// 		///
// 	}
//

// }

// type CreatedAt string
// type Date string
//
// Amount string|float
//
// func (c CreatedAt) Marshal/Unmarshal() {
// 	// convert c to epoch time
// }
//
//  json.Unmarshal(, v)

// unix time/ epoch

// call all the banks
func handler(c echo.Context) error {
	// for ID if needed
	// id := c.Param("id")
	// if id == "" {
	// 	return echo.NewHTTPError(http.StatusBadRequest, "ID cannot be empty")
	// }

	transChan := make(chan []transaction, len(bankNames))
	for _, b := range bankNames {
		go func(n string) {
			trans, err := getDataFromBank(n)
			if err != nil {
				fmt.Printf("[Error] Can get data from bank %s: %s\n", b, err.Error())
			}
			transChan <- trans
		}(b)
	}

	totalTrans := []transaction{}
	for i := 0; i < len(bankNames); i++ {
		if trans := <-transChan; trans != nil {
			totalTrans = append(totalTrans, trans...)
		}
	}

	return c.JSON(http.StatusOK, totalTrans)
}

func getDataFromBank(bankName string) ([]transaction, error) {
	bankAPIURL := fmt.Sprintf("%s/%s.json", bankBaseURL, bankName)

	res, err := http.Get(bankAPIURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	trans := []transaction{}

	d := json.NewDecoder(res.Body)
	err = d.Decode(&trans)

	// format(trans)

	return trans, err
}

// epoch time
