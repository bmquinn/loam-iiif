# LoamIIIF

⚠️ This project is in active development and is not yet stable. ⚠️

LoamIIIF is a terminal-based IIIF (International Image Interoperability Framework) manifest browser. It provides a simple, efficient interface for exploring IIIF collections and manifests through a beautiful terminal user interface.

## Features

- Browse IIIF collections and manifests in a clean, intuitive interface
- Support for multiple IIIF presentation API versions
- Real-time URL validation
- Direct browser integration for opening manifests
- Responsive design that adapts to terminal size
- Full keyboard navigation1
- Clean, modern terminal UI with dynamic updates

## Installation

```bash
go install github.com/bmquinn/loam-iiif@latest
```

## Usage

Launch LoamIIIF by running:

```bash
loam-iiif
```

### Basic Controls

- **Tab**: Switch focus between URL input and results list
- **Enter**: Submit URL for processing
- **O**: Open selected manifest in default browser
- **↑/↓ or j/k**: Navigate through results
- **Ctrl+C/Esc**: Quit application

### URL Input

1. Enter a valid IIIF collection or manifest URL in the input field
2. Press Enter to fetch and parse the IIIF data
3. Use Tab to switch to the results list
4. Navigate to desired manifest
5. Press 'O' to open in browser

## Supported IIIF Formats

LoamIIIF supports both IIIF Presentation API 2.0 and 3.0 formats, including:

- Collection manifests
- Single manifests
- Northwestern University style collections
- National Library of Scotland style collections

## Requirements

- Minimum terminal size: 60x20 characters
- Go 1.16 or higher
- Operating system with support for opening URLs in browser (Linux, macOS, Windows)

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components

## Error Handling

- Invalid URLs are caught and reported
- Network errors are displayed in the status bar
- Parse errors for invalid IIIF data are handled gracefully
- Window size constraints are enforced with clear error messages

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
