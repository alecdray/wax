---
name: inbox-pull
description: Check Gmail for unread wax-labeled inbox emails and save to tmp/inbox.md. Keywords: inbox, check email, wax emails, pull emails.
allowed-tools: [mcp__claude_ai_Gmail__gmail_search_messages, mcp__claude_ai_Gmail__gmail_read_message, Write, Bash]
---

# Inbox Skill

Check Gmail for unread emails with the "wax" label that are in the INBOX, save them to `./tmp/inbox.md`, and report what was found.

## Steps

1. Search Gmail using query: `label:wax in:inbox is:unread`
2. For each message returned, read the full message body
3. Sort newest first (by `internalDate` descending)
4. Write results to `./tmp/inbox.md`:
   - Strip "Sent from my iPhone" footers (and any trailing blank lines before them)
   - Format each email with From, Date, Subject, and body
   - Use `---` as a separator between emails
5. Report how many emails were found and a brief summary of each

## Output Format for tmp/inbox.md

```
# Wax Inbox

---

**From:** ...
**Date:** ...
**Subject:** ...

<body text, cleaned>

---
```

If no unread emails are found, tell the user and do not overwrite tmp/inbox.md.
