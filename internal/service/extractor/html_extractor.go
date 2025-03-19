package extractor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type HTMLExtractor struct {
	httpClient *http.Client
}

func NewHTMLExtractor() *HTMLExtractor {
	return &HTMLExtractor{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *HTMLExtractor) ExtractContent(url string) (title, description, content string, err error) {
	resp, err := e.httpClient.Get(url)

	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("failed to fetch URL, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}

	title = e.extractTitle(doc)
	description = e.extractMetaDescription(doc)
	content = e.extractMainContent(doc)
	return title, description, content, nil

}

func (e *HTMLExtractor) extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return n.FirstChild.Data
		}
		return ""
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := e.extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}

func (e *HTMLExtractor) extractMetaDescription(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var name, content string
		for _, a := range n.Attr {
			if a.Key == "name" && a.Val == "description" {
				name = a.Val
			} else if a.Key == "content" {
				content = a.Val
			}
		}
		if name == "description" && content != "" {
			return content
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if desc := e.extractMetaDescription(c); desc != "" {
			return desc
		}
	}
	return ""
}

func (e *HTMLExtractor) extractMainContent(n *html.Node) string {
	var content strings.Builder

	e.extractText(n, &content)
	return content.String()
}

func (e *HTMLExtractor) extractText(n *html.Node, content *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			content.WriteString(text)
			content.WriteString(" ")
		}
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "script", "style", "nav", "footer", "header", "aside":
			return
		case "p", "h1", "h2", "h3", "h4", "h5", "h6", "article", "section", "div":
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				e.extractText(c, content)
			}

			if n.Data == "p" || strings.HasPrefix(n.Data, "h") {
				content.WriteString("\n\n")
			}
			return
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.extractText(c, content)
	}
}

func (e *HTMLExtractor) GenerateSummary(content string) string {
	sentences := strings.Split(content, ".")

	maxSentences := 3
	if len(sentences) < maxSentences {
		maxSentences = len(sentences)
	}

	summary := strings.Join(sentences[:maxSentences], ".")
	if len(sentences) > 0 {
		summary += "."
	}

	return strings.TrimSpace(summary)
}
