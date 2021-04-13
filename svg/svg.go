package svg

import (
	"encoding/xml"
)

const (
	groupElement = "g"
	pathElement  = "path"
)

type svgElements struct {
	XMLName  xml.Name  `xml:"svg"`
	Elements []element `xml:",any"`
}

type groupElements struct {
	XMLName  xml.Name  `xml:"g"`
	Elements []element `xml:",any"`
}

type element struct {
	XMLName xml.Name
	ID      string `xml:"id,attr"`
	Data    []byte `xml:"d,attr"`
	Value   []byte `xml:",innerxml"`
}

var paths []Path

func ParsePath(data []byte) ([]Path, error) {
	svg := svgElements{}
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, err
	}

	paths = nil
	if err := parseElements(svg.Elements); err != nil {
		return nil, err
	}

	return paths, nil
}

func parseElements(elements []element) error {
	for _, e := range elements {
		var err error
		var newPaths []PathData

		switch e.XMLName.Local {
		case groupElement:
			err = parseGroup(e.Value)
		case pathElement:
			path := path{
				ID:   string(e.ID),
				Data: string(e.Data),
			}
			path.Clean()
			newPaths, err = path.Parse()
		}

		if err != nil {
			return err
		}
		if len(newPaths) > 0 {
			paths = append(paths, Path{
				ID:   string(e.ID),
				Data: newPaths,
			})
		}
	}

	return nil
}

func parseGroup(group []byte) error {
	group = append([]byte("<g>"), group...)
	group = append(group, []byte("</g>")...)
	g := groupElements{}
	if err := xml.Unmarshal(group, &g); err != nil {
		return err
	}

	return parseElements(g.Elements)
}
