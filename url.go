package hubur

import "net/url"

func ReplaceUri(u *url.URL, requestURI string) (*url.URL, error) {
	if requestURI == "" {
		requestURI = "/"
	}
	newURI, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, err
	}

	newURL := *u
	newURL.Path = newURI.Path
	newURL.RawPath = newURI.RawPath
	newURL.RawQuery = newURI.RawQuery
	newURL.Fragment = newURI.Fragment
	return &newURL, nil
}
