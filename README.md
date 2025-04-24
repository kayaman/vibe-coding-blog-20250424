# Creating a GO CLI to extract data from articles

# Slug Feature Update Instructions

I've updated the OpenGraph Metadata Extractor to include a slug property extracted from the provided URL. Here's how to apply and test these changes.

## Understanding the Changes

The following changes have been made to the code:

1. Added a `Slug` field to the `OGMetadata` struct
2. Created a new `extractSlug()` function that:
   - Removes protocol (http://, https://)
   - Removes query strings and fragments
   - Gets the last section of the URL by splitting on slashes
   - Handles edge cases like trailing slashes
3. Called `extractSlug()` when processing the URL

## How to Apply the Changes

If you've already set up the project following the previous instructions, follow these steps to update your code:

1. **Update the `OGMetadata` struct**

   Add the `Slug` field to your struct:

   ```go
   type OGMetadata struct {
       URL         string `json:"url"`
       Title       string `json:"title"`
       Description string `json:"description"`
       Image       string `json:"image"`
       Slug        string `json:"slug"`
   }
   ```

2. **Add the `extractSlug()` function**

   Add this new function to your code:

   ```go
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
   ```

3. **Modify the `extractOGMetadata()` function**

   Update the function to call the slug extraction:

   ```go
   func extractOGMetadata(url string) (OGMetadata, error) {
       metadata := OGMetadata{}

       // Extract slug from URL
       metadata.Slug = extractSlug(url)

       // Rest of the function remains the same...
   ```

4. **Rebuild the application**

   ```bash
   go build -o og-extractor
   ```

## Testing the Changes

1. **Test with various URL formats**

   Run the program with different types of URLs to test the slug extraction:

   ```bash
   # Simple URL
   ./og-extractor https://example.com/blog-post output/simple.json

   # URL with query string
   ./og-extractor https://example.com/article?id=123 output/query.json

   # URL with fragment
   ./og-extractor https://example.com/page#section output/fragment.json

   # URL with path
   ./og-extractor https://example.com/category/subcategory/post output/path.json

   # URL with trailing slash
   ./og-extractor https://example.com/blog/ output/trailing.json
   ```

2. **Verify the extracted slugs**

   For each test case, check that the JSON output contains the correct slug:

   - `simple.json` should have slug: `"blog-post"`
   - `query.json` should have slug: `"article"`
   - `fragment.json` should have slug: `"page"`
   - `path.json` should have slug: `"post"`
   - `trailing.json` should have slug: `"blog"`

3. **Edge case testing**

   Also test some edge cases:

   ```bash
   # Root domain
   ./og-extractor https://example.com output/root.json

   # Complex URL with multiple parameters
   ./og-extractor https://example.com/posts/how-to-code?source=newsletter&utm_medium=email#comments output/complex.json
   ```

   For these cases:

   - `root.json` should have slug derived from the domain (likely `"example"`)
   - `complex.json` should have slug: `"how-to-code"`

## Troubleshooting

- If you encounter any issues with the slug extraction, check if the URL format is unusual
- The extraction should handle most common URL formats, but very complex URLs might need additional logic
- Verify that the JSON output has the correct structure with all five properties: url, title, description, image, and slug

## Understanding the Slug Extraction Logic

The slug extraction follows these steps:

1. Clean the URL by removing protocol, query strings, and fragments
2. Split the remaining path by slashes
3. Take the last non-empty segment as the slug
4. If there's no valid path segment (e.g., for a root domain), use part of the domain name

This approach works for most common URL formats but might need adjustments for specific edge cases or requirements.
