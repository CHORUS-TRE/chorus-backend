-- +migrate Up

ALTER TABLE public.approval_requests
    ADD COLUMN approveridsbyarm JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE public.approval_requests
    ADD COLUMN armapprovals JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Backfill: existing approverids becomes the "download" arm so legacy
-- requests remain approvable by the same set of users.
UPDATE public.approval_requests
SET approveridsbyarm = jsonb_build_object('download', COALESCE(to_jsonb(approverids), '[]'::jsonb))
WHERE approverids IS NOT NULL AND array_length(approverids, 1) > 0;

-- The new column is queried in List / Count via a JSONB containment test,
-- so we provide a GIN index over the approveridsbyarm column.
CREATE INDEX approval_requests_active_approveridsbyarm_idx
    ON public.approval_requests USING GIN (approveridsbyarm)
    WHERE deletedat IS NULL;

DROP INDEX IF EXISTS approval_requests_active_approverids_idx;
ALTER TABLE public.approval_requests DROP COLUMN approverids;

-- +migrate Down

ALTER TABLE public.approval_requests ADD COLUMN approverids BIGINT[] DEFAULT '{}';

-- Best-effort restore: take the union of approver IDs across all arms.
UPDATE public.approval_requests
SET approverids = ARRAY(
    SELECT DISTINCT (elem)::BIGINT
    FROM jsonb_each(approveridsbyarm) AS arm,
         jsonb_array_elements_text(arm.value) AS elem
);

DROP INDEX IF EXISTS approval_requests_active_approveridsbyarm_idx;
ALTER TABLE public.approval_requests DROP COLUMN armapprovals;
ALTER TABLE public.approval_requests DROP COLUMN approveridsbyarm;

CREATE INDEX approval_requests_active_approverids_idx
    ON public.approval_requests USING GIN (approverids)
    WHERE deletedat IS NULL;
