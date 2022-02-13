package mdsend

type Mailer struct {
	// future replacement for Send function
}

type Option func(m *Mailer) error

func WithOptions(options ...Option) Option {
	return func(m *Mailer) (err error) {
		for _, option := range options {
			if err = option(m); err != nil {
				return err
			}
		}
		return nil
	}
}
