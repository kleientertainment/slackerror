package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func jsonPost(url URL, data interface{}) (err error) {
	var resp *http.Response
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if resp, err = http.Post(string(url), "application/json", bytes.NewBuffer(jsonData)); err != nil {
		return err
	}
	defer io.Copy(ioutil.Discard, resp.Body) // Throw away anything left in the response body so we can reuse the connection
	defer resp.Body.Close()                  // Close the response body
	switch resp.StatusCode {
	case 429:
		var response slackRateLimitResponse
		json.NewDecoder(resp.Body).Decode(&response)
		return &Non200ResponseError{Code: resp.StatusCode, slackRateLimitResponse: &response}
	case 200:
		return nil
	default:
		return &Non200ResponseError{Code: resp.StatusCode}
	}
}

type slackRateLimitResponse struct {
	OK             bool `json:"ok"`
	CountHourAgo   uint `json:"count_hour_ago"`
	CountMinuteAgo uint `json:"count_minute_ago"`
	CountSecondAgo uint `json:"count_second_ago"`
}

type Non200ResponseError struct {
	Code int
	*slackRateLimitResponse
}

func (x *Non200ResponseError) Error() string {
	return fmt.Sprintf("Slack returned non-200 response code %d", x.Code)
}

type messageWithErrCh struct {
	message Message
	errCh   chan error
}
