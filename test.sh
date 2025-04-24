#!/bin/bash

# Test script for OpenGraph Metadata Extractor with JSON Append
# This script tests the application with different URLs and verifies the output

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Create test directory
TEST_DIR="test_output"
mkdir -p $TEST_DIR

echo -e "${YELLOW}Building the application...${NC}"
go build -o og-extractor

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"

# Create a sample JSON file
echo -e "${YELLOW}Creating sample JSON file...${NC}"
cat > $TEST_DIR/articles.json << EOF
{
  "articles": [
    {
      "url": "https://example.com/sample-article",
      "title": "Sample Article",
      "description": "This is a sample article for testing",
      "image": "https://example.com/sample.jpg",
      "slug": "sample-article",
      "published_date": "2023-01-01"
    }
  ]
}
EOF

# Test 1: Basic functionality with GitHub
echo -e "\n${YELLOW}Test 1: Basic functionality with GitHub${NC}"
./og-extractor https://github.com/golang/go $TEST_DIR/articles.json

if [ $? -ne 0 ]; then
    echo -e "${RED}Test 1 failed!${NC}"
else
    echo -e "${GREEN}Test 1 passed!${NC}"
    echo -e "Content of updated file:"
    cat $TEST_DIR/articles.json | grep -A 10 "github"
fi

# Test 2: News site with publication date
echo -e "\n${YELLOW}Test 2: News site with publication date${NC}"
./og-extractor https://blog.golang.org/go1.16 $TEST_DIR/articles.json

if [ $? -ne 0 ]; then
    echo -e "${RED}Test 2 failed!${NC}"
else
    echo -e "${GREEN}Test 2 passed!${NC}"
    echo -e "Content of updated file:"
    cat $TEST_DIR/articles.json | grep -A 10 "go1.16"
fi

# Test 3: URL with date pattern
echo -e "\n${YELLOW}Test 3: URL with date pattern${NC}"
./og-extractor https://example.com/2023/04/24/sample-article $TEST_DIR/articles.json

if [ $? -ne 0 ]; then
    echo -e "${RED}Test 3 failed!${NC}"
else
    echo -e "${GREEN}Test 3 passed!${NC}"
    echo -e "Content of updated file:"
    cat $TEST_DIR/articles.json | grep -A 10 "2023-04-24"
fi

# Test 4: Create new file
echo -e "\n${YELLOW}Test 4: Create new file${NC}"
./og-extractor https://golang.org $TEST_DIR/new_articles.json

if [ $? -ne 0 ]; then
    echo -e "${RED}Test 4 failed!${NC}"
else
    echo -e "${GREEN}Test 4 passed!${NC}"
    echo -e "Content of new file:"
    cat $TEST_DIR/new_articles.json
fi

# Verify backup file exists
echo -e "\n${YELLOW}Checking for backup files:${NC}"
find $TEST_DIR -name "*.bkp" -type f

echo -e "\n${GREEN}Testing completed!${NC}"