package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
	"golang.org/x/net/html"
)

var matchNodeQuery = "MATCH (n:Item { id: $id, checked: false }) RETURN n"
var mergeNodeQuery = "MERGE (n:Item { id: $id, name: $name, checked: true }) RETURN n"
var mergeLinkQuery = "MATCH (n1:Item { id: $id1 }) MATCH (n2:Item { id: $id2 }) " +
	"MERGE (n1)-[:leads]->(n2) RETURN n1, n2"

func getPathnamesOnSite(url string) ([]string, error) {
	pathnames := []string{}

	resp, err := http.Get(url)
	if err != nil {
		return pathnames, fmt.Errorf("Cannot GET %s", url)
	}
	defer resp.Body.Close()

	tokenizer := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return pathnames, nil
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "a" {
				var key, value []byte
				moreAttr := true
				for moreAttr {
					key, value, moreAttr = tokenizer.TagAttr()
					stringKey := string(key)
					if stringKey == "href" {
						pathnames = append(pathnames, string(value))
					}
				}
			}
		}
	}
}

func runQuery(driver neo4j.Driver, command string, arguments map[string]interface{}) (neo4j.Result, error) {
	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		log.Fatal("Cannot create Neo4j Session")
		return nil, err
	}
	defer session.Close()

	result, err := session.Run(command, arguments)
	if err != nil {
		return nil, err
	}

	if err = result.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func processSite(driver neo4j.Driver, host string, basePathname string, counter int) error {
	fmt.Println("Processing " + basePathname)

	matchNodeArgs := map[string]interface{}{"id": basePathname}
	result, err := runQuery(driver, matchNodeQuery, matchNodeArgs)
	if err != nil {
		return err
	}

	if result.Next() || counter <= 0 {
		return fmt.Errorf("Already processed %s", basePathname)
	}

	pathnames, err := getPathnamesOnSite(host + basePathname)
	if err != nil {
		return err
	}

	mergeNodeArgs := map[string]interface{}{"id": basePathname, "name": basePathname}
	if _, err := runQuery(driver, mergeNodeQuery, mergeNodeArgs); err != nil {
		return err
	}

	var innerWaitgroup sync.WaitGroup

	for _, pathname := range pathnames {
		innerWaitgroup.Add(1)

		go func(pathname string) {
			defer innerWaitgroup.Done()

			mergeNodeArgs := map[string]interface{}{"id": pathname, "name": pathname}
			if _, err := runQuery(driver, mergeNodeQuery, mergeNodeArgs); err != nil {
				fmt.Println(fmt.Errorf("Merging error for %s: %s", pathname, err.Error()))
				return
			}

			mergeLinkArgs := map[string]interface{}{"id1": basePathname, "id2": pathname}
			if _, err := runQuery(driver, mergeLinkQuery, mergeLinkArgs); err != nil {
				fmt.Println(fmt.Errorf("Merging error for %s and %s: %s", basePathname, pathname, err.Error()))
				return
			}

			if err := processSite(driver, host, pathname, counter-1); err != nil {
				fmt.Println(err.Error())
				return
			}
		}(pathname)
	}

	innerWaitgroup.Wait()
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments")
	}

	fmt.Println("Waiting for Neo4j...")
	time.Sleep(10 * time.Second)
	fmt.Println("Running...")

	driver, err := neo4j.NewDriver("bolt://neo4j:7687", neo4j.BasicAuth("neo4j", "neo", ""))
	if err != nil {
		log.Fatal("Cannot create Neo4j Driver")
	}
	defer driver.Close()

	runQuery(driver, "MATCH (n) DETACH DELETE n", nil)
	processSite(driver, os.Args[1], "/", 2)

	fmt.Println("Finished")
}
