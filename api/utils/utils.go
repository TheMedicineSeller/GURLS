package utils

import (
    "github.com/TheMedicineSeller/GURLS/config"
    "strings"
)

func EnforceHTTP(url string) string {
    if url[:4] != "http" {
        return "http://" + url
    }
    return url
}

func RemoveDomainError (url string) bool {
    if url == config.DOMAIN {
        return false
    }
    // Checking if the user is trying the heckle the website with erorr causing url
    newURL := strings.Replace(url, "http://", "", 1)
    newURL = strings.Replace(url, "https://", "", 1)
    newURL = strings.Replace(url, "www.", "", 1)
    if strings.Split(newURL, "/")[0] == config.DOMAIN {
        return false
    }
    return true
}
