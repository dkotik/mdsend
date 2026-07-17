package address

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"mime"
	"os/exec"
	"regexp"
	"strings"
)

var reCollectOutputContentType = regexp.MustCompile(
	`(?i)^\s*Content-Type:([^\n]+)((\r?\n)+)`,
)

type ExecutableCommandError struct {
	Command string
	Cause   error
}

func (err ExecutableCommandError) Error() string {
	return fmt.Sprintf(
		"unable to execute file %q: %v",
		err.Command,
		err.Cause.Error(),
	)
}

func eachEntryFromExecutable(
	ctx context.Context,
	p string,
	fs fs.FS,
) iter.Seq2[any, error] {
	return func(yield func(any, error) bool) {
		cmd := exec.CommandContext(ctx, p)
		// cmd.Dir = filepath.Dir(p)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   err,
			})
			return
		}

		if err := cmd.Start(); err != nil {
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   err,
			})
			return
		}

		b := &bytes.Buffer{}
		if _, err = io.Copy(b, stdout); err != nil {
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   err,
			})
			return
		}

		if err = cmd.Wait(); err != nil {
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   err,
			})
			return
		}

		data := b.Bytes()
		m := reCollectOutputContentType.FindSubmatchIndex(data)
		if m == nil {
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   errors.New("first line of command output does not contain a Content-Type"),
			})
			return
		}

		contentType, _, err := mime.ParseMediaType(string(data[m[2]:m[3]]))
		data = data[m[5]:] // skip content type and free lines
		switch contentType := strings.ToLower(strings.TrimSpace(contentType)); contentType {
		case `application/yaml`:
			for entry, err := range eachEntryFromFileYAML(data) {
				if !yield(entry, err) {
					return
				}
			}
		case `application/toml`, `text/toml`:
			for entry, err := range eachEntryFromFileTOML(data) {
				if !yield(entry, err) {
					return
				}
			}
		case `application/json`:
			for entry, err := range eachEntryFromFileJSON(data) {
				if !yield(entry, err) {
					return
				}
			}
		case `application/cue`:
			for entry, err := range eachEntryFromFileCue(data) {
				if !yield(entry, err) {
					return
				}
			}
		case `text/csv`:
			for entry, err := range eachEntryFromFileCSV(data) {
				if !yield(entry, err) {
					return
				}
			}
		case ``:
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   errors.New("empty Content-Type"),
			})
			return
		default:
			yield(nil, ExecutableCommandError{
				Command: p,
				Cause:   fmt.Errorf("unsupported Content-Type: %v", contentType),
			})
			return
		}
	}
}
