# Interactive Reading Companion

An application that helps you process and understand book highlights by engaging you in a dialogue with an AI assistant.

## Overview

This application transforms the passive process of reviewing book highlights into an active, reflective dialogue. The system presents you with one highlight at a time and generates thought-provoking questions about it. Your answers are collected and compiled into a structured note file for use in personal knowledge bases.

## Current Status

✅ **MVP Core Features Implemented**
- Session creation and management
- Interactive dialogue with AI for each highlight
- Session history and continuation
- Session editing and deletion
- Basic UI for all core workflows

✅ **Advanced Features Implemented**
- Markdown export functionality
- Advanced session completion workflows
- Enhanced UI/UX polish
- Session content review with formatted display
- Toggle between JSON and Markdown rendering modes

## Technical Stack

- **Backend**: Go (Golang) with Clean Architecture
- **Frontend**: React (Vite) with TypeScript
- **Database**: PostgreSQL
- **Infrastructure**: Docker & Docker Compose
- **AI Integration**: LLM client with OpenAI-compatible API support

## Features

### ✅ Core Functionality
- Upload text files containing book highlights
- Interactive dialogue with AI for each highlight
- Export processed highlights as structured notes
- Session management to continue work later
- Session history browsing and management

### ✅ Advanced Features
- LLM-powered question generation
- Session renaming and deletion
- Session continuation from any point
- Review of completed sessions with formatted content display
- Toggle between JSON and Markdown rendering modes for session review
- Responsive web interface
- Two-panel layout for efficient workflow

### 🔄 Upcoming Features
- Full Markdown export functionality
- Enhanced session completion workflows
- Additional session management capabilities

## Getting Started

```bash
# Clone the repository
git clone <repository-url>

# Run the application
make run

# Or using Docker Compose directly
docker compose up -d
```

## Development

This project uses a monorepo structure with separate backend and frontend directories.

For detailed development instructions, see [CONVENTIONS.md](CONVENTIONS.md) and [SPECIFICATION.md](SPECIFICATION.md).

## Project Management

Development tasks are tracked in the [project_management](project_management/) directory.

See [PROGRESS_SUMMARY.md](project_management/PROGRESS_SUMMARY.md) for current implementation status.