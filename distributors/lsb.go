package distributors

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/OneOfOne/xxhash"

	"github.com/dkotik/mdsend/loaders"
	"github.com/dkotik/mdsend/loggers"
	"github.com/dkotik/mdsend/providers"
	"github.com/dkotik/mdsend/renderers"
)

// LockingSynchronousBufferingDistributor sends mail from one thread using memory buffer and prevents double deliveries using a hash lock file.
type LockingSynchronousBufferingDistributor struct {
	seed     uint64         // helps protect personal information
	handle   io.WriteCloser // lock file handle used by this distributor
	hashList []uint64       // buffered hashes of sent emails
	logger   loggers.Logger
}

func (d *LockingSynchronousBufferingDistributor) SetLogger(l loggers.Logger) {
	d.logger = l
}

func (d *LockingSynchronousBufferingDistributor) getSeedWithHashList(
	differentiators ...string) error {
	d.hashList = make([]uint64, 0)
	u, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	lockDirectory := filepath.Join(u, `mdsend.lock`)
	if err = os.MkdirAll(lockDirectory, 0700); err != nil {
		return err
	}
	h := xxhash.New64()
	for _, diff := range differentiators {
		h.WriteString(diff)
	}
	file := filepath.Join(lockDirectory, fmt.Sprintf(`%x.lock`, h.Sum64()))
	handle, err := os.Open(file)
	if err == nil {
		// Check file length.
		stat, err2 := handle.Stat()
		if err2 != nil {
			handle.Close()
			return err2
		}
		if stat.Size()%8 > 0 { // Checking if the lock file is corrupt.
			log.Fatalf(`Lock file "%s" is corrupt, because it is %d bytes too long. Delete the file to continue.`, file, stat.Size()%8)
		}
		binary.Read(handle, binary.LittleEndian, &d.seed)
		var hh uint64
		for {
			err = binary.Read(handle, binary.LittleEndian, &hh)
			if err != nil {
				break
			}
			d.hashList = append(d.hashList, hh)
		}
		handle.Close()
		// Flags are important here!
		d.logger.LogInfo(`Using lock file with seed "%x" and "%d" hashes.`, d.seed, len(d.hashList))
		d.handle, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		return err
	}
	d.seed = binary.LittleEndian.Uint64(RandomBytes(8))
	d.handle, err = os.Create(file)
	if err != nil {
		return err
	}
	d.logger.LogInfo(`Using new lock file with seed "%x."`, d.seed)
	return binary.Write(d.handle, binary.LittleEndian, d.seed)
}

// Generate hash, see if it is in the hash table.
func (d *LockingSynchronousBufferingDistributor) check(differentiators ...string) (uint64, bool) {
	h := xxhash.NewS64(d.seed)
	for _, diff := range differentiators {
		h.WriteString(diff)
	}
	hh := h.Sum64()
	for _, v := range d.hashList {
		if v == hh {
			return hh, true
		}
	}
	return hh, false
}

// Send delivers mail to all potential recepients.
func (d *LockingSynchronousBufferingDistributor) Send(
	m *loaders.Message, r renderers.Renderer, p providers.Provider, test bool) error {
	err := d.getSeedWithHashList(m.Subject, m.Date)
	if err != nil {
		return err
	}
	err = p.Open()
	if err != nil {
		d.logger.LogFail("provider error: %s", err.Error())
	}

	closure := func(u *loaders.Participant) error {
		t := u.String()
		hash, locked := d.check(u.Email)
		if locked {
			d.logger.LogSkip(`letter to %s with this subject and date had already been sent (lock #%x)`, u.Email, hash)
			return nil
		}
		m.Current = u
		buffer := bytes.NewBuffer(nil)
		err := r.Render(buffer, m, t)
		if err != nil {
			return err
		}

		var resp string
		for i := 0; i < 3; i++ { // multiple attempts
			// NewReader resets buffer cursor position on each attempt.
			resp, err = p.Send(t, ioutil.NopCloser(bytes.NewReader(buffer.Bytes())), test)
			if err == nil {
				d.hashList = append(d.hashList, hash)
				if test {
					d.logger.LogTest("<%s> %s", u.Email, resp)
					return nil
				}
				d.logger.LogSent("<%s> %s", u.Email, resp)
				return binary.Write(d.handle, binary.LittleEndian, hash)
			}
			d.logger.LogFail(`(attempt %d) Provider API error: %s.`, i+1, err.Error())
			time.Sleep(time.Second * 10)
		}
		return err
	}
	for _, p := range *m.To {
		err := closure(&p)
		if err != nil {
			return err
		}
	}
	for _, p := range *m.CC {
		err := closure(&p)
		if err != nil {
			return err
		}
	}
	for _, p := range *m.BCC {
		err := closure(&p)
		if err != nil {
			return err
		}
	}
	return p.Close()
}

// Close wraps up operations and closes handle.
func (d *LockingSynchronousBufferingDistributor) Close() error {
	if d.handle == nil { // Handle uninitialized.
		return nil
	}
	return d.handle.Close()
}
