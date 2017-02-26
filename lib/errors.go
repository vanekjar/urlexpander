package urlexpander

import "errors"

var ErrInvalidUrl = errors.New("Provided URL is not valid")
var ErrLongUrl = errors.New("Provided URL is not shortened")
var ErrDisallowedByRobotsTxt = errors.New("Provided URL is disallowed by robots.txt")
