package svg

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type Path struct {
	ID   string
	Data []PathData
}

type PathData struct {
	Start, End Point
	Control    [2]Point
}

type parserOptions struct {
	Data     []string
	Absolute bool
}

func newParserOptions(data string, start, end int, absolute bool) parserOptions {
	data = data[start:end]

	return parserOptions{
		Data:     strings.Split(strings.TrimSpace(data), " "),
		Absolute: absolute,
	}
}

type path struct {
	XMLName xml.Name `xml:"path"`
	ID      string   `xml:"id,attr"`
	Style   string   `xml:"style,attr"`
	Data    string   `xml:"d,attr"`
}

func (p *path) Clean() {
	// TODO: consider using regex. Might be better (more performant/less memory usage) than this
	p.Data = strings.Join(strings.Fields(strings.ReplaceAll(p.Data, ",", " ")), " ")
}

func (p path) Parse() ([]PathData, error) {
	var paths []PathData

	var currentAbsolute bool
	var start int
	var current, initial Point
	var parser = func(options parserOptions, current, initial *Point) ([]PathData, error) { return nil, nil }
	var updatePaths = func(end int) (err error) {
		options := newParserOptions(p.Data, start, end, currentAbsolute)

		var newPaths []PathData
		newPaths, err = parser(options, &current, &initial)
		paths = append(paths, newPaths...)

		return
	}
	for i, c := range p.Data {
		var absolute bool
		switch c {
		case 'M':
			absolute = true
			fallthrough
		case 'm':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			start = i + 1
			currentAbsolute = absolute
			parser = parseMoveTo
		case 'L':
			absolute = true
			fallthrough
		case 'l':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			start = i + 1
			currentAbsolute = absolute
			parser = parseLineTo
		case 'H':
			absolute = true
			fallthrough
		case 'h':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			start = i + 1
			currentAbsolute = absolute
			parser = parseHorizontalTo
		case 'V':
			absolute = true
			fallthrough
		case 'v':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			start = i + 1
			currentAbsolute = absolute
			parser = parseVerticalTo
		case 'C':
			absolute = true
			fallthrough
		case 'c':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			start = i + 1
			currentAbsolute = absolute
			parser = parseCurveTo
		case 'S', 's', 'Q', 'q', 'T', 't', 'A', 'a':
			return nil, newUnsupportedCommandError(string(c))
		case 'Z', 'z':
			if err := updatePaths(i); err != nil {
				return nil, err
			}

			parser = func(parserOptions, *Point, *Point) ([]PathData, error) { return nil, nil }
			paths = append(paths, parseClosePath(current, initial, &current))
		}
	}

	if err := updatePaths(len(p.Data)); err != nil {
		return nil, err
	}

	return paths, nil
}

func parseMoveTo(options parserOptions, point, initial *Point) ([]PathData, error) {
	command := func() string { return command(options, "M", "m") }

	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command())
	} else if len(options.Data)%2 != 0 {
		return nil, newInvalidCoordinateError(command(), options.Data)
	}

	if options.Absolute {
		point.Reset()
	}
	x, err := strconv.ParseFloat(options.Data[0], 0)
	if err != nil {
		return nil, newInvalidXError(command(), options.Data[0])
	}
	y, err := strconv.ParseFloat(options.Data[1], 0)
	if err != nil {
		return nil, newInvalidYError(command(), options.Data[1])
	}

	point.X += x
	point.Y += y

	initial.X = point.X
	initial.Y = point.Y

	previous := *point
	paths := make([]PathData, len(options.Data)/2-1)
	for i := 2; i < len(options.Data); i += 2 {
		if options.Absolute {
			point.Reset()
		}

		x, err = strconv.ParseFloat(options.Data[i], 0)
		if err != nil {
			return nil, newInvalidXError(command(), options.Data[i])
		}
		y, err = strconv.ParseFloat(options.Data[i+1], 0)
		if err != nil {
			return nil, newInvalidYError(command(), options.Data[i+1])
		}

		point.X += x
		point.Y += y

		current := *point
		middle := Point{X: 0.5 * (previous.X + current.X), Y: 0.5 * (previous.Y + current.Y)}
		paths[i/2] = PathData{
			Start:   previous,
			End:     current,
			Control: [2]Point{middle, middle},
		}
		previous = current
	}

	return paths, nil
}

func parseLineTo(options parserOptions, point, initial *Point) ([]PathData, error) {
	command := func() string { return command(options, "L", "l") }

	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command())
	} else if len(options.Data)%2 != 0 {
		return nil, newInvalidCoordinateError(command(), options.Data)
	}

	previous := *point
	paths := make([]PathData, len(options.Data)/2)
	for i := 0; i < len(options.Data); i += 2 {
		if options.Absolute {
			point.Reset()
		}

		x, err := strconv.ParseFloat(options.Data[i], 0)
		if err != nil {
			return nil, newInvalidXError(command(), options.Data[i])
		}
		y, err := strconv.ParseFloat(options.Data[i+1], 0)
		if err != nil {
			return nil, newInvalidYError(command(), options.Data[i+1])
		}

		point.X += x
		point.Y += y

		current := *point
		middle := Point{X: 0.5 * (previous.X + current.X), Y: 0.5 * (previous.Y + current.Y)}
		paths[i/2] = PathData{
			Start:   previous,
			End:     current,
			Control: [2]Point{middle, middle},
		}
		previous = current
	}

	return paths, nil
}

func parseHorizontalTo(options parserOptions, point, initial *Point) ([]PathData, error) {
	command := func() string { return command(options, "H", "h") }

	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command())
	}

	previous := point.X
	paths := make([]PathData, len(options.Data))
	for i, c := range options.Data {
		if options.Absolute {
			point.X = 0
		}

		x, err := strconv.ParseFloat(c, 0)
		if err != nil {
			return nil, newInvalidXError(command(), c)
		}

		point.X += x

		current := point.X
		middle := Point{X: 0.5 * (previous + current), Y: point.Y}
		paths[i] = PathData{
			Start:   Point{X: previous, Y: point.Y},
			End:     Point{X: current, Y: point.Y},
			Control: [2]Point{middle, middle},
		}
		previous = point.X
	}

	return paths, nil
}

func parseVerticalTo(options parserOptions, point, initial *Point) ([]PathData, error) {
	command := func() string { return command(options, "V", "v") }

	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command())
	}

	previous := point.Y
	paths := make([]PathData, len(options.Data))
	for i, c := range options.Data {
		if options.Absolute {
			point.Y = 0
		}

		y, err := strconv.ParseFloat(c, 0)
		if err != nil {
			return nil, newInvalidYError(command(), c)
		}

		point.Y += y

		current := point.Y
		middle := Point{X: point.X, Y: 0.5 * (previous + current)}
		paths[i] = PathData{
			Start:   Point{X: point.X, Y: previous},
			End:     Point{X: point.X, Y: current},
			Control: [2]Point{middle, middle},
		}
		previous = point.X
	}

	return paths, nil
}

func parseCurveTo(options parserOptions, point, initial *Point) ([]PathData, error) {
	command := func() string { return command(options, "C", "c") }

	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command())
	} else if len(options.Data)%6 != 0 {
		return nil, newInvalidCoordinateError(command(), options.Data)
	}

	previous := *point
	paths := make([]PathData, len(options.Data)/6)
	var err error
	for i := 0; i < len(options.Data); i += 6 {
		if options.Absolute {
			point.Reset()
		}

		var points [3]Point
		for j := range points {
			k := i + j*2
			points[j].X, err = strconv.ParseFloat(options.Data[k], 0)
			if err != nil {
				return nil, newInvalidXError(command(), options.Data[k])
			}
			points[j].Y, err = strconv.ParseFloat(options.Data[k+1], 0)
			if err != nil {
				return nil, newInvalidYError(command(), options.Data[k+1])
			}
		}

		current := *point
		end := current.Add(points[2])
		paths[i/6] = PathData{
			Start:   previous,
			End:     end,
			Control: [2]Point{current.Add(points[0]), current.Add(points[1])},
		}
		previous = end

		point.X = end.X
		point.Y = end.Y
	}

	return paths, nil
}

func parseClosePath(start, end Point, current *Point) PathData {
	middle := Point{X: 0.5 * (start.X + end.X), Y: 0.5 * (start.Y + end.Y)}
	current.X = start.X
	current.Y = start.Y

	return PathData{
		Start:   start,
		End:     end,
		Control: [2]Point{middle, middle},
	}
}

func command(options parserOptions, absoluteCommand, relativeCommand string) string {
	if options.Absolute {
		return absoluteCommand
	}

	return relativeCommand
}
