package svg

// For more information on the "d" attribute:
// - https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/d

import (
	"encoding/xml"
	"math"
	"strconv"
	"strings"

	"github.com/mindera-gaming/go-math/vector2"
)

// PathData represents the "d" attribute that defines a path to be drawn
type PathData struct {
	Start, End vector2.Point
	Control    [2]vector2.Point
}

// parserOptions are essential for the parse of the different commands
type parserOptions struct {
	Data           []string
	Absolute       bool
	SlopeTolerance float64
}

// newParserOptions creates and returns a new parser options structure
func newParserOptions(options ParserOptions, data string, start, end int, absolute bool) parserOptions {
	data = data[start:end]

	return parserOptions{
		Data:           strings.Split(strings.TrimSpace(data), " "),
		Absolute:       absolute,
		SlopeTolerance: options.SlopeTolerance,
	}
}

// path represents the structure of a path element
type path struct {
	XMLName xml.Name `xml:"path"`
	ID      string   `xml:"id,attr"`
	Style   string   `xml:"style,attr"`
	Data    string   `xml:"d,attr"`
}

// Clean the current path to facilitate further processes
func (p *path) Clean() {
	// TODO: consider using regex. Might be better (more performant/less memory usage) than this
	p.Data = strings.Join(strings.Fields(strings.ReplaceAll(p.Data, ",", " ")), " ")
}

// Parse the current path
func (p path) Parse(options ParserOptions) ([]PathData, error) {
	var paths []PathData

	var currentAbsolute bool
	var start int
	var current, initial vector2.Point
	var parser = func(options parserOptions, current, initial *vector2.Point) ([]PathData, error) { return nil, nil }
	var updatePaths = func(end int) (err error) {
		options := newParserOptions(options, p.Data, start, end, currentAbsolute)

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

			parser = func(parserOptions, *vector2.Point, *vector2.Point) ([]PathData, error) { return nil, nil }
			paths = append(paths, parseClosePath(current, initial, &current))
		}
	}

	if err := updatePaths(len(p.Data)); err != nil {
		return nil, err
	}

	return paths, nil
}

// parseMoveTo parses a "MoveTo" command
func parseMoveTo(options parserOptions, lastPoint, initial *vector2.Point) ([]PathData, error) {
	// represents the current command
	command := command(options.Absolute, "M", "m")

	// checks if there is no data to be parsed
	// or if the data has invalid coordinates
	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command)
	} else if len(options.Data)%2 != 0 {
		return nil, newInvalidCoordinateError(command, options.Data)
	}

	// checks the relativity of this command
	if options.Absolute {
		// resets the last (parsed) point
		lastPoint.Reset()
	}
	// parsing the initial point (called 'previous' to make it easier to distinguish further on)
	previous, err := parsePoint(options.Data[0], options.Data[1], command)
	if err != nil {
		return nil, err
	}
	// updating the last point
	lastPoint.Add(previous)

	// updating the initial point
	*initial = *lastPoint

	// updating of the previous point, since it corresponds to the last one
	previous = *lastPoint
	// contain all the parsed paths
	var paths []PathData
	// cycles through all the data this command contains
	for i := 2; i < len(options.Data); i += 2 {
		// current optimised point index
		currentIndex, err := optimizePoints(previous, *lastPoint, i, command, options)
		if err != nil {
			return nil, err
		}

		// checks if there has been any optimisation
		if i != currentIndex {
			// sets the index of the optimised point
			i = currentIndex
		}

		if options.Absolute {
			lastPoint.Reset()
		}
		// parsing the current optimised point
		current, err := parsePoint(options.Data[i], options.Data[i+1], command)
		if err != nil {
			return nil, err
		}
		// updating the last point
		lastPoint.Add(current)

		// updating of the current point, since it corresponds to the last one
		current = *lastPoint
		// represents the middle of the path
		middle := vector2.Point{X: 0.5 * (previous.X + current.X), Y: 0.5 * (previous.Y + current.Y)}
		// adding the new path
		paths = append(paths, PathData{
			Start:   previous,
			End:     current,
			Control: [2]vector2.Point{middle, middle},
		})

		// updating of the previous point, since it corresponds to the current one
		previous = current
	}

	return paths, nil
}

// parseLineTo parses a "LineTo" command
func parseLineTo(options parserOptions, lastPoint, initial *vector2.Point) ([]PathData, error) {
	// represents the current command
	command := command(options.Absolute, "L", "l")

	// checks if there is no data to be parsed
	// or if the data has invalid coordinates
	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command)
	} else if len(options.Data)%2 != 0 {
		return nil, newInvalidCoordinateError(command, options.Data)
	}

	// initial/previous point to next point (current)
	previous := *lastPoint
	// contain all the parsed paths
	var paths []PathData
	// cycles through all the data this command contains
	for i := 0; i < len(options.Data); i += 2 {
		// current optimised point index
		currentIndex, err := optimizePoints(previous, *lastPoint, i, command, options)
		if err != nil {
			return nil, err
		}

		// checks if there has been any optimisation
		if i != currentIndex {
			// sets the index of the optimised point
			i = currentIndex
		}

		if options.Absolute {
			lastPoint.Reset()
		}
		// parsing the current optimised point
		current, err := parsePoint(options.Data[i], options.Data[i+1], command)
		if err != nil {
			return nil, err
		}
		// updating the last point
		lastPoint.Add(current)

		// updating of the current point, since it corresponds to the last one
		current = *lastPoint
		// represents the middle of the path
		middle := vector2.Point{X: 0.5 * (previous.X + current.X), Y: 0.5 * (previous.Y + current.Y)}
		// adding the new path
		paths = append(paths, PathData{
			Start:   previous,
			End:     current,
			Control: [2]vector2.Point{middle, middle},
		})

		// updating of the previous point, since it corresponds to the current one
		previous = current
	}

	return paths, nil
}

// parseHorizontalTo parses a horizontal "LineTo" command
func parseHorizontalTo(options parserOptions, lastPoint, initial *vector2.Point) ([]PathData, error) {
	// represents the current command
	command := command(options.Absolute, "H", "h")

	// checks if there is no data to be parsed
	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command)
	}

	// initial point (called 'previous' to make it easier to distinguish further on)
	previous := lastPoint.X
	// contain all the parsed paths
	var paths []PathData
	// cycles through all the data this command contains
	for i := 0; i < len(options.Data); i++ {
		// current optimised point index
		currentIndex, err := optimizeHorizontalPoints(previous, *lastPoint, i, command, options)
		if err != nil {
			return nil, err
		}

		// checks if there has been any optimisation
		if i != currentIndex {
			// sets the index of the optimised point
			i = currentIndex
		}

		if options.Absolute {
			lastPoint.X = 0
		}
		// parsing the current optimised point
		current, err := parseAbscissa(options.Data[i], command)
		if err != nil {
			return nil, err
		}
		// updating the last point
		lastPoint.X += current

		// updating of the current point, since it corresponds to the last one
		current = lastPoint.X
		// represents the middle of the path
		middle := vector2.Point{X: 0.5 * (previous + current), Y: lastPoint.Y}
		// adding the new path
		paths = append(paths, PathData{
			Start:   vector2.Point{X: previous, Y: lastPoint.Y},
			End:     vector2.Point{X: current, Y: lastPoint.Y},
			Control: [2]vector2.Point{middle, middle},
		})

		// updating of the previous point, since it corresponds to the current one
		previous = current
	}

	return paths, nil
}

// parseVerticalTo parses a vertical "LineTo" command
func parseVerticalTo(options parserOptions, lastPoint, initial *vector2.Point) ([]PathData, error) {
	// represents the current command
	command := command(options.Absolute, "V", "v")

	// checks if there is no data to be parsed
	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command)
	}

	// initial point (called 'previous' to make it easier to distinguish further on)
	previous := lastPoint.Y
	// contain all the parsed paths
	var paths []PathData
	// cycles through all the data this command contains
	for i := 0; i < len(options.Data); i++ {
		// current optimised point index
		currentIndex, err := optimizeVerticalPoints(previous, *lastPoint, i, command, options)
		if err != nil {
			return nil, err
		}

		// checks if there has been any optimisation
		if i != currentIndex {
			// sets the index of the optimised point
			i = currentIndex
		}

		if options.Absolute {
			lastPoint.Y = 0
		}
		// parsing the current optimised point
		current, err := parseOrdinate(options.Data[i], command)
		if err != nil {
			return nil, err
		}
		// updating the last point
		lastPoint.Y += current

		// updating of the current point, since it corresponds to the last one
		current = lastPoint.Y
		// represents the middle of the path
		middle := vector2.Point{X: lastPoint.X, Y: 0.5 * (previous + current)}
		// adding the new path
		paths = append(paths, PathData{
			Start:   vector2.Point{X: lastPoint.X, Y: previous},
			End:     vector2.Point{X: lastPoint.X, Y: current},
			Control: [2]vector2.Point{middle, middle},
		})

		// updating of the previous point, since it corresponds to the current one
		previous = current
	}

	return paths, nil
}

// parseCurveTo parses a "Cubic Bézier Curve" command
func parseCurveTo(options parserOptions, lastPoint, initial *vector2.Point) ([]PathData, error) {
	// represents the current command
	command := command(options.Absolute, "C", "c")

	// checks if there is no data to be parsed
	// or if the data has invalid coordinates
	if len(options.Data) == 0 {
		return nil, newEmptyCoordinateError(command)
	} else if len(options.Data)%6 != 0 {
		return nil, newInvalidCoordinateError(command, options.Data)
	}

	// initial/previous point to next point (current)
	previous := *lastPoint
	// contain all the parsed paths
	paths := make([]PathData, len(options.Data)/6)
	var err error
	// cycles through all the data this command contains
	for i := 0; i < len(options.Data); i += 6 {
		// parsing the current point and its control points
		var points [3]vector2.Point
		for j := range points {
			k := i + j*2
			points[j], err = parsePoint(options.Data[k], options.Data[k+1], command)
			if err != nil {
				return nil, err
			}
		}

		if options.Absolute {
			lastPoint.Reset()
		}
		// last parsed point
		last := *lastPoint
		// current parsed point
		current := last.Add(points[2])
		// adding the new path
		paths[i/6] = PathData{
			Start:   previous,
			End:     current,
			Control: [2]vector2.Point{last.Add(points[0]), last.Add(points[1])},
		}

		// updating of the previous point, since it corresponds to the current one
		previous = current
		// updating of the last point, since it corresponds to the current one
		*lastPoint = current
	}

	return paths, nil
}

// parseClosePath parses a "ClosePath" command
func parseClosePath(start, end vector2.Point, current *vector2.Point) PathData {
	middle := vector2.Point{X: 0.5 * (start.X + end.X), Y: 0.5 * (start.Y + end.Y)}
	*current = start

	return PathData{
		Start:   start,
		End:     end,
		Control: [2]vector2.Point{middle, middle},
	}
}

// command returns the current command depending on its relativity
func command(absolute bool, absoluteCommand, relativeCommand string) string {
	if absolute {
		return absoluteCommand
	}
	return relativeCommand
}

// optimizePoints removes unnecessary points
func optimizePoints(previousPoint vector2.Point, lastPoint vector2.Point, currentIndex int, command string, options parserOptions) (int, error) {
	// temporary copy of the last point
	tempPoint := lastPoint
	if options.Absolute {
		tempPoint.Reset()
	}

	// parsing the current point
	currentPoint, err := parsePoint(options.Data[currentIndex], options.Data[currentIndex+1], command)
	if err != nil {
		return 0, err
	}
	currentPoint.Add(tempPoint)

	// current optimised point index
	optimisedPointIndex := currentIndex
	// cycles through the adjacent points to the current one
	for i := currentIndex + 2; i < len(options.Data); i += 2 {
		// temporary copy of the last point
		tempPoint = lastPoint
		if options.Absolute {
			tempPoint.Reset()
		}

		// parsing the current optimised point
		// used to check the possibility of replacing the current point
		currentOptimised, err := parsePoint(options.Data[i], options.Data[i+1], command)
		if err != nil {
			return 0, err
		}
		currentOptimised.Add(tempPoint)

		// represents the slope difference between the initial path and the "last" path being tested
		var slopeDifference float64

		// slope of the initial (previous + current) and "final" (current + lastPoint) path
		initialPathSlope := math.Abs(previousPoint.Slope(currentPoint))
		lastPathSlope := math.Abs(currentPoint.Slope(currentOptimised))

		// checking some special cases
		if math.IsInf(initialPathSlope, 1) && math.IsInf(lastPathSlope, 1) {
			// reaching here means that both paths are vertically aligned
			slopeDifference = 0
		} else {
			// slope difference calculation
			slopeDifference = math.Abs(lastPathSlope - initialPathSlope)
		}

		// checks if this path can be joined with the initial one
		// this means that, within the given tolerance, this point could be joined with the initial one without any loss of information
		if slopeDifference < options.SlopeTolerance {
			// updates the current optimised
			optimisedPointIndex = i
		} else {
			break
		}
	}
	return optimisedPointIndex, nil
}

// optimizeHorizontalPoints removes unnecessary horizontal points
func optimizeHorizontalPoints(previousAbscissa float64, lastPoint vector2.Point, currentIndex int, command string, options parserOptions) (int, error) {
	// temporary copy of the last point
	tempPoint := lastPoint
	if options.Absolute {
		tempPoint.X = 0
	}

	// parsing the current point
	currentPointAbscissa, err := parseAbscissa(options.Data[currentIndex], command)
	if err != nil {
		return 0, err
	}
	currentPoint := vector2.Point{
		X: currentPointAbscissa + tempPoint.X,
		Y: tempPoint.Y,
	}

	// initial point
	previousPoint := vector2.Point{
		X: previousAbscissa,
		Y: tempPoint.Y,
	}

	// current optimised point index
	optimisedPointIndex := currentIndex
	// cycles through the adjacent points to the current one
	for i := currentIndex + 1; i < len(options.Data); i++ {
		// temporary copy of the last point
		tempPoint := lastPoint
		if options.Absolute {
			tempPoint.X = 0
		}

		// parsing the current optimised point
		// used to check the possibility of replacing the current point
		currentOptimisedAbscissa, err := parseAbscissa(options.Data[i], command)
		if err != nil {
			return 0, err
		}
		currentOptimised := vector2.Point{
			X: currentOptimisedAbscissa + tempPoint.X,
			Y: tempPoint.Y,
		}

		// represents the slope difference between the initial path and the "last" path being tested
		var slopeDifference float64

		// slope of the initial (previous + current) and "final" (current + lastPoint) path
		initialPathSlope := math.Abs(previousPoint.Slope(currentPoint))
		lastPathSlope := math.Abs(currentPoint.Slope(currentOptimised))

		// checking some special cases
		if math.IsInf(initialPathSlope, 1) && math.IsInf(lastPathSlope, 1) {
			// reaching here means that both paths are vertically aligned
			slopeDifference = 0
		} else {
			// slope difference calculation
			slopeDifference = math.Abs(lastPathSlope - initialPathSlope)
		}

		// checks if this path can be joined with the initial one
		// this means that, within the given tolerance, this point could be joined with the initial one without any loss of information
		if slopeDifference < options.SlopeTolerance {
			// updates the current optimised
			optimisedPointIndex = i
		} else {
			break
		}
	}
	return optimisedPointIndex, nil
}

// optimizeVerticalPoints removes unnecessary vertical points
func optimizeVerticalPoints(previousOrdinate float64, lastPoint vector2.Point, currentIndex int, command string, options parserOptions) (int, error) {
	// temporary copy of the last point
	tempPoint := lastPoint
	if options.Absolute {
		tempPoint.Y = 0
	}

	// parsing the current point
	currentPointOrdinate, err := parseOrdinate(options.Data[currentIndex], command)
	if err != nil {
		return 0, err
	}
	currentPoint := vector2.Point{
		X: tempPoint.X,
		Y: currentPointOrdinate + tempPoint.Y,
	}

	// initial point
	previousPoint := vector2.Point{
		X: tempPoint.X,
		Y: previousOrdinate,
	}

	// current optimised point index
	optimisedPointIndex := currentIndex
	// cycles through the adjacent points to the current one
	for i := currentIndex + 1; i < len(options.Data); i++ {
		// temporary copy of the last point
		tempPoint := lastPoint
		if options.Absolute {
			tempPoint.Y = 0
		}

		// parsing the current optimised point
		// used to check the possibility of replacing the current point
		currentOptimisedOrdinate, err := parseOrdinate(options.Data[currentIndex], command)
		if err != nil {
			return 0, err
		}
		currentOptimised := vector2.Point{
			X: tempPoint.X,
			Y: currentOptimisedOrdinate + tempPoint.Y,
		}

		// represents the slope difference between the initial path and the "last" path being tested
		var slopeDifference float64

		// slope of the initial (previous + current) and "final" (current + lastPoint) path
		initialPathSlope := math.Abs(previousPoint.Slope(currentPoint))
		lastPathSlope := math.Abs(currentPoint.Slope(currentOptimised))

		// checking some special cases
		if math.IsInf(initialPathSlope, 1) && math.IsInf(lastPathSlope, 1) {
			// reaching here means that both paths are vertically aligned
			slopeDifference = 0
		} else {
			// slope difference calculation
			slopeDifference = math.Abs(lastPathSlope - initialPathSlope)
		}

		// checks if this path can be joined with the initial one
		// this means that, within the given tolerance, this point could be joined with the initial one without any loss of information
		if slopeDifference < options.SlopeTolerance {
			// updates the current optimised
			optimisedPointIndex = i
		} else {
			break
		}
	}
	return optimisedPointIndex, nil
}

// parsePoint parses the given x and y axes and returns a Point
func parsePoint(x, y, command string) (vector2.Point, error) {
	xAxis, err := parseAbscissa(x, command)
	if err != nil {
		return vector2.Point{}, err
	}
	yAxis, err := parseOrdinate(y, command)
	if err != nil {
		return vector2.Point{}, err
	}

	return vector2.Point{
		X: xAxis,
		Y: yAxis,
	}, nil
}

// parseAbscissa parses the given x-axes and returns its value
func parseAbscissa(x, command string) (float64, error) {
	axis, err := strconv.ParseFloat(x, 0)
	if err != nil {
		return 0, newInvalidXError(command, x)
	}

	return axis, nil
}

// parseOrdinate parses the given y-axes and returns its value
func parseOrdinate(y, command string) (float64, error) {
	axis, err := strconv.ParseFloat(y, 0)
	if err != nil {
		return 0, newInvalidYError(command, y)
	}

	return axis, nil
}
