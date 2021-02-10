package svg

import "encoding/xml"

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
	var err error
	var paths, newPaths []Path
	for _, e := range elements {
		switch e.XMLName.Local {
		case groupElement:
			newPaths, err = parseGroup(e.Value)
			if err != nil {
				return nil, err
			}
			if len(newPaths) == 0 {
				continue
			}

			paths = append(paths, newPaths...)
		case pathElement:
			path := path{Data: string(e.Data)}
			path.Clean()
			newPaths, err = path.Parse()
			if err != nil {
				return nil, err
			}
			if len(newPaths) == 0 {
				continue
			}

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
