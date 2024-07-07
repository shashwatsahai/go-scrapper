# Go Scraper for Google Maps Reviews

This project is a Go-based web scraper that fetches reviews from Google Maps for a list of hotel place IDs. The scraped reviews are then stored in a PostgreSQL database.

## Features

- Fetches hotel place IDs and addresses from a PostgreSQL database.
- Scrapes reviews from Google Maps using the Go-Rod library.
- Stores the scraped reviews in a PostgreSQL database.
- Uses concurrency to scrape and insert reviews efficiently.

## Prerequisites

- Go 1.16 or higher
- PostgreSQL database
- Chrome or Chromium browser

## Installation

1. **Clone the repository:**
   ```
   sh
   git clone https://github.com/shashwatsahai/go-scraper.git
   cd go-scraper
   ```
   
2. **Install dependencies:**
   ```go mod tidy```

3. **Set Up PostgreSQL**
   Create a database and a table hotels with columns id, place_id, and address.
   Create a table hotels_reviews with the following schema:
   
   ```
      CREATE TABLE hotels_reviews (
         id SERIAL PRIMARY KEY,
         reviews VARCHAR(255),
         place_id VARCHAR(255) UNIQUE NOT NULL
       );
   ```
   
5. **Configure the database connection:**
    Modify the db.ConnectDB function in db/db.go to include your PostgreSQL connection details.
   
7. **Run the scraper:**
   ```
      go run main.go
   ```

Usage
The scraper fetches hotel place IDs from the hotels table in the PostgreSQL database, navigates to their respective Google Maps pages, scrapes the reviews, and inserts them into the hotels_reviews table.

Contributing
Contributions are welcome! Please open an issue or submit a pull request.

