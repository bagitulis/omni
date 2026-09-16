package jobs

import (
	"fmt"
	"sort"

	"github.com/omni/backend/internal/models"
)

// TransitionPlan describes what the status guard would do to a set of jobs,
// without touching any of them.
//
// Adding the guard changes the behaviour of every existing status write in the
// codebase, and those writes run against live tenant schemas. A plan that can be
// inspected before the change reaches production is the difference between
// "we believe this is safe" and "we checked".
type TransitionPlan struct {
	From    models.JobStatus
	To      models.JobStatus
	Allowed bool
}

// PlanTransitions reports, for every current status, whether a move to the given
// target would be permitted. It performs no database access at all.
func PlanTransitions(to models.JobStatus) []TransitionPlan {
	var plans []TransitionPlan
	for from := range allowedTransitions {
		plans = append(plans, TransitionPlan{
			From:    from,
			To:      to,
			Allowed: transitionAllowed(from, to),
		})
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].From < plans[j].From
	})
	return plans
}

// DescribeTransitionPlan renders PlanTransitions as text for a dry run.
func DescribeTransitionPlan(to models.JobStatus) string {
	out := fmt.Sprintf("target status: %s\n", to)
	for _, p := range PlanTransitions(to) {
		verdict := "REFUSED"
		if p.Allowed {
			verdict = "allowed"
		}
		out += fmt.Sprintf("  %-10s -> %-10s %s\n", p.From, p.To, verdict)
	}
	return out
}
