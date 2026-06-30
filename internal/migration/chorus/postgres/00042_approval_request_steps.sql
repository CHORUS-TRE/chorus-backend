-- +migrate Up

ALTER TABLE public.approval_requests
    ADD COLUMN approveridsbystep JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE public.approval_requests
    ADD COLUMN stepdecisions JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Backfill: the legacy approverids list seeds every step the request's type
-- requires, so existing requests remain approvable by the same set of users.
UPDATE public.approval_requests
SET approveridsbystep = CASE
    WHEN type = 'data_transfer' THEN jsonb_build_object(
        'download', COALESCE(to_jsonb(approverids), '[]'::jsonb),
        'upload', COALESCE(to_jsonb(approverids), '[]'::jsonb))
    ELSE jsonb_build_object(
        'download', COALESCE(to_jsonb(approverids), '[]'::jsonb))
END
WHERE approverids IS NOT NULL AND array_length(approverids, 1) > 0;

-- The new column is queried in List / Count via a JSONB containment test,
-- so we provide a GIN index over the approveridsbystep column.
CREATE INDEX approval_requests_active_approveridsbystep_idx
    ON public.approval_requests USING GIN (approveridsbystep)
    WHERE deletedat IS NULL;

DROP INDEX IF EXISTS approval_requests_active_approverids_idx;
ALTER TABLE public.approval_requests DROP COLUMN approverids;

-- approvedbyid is redundant now that each entry in stepdecisions records its
-- own approver
ALTER TABLE public.approval_requests DROP COLUMN approvedbyid;

-- +migrate Down

ALTER TABLE public.approval_requests ADD COLUMN approvedbyid BIGINT;
ALTER TABLE public.approval_requests
    ADD CONSTRAINT approval_requests_approvedbycon FOREIGN KEY (approvedbyid) REFERENCES users(id);

ALTER TABLE public.approval_requests ADD COLUMN approverids BIGINT[] DEFAULT '{}';

-- Best-effort restore: take the union of approver IDs across all steps.
UPDATE public.approval_requests
SET approverids = ARRAY(
    SELECT DISTINCT (elem)::BIGINT
    FROM jsonb_each(approveridsbystep) AS step,
         jsonb_array_elements_text(step.value) AS elem
);

DROP INDEX IF EXISTS approval_requests_active_approveridsbystep_idx;
ALTER TABLE public.approval_requests DROP COLUMN stepdecisions;
ALTER TABLE public.approval_requests DROP COLUMN approveridsbystep;

CREATE INDEX approval_requests_active_approverids_idx
    ON public.approval_requests USING GIN (approverids)
    WHERE deletedat IS NULL;
