package PageMarkdownUtils

import (
	"regexp"
	"bytes"
)

func ProcessPage(body []byte, isView bool) (new []byte) {
	if isView {
		return forwardProcessPage(body)
	}else {
		return reverseProcessPage(body)
	}
}

func reverseProcessPage(body []byte) ([]byte) {
	boldStartRegex := regexp.MustCompile("<strong>")
	boldFinishRegex := regexp.MustCompile("<\\/strong>")

	italicsStartRegex := regexp.MustCompile("<em>")
	italicsFinishRegex := regexp.MustCompile("<\\/em>")

	linkBeginRegex := regexp.MustCompile("<a href=\"/view/")
	linkCenterRegex := regexp.MustCompile(">")
	linkFinishRegex := regexp.MustCompile("<\\/a>")

	body = bytes.Replace(body, []byte("<br>"), []byte("\n"), -1)

	body = boldStartRegex.ReplaceAll(body, []byte("*\\"))
	body = boldFinishRegex.ReplaceAll(body, []byte("/*"))

	body = italicsStartRegex.ReplaceAll(body, []byte("_\\"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("/_"))

	body = linkBeginRegex.ReplaceAll(body, []byte("\\["))
	body = linkFinishRegex.ReplaceAll(body, []byte("]/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("]["))

	return body
}

func forwardProcessPage(body []byte) ([]byte) {
	boldStartRegex := regexp.MustCompile("(\\*)\\\\")
	boldFinishRegex := regexp.MustCompile("\\/(\\*)")

	italicsStartRegex := regexp.MustCompile("_\\\\")
	italicsFinishRegex := regexp.MustCompile("\\/_")

	linkBeginRegex := regexp.MustCompile("\\\\\\[")
	linkCenterRegex := regexp.MustCompile("\\]\\[")
	linkFinishRegex := regexp.MustCompile("\\]\\/")

	body = bytes.Replace(body, []byte("\n"), []byte("<br>"), -1)

	body = boldStartRegex.ReplaceAll(body, []byte("<strong>"))
	body = boldFinishRegex.ReplaceAll(body, []byte("</strong>"))

	body = italicsStartRegex.ReplaceAll(body, []byte("<em>"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("</em>"))

	body = linkBeginRegex.ReplaceAll(body, []byte("<a href=\"/view/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("\">"))
	body = linkFinishRegex.ReplaceAll(body, []byte("</a>"))

	return body
}