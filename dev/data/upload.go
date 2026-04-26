// Package main seeds development data by uploading sample recipes via the API.
package main

import (
	"4ks/apps/api/dtos"
	"4ks/libs/go/models"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var index = 0

func parseArgs() (string, string, string) {
	var f string
	flag.StringVar(&f, "f", "", "input filename")
	var t string
	flag.StringVar(&t, "t", "", "token")
	var u string
	flag.StringVar(&u, "u", "", "url")
	flag.Parse()

	if f == "" {
		fmt.Fprintf(os.Stderr, "error: %v\n", fmt.Errorf("filename required. Use flag -f"))
		os.Exit(1)
	}
	if u == "" {
		fmt.Fprintf(os.Stderr, "error: %v\n", fmt.Errorf("API url must be provided. Use flag -u"))
		os.Exit(1)
	}
	if t == "" {
		fmt.Fprintf(os.Stderr, "error: %v\n", fmt.Errorf("bearer API Token must be provided. Use flag -t"))
		os.Exit(1)
	}

	return f, t, u
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func convertStrToSlice(data string) []string {
	data = strings.Replace(data, "\"[", "", -1)
	data = strings.Replace(data, "]\"", "", -1)
	parts := strings.Split(data, "\",")

	for i, v := range parts {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, "\"")
		parts[i] = v
	}

	return parts
}

func getIngredients(data string) []models.Ingredient {
	i := convertStrToSlice(data)
	o := []models.Ingredient{}

	for _, v := range i {
		p := strings.Split(v, ".")
		// fmt.Println(p)
		if len(p) > 1 {
			o = append(o, models.Ingredient{
				Name:     p[1],
				Quantity: p[0],
			})
		} else {
			o = append(o, models.Ingredient{
				Name: p[0],
			})
		}
	}

	return o
}

func getInstructions(data string) []models.Instruction {
	// fmt.Println(data)
	i := convertStrToSlice(data)
	o := []models.Instruction{}

	for _, v := range i {
		// p := strings.Split(v, ".")
		o = append(o, models.Instruction{
			Text: v,
		})
	}

	return o
}

func getRecipeFromLine(line string) dtos.CreateRecipe {
	line = strings.Replace(line, "\"\"", "\"", -1)

	// https://gobyexample.com/regular-expressions
	var re = regexp.MustCompile(`(?m)\"\[\"(.*?)\"\]\"`)
	match := re.FindAllString(line, -1)

	line = strings.Replace(line, match[0], "INGREDIENTS", -1)
	line = strings.Replace(line, match[1], "DIRECTIONS", -1)
	if len(match) > 2 {
		line = strings.Replace(line, match[2], "TAGS", -1)
	}
	res := strings.Split(line, ",")
	// fmt.Println(res)

	r := dtos.CreateRecipe{}
	r.Name = res[1]
	r.Link = res[4]
	r.Ingredients = getIngredients(match[0])
	r.Instructions = getInstructions(match[1])

	// fmt.Println(r)
	return r
}

func postRecipe(client http.Client, u string, t string, r dtos.CreateRecipe) {
	// fmt.Println(index, " => ", r.Name)
	index++
	data, _ := json.Marshal(r)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t)

	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	filename, t, u := parseArgs()
	fmt.Println("Uploading:", filename)

	client := http.Client{
		Timeout: 30 * time.Second,
		// trust insecure certs
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Text()
	// optionally, resize scanner's capacity for lines over 64K, see next example
	// https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go
	line := 0
	for scanner.Scan() {
		line++
		// skip header
		if line == 1 {
			continue
		}

		r := getRecipeFromLine(scanner.Text())
		postRecipe(client, u, t, r)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()

}
