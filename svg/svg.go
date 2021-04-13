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

func ParsePath(data []byte) ([]Path, error) {
	svg := svgElements{}
	if err := xml.Unmarshal(data, &svg); err != nil {
		return nil, err
	}

	return parseElements(svg.Elements)
}

func parseElements(elements []element) ([]Path, error) {
	var paths []Path
	for _, e := range elements {
		var err error
		var newPaths []Path

		switch e.XMLName.Local {
		case groupElement:
			newPaths, err = parseGroup(e.Value)
		case pathElement:
			path := path{
				ID:   string(e.ID),
				Data: string(e.Data),
			}
			path.Clean()
			var pathData []PathData
			pathData, err = path.Parse()
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

func parseGroup(group []byte) ([]Path, error) {
	group = append([]byte("<g>"), group...)
	group = append(group, []byte("</g>")...)
	g := groupElements{}
	if err := xml.Unmarshal(group, &g); err != nil {
		return nil, err
	}

	return parseElements(g.Elements)
}
