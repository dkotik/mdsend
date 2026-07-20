# Mdsend
[![Go Reference](https://camo.githubusercontent.com/f3bee28c74a644e266e819bedf0150b80af8a7d46292a8fa2837e42aff739ccc/68747470733a2f2f706b672e676f2e6465762f62616467652f6769746875622e636f6d2f5468726565446f74734c6162732f77617465726d696c6c2e737667)](https://pkg.go.dev/github.com/dkotik/mdsend)

Send markdown files as electronic mail. Maintain mailing lists as templated text files.

## Features

- **Durable:** mail queues are fault tolerant and atomic, brokered by <www.watermill.io> over SQLite3. Can handle any volume of mail without degredation.
- **Portable:** runs on many kinds of systems. Will include an embeddable HTTP service and Posgres support in the future.
- **Flexible:** select the mailer backend based on highest deliverability. Swap it out later without changing anything in your letters or templates.
  - Supports recipient list formats: `CSV`, `JSON`, `YAML`, `TOML`, and `Cue`.
  - Supports mailing services:
    - [x] **SMTP** (SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD)
    - [x] **Mailgun** (MG_API_KEY, MG_API_DOMAIN) or SMTP relay
    - [x] **Resend** via SMTP relay or (RESEND_API_KEY) 
    - [x] **Amazon SES** (requires AWS configuration present)
      - The author decoupled himself from AWS years ago and requests
        help testing this implementation.
    - [ ] SparkPost (soon)
    - [ ] Loops
  - Supports additional mailing services by SMTP relay:
    - [x] SMTP2GO
    - [x] Brevo
    - [x] Mailchimp
    - [x] Twilio
    - [x] SparkPost
    - [x] Postmark
    - [x] SendGrid
    - [x] ZeptoMail
    - [x] SendPulse
    - [x] MailTrap
    - [x] MailJet
    - [x] EmailLabs
    - [x] PurelyMail
    - [ ] HubSpot? (smtp.hubapi.com)

SMTP relay is often more robust than the provider API. For example, Resend API does not support multi-value headers, but its SMTP relay does.

## Examples

Compose a letter as a Markdown file. It must include the subject, the sender, and at least one recipient in the frontmatter:

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
export SMTP_HOST=...
export SMTP_PORT=...
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
mdsend send letter.md
```

More usage examples are here:

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
- Debian Package: [latest release](releases)
- Linux Binary: [latest release](releases)
- Windows Binary: [latest release](releases)
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```

## Development

Mdsend is under active development. Version 1.0.0 is expected by the end of 2026.

<details>
  <summary>Click here to see a list of planned features.</summary>

- [ ] attachments should be collected from templates as well
- [ ] Validate function should detect language and complain that `language` field is not set, if the content is not English.
- [ ] Beautify the default template. Add `dark.html` one.
- [ ] service package with HTTP unsubscribe endpoint package
- [ ] mdsend <file> should be equivalent to mdsend send <file>?
- [ ] Queue clean up scanner - should be first added to sqlite Watermill driver.
- [ ] Sending message with scheduling delay (instead of only queuing) should prompt a confirmation.
- [ ] list out templating functions in the example, including the default ones
- [ ] Write a better Goldmark 2.0 renderer that recognizes single-line youtube and image paragraphs, centers them.
- [ ] Event invitations markup: https://developers.google.com/gmail/markup/reference/event-reservation#basic_event_reminder_without_a_ticket
- [ ] https://github.com/charmbracelet/glamour
- [ ] Run some tests with <https://www.suped.com/tools/email-tester>
- [ ] modularize the proprietory mailers - they should not be in default dependencies
- [ ] add to https://github.com/rothgar/awesome-tuis and bubbletea list of apps

</details>

## E-mail Tools

- <https://www.caniemail.com/> - check what is template-appropriate
- E-mail message validators:
  - <https://www.mail-tester.com/>
  - <https://www.suped.com/tools/email-tester>
  - <https://www.htmlemailcheck.com/check/>
  - <https://github.com/mailhog/MailHog>
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
- Beautiful templates:
  - <https://mjml.io/templates>

## Similar Projects

- <https://github.com/domodwyer/mailyak>
- <https://github.com/charmbracelet/pop>
- <https://github.com/wneessen/go-mail>
- <https://github.com/go-gomail/gomail>
- <https://github.com/mailhog/mhsendmail>
- <https://sendune.com/>
