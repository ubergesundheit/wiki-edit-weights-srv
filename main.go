package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func connectRemoteWebsocket(urlStr string) (*websocket.Conn, error) {
	conn, resp, err := websocket.DefaultDialer.Dial(urlStr, http.Header{})
	if err != nil {
		return &websocket.Conn{}, err
	}

	if resp.StatusCode != 101 {
		return &websocket.Conn{}, fmt.Errorf("Error: Expected remote to return HTTP 101, received %d instead", resp.StatusCode)
	}

	return conn, nil
}

func aggregateByInterval(aggregationIntervalDuration, backlogDuration time.Duration, c *websocket.Conn) error {
	// set up backlog timestamp limit
	backlogTimestamp := time.Now().Add(-backlogDuration)
	// synchronize backlock timestamp with the aggregation interval
	// by subtracting the remainder of the division between the backlog timestamp
	// and the aggregation interval
	backlogTimestamp = backlogTimestamp.Add(-time.Duration(backlogTimestamp.UnixNano() % int64(aggregationIntervalDuration)))
	// set up step timestamp
	backlogStepTimestamp := backlogTimestamp.Add(aggregationIntervalDuration)
	changeSize := 0
	for _, msg := range messageCache {
		if msg.Timestamp.Before(backlogTimestamp) {
			continue
		}
		if msg.Timestamp.Before(backlogStepTimestamp) {
			changeSize += msg.ChangeSize
		} else if msg.Timestamp.After(time.Now().Add(-aggregationIntervalDuration)) {
			break
		} else {
			err := c.WriteJSON(EditMessage{Timestamp: backlogStepTimestamp, ChangeSize: changeSize})
			if err != nil {
				return err
			}
			changeSize = 0
			backlogStepTimestamp = backlogStepTimestamp.Add(aggregationIntervalDuration)
		}
	}

	// try to start the for loop at the start of a period
	time.Sleep(time.Duration(int64(aggregationIntervalDuration) - time.Now().UnixNano()%int64(aggregationIntervalDuration)))

	for {
		currTs := time.Now().Add(-aggregationIntervalDuration)
		currTs = currTs.Add(-time.Duration(currTs.UnixNano() % int64(aggregationIntervalDuration)))
		changeSize := 0

	AggregateLoop:
		for i := len(messageCache) - 1; i >= 0; i-- {
			// iterate backwards until the stored timestamp is no longer before
			// the current timestamp
			if messageCache[i].Timestamp.After(currTs) {
				changeSize += messageCache[i].ChangeSize
			} else {
				break AggregateLoop
			}
		}
		err := c.WriteJSON(EditMessage{Timestamp: currTs, ChangeSize: changeSize})
		if err != nil {
			return err
		}

		// sleep to match the next aggregation interval
		time.Sleep(time.Duration(int64(aggregationIntervalDuration) - time.Now().UnixNano()%int64(aggregationIntervalDuration)))
	}
}

func main() {
	flag.Parse()
	wsConn, err := connectRemoteWebsocket(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	// collect messages into the cache
	go func(conn *websocket.Conn) {
		for {
			var mesg EditMessage
			err := conn.ReadJSON(&mesg)
			if err == nil {
				messageCache = append(messageCache, mesg)
			}
		}
	}(wsConn)

	startServer(*addr)
}
