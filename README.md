# OpenGraph Metadata Extractor - Setup and Testing Instructions

## Prerequisites

- Go installed on your system (version 1.13 or higher recommended)
- Basic familiarity with terminal/command line
- Internet connection for fetching web pages and downloading dependencies

## Setup Instructions

1. **Create a new project directory**

   ```bash
   mkdir og-extractor
   cd og-extractor
   ```

2. **Initialize Go module**

   ```bash
   go mod init og-extractor
   ```

3. **Save the provided code as `main.go`**

   Copy the code from the artifact into a file named `main.go` in your project directory.

4. **Install dependencies**

   The code requires the `golang.org/x/net/html` package. Install it with:

   ```bash
   go get golang.org/x/net/html
   ```

5. **Build the application**

   ```bash
   go build -o og-extractor
   ```

   This will create an executable named `og-extractor` in your current directory.

## Testing Instructions

1. **Create a sample output directory**

   ```bash
   mkdir output
   ```

2. **Run the program with a test URL**

   Try extracting OpenGraph metadata from a website that uses OG tags, such as:

   ```bash
   # On Linux/macOS
   ./og-extractor https://www.github.com output/github.json

   # On Windows
   og-extractor.exe https://www.github.com output/github.json
   ```

3. **Verify the output**

   - Check that a JSON file was created at the specified path (`output/github.json`)
   - Verify that the console output shows the extracted metadata in JSON format
   - Open the JSON file to confirm it contains the expected OG metadata

4. **Test error handling**

   Try running the program without the required arguments:

   ```bash
   ./og-extractor
   ```

   It should display usage information.

5. **Try with different websites**

   Test with a few different websites to see how the metadata varies:

   ```bash
   ./og-extractor https://www.twitter.com output/twitter.json
   ./og-extractor https://www.nytimes.com output/nytimes.json
   ```

## Troubleshooting

- If you get an error about missing packages, ensure you ran `go get golang.org/x/net/html`
- If a website doesn't return any metadata, it might not have OG tags implemented
- Check that you have internet access if the program fails to fetch URLs
- Verify you have write permissions in the directory where you're trying to save the JSON file

## Notes on the Implementation

- The program extracts four OpenGraph tags: `og:url`, `og:title`, `og:description`, and `og:image`
- The extracted metadata is displayed in the console and saved to the specified JSON file
- The code uses the standard Go HTTP client to fetch web pages and the `golang.org/x/net/html` package to parse HTML
- Error handling is implemented to provide clear messages if anything goes wrong
