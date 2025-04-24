package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// OGMetadata struct to store Open Graph metadata
type OGMetadata struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Slug        string `json:"slug"`
	PublishDate string `json:"publishDate,omitempty"`
	Source      string `json:"source,omitempty"`
}

// ArticlesCollection represents the structure of the target JSON file
type ArticlesCollection struct {
	Articles []OGMetadata `json:"articles"`
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

	// Create backup and append to existing JSON file
	err = appendToJSONFile(metadata, jsonFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error appending to JSON file: %v\n", err)
		os.Exit(1)
	}

	// Print metadata to console
	printMetadata(metadata)
	fmt.Printf("\nSuccessfully appended to %s\n", jsonFilePath)
}

func printUsage() {
	fmt.Println("Usage: og-extractor <url> <json-file-path>")
	fmt.Println("  url:            URL of the web page to extract Open Graph metadata from")
	fmt.Println("  json-file-path: Path to the target JSON file to append the metadata to")
	fmt.Println("\nThe target JSON file must follow the structure: {\"articles\":[{...}]}")
	fmt.Println("A backup of the original file will be created before modification.")
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
				if attr.Key == "property" || attr.Key == "name" {
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
			case "og:site_name":
				metadata.Source = content
			case "article:published_time", "datePublished", "pubdate", "publishdate", "DC.date.issued", "article:modified_time":
				if metadata.PublishDate == "" {
					metadata.PublishDate = content
				}
			}
		}

		// Look for LD+JSON data that might contain publication date
		if n.Type == html.ElementNode && n.Data == "script" {
			var isJSON bool
			for _, attr := range n.Attr {
				if attr.Key == "type" && (attr.Val == "application/ld+json" || attr.Val == "application/json") {
					isJSON = true
					break
				}
			}

			if isJSON && n.FirstChild != nil {
				jsonContent := n.FirstChild.Data
				extractDateFromJSON(jsonContent, &metadata)
			}
		}

		// Recursively process all child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMetadata(c)
		}
	}

	extractMetadata(doc)
	
	// If we couldn't find a date in metadata, try to extract it from the URL
	if metadata.PublishDate == "" {
		metadata.PublishDate = extractDateFromURL(url)
	}
	
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

// extractDateFromJSON attempts to extract publication date from JSON-LD data
func extractDateFromJSON(jsonContent string, metadata *OGMetadata) {
	var data map[string]interface{}
	
	// Try to unmarshal the JSON
	err := json.Unmarshal([]byte(jsonContent), &data)
	if err != nil {
		return // Ignore errors, just continue
	}
	
	// Look for common date fields in schema.org and other formats
	dateFields := []string{"datePublished", "dateCreated", "publishedTime", "dateModified", "pubDate"}
	
	for _, field := range dateFields {
		if dateStr, ok := data[field].(string); ok && metadata.PublishDate == "" {
			metadata.PublishDate = dateStr
			return
		}
	}
	
	// Check for nested objects like Article type
	if article, ok := data["@type"]; ok && (article == "Article" || article == "NewsArticle") {
		for _, field := range dateFields {
			if dateStr, ok := data[field].(string); ok && metadata.PublishDate == "" {
				metadata.PublishDate = dateStr
				return
			}
		}
	}
}

// extractDateFromURL attempts to find a date pattern in the URL
func extractDateFromURL(urlStr string) string {
	// Common date patterns in URLs
	patterns := []struct {
		regex   *regexp.Regexp
		format  string
		example string
	}{
		// YYYY/MM/DD pattern (e.g., example.com/2023/05/15/article-title)
		{
			regex:   regexp.MustCompile(`/(\d{4})/(\d{2})/(\d{2})/`),
			format:  "%s-%s-%s",
			example: "2023/05/15",
		},
		// YYYY-MM-DD pattern (e.g., example.com/article/2023-05-15-title)
		{
			regex:   regexp.MustCompile(`/(\d{4})-(\d{2})-(\d{2})`),
			format:  "%s-%s-%s",
			example: "2023-05-15",
		},
		// DD-MM-YYYY pattern (e.g., example.com/article/15-05-2023)
		{
			regex:   regexp.MustCompile(`/(\d{2})-(\d{2})-(\d{4})`),
			format:  "%s-%s-%s",
			example: "15-05-2023",
		},
		// YYYYMMDD pattern (e.g., example.com/article/20230515)
		{
			regex:   regexp.MustCompile(`/(\d{4})(\d{2})(\d{2})`),
			format:  "%s-%s-%s",
			example: "20230515",
		},
	}
	
	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(urlStr)
		if len(matches) >= 4 {
			// For YYYY/MM/DD and YYYY-MM-DD formats
			if pattern.example == "2023/05/15" || pattern.example == "2023-05-15" {
				dateStr := fmt.Sprintf(pattern.format, matches[1], matches[2], matches[3])
				// Validate the date
				if validateDate(dateStr) {
					return dateStr
				}
			}
			// For DD-MM-YYYY format
			if pattern.example == "15-05-2023" {
				dateStr := fmt.Sprintf(pattern.format, matches[3], matches[2], matches[1])
				if validateDate(dateStr) {
					return dateStr
				}
			}
			// For YYYYMMDD format
			if pattern.example == "20230515" {
				dateStr := fmt.Sprintf(pattern.format, matches[1], matches[2], matches[3])
				if validateDate(dateStr) {
					return dateStr
				}
			}
		}
	}
	
	return ""
}

// validateDate checks if a date string in YYYY-MM-DD format is valid
func validateDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// appendToJSONFile reads the existing JSON file, creates a backup, and appends the new metadata
func appendToJSONFile(metadata OGMetadata, filePath string) error {
	var collection ArticlesCollection
	
	// Check if the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If file exists, read it and create a backup
	if err == nil && fileInfo.Size() > 0 {
		// Create backup with timestamp
		backupPath := createBackupPath(filePath)
		err = createBackupFile(filePath, backupPath)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		
		// Read existing content
		fileContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
		
		// Parse JSON
		err = json.Unmarshal(fileContent, &collection)
		if err != nil {
			return fmt.Errorf("invalid JSON format in existing file: %w", err)
		}
	} else {
		// Initialize new collection if file doesn't exist
		collection = ArticlesCollection{
			Articles: []OGMetadata{},
		}
	}
	
	// Append new metadata to articles array
	collection.Articles = append(collection.Articles, metadata)
	
	// Write back to file with indentation
	jsonData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	
	return nil
}

// createBackupPath generates a backup file path with timestamp
func createBackupPath(filePath string) string {
	now := time.Now()
	timestamp := now.Format("20060102") // YYYYMMDD format
	
	ext := filepath.Ext(filePath)
	basePath := filePath[:len(filePath)-len(ext)]
	
	return fmt.Sprintf("%s%s.%s.bkp", basePath, ext, timestamp)
}

// createBackupFile creates a backup copy of the original file
func createBackupFile(sourcePath, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	return nil
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
