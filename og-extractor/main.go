package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// OGMetadata struct to store Open Graph metadata
type OGMetadata struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Slug        string `json:"slug"`
}

func main() {
	// Check if correct number of arguments is provided
	if len(os.Args) != 3 {
		printUsage()
		os.Exit(1)
	}

	url := os.Args[1]
	jsonFilePath := os.Args[2]

	// Fetch and extract metadata from URL
	metadata, err := extractOGMetadata(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting metadata: %v\n", err)
		os.Exit(1)
	}

	// Save metadata to JSON file
	err = saveToJSON(metadata, jsonFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving to JSON file: %v\n", err)
		os.Exit(1)
	}

	// Print metadata to console
	printMetadata(metadata)
}

func printUsage() {
	fmt.Println("Usage: og-extractor <url> <json-file-path>")
	fmt.Println("  url:            URL of the web page to extract Open Graph metadata from")
	fmt.Println("  json-file-path: Path to save the extracted metadata as JSON")
}

func extractOGMetadata(url string) (OGMetadata, error) {
	metadata := OGMetadata{}
	
	// Extract slug from URL
	metadata.Slug = extractSlug(url)

	// Fetch the web page
	resp, err := http.Get(url)
	if err != nil {
		return metadata, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return metadata, fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return metadata, err
	}

	// Extract Open Graph metadata
	var extractMetadata func(*html.Node)
	extractMetadata = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, content string
			for _, attr := range n.Attr {
				if attr.Key == "property" {
					property = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}

			switch property {
			case "og:url":
				metadata.URL = content
			case "og:title":
				metadata.Title = content
			case "og:description":
				metadata.Description = content
			case "og:image":
				metadata.Image = content
			}
		}

		// Recursively process all child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMetadata(c)
		}
	}

	extractMetadata(doc)
	return metadata, nil
}

// extractSlug extracts the slug from a URL
func extractSlug(url string) string {
	// Remove protocol (http://, https://)
	cleanURL := url
	if idx := strings.Index(cleanURL, "://"); idx != -1 {
		cleanURL = cleanURL[idx+3:]
	}
	
	// Remove query string and fragment
	if idx := strings.Index(cleanURL, "?"); idx != -1 {
		cleanURL = cleanURL[:idx]
	}
	if idx := strings.Index(cleanURL, "#"); idx != -1 {
		cleanURL = cleanURL[:idx]
	}
	
	// Remove trailing slash if present
	cleanURL = strings.TrimSuffix(cleanURL, "/")
	
	// Split by slashes and take the last section
	parts := strings.Split(cleanURL, "/")
	if len(parts) > 0 && parts[len(parts)-1] != "" {
		return parts[len(parts)-1]
	} else if len(parts) > 1 {
		// If the URL ends with a slash, take the second-to-last non-empty part
		for i := len(parts) - 2; i >= 0; i-- {
			if parts[i] != "" {
				return parts[i]
			}
		}
	}
	
	// If we can't find a valid slug, return the domain
	domainParts := strings.Split(cleanURL, ".")
	if len(domainParts) > 0 {
		return domainParts[0]
	}
	
	return ""
}

func saveToJSON(metadata OGMetadata, filePath string) error {
	// Marshal metadata to JSON with indentation
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	// Write JSON to file
	return ioutil.WriteFile(filePath, jsonData, 0644)
}

func printMetadata(metadata OGMetadata) {
	// Marshal metadata to JSON with indentation for console output
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return
	}

	// Print JSON to console
	fmt.Println(string(jsonData))
}