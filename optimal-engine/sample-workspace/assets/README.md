# Assets

Binary attachments referenced by signals — images, PDFs, audio,
video. The engine doesn't index binary content at this layer; it
references the file by path and leaves extraction to parsers.

## Convention

- `media/` — audio/video episodes for the `08-media` node
- `screenshots/` — image attachments referenced from any node
- `pdfs/` — PDF documents (parser extracts text during ingest)

## Privacy

Assets inherit the ACLs of the signals that cite them. If a signal is
gated to `audience: legal`, its attached PDF is too.
