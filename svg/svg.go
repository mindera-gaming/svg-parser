package svg

import (
	"encoding/xml"

	"github.com/mindera-gaming/go-math/mathf"
)

// SVG tags
const (
	groupElementTag = "g"
	pathElementTag  = "path"
)

// svg represents the structure of an SVG file
type svg struct {
	XMLName  xml.Name  `xml:"svg"`
	Elements []element `xml:",any"`
}

// groupElement represents a group element
type groupElement struct {
	XMLName  xml.Name  `xml:"g"`
	Elements []element `xml:",any"`
}

// element represents an SVG element
type element struct {
	XMLName xml.Name
	ID      string `xml:"id,attr"`
	Data    []byte `xml:"d,attr"`
	Value   []byte `xml:",innerxml"`
}

// ParserOptions are used to configure the parse of the SVG
type ParserOptions struct {
	// tolerance to ignore path nodes that are not visible to the naked eye
	SlopeTolerance float64
}

// ParsePath deserialises the SVG data and returns a set of paths
func ParsePath(data []byte, options ParserOptions) ([]Path, error) {
	svg := svg{}
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, err
	}

	options.SlopeTolerance = mathf.Max(0, options.SlopeTolerance)

	return parseElements(svg.Elements, options)
}

// parseElements deserialises a SVG element
func parseElements(elements []element, options ParserOptions) ([]Path, error) {
	var paths []Path
	for _, e := range elements {
		var err error
		var newPaths []Path

		switch e.XMLName.Local {
		case groupElementTag:
			newPaths, err = parseGroup(e.Value, options)
		case pathElementTag:
			path := path{
				ID:   string(e.ID),
				Data: string(e.Data),
			}
			path.Clean()
			var pathData []PathData
			pathData, err = path.Parse(options.SlopeTolerance)
			newPaths = append(newPaths, Path{
				ID:   path.ID,
				Data: pathData,
			})
		}

		if err != nil {
			return nil, err
		}
		if len(newPaths) > 0 {
			paths = append(paths, newPaths...)
		}
	}

	return paths, nil
}

// parseGroup deserialises an element of type group
func parseGroup(group []byte, options ParserOptions) ([]Path, error) {
	group = append([]byte("<g>"), group...)
	group = append(group, []byte("</g>")...)
	g := groupElement{}
	if err := xml.Unmarshal(group, &g); err != nil {
		return nil, err
	}

	return parseElements(g.Elements, options)
}
