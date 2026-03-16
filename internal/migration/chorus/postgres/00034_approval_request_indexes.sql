-- +migrate Up

CREATE INDEX approval_requests_active_requester_idx
    ON public.approval_requests (tenantid, requesterid)
    WHERE deletedat IS NULL;

CREATE INDEX approval_requests_active_approverids_idx
    ON public.approval_requests USING GIN (approverids)
    WHERE deletedat IS NULL;
