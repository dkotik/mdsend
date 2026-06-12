# Mdsend

Project status: ALPHA.

> Send markdown files as MIME-encoded electronic mail.

## Installation

- Go:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```
- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```

## Example

```markdown
---
subject: "Test Email"
from: "Test Account <test@gmail.com>"
to: "Test Account <test@gmail.com>"
to: "./localfile.yaml|toml|json|csv|txt"
---

# Title

Message body.
```

EmailSend(file.md, templating engine, provider engine)

## Delivery Locking

## Distributors

- [x] Mailgun
- [ ] Resend
- [ ] Loops
- [ ] https://purelymail.com/

## Development Roadmap

- [ ] Anticipate circular imports for recipient lists.
- [ ] Drop `yaml.v2` dependency.
- [ ] Bump `github.com/pelletier/go-toml` to v2.
- [ ] Text-part writer should minify html! minified html will have <img src=url> without quotes! (change inline detector)
- [ ] Should write a better markdown renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Email validation: <https://github.com/reacherhq/check-if-email-exists>.
- [ ] Beautify the default template. Add `dark.html` one.
- [ ] Drop Hashcorp multierr.

## Tools

- <https://github.com/AfterShip/email-verifier>
- <https://www.htmlemailcheck.com/check/>
- Use <https://www.mail-tester.com/> to check the deliverability of your mail.

## Similar Projects

- <https://github.com/charmbracelet/pop>
