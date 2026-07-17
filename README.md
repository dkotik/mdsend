# Mdsend
[![Go Reference](https://camo.githubusercontent.com/f3bee28c74a644e266e819bedf0150b80af8a7d46292a8fa2837e42aff739ccc/68747470733a2f2f706b672e676f2e6465762f62616467652f6769746875622e636f6d2f5468726565446f74734c6162732f77617465726d696c6c2e737667)](https://pkg.go.dev/github.com/dkotik/mdsend)

Send markdown files as electronic mail.

## Features

- **Durable:** mail queues are fault tolerant and atomic, brokered by <Watermill.io> over SQLite3. Can handle any volume of mail without degredation.
- **Flexible:** select the mailer backend based on highest deliverability. Swap it out later without changing anything in your letters or templates.
  - Supports recipient list formats: `CSV`, `JSON`, `YAML`, `TOML`, and `Cue`.
  - Supports mailing services:
    - [x] **SMTP**
    - [x] **Mailgun**
    - [ ] Resend (soon)
    - [ ] Loops (soon)
    - [ ] SendGrid (soon)
    - [ ] Amazon SES (soon)
    - [ ] Postmark (soon)
    - [ ] Sparkpost (soon)
    - [ ] Brevo (soon)
    - [ ] SendGrid (soon)
    - [ ] Mailchimp (soon)
    - [ ] HubSpot (soon)
    - [ ] Twilio (soon)
    - [ ] <https://purelymail.com/> (soon)
    - [ ] <https://emaillabs.io/en/product/> (soon)
    - [ ] <https://mailtrap.io/> (soon)
- **Portable:** runs on many kinds of systems. Will include an embeddable HTTP service and Posgres support in the future.

## Examples

Compose a letter as Markdown file. It must include the subject, the sender, and at least one recipient in the frontmatter:

```markdown
---
subject: "Test Email"
from: "Sender <test@gmail.com>"
to: "Recipient <test@gmail.com>"
---

# Title of the Example Letter

Write text in **Markdown** notation.
```

Provide server credentials and point `mdsend` at the saved file.

```sh
export SMTP_SERVER=...
export SMTP_PORT=...
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
mdsend send letter.md
```

Annotated examples are the documentation:

- [1-minimal.md](examples/1-minimal.md?plain=1)
- [2-attachments.md](examples/2-attachments.md?plain=1)
- [3-scheduling.md](examples/3-scheduling.md?plain=1)
- [4-templating.md](examples/4-templating.md?plain=1)
- [5-list.md](examples/5-list.md?plain=1)
- [6-extending.md](examples/6-extending.md?plain=1)

## Installation

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install curl mdsend
  ```
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```

## Development

Mdsend is under active development. Version 1.0.0 is expected by the end of 2026.

<details>
  <summary>Click here to see a list of planned features.</summary>

- [ ] Queue clean up scanner - should be first added to sqlite Watermill driver.
- [ ] Sending message with scheduling delay (instead of only queuing) should prompt a confirmation.
- [ ] Ensure carbon copy list is in header.
- [ ] Write a better Goldmark 2.0 renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Beautify the default template. Add `dark.html` one.
- [x] Anticipate circular imports for recipient lists and extensions.
- [x] Add queue.Marshaler interface and a JSON implementation?
- [ ] Validate function should detect language and complain that `language` field is not set, if the content is not English.
- [ ] Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
- [ ] https://github.com/charmbracelet/glamour
- [ ] add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps
- [ ] Publish Debian package to JFrog Artifactory or similar.

</details>

## E-mail Tools

- <https://www.caniemail.com/> - check what is template-appropriate
- E-mail message validators:
  - <https://www.mail-tester.com/>
  - <https://www.suped.com/tools/email-tester>
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
