# Creating a GO CLI to extract data from articles

# Publication Date Extraction Feature

I've enhanced the OpenGraph Metadata Extractor to include publication date extraction. The implementation uses multiple strategies to extract the publication date from articles, attempting each method in order until a date is found.

## Understanding the Implementation

The date extraction uses these strategies in order of preference:

1. **OpenGraph and Meta Tags**

   - Checks for standard meta tags like `article:published_time`, `datePublished`, `pubdate`, etc.
   - This is the most reliable method when available

2. **JSON-LD Structured Data**

   - Looks for JSON-LD data in `<script type="application/ld+json">` tags
   - Extracts dates from Schema.org Article or NewsArticle markup
   - Common in news sites and blogs

3. **URL Pattern Matching**
   - Falls back to extracting dates from URL patterns
   - Handles common formats like:
     - `/2023/05/15/article-title`
     - `/article/2023-05-15-title`
     - `/article/15-05-2023`
     - `/article/20230515`
   - Validates extracted dates to ensure they're legitimate

## How to Apply the Changes

If you've already set up the project with the slug feature, follow these steps:

1. **Update the imports**

   Add the required packages:

   ```go
   import (
       "encoding/json"
       "fmt"
       "io/ioutil"
       "net/http"
       "os"
       "regexp"
       "strings"
       "time"

       "golang.org/x/net/html"
   )
   ```

2. **Update the `OGMetadata` struct**

   Add the `PublishedDate` field:

   ```go
   type OGMetadata struct {
       URL           string `json:"url"`
       Title         string `json:"title"`
       Description   string `json:"description"`
       Image         string `json:"image"`
       Slug          string `json:"slug"`
       PublishedDate string `json:"published_date,omitempty"`
   }
   ```

3. **Modify the `extractMetadata` function**

   Update to check for date meta tags and JSON-LD data:

   ```go
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
           case "article:published_time", "datePublished", "pubdate", "publishdate", "DC.date.issued", "article:modified_time":
               if metadata.PublishedDate == "" {
                   metadata.PublishedDate = content
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
   if metadata.PublishedDate == "" {
       metadata.PublishedDate = extractDateFromURL(url)
   }
   ```

4. **Add the helper functions**

   Add these new functions:

   ```go
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
           if dateStr, ok := data[field].(string); ok && metadata.PublishedDate == "" {
               metadata.PublishedDate = dateStr
               return
           }
       }

       // Check for nested objects like Article type
       if article, ok := data["@type"]; ok && (article == "Article" || article == "NewsArticle") {
           for _, field := range dateFields {
               if dateStr, ok := data[field].(string); ok && metadata.PublishedDate == "" {
                   metadata.PublishedDate = dateStr
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
   ```

5. **Rebuild the application**

   ```bash
   go build -o og-extractor
   ```

## Testing the Publication Date Extraction

1. **Test with various news sites**

   Run the program with URLs from different news sites and blogs:

   ```bash
   # News website
   ./og-extractor https://www.nytimes.com/2023/04/15/world/europe/ukraine-war-russia.html output/news.json

   # Blog post
   ./og-extractor https://medium.com/golang/go-modules-in-2023-a4a44971d12b output/blog.json

   # Tech article
   ./og-extractor https://techcrunch.com/2023/05/10/google-io-2023-announcements/ output/tech.json
   ```

2. **Check date extraction from URLs**

   For URLs that contain dates:

   ```bash
   # URL with YYYY/MM/DD format
   ./og-extractor https://example.com/2023/05/15/my-article output/date-url1.json

   # URL with YYYY-MM-DD format
   ./og-extractor https://example.com/posts/2023-05-15-article-title output/date-url2.json
   ```

3. **Test with sites using structured data**

   Many modern news sites use JSON-LD structured data:

   ```bash
   # Site with Schema.org markup
   ./og-extractor https://www.wired.com/story/artificial-intelligence-generative-ai-concerns/ output/jsonld.json
   ```

## Possible Improvements

1. **Date format standardization**: Currently, dates are returned in whatever format they're found. You could add date parsing and standardization to ensure all dates are in ISO 8601 format (YYYY-MM-DD).

2. **More fallback methods**: You could scan the HTML content for dates in specific patterns or near bylines.

3. **Custom date matchers for specific sites**: For frequently accessed sites, you could add custom extractors that know where to look.

4. **Handling multiple dates**: Some articles have both published and modified dates. You could extract both.

## Limitations

1. **Not always available**: Many websites don't include clear publication dates or use dynamic JavaScript to display them.

2. **Format variations**: Dates can come in many formats, making parsing challenging.

3. **False positives**: Date-like patterns in URLs might not be publication dates (could be version numbers, etc.).

4. **HTML structure changes**: Websites change their HTML structure frequently, which can break extraction methods.

## Notes on Implementation

The current implementation tries multiple strategies in order of reliability:

1. OpenGraph and meta tags (most reliable when available)
2. JSON-LD structured data (common in news sites)
3. URL pattern matching (least reliable, but often works)

If all methods fail, the `published_date` field will be omitted from the JSON output.
