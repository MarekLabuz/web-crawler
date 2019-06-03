# web-crawler

Simple web crawler written in Go. Given a website, it creates a graph in Neo4j database
with all subpages it is able to reach. Directed edges in the graph describe which
website is directly accessible from the other.

![graph_screenshot](https://raw.githubusercontent.com/MarekLabuz/web-crawler/master/graph_screenshot.png)
