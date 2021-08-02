package vite

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

const (
	gviteLogger       = "gvite"
	gviteStdErrLogger = "gvite err"
)

// logPipe prints out logs from gvite. We don't end when context
// is canceled beacause there are often logs printed after this.
func logPipe(pipe io.ReadCloser, identifier string) error {
	reader := bufio.NewReader(pipe)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Println("closing", identifier, err)
			return err
		}

		message := strings.ReplaceAll(str, "\n", "")
		log.Println(identifier, message)
	}
}

// StartGvite starts a gvite daemon in another goroutine
// and logs the results to the console.
func StartGvite(ctx context.Context, arguments string, g *errgroup.Group) error {
	parsedArgs := strings.Split(arguments, " ")
	cmd := exec.Command(
		"/app/gvite",
		parsedArgs...,
	) // #nosec G204

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	g.Go(func() error {
		return logPipe(stdout, gviteLogger)
	})

	g.Go(func() error {
		return logPipe(stderr, gviteStdErrLogger)
	})

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: unable to start gvite", err)
	}

	g.Go(func() error {
		<-ctx.Done()

		log.Println("sending interrupt to gvite")
		return cmd.Process.Signal(os.Interrupt)
	})

	return cmd.Wait()
}
