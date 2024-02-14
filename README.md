# Mdsend

Project status: ALPHA DRAFT

> Send markdown files as MIME-encoded electronic mail.

## Installation

- Go:
  ```bash
  go install github.com/dkotik/mdsend@latest
  ```
- Macintosh:
  ```bash
  brew tap dkotik/tap
  brew install mdsend
  ```

## Example

```markdown
---
subject: "Test Email"
to: "Test Account <test@gmail.com>"
to: "./localfile.yaml|toml|json|csv|txt"
---

# Title

Message body.
```

Easily hackable.

EmailSend(file.md, templating engine, provider engine)

## Delivery Locking

## Distributors

- https://purelymail.com/

## Planned Features

- Text-part writer should minify html! minified html will have <img src=url> without quotes! (change inline detector)
- Should write a better markdown renderer that recognizes single-line youtube and image paragraphs, centers them.
- Email validation: https://github.com/reacherhq/check-if-email-exists
