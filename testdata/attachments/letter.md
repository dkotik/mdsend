---
from: Test Sender <mock@test.com>
subject: "test regular and inline attachments"
attachments: ["../../internal/testdata/image/cat.jpg", "attachment.txt"]
templates: ["template.html"]
---

# Test Attachments

Inline attachment here: ![panda.jpg](../../internal/testdata/image/panda.jpg)

Repeating should not fluke the test: ![panda.jpg](../../internal/testdata/image/panda.jpg)
