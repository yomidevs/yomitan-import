package yomitan

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

func dictionaryXMLRoot(reader io.Reader) (xml.StartElement, error) {
	decoder := xml.NewDecoder(io.LimitReader(reader, 1<<20))
	for {
		token, err := decoder.Token()
		if err != nil {
			return xml.StartElement{}, err
		}
		if root, ok := token.(xml.StartElement); ok {
			return root, nil
		}
	}
}

// Read metadata not exposed by jmdict-go, then rewind for its full XML parser.
func dictionaryXMLMetadata(file *os.File, expectedRoot string) (created, version string, err error) {
	root, err := dictionaryXMLRoot(file)
	if err != nil {
		return "", "", err
	}
	if root.Name.Local != expectedRoot {
		return "", "", fmt.Errorf("expected <%s>, found <%s>", expectedRoot, root.Name.Local)
	}
	for _, attr := range root.Attr {
		switch attr.Name {
		case xml.Name{Local: "created"}:
			created = attr.Value
		case xml.Name{Local: "version"}:
			version = attr.Value
		}
	}
	_, err = file.Seek(0, io.SeekStart)
	return created, version, err
}
