# OpenGraph Metadata Extractor with JSON Append - Documentation

## Overview

This Go application extracts OpenGraph metadata from web pages and appends it to an existing JSON file that contains a collection of article metadata. The application handles backing up the original JSON file before modification and ensures that the resulting JSON file is valid.

## Features

1. **OpenGraph Tag Extraction**: Extracts standard OG tags (url, title, description, image, site_name)
2. **Slug Extraction**: Extracts a slug from the URL for SEO and identification purposes
3. **Publication Date Extraction**: Uses multiple strategies to extract article publication dates
4. **JSON Append**: Appends the extracted metadata to an existing JSON file with articles collection
5. **Automatic Backup**: Creates a timestamped backup of the original JSON file before modification

## Requirements

- Go 1.13 or higher
- `golang.org/x/net/html` package for HTML parsing

## Installation

1. **Create a new project directory**

   ```bash
   mkdir og-extractor
   cd og-extractor
   ```

2. **Initialize Go module**

   ```bash
   go mod init og-extractor
   ```

3. **Install dependencies**

   ```bash
   go get golang.org/x/net/html
   ```

4. **Save the source code**

   Create a file named `main.go` and paste the complete source code from the provided artifact.

5. **Build the application**

   ```bash
   go build -o og-extractor
   ```

## Usage

```bash
./og-extractor <url> <json-file-path>
```

- `<url>`: The URL of the web page to extract metadata from
- `<json-file-path>`: Path to the target JSON file to append the metadata to

### Example

```bash
./og-extractor https://example.com/article data/articles.json
```

## File Structure

The target JSON file must follow this structure:

```json
{
  "articles": [
    {
      "url": "https://example.com/article1",
      "title": "Article 1 Title",
      "description": "Article 1 Description",
      "image": "https://example.com/image1.jpg",
      "slug": "article1",
      "publishDate": "2023-04-15",
      "source": "Example Blog"
    },
    // Additional articles...
  ]
}
```

## Detailed Functionality

### 1. OpenGraph Metadata Extraction

The application extracts the following OpenGraph metadata:

- `og:url`: The canonical URL of the page
- `og:title`: The title of the page
- `og:description`: A brief description of the page content
- `og:image`: An image URL representing the page
- `og:site_name`: The name of the site (stored as "source")

### 2. Slug Extraction

The slug is extracted from the URL using the following algorithm:

1. Remove protocol prefixes (http://, https://)
2. Remove query strings and fragments
3. Split by slashes and take the last segment
4. Handle special cases like trailing slashes

For example:
- `https://example.com/my-article` → `my-article`
- `https://example.com/blog/2023/05/my-article?utm=source` → `my-article`

### 3. Publication Date Extraction

The application uses multiple strategies to extract publication dates:

1. **Meta Tags**: Checks common meta tags like `article:published_time`
2. **JSON-LD Data**: Parses structured data for publication dates
3. **URL Pattern**: Extracts dates from URL patterns like `/2023/05/15/article-title`

Extracted dates are stored in the `publishDate` field.

### 4. JSON File Handling

The application processes the target JSON file as follows:

1. **Read Existing File**: If the file exists, reads its content and parses the JSON
2. **Create Backup**: Creates a backup of the original file with format `<filename>.json.YYYYMMDD.bkp`
3. **Append Metadata**: Adds the new metadata to the `articles` array
4. **Write Valid JSON**: Writes the updated collection back to the file with proper indentation

If the target file doesn't exist or is empty, a new file with the proper structure is created.

## Code Structure

The application is organized into these main components:

### Data Structures

- **OGMetadata**: Struct for storing Open Graph metadata including:
  - url
  - title
  - description
  - image
  - slug
  - publishDate
  - source
- **ArticlesCollection**: Struct representing the target JSON file structure

### Core Functions

- **main()**: Entry point that processes arguments and orchestrates the workflow
- **extractOGMetadata()**: Fetches and parses the web page to extract metadata
- **extractSlug()**: Extracts a slug from the URL
- **extractDateFromJSON()**: Parses JSON-LD data for publication dates
- **extractDateFromURL()**: Finds date patterns in URLs
- **validateDate()**: Validates extracted date strings

### File Handling Functions

- **appendToJSONFile()**: Main function for appending to the JSON file with backup
- **createBackupPath()**: Generates the backup file path with timestamp
- **createBackupFile()**: Creates a backup copy of the original file
- **printMetadata()**: Formats and prints the extracted metadata to console

## Error Handling

The application handles various error conditions:

1. **Invalid Arguments**: Displays usage information
2. **Network Errors**: Reports errors when fetching web pages
3. **HTML Parsing Errors**: Handles issues with parsing HTML content
4. **File System Errors**: Manages problems with reading/writing files
5. **JSON Parsing Errors**: Reports when target file has invalid JSON format

## Backup Strategy

Before modifying the target JSON file, the application creates a backup with this naming convention:

```
<original-filename>.json.YYYYMMDD.bkp
```

For example, if the target file is `articles.json`, the backup might be `articles.json.20250424.bkp`.

This allows for easy recovery if needed, with the backup clearly showing the date it was created.

## Limitations and Considerations

1. **HTML-only Support**: The application only extracts data from static HTML, not JavaScript-rendered content
2. **No Duplicate Checking**: The application doesn't check for duplicate entries in the articles collection
3. **Date Format Variations**: Publication dates are stored in whatever format they're found
4. **File Locking**: No file locking mechanism is implemented for concurrent access

## Examples

### Example 1: Basic Usage

```bash
./og-extractor https://blog.golang.org/go1.16 articles.json
```

This will:
1. Extract metadata from the Go blog article
2. Back up the existing `articles.json` file to something like `articles.json.20250424.bkp`
3. Append the new article metadata to the collection
4. Save the updated collection back to `articles.json`

### Example 2: First-time Usage (No Existing File)

```bash
./og-extractor https://blog.golang.org/go1.16 new-collection.json
```

This will:
1. Extract metadata from the Go blog article
2. Create a new file `new-collection.json` with the proper structure
3. Add the article metadata as the first item in the collection

## Troubleshooting

### Common Issues

1. **"Error extracting metadata"**: 
   - Check network connectivity
   - Verify the URL is accessible
   - Ensure the website returns valid HTML

2. **"Error appending to JSON file"**:
   - Check file permissions
   - Verify the target directory exists
   - Ensure any existing JSON file has the correct format

3. **Missing metadata fields**:
   - Not all websites implement complete OpenGraph tags
   - Some fields might be empty in the extracted metadata

### Debugging Tips

1. **Examine the console output**: The extracted metadata is printed to the console
2. **Check the backup file**: Compare with the updated file to see the changes
3. **Validate the JSON**: Use tools like `jq` to validate the JSON structure

```bash
jq '.' articles.json
```

## Future Enhancements

Potential improvements for the application:

1. **Duplicate Detection**: Add option to prevent duplicate entries
2. **Concurrent Processing**: Support for processing multiple URLs in parallel
3. **Date Standardization**: Convert all dates to a standard format
4. **Enhanced Error Recovery**: Better handling of partial failures
5. **Content Extraction**: Add support for extracting article content
6. **Configuration File**: Support for customizing behavior via config file
