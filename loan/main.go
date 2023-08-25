package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type LoanApplication struct {
	AppVersion     string
	BackendVersion string
	BackendHost    string
	BackendPort    string
	InterestRate   int
	LoanAmount     int
}

func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	loanApp := LoanApplication{}

	loanApp.AppVersion = os.Getenv("APP_VERSION")
	if len(loanApp.AppVersion) == 0 {
		loanApp.AppVersion = "dev"
	}

	loanApp.BackendHost = os.Getenv("BACKEND_HOST")
	if len(loanApp.BackendHost) == 0 {
		loanApp.BackendHost = "interest"
	}

	loanApp.BackendPort = os.Getenv("BACKEND_PORT")
	if len(loanApp.BackendPort) == 0 {
		loanApp.BackendPort = "8080"
	}

	http.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		_, err := getInterestRate()
		if err != nil {
			http.Error(w, "down!", http.StatusServiceUnavailable)
		} else {
			fmt.Fprintln(w, "up")
		}

		// fmt.Fprintln(w, "up")
	})

	http.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		// _, err := getInterestRate()
		// if err != nil {
		// 	http.Error(w, "nope!", http.StatusServiceUnavailable)
		// } else {
		// 	fmt.Fprintln(w, "yes")
		// }

		fmt.Fprintln(w, "yes")

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		loanAmount := parseLoanAmount(r)

		quote := ""
		interestFound, err := getInterestRate()
		if err != nil {
			log.Println("Interest error :", err)
			quote = "Could not get interest. Sorry!"
		} else {
			quote = offerQuote(loanAmount, interestFound)
		}

		fmt.Fprintf(w, `<html>
		<form method="post">
		Enter your loan amount to see the interest. $<input name="loan" type="number" value="%d">
		<br/>
		<input type="submit">
		</form>
		<br/>
		%s
		</html>
		`, loanAmount, quote)
	})

	fmt.Printf("Frontend version %s is listening now at port %s\n", loanApp.AppVersion, port)
	err := http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}

func parseLoanAmount(r *http.Request) int {

	err := r.ParseForm() // Parses the request body
	if err != nil {
		return 0
	}

	loanPostParameter := r.Form.Get("loan") // x will be "" if parameter is not set

	loanAmount, err := strconv.Atoi(loanPostParameter)
	if err != nil {
		return 0
	}
	return loanAmount

}

func offerQuote(loan int, interest int) string {
	if loan <= 0 {
		return ""
	}

	total := loan * interest / 100
	return fmt.Sprintf("With rate %d%% you will pay  %d extra interest", interest, total)

}

func getInterestRate() (rate int, err error) {
	url := "http://interest:8080/api/v1/interest"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Could not access %s, got %s\n ", url, err)
		return 0, errors.New("Could not access " + url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Non-OK HTTP status:", resp.StatusCode)
		return 0, errors.New("Could not access " + url)
	}

	log.Printf("Response status of %s: %s\n", url, resp.Status)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return 0, err
	}
	log.Println("Found interest rate " + buf.String())
	return strconv.Atoi(buf.String())
}