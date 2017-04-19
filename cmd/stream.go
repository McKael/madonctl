// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

/*
var streamOpts struct {
	local bool
}
*/

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
  madonctl stream #madonctl`,
	RunE:       streamRunE,
	ValidArgs:  []string{"user", "public"},
	ArgAliases: []string{"home"},
}

func init() {
	RootCmd.AddCommand(streamCmd)

	//streamCmd.Flags().BoolVar(&streamOpts.local, "local", false, "Events from the local instance")
}

func streamRunE(cmd *cobra.Command, args []string) error {
	streamName := "user"
	tag := ""

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
		}
	}

	if err := madonInit(true); err != nil {
		return err
	}

	evChan := make(chan madon.StreamEvent, 10)
	stop := make(chan bool)
	done := make(chan bool)

	// StreamListener(name string, hashTag string, events chan<- madon.StreamEvent, stopCh <-chan bool, doneCh chan<- bool) error
	err := gClient.StreamListener(streamName, tag, evChan, stop, done)
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		close(stop)
		<-done
		close(evChan)
		return err
	}

LISTEN:
	for {
		select {
		case _, ok := <-done:
			if !ok { // done is closed, end of streaming
				done = nil
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
				s := ev.Data.(madon.Status)
				p.PrintObj(&s, nil, "")
				continue
			case "notification":
				n := ev.Data.(madon.Notification)
				p.PrintObj(&n, nil, "")
				continue
			case "delete":
				// TODO PrintObj ?
				errPrint("Event: [%s] Status %d was deleted", ev.Event, ev.Data.(int))
			default:
				errPrint("Unhandled event: [%s] %T", ev.Event, ev.Data)
			}
		}
	}
	close(evChan)
	return nil
}
