# Mdsend (beta)

Send markdown files as electronic mail.

## Features

- **Durable:** mail queues are fault tolerant and atomic, brokered by <Watermill.io> over SQLite3. Can handle any volume of mail without degredation.
- **Flexible:** select the mailer backend based on highest deliverability. Swap it out later without changing anything in your letters or templates.
  - Supports recipient lists formats: CSV, JSON, YAML, TOML, and Cue.
- **Portable:** runs on many kinds of systems. Will include an embeddable HTTP service and Posgres support in the future.

## Examples

```markdown
---
subject: "Test Email"
from: "Sender <test@gmail.com>"
to: "Recipient <test@gmail.com>"
---

# Title of the Example Letter

Write text in **Markdown** notation.
```

```sh
cd examples
mdsend send 1-minimal.md
```

Annotated examples are the documentation:

- [1-minimal.md](examples/minimal.md)
- [2-attachments.md](examples/attachments.md)
- [3-scheduling.md](examples/scheduling.md)
- [4-templating.md](examples/templating.md)
- [5-list.md](examples/list.md)
- [6-extending.md](examples/extending.md)

## Installation

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```

## Development Roadmap

Mdsend is under active development.

<details>
  <summary>Click here to see a list of planned features.</summary>

- [ ] Bug: example six extending example five produces a <nil> recipient.
- [ ] Ensure carbon copy list is in header.
- [ ] Write a better Goldmark 2.0 renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Beautify the default template. Add `dark.html` one.
- [x] Anticipate circular imports for recipient lists and extensions.
- [x] Add queue.Marshaler interface and a JSON implementation?
- [ ] Validate function should detect language and complain that `language` field is not set, if the content is not English.
- [ ] Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
- [x] Mailgun
- [ ] Resend
- [ ] Loops
- [ ] https://purelymail.com/
- [ ] SendGrid
- [ ] Amazon SES
- [ ] Postmark
- [ ] Sparkpost
- [ ] Brevo
- [ ] SendGrid
- [ ] Mailchimp
- [ ] HubSpot
- [ ] Twilio
- [ ] <https://emaillabs.io/en/product/>
- [ ] <https://mailtrap.io/>
- [ ] https://github.com/charmbracelet/glamour
- [ ] add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps

</details>

## E-mail Tools

- <https://www.caniemail.com/> - check what is template-appropriate
- Use <https://www.mail-tester.com/> to check the deliverability of your mail.
- <https://www.htmlemailcheck.com/check/>
- Address verification:
  - <https://github.com/AfterShip/email-verifier>
  - <https://github.com/reacherhq/check-if-email-exists>
  - <https://hunter.io>
  - <https://verify-email.org>
  - <https://email-checker.net>
  - <https://github.com/mailcheck/mailcheck>
  - <https://github.com/ivolo/disposable-email-domains>
  - <https://github.com/willwhite/freemail>
- Hosting:
  - <https://github.com/modoboa/modoboa>

## Similar Projects

- <https://github.com/domodwyer/mailyak>
- <https://sendune.com/>
- <https://github.com/charmbracelet/pop>
- <https://github.com/wneessen/go-mail>
