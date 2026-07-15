---
subject: minimal example message
from: Test Author <joedoe@test.com>
to: to@test.com
---

# Introduction

This is the simplest letter that can be sent. It has one recipient.

- _Letter_: this file.
- _Message_: a rendered copy of this file for each recipient.

## 1. Install `mdsend`

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install mdsend
  ```
- Build from source:
  ```sh
  go install github.com/dkotik/mdsend@latest
  ```


## 2. Send the Letter

Sending requires a set of mailing service credentials provided as environment variables.

```sh
export SMTP_SERVER=...
export SMTP_PORT=...
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
mdsend send 1-minimal.md
```

## 3. Separate Queuing

Send command will continued to deliver any queued messages, which can be added to the queue with a separate command.

```sh
mdsend queue add 1-minimal.md
mdsend queue add 2-attachments.md
mdsend send
```

# Sample Content

> A Markdown file requires valid 'from', 'to', and 'subject' fields
> in the frontmatter.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce commodo sapien sed magna eleifend, a aliquam nunc condimentum. Sed feugiat nibh at lorem malesuada, lobortis tincidunt urna bibendum. Maecenas pulvinar, quam id tincidunt auctor, sem est tincidunt nulla, ac fringilla velit sapien eget nisl.

Integer fringilla ipsum in elit `tempus varius`. Phasellus viverra velit in pulvinar bibendum. Pellentesque non varius est, in faucibus felis. Praesent sem tellus, consequat id sem nec, tempor efficitur mi. Proin et sem condimentum, dictum dolor vitae, blandit mi. Vestibulum eget diam ante. Etiam facilisis mi non augue malesuada placerat. Maecenas porttitor nulla a sem mattis eleifend. Fusce at neque augue.

Integer at odio laoreet orci iaculis hendrerit. Curabitur sollicitudin volutpat dui at sodales. Vestibulum elit magna, semper nec dolor eget, pharetra tempus augue.

Suspendisse iaculis tortor at massa congue condimentum. Ut interdum leo vel dignissim feugiat. Pellentesque volutpat molestie interdum. Vivamus vel purus sem. In eu hendrerit purus, vel gravida libero.
