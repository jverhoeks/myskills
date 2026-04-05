---
name: authoring-ebooks-with-pandoc
description: >
  Guide ebook authoring from Markdown sources through Pandoc to PDF, EPUB, and HTML
  output. Use when the user wants to write a book, create an ebook, build a PDF from
  Markdown, set up a Pandoc pipeline, add diagrams to a book with Mermaid or D2,
  configure LaTeX templates, manage chapters, or mentions Pandoc, ebook, manuscript,
  or book project structure.
license: MIT
metadata:
  author: jverhoeks
  version: "1.0.0"
  team: data-ai
---

# Authoring Ebooks with Pandoc

End-to-end workflow for writing, structuring, and producing ebooks from Markdown source
files using Pandoc. Covers project scaffolding, chapter management, diagram embedding,
cross-references, styling, and multi-format output (PDF via LaTeX, EPUB, HTML).

## When to Use

- User wants to start writing a book or ebook
- User has Markdown files and wants to produce PDF / EPUB / HTML
- User asks about Pandoc configuration, LaTeX templates, or EPUB metadata
- User wants to embed Mermaid or D2 diagrams in a book
- User mentions chapters, frontmatter, table of contents, or manuscript structure

## Prerequisites

Verify these are installed before proceeding. Suggest installation commands if missing.

| Tool | Purpose | Check command |
|------|---------|---------------|
| `pandoc` | Document conversion | `pandoc --version` |
| `pdflatex` or `xelatex` | PDF generation | `xelatex --version` |
| `d2` | D2 diagrams (optional) | `d2 --version` |
| `mmdc` (mermaid-cli) | Mermaid diagrams (optional) | `mmdc --version` |
| `pandoc-crossref` | Cross-references (optional) | `pandoc-crossref --version` |

## Project Structure

Scaffold this structure for any new book project:

```
book-title/
├── metadata.yaml          # Book metadata and Pandoc config
├── Makefile               # Build commands
├── chapters/
│   ├── 00-preface.md
│   ├── 01-introduction.md
│   ├── 02-chapter-name.md
│   └── ...
├── diagrams/
│   ├── src/               # D2 or Mermaid source files
│   │   ├── architecture.d2
│   │   └── flow.mmd
│   └── out/               # Rendered PNGs/SVGs (gitignored, generated)
├── assets/
│   ├── cover.png          # Cover image
│   ├── images/            # Static images
│   └── fonts/             # Custom fonts (optional)
├── templates/
│   ├── template.tex       # LaTeX template for PDF
│   └── style.css          # CSS for EPUB/HTML
├── output/                # Generated files (gitignored)
│   ├── book.pdf
│   ├── book.epub
│   └── book.html
└── .gitignore
```

### Scaffold Command

When the user asks to start a new book, create this structure:

```bash
mkdir -p book-title/{chapters,diagrams/{src,out},assets/{images,fonts},templates,output}
```

Then generate the starter files described in the following sections.

## metadata.yaml

This file drives Pandoc. Generate it with the user's book details:

```yaml
---
title: "Book Title"
subtitle: "Optional Subtitle"
author:
  - "Author Name"
date: "2026"
lang: en
rights: "© 2026 Author Name. All rights reserved."

# Output settings
documentclass: book
papersize: a5
fontsize: 11pt
geometry:
  - margin=2cm
linestretch: 1.3
toc: true
toc-depth: 2
numbersections: true

# PDF (LaTeX) settings
mainfont: "DejaVu Serif"
sansfont: "DejaVu Sans"
monofont: "DejaVu Sans Mono"
colorlinks: true
linkcolor: "NavyBlue"
urlcolor: "NavyBlue"

# EPUB settings
cover-image: assets/cover.png
epub-chapter-level: 1

# Cross-references (if using pandoc-crossref)
figPrefix: "Figure"
tblPrefix: "Table"
secPrefix: "Section"
---
```

## Chapter File Format

Each chapter file should follow this pattern:

```markdown
# Chapter Title {#ch-slug}

Chapter introduction paragraph.

## First Section {#sec-slug}

Body text with a reference to [@fig:architecture] or [@sec:other-slug].

### Subsection

More content here.

::: {.callout}
**Key Takeaway:** Important point highlighted in a callout box.
:::
```

### Chapter Conventions

- Prefix filenames with two-digit numbers for ordering: `01-`, `02-`, etc.
- Use `{#ch-slug}` identifiers on chapter headings for cross-references
- Use `{#sec-slug}` on sections you want to reference elsewhere
- One H1 (`#`) per file — this becomes the chapter title
- Keep chapters focused: 3,000-6,000 words each is a good target

## Diagram Workflow

### D2 Diagrams

Source files go in `diagrams/src/*.d2`, rendered to `diagrams/out/`.

Render command:
```bash
d2 --theme 200 --pad 20 diagrams/src/architecture.d2 diagrams/out/architecture.png
```

Reference in Markdown:
```markdown
![System architecture](diagrams/out/architecture.png){#fig:architecture width=90%}
```

### Mermaid Diagrams

Source files go in `diagrams/src/*.mmd`, rendered with mermaid-cli.

Render command:
```bash
mmdc -i diagrams/src/flow.mmd -o diagrams/out/flow.png -w 1200
```

Reference in Markdown:
```markdown
![Process flow](diagrams/out/flow.png){#fig:flow width=80%}
```

### Batch Render All Diagrams

Add this to the Makefile (see Makefile section below).

## Makefile

Generate this Makefile for the project:

```makefile
CHAPTERS := $(sort $(wildcard chapters/*.md))
META     := metadata.yaml
FILTERS  :=
OUT      := output

# Add pandoc-crossref if available
ifneq ($(shell which pandoc-crossref 2>/dev/null),)
  FILTERS += --filter pandoc-crossref
endif

.PHONY: all pdf epub html diagrams clean

all: diagrams pdf epub

pdf: diagrams $(CHAPTERS) $(META)
	pandoc $(META) $(CHAPTERS) \
		$(FILTERS) \
		--pdf-engine=xelatex \
		--template=templates/template.tex \
		--resource-path=.:assets:diagrams/out \
		--highlight-style=tango \
		-o $(OUT)/book.pdf

epub: diagrams $(CHAPTERS) $(META)
	pandoc $(META) $(CHAPTERS) \
		$(FILTERS) \
		--css=templates/style.css \
		--resource-path=.:assets:diagrams/out \
		--epub-embed-font=assets/fonts/*.ttf \
		--highlight-style=tango \
		-o $(OUT)/book.epub

html: diagrams $(CHAPTERS) $(META)
	pandoc $(META) $(CHAPTERS) \
		$(FILTERS) \
		--standalone \
		--css=templates/style.css \
		--resource-path=.:assets:diagrams/out \
		--highlight-style=tango \
		--toc \
		-o $(OUT)/book.html

diagrams:
	@mkdir -p diagrams/out
	@for f in diagrams/src/*.d2 2>/dev/null; do \
		[ -f "$$f" ] && d2 --theme 200 --pad 20 "$$f" \
			"diagrams/out/$$(basename $$f .d2).png"; \
	done; true
	@for f in diagrams/src/*.mmd 2>/dev/null; do \
		[ -f "$$f" ] && mmdc -i "$$f" \
			-o "diagrams/out/$$(basename $$f .mmd).png" -w 1200; \
	done; true

clean:
	rm -rf $(OUT)/* diagrams/out/*
```

## Build Commands

| Command | Output |
|---------|--------|
| `make pdf` | `output/book.pdf` |
| `make epub` | `output/book.epub` |
| `make html` | `output/book.html` |
| `make all` | PDF + EPUB |
| `make clean` | Remove all generated files |

## .gitignore

```
output/
diagrams/out/
*.aux
*.log
*.out
*.toc
*.fdb_latexmk
*.fls
*.synctex.gz
```

## Common Tasks

### Adding a New Chapter

1. Create `chapters/NN-chapter-slug.md` with the next sequence number
2. Add the H1 heading with an ID: `# Chapter Title {#ch-slug}`
3. Run `make pdf` — Pandoc picks up files alphabetically via the wildcard

### Reordering Chapters

Rename the numeric prefixes. The Makefile uses `$(sort ...)` so alphabetical order
determines chapter sequence.

### Adding a Callout / Aside Box

Use Pandoc's fenced divs:
```markdown
::: {.callout}
**Note:** This is a highlighted callout box.
:::
```

Map this to a styled box in both the LaTeX template and CSS.

### Adding Epigraphs

```markdown
::: {.epigraph}
*"The supreme art of war is to subdue the enemy without fighting."*
— Sun Tzu
:::
```

### Code Listings with Captions

````markdown
```{#lst:example .python caption="Example: loading data"}
import pandas as pd
df = pd.read_csv("data.csv")
```
````

### Cross-References (with pandoc-crossref)

- Figures: `[@fig:architecture]` renders as "Figure 1"
- Tables: `[@tbl:comparison]` renders as "Table 2"
- Sections: `[@sec:design]` renders as "Section 3.1"
- Listings: `[@lst:example]` renders as "Listing 1"

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| Missing images in PDF | Resource path wrong | Check `--resource-path` includes diagram output dir |
| Fonts not found | Font not installed system-wide | Use `--pdf-engine=xelatex` and specify `mainfont` in metadata |
| Cross-refs show `??` | `pandoc-crossref` not installed or IDs mistyped | Install filter and check `{#fig:id}` syntax |
| EPUB validation errors | Invalid metadata | Run `epubcheck output/book.epub` and fix warnings |
| Chapter order wrong | File naming gaps | Ensure sequential `01-`, `02-`, `03-` prefixes |
| D2 render fails | Theme not available | Remove `--theme` flag or use numeric theme ID |
| Mermaid too small | Default width low | Increase `-w` value in mmdc command |

## Tips

- Write in plain Markdown, keep formatting minimal — Pandoc and templates handle the look
- Commit diagram sources (`diagrams/src/`), not rendered outputs (`diagrams/out/`)
- Use `xelatex` over `pdflatex` for Unicode and custom font support
- For print-ready output, set `papersize: a5` and adjust margins
- Preview quickly with `make html` — it builds fastest
- Keep the LaTeX template minimal; override only what you need from Pandoc defaults
