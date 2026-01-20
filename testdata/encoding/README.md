# Textivus Encoding Test Data (v2)

This corpus contains test files with **known encodings** for validating Textivus:
- encoding detection / selection
- rendering and glyph handling
- newline handling (LF/CRLF/mixed)
- invalid UTF-8 error paths

Most files are expanded to **~16384 bytes** to improve charset detector confidence (e.g., chardet/uchardet).

## Missing encodings addressed in v2
- ISO-8859-15
- GBK
- GB18030
- EUC-KR

## Notes on optional legacy encodings
This corpus includes **Big5** and **KOI8-R** because they still appear in real-world text archives.
If you do not plan to support them, you can delete:
- `chinese/big5_lf.txt`
- `cyrillic/koi8-r_lf.txt`

## Manifest
See `manifest.json` for the authoritative list including SHA-256 hashes and byte sizes.

Generated: 2026-01-20 15:52:28Z
