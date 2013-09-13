# Go Handle Mail

## Implementation status:

`inc` will incorporate mail from the IMAP mailbox "inbox"; it
will only download messages that it is not already seen.

`show`, `next`, and `prev` have bare-bones implementations; but
that does include a start on multipart/alternative messages.

`scan` has a bare-bones implementation.

## Environment Variables

`GOHM_PATH` directory containing mail folders. Usually `~/Mail`.
