// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var streamOpts struct {
	command           string
	notificationsOnly bool
	notificationTypes string
}

// Maximum number of websockets (1 hashtag <=> 1 ws)
const maximumHashtagStreamWS = 4

// streamCmd represents the stream command
var streamCmd = &cobra.Command{
	Use:   "stream [user|local|public|:HASHTAG]",
	Short: "Listen to an event stream",
	Long: `
The stream command stays connected to the server and listen to a stream of
events (user, local or federated).
It can also get a hashtag-based stream if the keyword or prefixed with
':' or '#'.`,
	Example: `  madonctl stream           # User timeline stream
  madonctl stream local     # Local timeline stream
  madonctl stream public    # Public timeline stream
  madonctl stream :mastodon # Hashtag
  madonctl stream #madonctl
  madonctl stream --notifications-only
  madonctl stream --notifications-only --notification-types mentions,follows

Several (up to 4) hashtags can be given.
Note: madonctl will use 1 websocket per hashtag stream.
  madonctl stream #madonctl,#mastodon,#golang
  madonctl stream :madonctl,mastodon,api`,
	RunE:       streamRunE,
	ValidArgs:  []string{"user", "public"},
	ArgAliases: []string{"home"},
}

func init() {
	RootCmd.AddCommand(streamCmd)

	streamCmd.Flags().StringVar(&streamOpts.command, "command", "", "Execute external command")
	streamCmd.Flags().BoolVar(&streamOpts.notificationsOnly, "notifications-only", false, "Display only notifications (user stream)")
	streamCmd.Flags().StringVar(&streamOpts.notificationTypes, "notification-types", "", "Filter notifications (mentions, favourites, reblogs, follows)")
}

func streamRunE(cmd *cobra.Command, args []string) error {
	streamName := "user"
	tag := ""
	var hashTagList []string

	if len(args) > 0 {
		if len(args) != 1 {
			return errors.New("too many parameters")
		}
		arg := args[0]
		switch arg {
		case "", "user":
		case "public":
			streamName = arg
		case "local":
			streamName = "public:local"
		default:
			if arg[0] != ':' && arg[0] != '#' {
				return errors.New("invalid argument")
			}
			streamName = "hashtag"
			tag = arg[1:]
			if len(tag) == 0 {
				return errors.New("empty hashtag")
			}
			hashTagList = strings.Split(tag, ",")
			for i, h := range hashTagList {
				if h[0] == ':' || h[0] == '#' {
					hashTagList[i] = h[1:]
				}
				if h == "" {
					return errors.New("empty hashtag")
				}
			}
			if len(hashTagList) > maximumHashtagStreamWS {
				return errors.Errorf("too many hashtags, maximum is %d", maximumHashtagStreamWS)
			}
		}
	}

	if err := madonInit(true); err != nil {
		return err
	}

	var filterMap *map[string]bool
	if streamOpts.notificationTypes != "" {
		var err error
		filterMap, err = buildFilterMap(streamOpts.notificationTypes)
		if err != nil {
			return err
		}
	}

	evChan := make(chan madon.StreamEvent, 10)
	stop := make(chan bool)
	done := make(chan bool)
	var err error

	if streamName != "hashtag" || len(hashTagList) <= 1 { // Usual case: Only 1 stream
		err = gClient.StreamListener(streamName, tag, evChan, stop, done)
	} else { // Several streams
		n := len(hashTagList)
		tagEvCh := make([]chan madon.StreamEvent, n)
		tagDoneCh := make([]chan bool, n)
		for i, t := range hashTagList {
			if verbose {
				errPrint("Launching listener for tag '%s'", t)
			}
			tagEvCh[i] = make(chan madon.StreamEvent)
			tagDoneCh[i] = make(chan bool)
			e := gClient.StreamListener(streamName, t, tagEvCh[i], stop, tagDoneCh[i])
			if e != nil {
				if i > 0 { // Close previous connections
					close(stop)
				}
				err = e
				break
			}
			// Forward events to main ev channel
			go func(i int) {
				for {
					select {
					case _, ok := <-tagDoneCh[i]:
						if !ok { // end of streaming for this tag
							done <- true
							return
						}
					case ev := <-tagEvCh[i]:
						evChan <- ev
					}
				}
			}(i)
		}
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	p, err := getPrinter()
	if err != nil {
		close(stop)
		<-done
		close(evChan)
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	// Set up external command
	p.setCommand(streamOpts.command)

LISTEN:
	for {
		select {
		case v, ok := <-done:
			if !ok || v == true { // done is closed, end of streaming
				break LISTEN
			}
		case ev := <-evChan:
			switch ev.Event {
			case "error":
				if ev.Error != nil {
					if ev.Error == io.ErrUnexpectedEOF {
						errPrint("The stream connection was unexpectedly closed")
						continue
					}
					errPrint("Error event: [%s] %s", ev.Event, ev.Error)
					continue
				}
				errPrint("Event: [%s]", ev.Event)
			case "update":
				if streamOpts.notificationsOnly {
					continue
				}
				s := ev.Data.(madon.Status)
				if err = p.printObj(&s); err != nil {
					break LISTEN
				}
				continue
			case "notification":
				n := ev.Data.(madon.Notification)
				if filterMap != nil && !(*filterMap)[n.Type] {
					continue
				}
				if p.printObj(&n); err != nil {
					break LISTEN
				}
				continue
			case "delete":
				if streamOpts.notificationsOnly {
					continue
				}
				// TODO PrintObj ?
				errPrint("Event: [%s] Status %s was deleted", ev.Event, ev.Data.(string))
			default:
				errPrint("Unhandled event: [%s] %T", ev.Event, ev.Data)
			}
		}
	}
	close(stop)
	close(evChan)
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return nil
}
