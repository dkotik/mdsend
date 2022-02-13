# Hackable Markdown to Email Bridge

Project status: ALPHA DRAFT

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
