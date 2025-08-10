# Interactive Reading Companion

An application that helps you process and understand book highlights by engaging you in a dialogue with an AI assistant.

## Overview

This application transforms the passive process of reviewing book highlights into an active, reflective dialogue. The system presents you with one highlight at a time and generates thought-provoking questions about it. Your answers are collected and compiled into a structured note file for use in personal knowledge bases.

## Features

- Upload text files containing book highlights
- Interactive dialogue with AI for each highlight
- Export processed highlights as structured notes
- Session management to continue work later

## Technical Stack

- **Backend**: Go (Golang) with Clean Architecture
- **Frontend**: React (Vite)
- **Database**: PostgreSQL
- **Infrastructure**: Docker & Docker Compose

## Getting Started

```bash
# Clone the repository
git clone <repository-url>

# Run the application
make run
```

## Development

This project uses a monorepo structure with separate backend and frontend directories.

For detailed development instructions, see [CONVENTIONS.md](CONVENTIONS.md) and [SPECIFICATION.md](SPECIFICATION.md).

## Project Management

Development tasks are tracked in the [project_management](project_management/) directory.