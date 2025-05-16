# Markdown to PDF Converter

A Terminal User Interface (TUI) application that converts Markdown files to PDF using Gotenberg.

## Prerequisites

1. Go 1.21 or later
2. Gotenberg running on port 3000

## Installation

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod download
   ```

## Running Gotenberg

Make sure Gotenberg is running on port 3000. You can run it using Docker:

```bash
docker run --rm -p 3000:3000 thecodingmachine/gotenberg:7
```

## Usage

1. Run the application:
   ```bash
   go run main.go
   ```

2. Use the arrow keys to navigate through the list of Markdown files
3. Press Enter to convert the selected file to PDF
4. Press 'q' to quit the application

The converted PDF will be saved in the same directory as the original Markdown file with the same name but with a .pdf extension.
