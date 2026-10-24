package mail_boxes

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/service/mail_boxes"
)

func (c *ControllerV1) DeleteMailbox(ctx context.Context, req *v1.DeleteMailboxReq) (res *v1.DeleteMailboxRes, err error) {
	res = &v1.DeleteMailboxRes{}

	if len(req.Emails) == 0 {
		return nil, gerror.New("Email addresses cannot be empty")
	}

	// An empty username is a legal (if unwanted) primary key left behind by old
	// imports, so it must stay addressable for deletion (#378). Only duplicates
	// are dropped.
	seen := make(map[string]bool, len(req.Emails))
	var validEmails []string
	for _, email := range req.Emails {
		email = strings.TrimSpace(email)
		if seen[email] {
			continue
		}
		seen[email] = true
		validEmails = append(validEmails, email)
	}

	affected, err := mail_boxes.DeleteBatch(ctx, validEmails)
	if err != nil {
		return nil, err
	}

	for _, email := range validEmails {
		_ = public.WriteLog(ctx, public.LogParams{
			Type: consts.LOGTYPE.Mailboxes,
			Log:  "Deleted mailbox:" + email + " successfully",
		})
	}

	if affected == 0 {
		res.SetSuccess("No mailboxes were deleted (they may not exist)")
	} else {
		res.SetSuccess(fmt.Sprintf("Successfully deleted %d mailbox(es)", affected))
	}

	return
}
