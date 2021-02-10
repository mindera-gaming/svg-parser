# SVG Parser

This package contains a library for parsing SVG 1.1 data. It's currently very incomplete,
as it only supports parsing path data, with no support for curve commands, except for the cubic Bézier curve command **C**.

### Installation

Simply run the following command to install this package to your GOPATH:
```shell
go get github.com/mindera-gaming/svg-parser
```

### Usage

The `svg` package API exposes one function:

```go
func ParsePath(data []byte) ([]Path, error)
```

`ParsePath` takes an entire XML file as a `[]byte` and returns a `[]Path`.

```go
type Path struct {
    Start, End Point
    Control    [2]Point
}

type Point struct {
    X, Y float64
}
```

Each `Path` represents a path segment, which contains the endpoints and the control points.
The latter are generated for commands that generate straight lines between the endpoints,
resulting in points that are halfway between the endpoints.

### Planned Features
- Add parsing support of curve commands **S**, **Q**, **T** and **A**;
- Add parsing support for transformations (matrix, translate, scale, rotate, skewX and skewY);
- Add parsing support for shapes (rect, circle, ellipse, line, polyline and polygon);
- Improve error handling (error information/description on error location);
- Disallow paths that do not begin with a **M** command;
