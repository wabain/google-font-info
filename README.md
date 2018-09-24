# google-font-info

This is a script to extract typographic metrics from fonts served by the Google Fonts project.

## Why?

This information can be leveraged to work around shortcomings of browser's font handling. I was motivated to write this after hitting [inconsistencies in how browsers handle vertical alignment in canvases](https://github.com/whatwg/html/issues/2470).

## Usage

**Dependencies**:

* Go 1.11+
* libfreetype2
* The Protocol Buffer compiler, with [Go support](https://github.com/golang/protobuf)

```bash
# Build executable google-font-info
make

# The executable clones the Google Fonts git repository and extracts font
# metrics for each font. Currently, these metrics are written to stdout.
./google-font-info
```
