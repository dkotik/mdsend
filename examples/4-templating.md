---
subject: example message for {{ .Recipient.name }}
from: Test Author <joedoe@test.com>
to:
  - name: Full Name
    email: to@test.com
    first_name: FirstName
someVariable: test
templates:
  - ../internal/template/html/default.html
headers:
  X-Template-Test: "{{ .Frontmatter.someVariable }}"
---

Hello {{ .Recipient.first_name }},

# {{ .Frontmatter.subject }}

The Markdown content itself is a `Golang` text template.
To wrap the output of a letter into an `HTML` template, specify
the list of files to load using the `template` field in the
frontmatter.

The frontmatter subject and the header values are also treated
as text templates. If the header template is rendered empty,
it will be ommitted from the sent message.

## Context

Template execution context contains the following fields:

- **Frontmatter:** the raw Markdown frontmatter as `map[string]any`.
- **Recipient:** the raw message target as `map[string]any`. It
  will contain `name` and `email` fields and any other fields
  fromt the contact source.
- **Content:** rendered Markdown body.
- **Schedule:** parsed scheduling directives.

## Reify Templates

Special `reify` function stores the output of a named `Golang` block
or template. Calling `reify` again on the same template, returns
the initial output. Each recipient starts with an empty output cache.

Here is an example of declaring a template that encodes recipient
address as a URL-safe field once to be used several times:

```
{{ define "token" }}
  {{ urlQuery .Recipient.email }}
{{ end }}

URL: https://test.com/?unsubscribe={{ reify "token" }}
```
