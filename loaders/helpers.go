package loaders

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/mail"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// DocumentBoundary determines where markdown files are chunked.
var DocumentBoundary = []byte("\n---\n")

// Loader parses recepients out of data streams.
type Loader interface {
	Load(source string, r io.Reader) (*Message, error)
}

// PathAutoJoin appends prefix to file path, if file path is relative, handles URL paths correctly.
func PathAutoJoin(prefix, file string) string {
	if prefix == `` || file == `` || file[0] == filepath.Separator {
		return file
	}
	// if strings.HasPrefix(prefix, `http://`)
	return filepath.Join(prefix, file)
}

func participantsFromString(pathPrefix string, l *[]Participant, s string) error {
	a, err := mail.ParseAddress(s)
	if err != nil {
		temp := make([]map[string]interface{}, 0)
		// err = configToObject(&temp, PathAutoJoin(pathPrefix, s))
		yamlFile, err := ioutil.ReadFile(PathAutoJoin(pathPrefix, s))
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(yamlFile, &temp)
		if err != nil {
			return fmt.Errorf(`failed to parse an address from "%s": %s`, s, err.Error())
		}
		for _, t := range temp {
			a, err = mail.ParseAddress(fmt.Sprintf(`%v`, t[`email`]))
			if err != nil {
				return fmt.Errorf(`Failed to parse email address in entry "%v."`, t)
			}
			*l = append(*l, Participant{
				Name:   fmt.Sprintf(`%v`, t[`name`]),
				Email:  a.Address,
				Data:   t,
				Source: s})
		}
		return nil
	}
	*l = append(*l, Participant{Name: a.Name, Email: a.Address, Source: `frontmatter`})
	return nil
}

func participantsFromInterface(pathPrefix string, raw interface{}) *[]Participant {
	l := make([]Participant, 0)
	switch value := raw.(type) {
	case string:
		err := participantsFromString(pathPrefix, &l, value)
		if err != nil {
			log.Fatal(err.Error())
		}
	case []interface{}:
		for _, v := range value {
			err := participantsFromString(pathPrefix, &l, v.(string))
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
	return &l
}
