# simpleWebScrapper
A tool for scraping simple website details.

A Go-based web application that accepts a URL, scrapes its content, and provides a summarized analysis of its HTML structure, headings, links, and accessibility.

---

## Main Steps to Build and Deploy

### Prerequisites
* **Go:** Ensure you have Go 1.18 or higher installed. You can verify this by running `go version`.
* **Git:** Ensure Git is installed for version control handling.

## Steps to Run the Application 

Follow these steps to get the application up and running on your local machine using Visual Studio Code.

### 1. Open the Project
1. Open **VS Code**.
2. Go to `File` > `Open Folder...` and select your `web-analyzer` directory.

### 2. Open the Integrated Terminal
Open the terminal inside VS Code by using the shortcut `Ctrl + ` ` (backtick) or by going to `Terminal` > `New Terminal` in the top menu.

### 3. Fetch Dependencies
Before running, you need to ensure the external HTML parsing library is downloaded. In the terminal, run:
```bash
go mod tidy
```

### 4. Run the Application
You can run the application directly without manual compilation by typing the following command in the terminal:

```bash
go run main.go
```

# Steps to Run Tests

This project includes a comprehensive suite of unit and mock integration tests to verify DOM parsing, heading counts, and parallel network requests without touching the live internet.

Follow these steps to execute the tests in VS Code.

### 1. Open the Integrated Terminal
If it is not already open, access the terminal in VS Code by using the shortcut `Ctrl + ` ` (backtick) or by navigating to `Terminal` > `New Terminal` in the top menu.

### 2. Run All Tests
To run all tests in the project and see a pass/fail summary, execute the following command:
```bash
go test ./internal/...
```

### 3. Run Tests with Verbose Output
To see exactly which specific tests are being executed (including all the table-driven edge cases and mock page scans), run:

```bash
go test ./internal/... -v
```

### 4. Check Code Coverage
To check what percentage of the analyzer statement logic is successfully simulated and covered by these tests, run:

```bash
go test ./internal/... -cover
```