# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development

### Commands
- **Build JS/CSS assets**: `npm run build` (uses gulp)
- **Lint TypeScript**: `npm run lint`
- **Typecheck**: `npm run typecheck`
- **Run locally**: Use Docker
  ```bash
  docker build -t schedulemaker .
  docker run --rm -i -t -p 5000:8080 --name=schedulemaker schedulemaker
  ```
  Then visit `http://localhost:5000`.

### Development Notes
- **Version Management**: Increment the version number in `package.json` after updating JS/CSS files to ensure cache busting. Ensure at least the patch number is incremented in PRs touching Javascript/CSS.
- **Configuration**: Configuration is handled via `config.php` in `/inc/` or environment variables defined in `config.env.php`.

## Architecture

The project is a web application with a split architecture:

### Backend (PHP)
- **API**: Located in the `/api/` directory. It serves as the backend interface, handling tasks like course generation, scheduling, and image management.
- **Core Logic**: Many core services are encapsulated in the `api/src/` directory.
- **Configuration**: Uses `/inc/` for configuration and shared PHP logic.

### Frontend (TypeScript/AngularJS)
- **Framework**: Uses AngularJS (1.5) with TypeScript.
- **Structure**: Assets are located in `assets/src/modules/sm/`.
  - **Controllers**: Business logic for specific modules (e.g., `BrowseController.ts`, `GenerateController.ts`).
  - **Directives**: Reusable UI components and behaviors (e.g., `loadingButtonDirective.ts`, `paginationControlsDirective.ts`).
  - **Providers**: Services and factory-like objects for dependency injection (e.g., `entityDataRequestFactory.ts`, `localStorageFactory.ts`).
  - **Templates**: HTML templates for the different modules.
- **Build Process**: Gulp is used to compile TypeScript, process Less/CSS, and minify assets into `assets/prod/`.
